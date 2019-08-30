package core

import (
	"github.com/sentinel-group/sentinel-golang/core/util"
	"math"
	"sync/atomic"
	"time"
)

// A universal interface for traffic shaping controller.
type TrafficShapingController interface {
	// check whether given resource entry can pass with provided count.
	CanPass(ctx *Context, node node, acquire uint64) bool
}

// Default throttling controller (immediately reject strategy).
// When the QPS had exceeded the threshold of any rule, the request will be rejected.
// This Controller used to apply to the situation that the exact capacity of a system is known
type DefaultController struct {
	grade FlowGradeType
	count uint64
}

func (dc *DefaultController) CanPass(ctx *Context, node node, acquire uint64) bool {
	curCount := dc.avgUsedTokens(node)
	if (curCount + acquire) > dc.count {
		return false
	}
	return true
}

func (dc *DefaultController) avgUsedTokens(node node) uint64 {
	if node == nil {
		return 0
	}
	if dc.grade == FlowGradeGoroutine {
		return uint64(node.CurGoroutineNum())
	}
	return node.PassQps()
}

// RateLimiterController will strictly control the interval time of request passing, that is,
// let the request pass at an even speed, similar to the leaky bucket algorithm.
//
// This approach is mainly used to handle intermittent bursts of traffic, such as message queues.
// Imagine a situation where a large number of requests arrive in one second and the next few seconds are idle,
// and we want the system to be able to process these requests gradually during the rest of the idle period,
// rather than rejecting the extra requests directly in the first second.
// Leaky Bucket algorithm combined with virtual queue waiting mechanism
type RateLimiterController struct {
	// 1000/count represent the interval time of each request passing.
	count uint64
	// max waiting time used for virtual queue
	maxQueueingTimeMs int64
	// last pass timestamp of last request
	latestPassedTime int64
}

func (rlc *RateLimiterController) CanPass(ctx *Context, node node, acquire uint64) bool {
	if acquire <= 0 {
		return true
	}
	// Reject when count is less or equal than 0.
	// Otherwise,the costTime will be max of long and waitTime will overflow in some cases.
	if rlc.count <= 0 {
		return false
	}

	currentTime := int64(util.GetTimeMilli())
	// Calculate the interval between every two requests.
	costTime := int64(math.Round(float64(acquire) / float64(rlc.count) * 1000))
	// Expected pass time of this request.
	expectedTime := costTime + atomic.LoadInt64(&rlc.latestPassedTime)

	if expectedTime < currentTime {
		// Contention may exist here, but it's okay.
		// current time has exceeded the expected time, return pass
		atomic.CompareAndSwapInt64(&rlc.latestPassedTime, rlc.latestPassedTime, currentTime)
		return true
	} else {
		// recalculate the time to wait.
		waitTime := costTime + atomic.LoadInt64(&rlc.latestPassedTime) - int64(util.GetTimeMilli())
		if waitTime > rlc.maxQueueingTimeMs {
			// exceed the max waiting time
			return false
		} else {
			// expect passing time
			oldTime := atomic.AddInt64(&rlc.latestPassedTime, costTime)
			waitTime = oldTime - int64(util.GetTimeMilli())
			if waitTime > rlc.maxQueueingTimeMs {
				atomic.AddInt64(&rlc.latestPassedTime, -costTime)
				return false
			}
			if waitTime > 0 {
				time.Sleep(time.Duration(waitTime) * time.Millisecond)
			}
			return true
		}
	}
}

// WarmUpController aka Preheat/cold start mode
//
type WarmUpController struct {
}

func (wpc WarmUpController) CanPass(ctx *Context, node node, acquire uint64) bool {
	return true
}

//
type WarmUpRateLimiterController struct {
}

func (wpc WarmUpRateLimiterController) CanPass(ctx *Context, node node, acquire uint64) bool {
	return true
}
