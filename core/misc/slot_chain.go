package misc

import (
	"sync"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/core/isolation"
	"github.com/alibaba/sentinel-golang/core/log"
	"github.com/alibaba/sentinel-golang/core/stat"
	"github.com/alibaba/sentinel-golang/core/system"
)

var (
	globalDefaultSlotChain = BuildDefaultSlotChain()

	rsSlotChainLock sync.RWMutex
	rsSlotChain     = make(map[string]*base.SlotChain, 8)

	globalStatPrepareSlots = make([]base.StatPrepareSlot, 0, 8)
	globalRuleCheckSlots   = make([]base.RuleCheckSlot, 0, 8)
	globalStatSlot         = make([]base.StatSlot, 0, 8)
)

// RegisterGlobalStatPrepareSlot is not thread safe, and user must call RegisterGlobalStatPrepareSlot when initializing sentinel running environment
func RegisterGlobalStatPrepareSlot(slot base.StatPrepareSlot) {
	for _, s := range globalStatPrepareSlots {
		if s.Name() == slot.Name() {
			return
		}
	}
	globalStatPrepareSlots = append(globalStatPrepareSlots, slot)
}

// RegisterGlobalRuleCheckSlot is not thread safe, and user must call RegisterGlobalRuleCheckSlot when initializing sentinel running environment
func RegisterGlobalRuleCheckSlot(slot base.RuleCheckSlot) {
	for _, s := range globalRuleCheckSlots {
		if s.Name() == slot.Name() {
			return
		}
	}
	globalRuleCheckSlots = append(globalRuleCheckSlots, slot)
}

// RegisterGlobalStatSlot is not thread safe, and user must call RegisterGlobalStatSlot when initializing sentinel running environment
func RegisterGlobalStatSlot(slot base.StatSlot) {
	for _, s := range globalStatSlot {
		if s.Name() == slot.Name() {
			return
		}
	}
	globalStatSlot = append(globalStatSlot, slot)
}

func GlobalDefaultSlotChain() *base.SlotChain {
	return globalDefaultSlotChain
}

func BuildDefaultSlotChain() *base.SlotChain {
	sc := base.NewSlotChain()
	sc.AddStatPrepareSlotLast(stat.DefaultResourceNodePrepareSlot)
	for _, s := range globalStatPrepareSlots {
		sc.AddStatPrepareSlotLast(s)
	}

	sc.AddRuleCheckSlotLast(system.DefaultAdaptiveSlot)
	sc.AddRuleCheckSlotLast(flow.DefaultSlot)
	sc.AddRuleCheckSlotLast(isolation.DefaultSlot)
	sc.AddRuleCheckSlotLast(circuitbreaker.DefaultSlot)
	sc.AddRuleCheckSlotLast(hotspot.DefaultSlot)
	for _, s := range globalRuleCheckSlots {
		sc.AddRuleCheckSlotLast(s)
	}

	sc.AddStatSlotLast(stat.DefaultSlot)
	sc.AddStatSlotLast(log.DefaultSlot)
	sc.AddStatSlotLast(circuitbreaker.DefaultMetricStatSlot)
	sc.AddStatSlotLast(hotspot.DefaultConcurrencyStatSlot)
	sc.AddStatSlotLast(flow.DefaultStandaloneStatSlot)
	for _, s := range globalStatSlot {
		sc.AddStatSlotLast(s)
	}

	return sc
}

func newSlotChain() *base.SlotChain {
	sc := base.NewSlotChain()
	sc.AddStatPrepareSlotLast(stat.DefaultResourceNodePrepareSlot)
	for _, s := range globalStatPrepareSlots {
		sc.AddStatPrepareSlotLast(s)
	}

	for _, s := range globalRuleCheckSlots {
		sc.AddRuleCheckSlotLast(s)
	}

	sc.AddStatSlotLast(stat.DefaultSlot)
	sc.AddStatSlotLast(log.DefaultSlot)
	for _, s := range globalStatSlot {
		sc.AddStatSlotLast(s)
	}

	return sc
}

func validateRuleCheckSlot(sc *base.SlotChain, s base.RuleCheckSlot) bool {
	flag := false
	f := func(slot base.RuleCheckSlot) {
		if slot.Name() == s.Name() {
			flag = true
		}
	}
	sc.RangeRuleCheckSlot(f)

	return flag
}

func validateStatSlot(sc *base.SlotChain, s base.StatSlot) bool {
	flag := false
	f := func(slot base.StatSlot) {
		if slot.Name() == s.Name() {
			flag = true
		}
	}
	sc.RangeStatSlot(f)

	return flag
}

func RegisterResourceRuleCheckSlot(rsName string, slot base.RuleCheckSlot) {
	rsSlotChainLock.Lock()
	defer rsSlotChainLock.Unlock()

	sc, ok := rsSlotChain[rsName]
	if !ok {
		sc = newSlotChain()
	}

	if !validateRuleCheckSlot(sc, slot) {
		sc.AddRuleCheckSlotLast(slot)
	}
	if !ok {
		rsSlotChain[rsName] = sc
	}
}

func RegisterResourceStatSlot(rsName string, slot base.StatSlot) {
	rsSlotChainLock.Lock()
	defer rsSlotChainLock.Unlock()

	sc, ok := rsSlotChain[rsName]
	if !ok {
		sc = newSlotChain()
	}

	if !validateStatSlot(sc, slot) {
		sc.AddStatSlotLast(slot)
	}
	if !ok {
		rsSlotChain[rsName] = sc
	}
}

func GetResourceSlotChain(rsName string) *base.SlotChain {
	rsSlotChainLock.RLock()
	defer rsSlotChainLock.RUnlock()

	sc, ok := rsSlotChain[rsName]
	if !ok {
		return nil
	}

	return sc
}
