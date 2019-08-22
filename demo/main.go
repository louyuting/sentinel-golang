package main

import (
	"fmt"
	"github.com/sentinel-group/sentinel-golang/core"
	"github.com/sentinel-group/sentinel-golang/core/log"
)

func main() {
	fmt.Println("=================start=================")
	log.Init()
	core.Entry1("aaaaaa")
	fmt.Println("=================end=================")
}
