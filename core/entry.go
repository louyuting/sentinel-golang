package core

import "github.com/sentinel-group/sentinel-golang/core/util"

// EntryType describe the traffic type: Inbound or OutBound
type EntryType int32

const (
	InBound EntryType = iota
	OutBound
)

type ResourceWrapper struct {
	// global unique resource name
	ResourceName string
	// InBound or OutBound
	FlowType EntryType
}

// CtEntry means Context Entry,
type CtEntry struct {
	CreateTime uint64
	rs         *ResourceWrapper
	// one entry with one context
	ctx *Context
	// each entry holds a slot chain.
	// it means this entry will go through the sc
	sc *SlotChain
	// caller node
	originNode *ResourceNode
	// current resource node
	currentNode *ResourceNode
}

type AsyncEntry struct {
	CtEntry
	asyncContext *Context
}

func NewCtEntry(ctx *Context, rw *ResourceWrapper, sc *SlotChain, cn *ResourceNode) *CtEntry {
	return &CtEntry{
		CreateTime:  util.GetTimeMilli(),
		rs:          rw,
		ctx:         ctx,
		sc:          sc,
		currentNode: cn,
	}
}

func (e *CtEntry) Exit1() {
	e.Exit2(1)
}

func (e *CtEntry) Exit2(count int32) {
	e.exitForContext(e.ctx, count)
}

func (e *CtEntry) exitForContext(ctx *Context, count int32) {
	if e.sc != nil {
		e.sc.Exit(ctx)
	}
}

func (e *CtEntry) GetCurrentNode() *ResourceNode {
	return e.currentNode
}
