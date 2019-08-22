package core

import (
	"github.com/sentinel-group/sentinel-golang/core/util"
	"log"
)

type StatisticSlot struct {
	sc *SlotChain
}

func NewStatisticSlot(sc *SlotChain) *StatisticSlot {
	return &StatisticSlot{
		sc: sc,
	}
}

func (fs *StatisticSlot) OnEntryPassed(ctx *Context) {
	log.Println("StatisticSlot, OnEntryPassed")
	defer func() {
		if e := recover(); e != nil {
			log.Println("StatisticSlot,log error, e=", e)
			// TODO, if throw panic, how to handle.
		}
	}()

	// do statistic
	node := ctx.Node
	count := ctx.Count
	rw := ctx.ResWrapper

	node.IncreaseGoroutineNum()
	node.IncreasePassRequest(count)

	if rw.FlowType == InBound {
		InBoundEntryNode.IncreaseGoroutineNum()
		InBoundEntryNode.IncreasePassRequest(count)
	}
}

func (fs *StatisticSlot) OnEntryBlocked(ctx *Context) {
	log.Println("StatisticSlot, OnEntryBlocked")
	defer func() {
		if e := recover(); e != nil {
			log.Println("StatisticSlot,log error, e=", e)
			// TODO, if throw panic, how to handle.
		}
	}()

	// do statistic
	node := ctx.Node
	count := ctx.Count
	rw := ctx.ResWrapper

	node.IncreaseGoroutineNum()
	node.IncreaseBlockRequest(count)
	if rw.FlowType == InBound {
		InBoundEntryNode.IncreaseGoroutineNum()
		InBoundEntryNode.IncreaseBlockRequest(count)
	}
	return
}

func (fs *StatisticSlot) OnCompleted(ctx *Context) {
	log.Println("StatisticSlot, Completed")
	node := ctx.Node
	count := ctx.Count
	rw := ctx.ResWrapper
	entry := ctx.Entry

	rt := util.GetTimeMilli() - entry.CreateTime
	node.IncreaseSuccessRequest(count)
	node.AddRt(rt)
	node.DecreaseGoroutineNum()

	if rw.FlowType == InBound {
		InBoundEntryNode.IncreaseSuccessRequest(count)
		InBoundEntryNode.AddRt(rt)
		InBoundEntryNode.DecreaseGoroutineNum()
	}
}
