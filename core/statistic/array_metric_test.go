package statistic

import (
	"fmt"
	"sync"
	"testing"
)

const (
	sampleCount  uint32 = 10
	intervalInMs uint32 = 1000
)

func TestArrayMetric_Normal(t *testing.T) {
	am := NewArrayMetric(sampleCount, intervalInMs)

	for i := 0; i < 100; i++ {
		if i%5 == 0 {
			am.AddPass(1)
		} else if i%5 == 1 {
			am.AddBlock(1)
		} else if i%5 == 2 {
			am.AddSuccess(1)
		} else if i%5 == 3 {
			am.AddError(1)
		} else if i%5 == 4 {
			am.AddRt(100)
		} else {
			t.Error("unexpect count")
		}
	}

	if am.Pass() != 20 {
		t.Error("unexpect count MetricEventBlock")
	}
	if am.Block() != 20 {
		t.Error("unexpect count MetricEventBlock")
	}
	if am.Success() != 20 {
		t.Error("unexpect count MetricEventSuccess")
	}
	if am.Error() != 20 {
		t.Error("unexpect count MetricEventError")
	}
	if am.Rt() != 20*100 {
		t.Error("unexpect count MetricEventRt")
	}
}

func TestArrayMetric_Concurrent(t *testing.T) {
	am := NewArrayMetric(sampleCount, intervalInMs)
	wg := &sync.WaitGroup{}
	wg.Add(5000)

	for i := 0; i < 1000; i++ {
		go func() {
			am.AddPass(1)
			wg.Done()
		}()
	}

	for i := 0; i < 1000; i++ {
		go func() {
			am.AddBlock(2)
			wg.Done()
		}()
	}

	for i := 0; i < 1000; i++ {
		go func() {
			am.AddSuccess(3)
			wg.Done()
		}()
	}

	for i := 0; i < 1000; i++ {
		go func() {
			am.AddError(4)
			wg.Done()
		}()
	}

	for i := 0; i < 1000; i++ {
		go func(c uint64) {
			am.AddRt(uint64(c))
			wg.Done()
		}(uint64(i))
	}
	wg.Wait()

	if am.Pass() != 1000 {
		t.Error("unexpect count MetricEventBlock")
	}
	if am.Block() != 2000 {
		t.Error("unexpect count MetricEventBlock")
	}
	if am.Success() != 3000 {
		t.Error("unexpect count MetricEventSuccess")
	}
	if am.Error() != 4000 {
		t.Error("unexpect count MetricEventError")
	}

	trt := (0 + 999) * 1000 / 2
	fmt.Println("expect:", trt)

	fmt.Println("rtt:", am.Rt())
	if am.Rt() != uint64(trt) {
		t.Error("unexpect count MetricEventRt")
	}
}
