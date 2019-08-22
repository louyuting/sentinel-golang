package core

type FlowSlot struct {
	sc *SlotChain
	rm *RuleManager
}

func NewFlowSlot(sc *SlotChain) *FlowSlot {
	return &FlowSlot{
		sc: sc,
	}
}

func (fs *FlowSlot) Check(ctx *Context) *RuleCheckResult {
	// no rule return pass
	if fs.rm == nil {
		return NewSlotResultPass()
	}
	rw := ctx.ResWrapper
	node := ctx.Node
	count := 0
	rules := fs.rm.getRuleBySource(rw.ResourceName)
	if len(rules) == 0 {
		return NewSlotResultPass()
	}
	success := checkFlow(ctx, rw, rules, node, count)
	if success {
		return NewSlotResultPass()
	} else {
		return NewSlotResultBlock("check fail")
	}
}

func (fs *FlowSlot) Exit(ctx *Context) {
}

func checkFlow(ctx *Context, resourceWrap *ResourceWrapper, rules []*rule, node *ResourceNode, count int) bool {
	if rules == nil {
		return true
	}
	for _, rule := range rules {
		if !canPass(ctx, resourceWrap, rule, node, uint32(count)) {
			return false
		}
	}
	return true
}

func canPass(ctx *Context, resourceWrap *ResourceWrapper, rule *rule, node *ResourceNode, count uint32) bool {
	return rule.controller_.CanPass(ctx, node, count)
}
