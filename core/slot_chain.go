package core

import (
	"github.com/sentinel-group/sentinel-golang/core/util"
	"log"
	"sync"
)

var DefaultSlotChain = buildDefaultSlotChain()

type StatPrepareSlot interface {
	// Prepare function do some initialization
	// Such as: init statistic structure、node、etc
	// All StatPrepareSlots execute sequentially
	// Prepare function should not throw panic.
	Prepare(ctx *Context)
}

type RuleCheckSlot interface {
	// Check function do some validation
	// It can break off slot pipeline
	// Each RuleCheckSlot will return slot check result
	// The upper logic will control pipeline according to SlotResult.
	Check(ctx *Context) *RuleCheckResult
}

type StatSlot interface {
	// OnEntryPass function will be invoked when StatPrepareSlots and RuleCheckSlots execute pass
	// StatSlots will do some statistic logic, such as QPS、log、etc
	OnEntryPassed(ctx *Context)
	// OnEntryBlocked function will be invoked when StatPrepareSlots and RuleCheckSlots execute fail
	// StatSlots will do some statistic logic, such as QPS、log、etc
	OnEntryBlocked(ctx *Context)
	// onComplete function will be invoked when chain exits.
	OnCompleted(ctx *Context)
}

type SlotType int8

const (
	StatisticPrepare SlotType = iota
	RuleCheck
	Statistic
)

//
type SlotChain struct {
	statPres   []StatPrepareSlot
	ruleChecks []RuleCheckSlot
	stats      []StatSlot

	// Context Pool
	pool sync.Pool
}

func newSlotChain() *SlotChain {
	return &SlotChain{
		statPres:   make([]StatPrepareSlot, 0, 10),
		ruleChecks: make([]RuleCheckSlot, 0, 10),
		stats:      make([]StatSlot, 0, 10),
		pool: sync.Pool{
			New: func() interface{} {
				return NewContext()
			},
		},
	}
}

func buildDefaultSlotChain() *SlotChain {
	sc := newSlotChain()
	// insert slots
	sc.addStatPrepareSlotLast(NewResourceBuilderSlot(sc))
	sc.addRuleCheckSlotLast(NewFlowSlot(sc))
	sc.addRuleCheckSlotLast(NewDegradeSlot(sc))
	sc.addStatSlotLast(NewStatisticSlot(sc))
	return sc
}

func GetDefaultSlotChain() *SlotChain {
	return DefaultSlotChain
}

func (sc *SlotChain) GetContext() *Context {
	ctx := sc.pool.Get().(*Context)
	defer sc.pool.Put(ctx)
	return ctx
}

func (sc *SlotChain) addStatPrepareSlotFirst(s StatPrepareSlot) {
	ns := make([]StatPrepareSlot, 0, len(sc.statPres)+1)
	// add to first
	ns = append(ns, s)
	sc.statPres = append(ns, sc.statPres...)
}

func (sc *SlotChain) addStatPrepareSlotLast(s StatPrepareSlot) {
	sc.statPres = append(sc.statPres, s)
}

func (sc *SlotChain) addRuleCheckSlotFirst(s RuleCheckSlot) {
	ns := make([]RuleCheckSlot, 0, len(sc.ruleChecks)+1)
	ns = append(ns, s)
	sc.ruleChecks = append(ns, sc.ruleChecks...)
}

func (sc *SlotChain) addRuleCheckSlotLast(s RuleCheckSlot) {
	sc.ruleChecks = append(sc.ruleChecks, s)
}

func (sc *SlotChain) addStatSlotFirst(s StatSlot) {
	ns := make([]StatSlot, 0, len(sc.stats)+1)
	ns = append(ns, s)
	sc.stats = append(ns, sc.stats...)
}

func (sc *SlotChain) addStatSlotLast(s StatSlot) {
	sc.stats = append(sc.stats, s)
}

// The entrance of Slot Chain
func (sc *SlotChain) Entry(ctx *Context) {
	log.Println("slot chain entry")
	startTime := util.GetTimeMilli()
	// prepare slot
	log.Println("prepare slot")
	sps := sc.statPres
	if len(sps) > 0 {
		for _, s := range sps {
			s.Prepare(ctx)
		}
	}

	// rule base check
	log.Println("rule base slot check")
	rcs := sc.ruleChecks
	ruleCheckRet := NewSlotResultPass()
	if len(rcs) > 0 {
		for _, s := range rcs {
			sr := s.Check(ctx)
			// check slot result
			if sr.Status == ResultStatusPass {
				continue
			}
			// block or other logic
			log.Printf("%v check fail, reason is %s \n", s, sr.BlockedMsg)
			ruleCheckRet.Status = ResultStatusBlocked
			break
		}
	}

	// statistic slot
	log.Println("statistic slot")
	ss := sc.stats
	if len(ss) > 0 {
		for _, s := range ss {
			// indicate the result of rule base check slot.
			if ruleCheckRet.Status == ResultStatusPass {
				s.OnEntryPassed(ctx)
			} else {
				s.OnEntryBlocked(ctx)
			}
		}
	}
	LogAccess(ctx, startTime)
}

func (sc *SlotChain) Exit(ctx *Context) {
	log.Println("slot chain exit")
	startTime := util.GetTimeMilli()
	for _, s := range sc.stats {
		s.OnCompleted(ctx)
	}
	LogAccess(ctx, startTime)
}

func LogAccess(ctx *Context, startTime uint64) {
	log.Println("start:", startTime, "end:", util.GetTimeMilli())
}
