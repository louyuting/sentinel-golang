package server

import (
	"context"
	"sync"
	"sync/atomic"

	sbase "github.com/alibaba/sentinel-golang/core/stat/base"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

const InvalidTokenCount = -1

type RedisTokenServer struct {
	redisCli                *redis.Client
	rwMux                   sync.RWMutex
	resClusterTokenResetMap map[string]*redisTokenResetSlidingWindow
}

// redisTokenResetSlidingWindow is responsible for cluster token reset based on time
type redisTokenResetSlidingWindow struct {
	redisCli *redis.Client
	resource string
	data     *sbase.LeapArray
}

func (r *redisTokenResetSlidingWindow) NewEmptyBucket() interface{} {
	return struct{}{}
}

func (r *redisTokenResetSlidingWindow) ResetBucketTo(bw *sbase.BucketWrap, startTime uint64) *sbase.BucketWrap {
	redisCli := r.redisCli

	ctx, cancel := context.WithTimeout(context.Background(), redisCli.Options().WriteTimeout)
	defer cancel()
	_, err := redisCli.Del(ctx, r.resource).Result()
	// if fail to clear resource token, don't reset bucket start time and wait for next reset operation
	if err != nil {
		logging.Error(err, "Fail to clear resource token in redisTokenResetSlidingWindow#ResetBucketTo", "resource", r.resource)
		return bw
	}
	atomic.StoreUint64(&bw.BucketStart, startTime)
	return bw
}

func newResourceTokenResetSlidingWindow(redisCli *redis.Client, resource string, intervalMs uint32) *redisTokenResetSlidingWindow {
	redisTokenWindow := &redisTokenResetSlidingWindow{
		redisCli: redisCli,
		resource: resource,
		data:     nil,
	}
	// TODO, LeapArray pre init start time, maybe need to improve it.
	la, _ := sbase.NewLeapArray(1, intervalMs, redisTokenWindow)
	redisTokenWindow.data = la
	return redisTokenWindow
}

func (r *RedisTokenServer) newResourceTokenReset(resource string, intervalMs uint32) *redisTokenResetSlidingWindow {
	return newResourceTokenResetSlidingWindow(r.redisCli, resource, intervalMs)
}

func (r *RedisTokenServer) AcquireFlowToken(resource string, tokenCount uint32, statIntervalInMs uint32) (curCount int64, err error) {
	// 1. general checking logic
	if len(resource) == 0 {
		return InvalidTokenCount, errors.New("empty resource")
	}
	if tokenCount == 0 {
		return InvalidTokenCount, errors.New("token count is zero")
	}

	// check whether reset token count
	r.rwMux.RLock()
	resourceTokenWindow, ok := r.resClusterTokenResetMap[resource]
	r.rwMux.RUnlock()
	if ok {
		// check token reset logic
		_, e := resourceTokenWindow.data.CurrentBucket(resourceTokenWindow)
		if e != nil {
			logging.Error(e, "Fail to check cluster token reset in RedisTokenServer#AcquireToken")
		}
	} else {
		r.rwMux.Lock()
		resourceTokenWindow, ok = r.resClusterTokenResetMap[resource]
		if !ok {
			// double check
			resourceTokenWindow = r.newResourceTokenReset(resource, statIntervalInMs)
			r.resClusterTokenResetMap[resource] = resourceTokenWindow
		}
		r.rwMux.Unlock()
		_, e := resourceTokenWindow.data.CurrentBucket(resourceTokenWindow)
		if e != nil {
			logging.Error(e, "Fail to check cluster token reset in RedisTokenServer#AcquireToken")
		}
	}

	redisCli := r.redisCli

	ctx, cancel := context.WithTimeout(context.Background(), redisCli.Options().ReadTimeout)
	defer cancel()
	currentVal, err := redisCli.Incr(ctx, resource).Result()
	if err == nil || err == redis.Nil {
		return currentVal, nil
	}
	return InvalidTokenCount, err
}
