package statistic

import (
	"errors"
	"fmt"
	"github.com/sentinel-group/sentinel-golang/core/util"
	"log"
	"math"
	"runtime"
)

// windowWrap means window slot, each slot represent a data structure to record metrics.
type windowWrap struct {
	windowLengthInMs uint32
	windowStart      uint64
	// the data structure of the actual record fields
	value interface{}
}

func (ww *windowWrap) resetTo(startTime uint64) {
	ww.windowStart = startTime
}

func (ww *windowWrap) isTimeInWindow(timeMillis uint64) bool {
	return ww.windowStart <= timeMillis && timeMillis < ww.windowStart+uint64(ww.windowLengthInMs)
}

// The basic data structure of sliding windows
// leapArray hold the windowWrap array(aka slide window) to record metrics
type leapArray struct {
	windowLengthInMs uint32
	sampleCount      uint32
	intervalInMs     uint32
	array            []*windowWrap     //window slots
	mux              util.TriableMutex //lock
}

func (la *leapArray) CurrentWindow(sw bucketGenerator) (*windowWrap, error) {
	return la.CurrentWindowWithTime(util.GetTimeMilli(), sw)
}

func (la *leapArray) CurrentWindowWithTime(timeMillis uint64, sw bucketGenerator) (*windowWrap, error) {
	if timeMillis < 0 {
		return nil, errors.New("timeMillion is less than 0")
	}

	idx := la.calculateTimeIdx(timeMillis)
	windowStart := la.calculateStartTime(timeMillis)

	for {
		old := la.array[idx]
		if old == nil {
			newWrap := &windowWrap{
				windowLengthInMs: la.windowLengthInMs,
				windowStart:      windowStart,
				value:            sw.newEmptyBucket(windowStart),
			}
			// must be concurrent safe,
			// some extreme condition,may newer override old empty windowWrap
			// confirm the array[idx] is nil before update array[idx].
			// la.mux.Lock()
			//if la.array[idx] == nil {
			//	la.array[idx] = newWrap
			//}
			//la.mux.Unlock()
			//return la.array[idx], nil
			if la.mux.TryLock() && la.array[idx] == nil {
				la.array[idx] = newWrap
				la.mux.Unlock()
				return la.array[idx], nil
			} else {
				runtime.Gosched()
			}
		} else if windowStart == old.windowStart {
			return old, nil
		} else if windowStart > old.windowStart {
			// reset windowWrap
			if la.mux.TryLock() {
				old, _ = sw.resetWindowTo(old, windowStart)
				la.mux.Unlock()
				return old, nil
			} else {
				runtime.Gosched()
			}
		} else if windowStart < old.windowStart {
			// Should not go through here,
			return nil, errors.New(fmt.Sprintf("provided time timeMillis=%d is already behind old.windowStart=%d", windowStart, old.windowStart))
		}
	}
}

func (la *leapArray) calculateTimeIdx(timeMillis uint64) uint32 {
	timeId := (int)(timeMillis / uint64(la.windowLengthInMs))
	return uint32(timeId % len(la.array))
}

func (la *leapArray) calculateStartTime(timeMillis uint64) uint64 {
	return timeMillis - (timeMillis % uint64(la.windowLengthInMs))
}

//  Get all the bucket in sliding window for current time;
func (la *leapArray) Values() []*windowWrap {
	return la.valuesWithTime(util.GetTimeMilli())
}

func (la *leapArray) valuesWithTime(timeMillis uint64) []*windowWrap {
	if timeMillis <= 0 {
		return nil
	}
	wwp := make([]*windowWrap, 0)
	for _, wwPtr := range la.array {
		if wwPtr == nil {
			//log.Printf("current bucket is nil, index is %d \n", idx)
			wwPtr = &windowWrap{
				windowLengthInMs: la.windowLengthInMs,
				windowStart:      util.GetTimeMilli(),
				value:            newMetricBucket(),
			}
			wwp = append(wwp, wwPtr)
			continue
		}
		newWW := &windowWrap{
			windowLengthInMs: wwPtr.windowLengthInMs,
			windowStart:      wwPtr.windowStart,
			value:            wwPtr.value,
		}
		wwp = append(wwp, newWW)
	}
	return wwp
}

type bucketGenerator interface {
	// 根据开始时间，创建一个新的统计bucket, bucket的具体数据结构可以有多个
	newEmptyBucket(startTime uint64) interface{}

	// 将窗口ww重置startTime和空的统计bucket
	resetWindowTo(ww *windowWrap, startTime uint64) (*windowWrap, error)
}

// The implement of sliding window based on struct leapArray
// slidingWindow use metricBucket to record statistic metrics
type slidingWindow struct {
	data             *leapArray
	MetricBucketType string
}

func newSlidingWindow(sampleCount uint32, intervalInMs uint32) *slidingWindow {
	if intervalInMs%sampleCount != 0 {
		panic(fmt.Sprintf("invalid parameters, intervalInMs is %d, sampleCount is %d.", intervalInMs, sampleCount))
	}
	winLengthInMs := intervalInMs / sampleCount
	arr := make([]*windowWrap, sampleCount)
	return &slidingWindow{
		data: &leapArray{
			windowLengthInMs: winLengthInMs,
			sampleCount:      sampleCount,
			intervalInMs:     intervalInMs,
			array:            arr,
		},
		MetricBucketType: "array_metric_bucket",
	}
}

func (sw *slidingWindow) newEmptyBucket(startTime uint64) interface{} {
	return newMetricBucket()
}

func (sw *slidingWindow) resetWindowTo(ww *windowWrap, startTime uint64) (*windowWrap, error) {
	ww.windowStart = startTime
	ww.value = newMetricBucket()
	return ww, nil
}

func (sw *slidingWindow) Count(event MetricEventType) uint64 {
	_, err := sw.data.CurrentWindow(sw)
	if err != nil {
		log.Println("sliding window fail to record success")
	}
	count := uint64(0)
	for _, ww := range sw.data.Values() {
		mb, ok := ww.value.(*metricBucket)
		if !ok {
			log.Println("assert fail")
			continue
		}
		count += mb.Get(event)
	}
	return count
}

func (sw *slidingWindow) AddCount(event MetricEventType, count uint64) {
	curWindow, err := sw.data.CurrentWindow(sw)
	if err != nil || curWindow == nil || curWindow.value == nil {
		log.Println("sliding window fail to record success")
		return
	}

	mb, ok := curWindow.value.(*metricBucket)
	if !ok {
		log.Println("assert fail")
		return
	}
	mb.Add(event, count)
}

func (sw *slidingWindow) MaxSuccess() uint64 {

	_, err := sw.data.CurrentWindow(sw)
	if err != nil {
		log.Println("sliding window fail to record success")
	}

	succ := uint64(0)
	for _, ww := range sw.data.Values() {
		mb, ok := ww.value.(*metricBucket)
		if !ok {
			log.Println("assert fail")
			continue
		}
		s := mb.Get(MetricEventSuccess)
		if err != nil {
			log.Println("get success counter fail, reason: ", err)
		}
		succ = uint64(math.Max(float64(succ), float64(s)))
	}
	return succ
}

func (sw *slidingWindow) MinSuccess() uint64 {

	_, err := sw.data.CurrentWindow(sw)
	if err != nil {
		log.Println("sliding window fail to record success")
	}

	succ := uint64(0)
	for _, ww := range sw.data.Values() {
		mb, ok := ww.value.(*metricBucket)
		if !ok {
			log.Println("assert fail")
			continue
		}
		s := mb.Get(MetricEventSuccess)
		succ = uint64(math.Min(float64(succ), float64(s)))
	}
	return succ
}
