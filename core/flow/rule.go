// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package flow

import (
	"encoding/json"
	"fmt"

	"github.com/alibaba/sentinel-golang/core/system_metric"
	"github.com/alibaba/sentinel-golang/util"
)

// RelationStrategy indicates the flow control strategy based on the relation of invocations.
type RelationStrategy int32

const (
	// CurrentResource means flow control by current resource directly.
	CurrentResource RelationStrategy = iota
	// AssociatedResource means flow control by the associated resource rather than current resource.
	AssociatedResource
)

func (s RelationStrategy) String() string {
	switch s {
	case CurrentResource:
		return "CurrentResource"
	case AssociatedResource:
		return "AssociatedResource"
	default:
		return "Undefined"
	}
}

type TokenCalculateStrategy int32

const (
	Direct TokenCalculateStrategy = iota
	WarmUp
)

func (s TokenCalculateStrategy) String() string {
	switch s {
	case Direct:
		return "Direct"
	case WarmUp:
		return "WarmUp"
	default:
		return "Undefined"
	}
}

type ControlBehavior int32

const (
	Reject ControlBehavior = iota
	Throttling
)

func (s ControlBehavior) String() string {
	switch s {
	case Reject:
		return "Reject"
	case Throttling:
		return "Throttling"
	default:
		return "Undefined"
	}
}

type AdaptiveType uint32

const (
	// Load represents system load1 in Linux/Unix.
	MetricDefault = AdaptiveType(0)
	CPU           = AdaptiveType(1)
	Memory        = AdaptiveType(2)
	AdaptiveMax   = AdaptiveType(3)
)

func (t AdaptiveType) String() string {
	switch t {
	case MetricDefault:
		return "default"
	case CPU:
		return "cpu"
	case Memory:
		return "memory"
	default:
		return fmt.Sprintf("unknown(%d)", t)
	}
}

func (t AdaptiveType) Validate() bool {
	return t < AdaptiveMax
}

// Rule describes the strategy of flow control, the flow control strategy is based on QPS statistic metric
type Rule struct {
	// ID represents the unique ID of the rule (optional).
	ID string `json:"id,omitempty"`
	// Resource represents the resource name.
	Resource               string                 `json:"resource"`
	TokenCalculateStrategy TokenCalculateStrategy `json:"tokenCalculateStrategy"`
	ControlBehavior        ControlBehavior        `json:"controlBehavior"`
	// Threshold means the threshold during StatIntervalInMs
	// If StatIntervalInMs is 1000(1 second), Threshold means QPS
	Threshold        float64          `json:"threshold"`
	RelationStrategy RelationStrategy `json:"relationStrategy"`
	RefResource      string           `json:"refResource"`
	// MaxQueueingTimeMs only takes effect when ControlBehavior is Throttling.
	// When MaxQueueingTimeMs is 0, it means Throttling only controls interval of requests,
	// and requests exceeding the threshold will be rejected directly.
	MaxQueueingTimeMs uint32 `json:"maxQueueingTimeMs"`
	WarmUpPeriodSec   uint32 `json:"warmUpPeriodSec"`
	WarmUpColdFactor  uint32 `json:"warmUpColdFactor"`
	// StatIntervalInMs indicates the statistic interval and it's the optional setting for flow Rule.
	// If user doesn't set StatIntervalInMs, that means using default metric statistic of resource.
	// If the StatIntervalInMs user specifies can not reuse the global statistic of resource,
	// 		sentinel will generate independent statistic structure for this rule.
	StatIntervalInMs uint32 `json:"statIntervalInMs"`

	// adaptive flow control alg
	//
	// If AdaptiveEnable is true, flow rule's threshold is calculated by FlowRule.GetThreshold()
	// instead of FlowRule.Threshold. Sentinel gets the metric watermark based on Rule.AdaptiveType
	// which maybe cpu or memory. And then Sentinel will calculate the real threshold.
	//
	// Alg:
	// If the watermark is less than Rule.LowWaterMark, the threshold is equal to Rule.SafeThreshold.
	// If the watermark is greater than Rule.LowWaterMark, the threshold is equal to Rule.SafeThreshold.
	// Otherwise, the threshold is equal to (watermark - LowWaterMark) / (HighWaterMark - LowWaterMark)) *
	//	(RiskThreshold - SafeThreshold) + SafeThreshold.
	AdaptiveEnable bool         `json:"adaptiveEnable"`
	AdaptiveType   AdaptiveType `json:"adaptiveType"`
	SafeThreshold  uint64       `json:"safeThreshold"`
	RiskThreshold  uint64       `json:"riskThreshold"`
	LowWaterMark   uint64       `json:"lowWatermark"`
	HighWaterMark  uint64       `json:"highWatermark"`
}

