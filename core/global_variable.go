package core

var (
	TotalInBoundResourceName = "__total_inbound_traffic__"
	InBoundEntryNode         = NewResourceNode(&ResourceWrapper{
		ResourceName: TotalInBoundResourceName,
		FlowType:     InBound,
	})
)
