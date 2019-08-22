package core

type SlotResultStatus int8

const (
	ResultStatusPass SlotResultStatus = iota
	ResultStatusBlocked
	ResultStatusError
)

type RuleCheckResult struct {
	Status     SlotResultStatus
	BlockedMsg string
	ErrorMsg   string
}

func NewSlotResultPass() *RuleCheckResult {
	return &RuleCheckResult{Status: ResultStatusPass}
}

func NewSlotResultBlock(blockedReason string) *RuleCheckResult {
	return &RuleCheckResult{Status: ResultStatusBlocked, BlockedMsg: blockedReason}
}

func NewSlotResultError(errorMsg string) *RuleCheckResult {
	return &RuleCheckResult{Status: ResultStatusError, ErrorMsg: errorMsg}
}
