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
			//double confirm
			n = NewResourceNode(rw)
			cacheNodeHold := make(map[string]*ResourceNode)
			for k, v := range NodeHolder {
				cacheNodeHold[k] = v
			}
			cacheNodeHold[rw.ResourceName] = n
			NodeHolder = cacheNodeHold
		}
		lock.Unlock()
		return NodeHolder[rw.ResourceName]
	} else {
		return node
	}
}
