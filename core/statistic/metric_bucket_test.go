package statistic

import (
	"fmt"
	"sync"
	"testing"
)

func TestMetricBucket_Normal(t *testing.T) {
	mb := newMetricBucket()

	for i := 0; i < 100; i++ {
		if i%5 == 0 {
			mb.AddPass(1)
		} else if i%5 == 1 {
			mb.AddBlock(1)
		} else if i%5 == 2 {
			mb.AddSuccess(1)
		} else if i%5 == 3 {
			mb.AddError(1)
		} else if i%5 == 4 {
			mb.AddRt(100)
		} else {
			t.Error("unexpect count")
		}
	}

	if mb.Get(MetricEventPass) != 20 {
		t.Error("unexpect count MetricEventBlock")
	}
	if mb.Get(MetricEventBlock) != 20 {
		t.Error("unexpect count MetricEventBlock")
	}
	if mb.Get(MetricEventSuccess) != 20 {
		t.Error("unexpect count MetricEventSuccess")
	}
	if mb.Get(MetricEventError) != 20 {
		t.Error("unexpect count MetricEventError")
	}
	if mb.Get(MetricEventRt) != 20*100 {
		t.Error("unexpect count MetricEventRt")
	}
	if mb.MinRt() != 100 {
		t.Error("unexpect min rt")
	}
}

func TestMetricBucket_Concurrent(t *testing.T) {
	mb := newMetricBucket()
	wg := &sync.WaitGroup{}
	wg.Add(5000)

	for i := 0; i < 1000; i++ {
		go func() {
			mb.AddPass(1)
			wg.Done()
		}()
	}

	for i := 0; i < 1000; i++ {
		go func() {
			mb.AddBlock(2)
			wg.Done()
		}()
	}

	for i := 0; i < 1000; i++ {
		go func() {
			mb.AddSuccess(3)
			wg.Done()
		}()
	}

	for i := 0; i < 1000; i++ {
		go func() {
			mb.AddError(4)
			wg.Done()
		}()
	}

	for i := 0; i < 1000; i++ {
		go func(c uint64) {
			mb.AddRt(uint64(c))
			wg.Done()
		}(uint64(i))
	}
	wg.Wait()

	if mb.Get(MetricEventPass) != 1000 {
		t.Error("unexpect count MetricEventBlock")
	}
	if mb.Get(MetricEventBlock) != 2000 {
		t.Error("unexpect count MetricEventBlock")
	}
	if mb.Get(MetricEventSuccess) != 3000 {
		t.Error("unexpect count MetricEventSuccess")
	}
	if mb.Get(MetricEventError) != 4000 {
		t.Error("unexpect count MetricEventError")
	}

	trt := (0 + 999) * 1000 / 2
	fmt.Println("expect:", trt)

	fmt.Println("rtt:", mb.Get(MetricEventRt))
	if mb.Get(MetricEventRt) != uint64(trt) {
		t.Error("unexpect count MetricEventRt")
	}
	if mb.MinRt() != 0 {
		t.Error("unexpect min rt")
	}
}