func (r *Rule) isEqualsTo(newRule *Rule) bool {
	if newRule == nil {
		return false
	}
	if !(r.Resource == newRule.Resource && r.RelationStrategy == newRule.RelationStrategy &&
		r.RefResource == newRule.RefResource && r.StatIntervalInMs == newRule.StatIntervalInMs &&

		r.TokenCalculateStrategy == newRule.TokenCalculateStrategy && r.ControlBehavior == newRule.ControlBehavior &&
		util.Float64Equals(r.Threshold, newRule.Threshold) &&

		r.MaxQueueingTimeMs == newRule.MaxQueueingTimeMs && r.WarmUpPeriodSec == newRule.WarmUpPeriodSec &&
		r.WarmUpColdFactor == newRule.WarmUpColdFactor &&

		r.AdaptiveEnable == newRule.AdaptiveEnable && r.AdaptiveType == newRule.AdaptiveType &&
		r.SafeThreshold == newRule.SafeThreshold && r.RiskThreshold == newRule.RiskThreshold &&
		r.LowWaterMark == newRule.LowWaterMark && r.HighWaterMark == newRule.HighWaterMark) {

		return false
	}

	return true
}

// GetThreshold gets the real flow threshold.
func (r *Rule) GetThreshold() float64 {
	threshold := r.Threshold

	if r.AdaptiveEnable {
		if r.AdaptiveType == Memory {
			m := uint64(system_metric.CurrentMemoryUsage())
			if m <= r.LowWaterMark {
				threshold = float64(r.SafeThreshold)
			} else if m >= r.HighWaterMark {
				threshold = float64(r.RiskThreshold)
			} else {
				threshold = ((float64(m-r.LowWaterMark))/(float64(r.HighWaterMark-r.LowWaterMark)))*
					(float64(r.SafeThreshold-r.RiskThreshold)) + float64(r.RiskThreshold)
			}
		}
	}

	return threshold
}

func (r *Rule) isStatReusable(newRule *Rule) bool {
	if newRule == nil {
		return false
	}
	return r.Resource == newRule.Resource && r.RelationStrategy == newRule.RelationStrategy &&
		r.RefResource == newRule.RefResource && r.StatIntervalInMs == newRule.StatIntervalInMs &&
		r.needStatistic() && newRule.needStatistic()
}

func (r *Rule) needStatistic() bool {
	return !(r.TokenCalculateStrategy == Direct && r.ControlBehavior == Throttling)
}

func (r *Rule) String() string {
	b, err := json.Marshal(r)
	if err != nil {
		// Return the fallback string
		return fmt.Sprintf("Rule{Resource=%s, TokenCalculateStrategy=%s, ControlBehavior=%s, "+
			"Threshold=%.2f, RelationStrategy=%s, RefResource=%s, MaxQueueingTimeMs=%d, WarmUpPeriodSec=%d, WarmUpColdFactor=%d, StatIntervalInMs=%d, "+
			"AdaptiveEnable=%v, AdaptiveType=%s, SafeThreshold=%v, RiskThreshold=%v, LowWaterMark=%v, HighWaterMark=%v}",
			r.Resource, r.TokenCalculateStrategy, r.ControlBehavior, r.Threshold, r.RelationStrategy, r.RefResource,
			r.MaxQueueingTimeMs, r.WarmUpPeriodSec, r.WarmUpColdFactor, r.StatIntervalInMs,
			r.AdaptiveEnable, r.AdaptiveType, r.SafeThreshold, r.RiskThreshold, r.LowWaterMark, r.HighWaterMark)
	}
	return string(b)
}

func (r *Rule) ResourceName() string {
	return r.Resource
}
