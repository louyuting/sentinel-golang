package core

import (
	"github.com/sentinel-group/sentinel-golang/core/statistic"
	"github.com/sentinel-group/sentinel-golang/core/util"
	"sync/atomic"
)

const (
	windowLengthImMs_ uint32 = 200
	sampleCount_      uint32 = 5
	intervalInMs_     uint32 = 1000
)

type ResourceNode struct {
	rollingCounterInSecond *statistic.ArrayMetric
	rollingCounterInMinute *statistic.ArrayMetric
	currentGoroutineNum    int64
	lastFetchTime          uint64
	resourceWrapper        *ResourceWrapper
}

func NewResourceNode(wrapper *ResourceWrapper) *ResourceNode {
	return &ResourceNode{
		resourceWrapper:        wrapper,
		currentGoroutineNum:    0,
		lastFetchTime:          util.GetTimeMilli(),
		rollingCounterInSecond: statistic.NewArrayMetric(sampleCount_, intervalInMs_),
		rollingCounterInMinute: statistic.NewArrayMetric(sampleCount_, 60*1000),
	}
}

func (n *ResourceNode) IncreasePassRequest(count uint64) {
	n.rollingCounterInSecond.AddPass(count)
	n.rollingCounterInMinute.AddPass(count)
}
func (n *ResourceNode) IncreaseBlockRequest(count uint64) {
	n.rollingCounterInSecond.AddBlock(count)
	n.rollingCounterInMinute.AddBlock(count)
}
func (n *ResourceNode) IncreaseErrorRequest(count uint64) {
	n.rollingCounterInSecond.AddError(count)
	n.rollingCounterInMinute.AddError(count)
}
func (n *ResourceNode) IncreaseSuccessRequest(count uint64) {
	n.rollingCounterInSecond.AddSuccess(count)
	n.rollingCounterInMinute.AddSuccess(count)
}

func (n *ResourceNode) TotalRequest() uint64 {
	return n.rollingCounterInMinute.Request()
}
func (n *ResourceNode) TotalPass() uint64 {
	return n.rollingCounterInMinute.Pass()
}
func (n *ResourceNode) TotalBlock() uint64 {
	return n.rollingCounterInMinute.Block()
}
func (n *ResourceNode) TotalSuccess() uint64 {
	return n.rollingCounterInMinute.Success()
}
func (n *ResourceNode) TotalError() uint64 {
	return n.rollingCounterInMinute.Error()
}

func (n *ResourceNode) RequestQps() uint64 {
	return n.PassQps() + n.BlockQps()
}
func (n *ResourceNode) PassQps() uint64 {
	return n.rollingCounterInSecond.Pass() / uint64(n.rollingCounterInSecond.GetIntervalInMs())
}
func (n *ResourceNode) BlockQps() uint64 {
	return n.rollingCounterInSecond.Block() / uint64(n.rollingCounterInSecond.GetIntervalInMs())
}
func (n *ResourceNode) SuccessQps() uint64 {
	return n.rollingCounterInSecond.Success() / uint64(n.rollingCounterInSecond.GetIntervalInMs())
}
func (n *ResourceNode) ErrorQps() uint64 {
	return n.rollingCounterInSecond.Error() / uint64(n.rollingCounterInSecond.GetIntervalInMs())
}

// unit is millisecond
func (n *ResourceNode) AddRt(rt uint64) {
	n.rollingCounterInSecond.AddRt(rt)
	n.rollingCounterInMinute.AddRt(rt)
}
func (n *ResourceNode) AvgRt() uint64 {
	succCount := n.TotalSuccess()
	if succCount == 0 {
		return 0
	}
	return n.rollingCounterInSecond.Rt() / succCount
}

//func (n *ResourceNode) MinRt() uint64 {
//	return 0
//}
func (n *ResourceNode) IncreaseGoroutineNum() {
	atomic.AddInt64(&n.currentGoroutineNum, 1)
}
func (n *ResourceNode) DecreaseGoroutineNum() {
	atomic.AddInt64(&n.currentGoroutineNum, -1)
}
func (n *ResourceNode) CurGoroutineNum() int64 {
	return atomic.LoadInt64(&n.currentGoroutineNum)
}
