package core

import (
	"fmt"
	"github.com/sentinel-group/sentinel-golang/core/util"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
)

func TestDefaultController_PassQps(t *testing.T) {
	node := &NodeMock{}
	threshold := 100

	c := &DefaultController{
		grade: FlowGradeQps,
		count: uint64(threshold),
	}
	cnt := new(int)
	*cnt = 0
	call := node.On("PassQps").Return(cnt)

	for i := 1; i < threshold*2; i++ {
		pass := c.CanPass(nil, node, 1)
		c := call.ReturnArguments.Get(0).(*int)
		*c = *c + 1
		if i <= threshold {
			assert.True(t, pass)
		} else {
			assert.True(t, !pass)
		}

	}
}

func TestDefaultController_GoroutinePass(t *testing.T) {
	node := &NodeMock{}
	threshold := 100

	c := &DefaultController{
		grade: FlowGradeGoroutine,
		count: uint64(threshold),
	}
	cnt := new(int)
	*cnt = 0
	call := node.On("CurGoroutineNum").Return(cnt)

	for i := 1; i < threshold*2; i++ {
		pass := c.CanPass(nil, node, 1)
		c := call.ReturnArguments.Get(0).(*int)
		*c = *c + 1
		if i <= threshold {
			assert.True(t, pass)
		} else {
			assert.True(t, !pass)
		}
	}
}

func TestRateLimiterController_CanPass(t *testing.T) {
	c := &RateLimiterController{
		count:             10,
		maxQueueingTimeMs: 500,
		latestPassedTime:  int64(util.GetTimeMilli()),
	}

	count := new(int64)
	*count = 0
	wg := &sync.WaitGroup{}

	num := 5
	wg.Add(num)
	start := util.GetTimeMilli()
	fmt.Println("start:", start)
	for i := 0; i < num; i++ {
		go func(cnt *int64, wg_ *sync.WaitGroup) {
			assert.True(t, c.CanPass(nil, nil, 1))
			wg_.Done()
		}(count, wg)
	}
	wg.Wait()
	end := util.GetTimeMilli()
	fmt.Println("end:", end)

	assert.True(t, (end-start) > 400)
}

func TestRateLimiterController_CanPass_Timeout(t *testing.T) {
	c := &RateLimiterController{
		count:             10,
		maxQueueingTimeMs: 500,
		latestPassedTime:  int64(util.GetTimeMilli()),
	}

	pass := new(int64)
	*pass = 0
	block := new(int64)
	*block = 0

	wg := &sync.WaitGroup{}

	num := 10
	wg.Add(num)
	start := util.GetTimeMilli()
	fmt.Println("start:", start)
	for i := 0; i < num; i++ {
		go func(cnt *int64, wg_ *sync.WaitGroup) {
			if c.CanPass(nil, nil, 1) {
				atomic.AddInt64(pass, 1)
			} else {
				atomic.AddInt64(block, 1)
			}
			wg_.Done()
		}(pass, wg)
	}
	wg.Wait()
	end := util.GetTimeMilli()
	fmt.Println("end:", end)
	assert.True(t, *block > 0)
}
