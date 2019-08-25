package statistic

import (
	"math"
	"sync/atomic"
)

type MetricEventType int8

const (
	// PASS + BLOCK == all traffic
	// sentinel rules check pass
	MetricEventPass MetricEventType = iota
	// sentinel rules check block
	MetricEventBlock
	// Success + Error == Pass
	MetricEventSuccess
	// app execute error
	MetricEventError
	// request execute rt, unit is millisecond
	MetricEventRt
	// hack for getting length of enum
	metricEventNum
)

// metricBucket is storage entity for the metric statistic
// event type contains (MetricEventPass、MetricEventBlock、MetricEventError、MetricEventSuccess、MetricEventRt)
// the statistic of metricBucket must be concurrent safe.
type metricBucket struct {
	// value of statistic
	counter [metricEventNum]uint64
	// record the min rt for this bucket
	minRt uint64
}

func newMetricBucket() *metricBucket {
	mb := &metricBucket{
		minRt: math.MaxUint64,
	}
	return mb
}

func (mb *metricBucket) Add(event MetricEventType, count uint64) {
	if event > metricEventNum {
		panic("event is bigger then metricEventNum")
	}
	atomic.AddUint64(&mb.counter[event], count)
}

func (mb *metricBucket) Get(event MetricEventType) uint64 {
	if event > metricEventNum {
		panic("event is bigger then metricEventNum")
	}
	return mb.counter[event]
}

func (mb *metricBucket) MinRt() uint64 {
	return mb.minRt
}

func (mb *metricBucket) Reset() {
	for i := 0; i < int(metricEventNum); i++ {
		atomic.StoreUint64(&mb.counter[i], 0)
	}
	atomic.StoreUint64(&mb.minRt, math.MaxUint64)
}

func (mb *metricBucket) AddPass(n uint64) {
	mb.Add(MetricEventPass, n)
}

func (mb *metricBucket) Pass() uint64 {
	return mb.Get(MetricEventPass)
}

func (mb *metricBucket) AddBlock(n uint64) {
	mb.Add(MetricEventBlock, n)
}

func (mb *metricBucket) Block() uint64 {
	return mb.Get(MetricEventBlock)
}

func (mb *metricBucket) AddSuccess(n uint64) {
	mb.Add(MetricEventSuccess, n)
}

func (mb *metricBucket) Success() uint64 {
	return mb.Get(MetricEventSuccess)
}

func (mb *metricBucket) AddError(n uint64) {
	mb.Add(MetricEventError, n)
}

func (mb *metricBucket) Error() uint64 {
	return mb.Get(MetricEventError)
}

func (mb *metricBucket) AddRt(rt uint64) {
	mb.Add(MetricEventRt, rt)
	// Not thread-safe, but it's okay.
	if rt < mb.minRt {
		mb.minRt = rt
	}
}

func (mb *metricBucket) Rt() uint64 {
	return mb.Get(MetricEventRt)
}
