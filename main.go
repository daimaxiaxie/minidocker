package main

import (
	"fmt"
	"minidocker/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
