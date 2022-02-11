//go:build debug
// +build debug

package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
)

func init() {
	fmt.Println("starting pprof server")
	go http.ListenAndServe(pprofAddr, nil)
}
