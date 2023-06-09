package main

import (
	"fmt"
	"minidocker/cmd"
	"runtime/debug"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		fmt.Println(string(debug.Stack()))
	}
}
