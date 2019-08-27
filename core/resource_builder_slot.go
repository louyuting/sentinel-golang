package core

import "log"

type ResourceBuilderSlot struct {
	sc *SlotChain
}

func NewResourceBuilderSlot(sc *SlotChain) *ResourceBuilderSlot {
	return &ResourceBuilderSlot{
		sc: sc,
	}
}

func (rbs *ResourceBuilderSlot) Prepare(ctx *Context) {
	log.Println("ResourceBuilderSlot#Prepare")
	return
}
