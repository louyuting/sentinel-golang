package core

type DegradeSlot struct {
	sc *SlotChain
}

func NewDegradeSlot(sc *SlotChain) *DegradeSlot {
	return &DegradeSlot{
		sc: sc,
	}
}

func (ds *DegradeSlot) Check(ctx *Context) *RuleCheckResult {
	return NewSlotResultPass()
}
