package main

import (
	"fmt"
	"go_keepalived/cfgparser"
	"os"
	"sync"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go_keepalived <cfg.file>")
		os.Exit(-1)
	}
	slList := cfgparser.ReadCfg(os.Args[1])
	fmt.Println(*slList)
	slList.Start()
	wg := sync.WaitGroup{}
	(&wg).Add(1)
	(&wg).Wait()
}
