package core

import (
	"sync"
)

var (
	lock       = sync.Mutex{}
	NodeHolder = make(map[string]*ResourceNode)
)

func FindNode(rw *ResourceWrapper) *ResourceNode {
	node, ok := NodeHolder[rw.ResourceName]
	if !ok {
		// node is not existed
		// must be concurrent safe.
		lock.Lock()
		n, ok := NodeHolder[rw.ResourceName]
		if !ok {
			n = NewResourceNode(rw)
			NodeHolder[rw.ResourceName] = n
		}
		node = n
		lock.Unlock()
	}
	return node
}
