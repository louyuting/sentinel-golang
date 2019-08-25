package main

import (
	"github.com/sentinel-group/sentinel-golang/core"
	internalLog "github.com/sentinel-group/sentinel-golang/core/log"
	"log"
	"time"
)

func main() {
	internalLog.Init()
	log.Println("=================start=================")
	cxe := core.Entry("aaaaaa")
	//
	time.Sleep(time.Second * 1)
	log.Println("Call service")

	cxe.Exit1()

	if cxe.GetCurrentNode().TotalSuccess() == 1 {
		log.Println("total success == 1")
	}
	if cxe.GetCurrentNode().TotalRequest() == 1 {
		log.Println("total request == 1")
	}
	if cxe.GetCurrentNode().TotalPass() == 1 {
		log.Println("total pass == 1")
	}
	log.Println("=================end=================")
}
