package flow

import (
	"math"
	"sync/atomic"

	"github.com/alibaba/sentinel-golang/logging"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/util"
)

type WarmUpTrafficShapingCalculator struct {
	threshold         float64
	warmUpPeriodInSec uint32
	coldFactor        uint32
	warningToken      uint64
	maxToken          int64
	slope             float64
	storedTokens      *int64
	lastFilledTime    *uint64
}

func NewWarmUpTrafficShapingCalculator(rule *FlowRule) *WarmUpTrafficShapingCalculator {
	if rule.WarmUpColdFactor == 0 {
		rule.WarmUpColdFactor = config.DefaultWarmUpColdFactor
		logging.Warnf("[NewWarmUpTrafficShapingCalculator] No set WarmUpColdFactor,use default values: %d", config.DefaultWarmUpColdFactor)
	}

	warningToken := int64((float64(rule.WarmUpPeriodSec) * rule.Count) / float64(rule.WarmUpColdFactor-1))

	maxToken := warningToken + int64(2*float64(rule.WarmUpPeriodSec)*rule.Count/float64(1.0+rule.WarmUpColdFactor))

	slope := float64(rule.WarmUpColdFactor-1.0) / rule.Count / float64(maxToken-warningToken)

	warmUpTrafficShapingCalculator := &WarmUpTrafficShapingCalculator{
		warmUpPeriodInSec: rule.WarmUpPeriodSec,
		coldFactor:        rule.WarmUpColdFactor,
		warningToken:      uint64(warningToken),
		maxToken:          maxToken,
		slope:             slope,
		threshold:         rule.Count,
		storedTokens:      new(int64),
		lastFilledTime:    new(uint64),
	}

	return warmUpTrafficShapingCalculator
}

func (d *WarmUpTrafficShapingCalculator) CalculateAllowedTokens(node base.StatNode, acquireCount uint32, flag int32) float64 {
	previousQps := node.GetPreviousQPS(base.MetricEventPass)
	d.syncToken(previousQps)

	restToken := atomic.LoadInt64(d.storedTokens)
	if restToken >= int64(d.warningToken) {
		aboveToken := restToken - int64(d.warningToken)
		warningQps := math.Nextafter(1.0/(float64(aboveToken)*d.slope+1.0/d.threshold), math.MaxFloat64)
		return warningQps
	} else {
		return d.threshold
	}
}

func (d *WarmUpTrafficShapingCalculator) syncToken(passQps float64) {
	currentTime := util.CurrentTimeMillis()
	currentTime = currentTime - currentTime%1000

	oldLastFillTime := atomic.LoadUint64(d.lastFilledTime)
	if currentTime <= oldLastFillTime {
		return
	}

	oldValue := atomic.LoadInt64(d.storedTokens)
	newValue := d.coolDownTokens(currentTime, passQps)

	if atomic.CompareAndSwapInt64(d.storedTokens, oldValue, newValue) {
		//sub
		if currentValue := atomic.AddInt64(d.storedTokens, ^(int64(passQps) - 1)); currentValue < 0 {
			atomic.StoreInt64(d.storedTokens, 0)
		}
		atomic.StoreUint64(d.lastFilledTime, currentTime)
	}
}

func (d *WarmUpTrafficShapingCalculator) coolDownTokens(currentTime uint64, passQps float64) int64 {
	oldValue := atomic.LoadInt64(d.storedTokens)
	newValue := oldValue

	// Prerequisites for adding a token:
	// When token consumption is much lower than the warning line
	if oldValue < int64(d.warningToken) {
		newValue = int64(float64(oldValue) + (float64(currentTime)-float64(atomic.LoadUint64(d.lastFilledTime)))*d.threshold/1000)
	} else if oldValue > int64(d.warningToken) {
		if passQps < float64(uint32(d.threshold)/d.coldFactor) {
			newValue = int64(float64(oldValue) + float64(currentTime-atomic.LoadUint64(d.lastFilledTime))*d.threshold/1000)
		}
	}

	if newValue <= d.maxToken {
		return newValue
	} else {
		return d.maxToken
	}
}
