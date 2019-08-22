package core

type SentinelInput struct {
	Context *Context
	// store some values in this context when calling context in slot.
	data map[interface{}]interface{}
}

func NewInput() *SentinelInput {
	return &SentinelInput{}
}

type SentinelOutput struct {
	CheckResult *RuleCheckResult
	msg         string
	// store output data.
	data map[interface{}]interface{}
}

func NewOutput() *SentinelOutput {
	return &SentinelOutput{}
}

type Context struct {
	ResWrapper  *ResourceWrapper
	Entry       *CtEntry
	Node        *ResourceNode
	Count       uint64
	Input       *SentinelInput
	Output      *SentinelOutput
	FeatureData map[interface{}]interface{}
}

func NewContext() *Context {
	return &Context{
		Input:  NewInput(),
		Output: NewOutput(),
	}
}

// Reset init Context,
func (ctx *Context) Reset() {
	// reset all fields of ctx
}

// Abort stops this transaction.
func (ctx *Context) Abort() {
	// abort the slot chain
}
