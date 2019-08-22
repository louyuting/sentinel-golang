package core

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
)

func TestFindNodeNewNode(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(100)
	for i := 1; i <= 100; i++ {
		key := "a" + strconv.Itoa(i)
		go func(wgp *sync.WaitGroup) {
			FindNode(&ResourceWrapper{
				ResourceName: key,
				FlowType:     InBound,
			})
			wg.Done()
		}(wg)
	}
	wg.Wait()
	if 100 != len(NodeHolder) {
		t.Error("TestFindNodeNewNode fail")
	}
	fmt.Printf("NodeHolder: %v \n", NodeHolder)
}

func TestFindNodeConcurrent(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(5000)
	for i := 1; i <= 5000; i++ {
		key := "a" + strconv.Itoa(i)
		go func(wgp *sync.WaitGroup) {
			FindNode(&ResourceWrapper{
				ResourceName: key,
				FlowType:     InBound,
			})
			wgp.Done()
		}(wg)
	}

	wg2 := &sync.WaitGroup{}
	wg2.Add(5000)
	for i := 1; i <= 5000; i++ {
		key := "a" + strconv.Itoa(i)
		go func(wgp *sync.WaitGroup) {
			FindNode(&ResourceWrapper{
				ResourceName: key,
				FlowType:     InBound,
			})
			wgp.Done()
		}(wg2)
	}

	wg.Wait()
	wg2.Wait()

	if 5000 != len(NodeHolder) {
		t.Error("TestFindNodeNewNode fail")
	}
	fmt.Printf("NodeHolder: %v \n", NodeHolder)
}
