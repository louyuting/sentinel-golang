package core

// Check all rules about the resource
func Entry(name string) *CtEntry {
	return Entry2(name, 1)
}

func Entry2(name string, count uint64) *CtEntry {
	return Entry3(name, count, InBound)
}

func Entry3(name string, count uint64, entryType EntryType) *CtEntry {
	rw := &ResourceWrapper{
		ResourceName: name,
		FlowType:     entryType,
	}

	sc := GetDefaultSlotChain()
	// get context
	ctx := sc.GetContext()
	ctx.ResWrapper = rw
	ctx.Node = FindNode(rw)
	ctx.Count = count
	ctx.Entry = NewCtEntry(ctx, rw, sc, ctx.Node)

	sc.Entry(ctx)

	return ctx.Entry
}
