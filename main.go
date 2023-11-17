package main

import (
	"fmt"

	"github.com/gboncoffee/egg/riscv"
)

func main() {
	fmt.Println("Hello, World!")

	if riscv.CreateMachine() {
		fmt.Println("yes")
	} else {
		fmt.Println("no")
	}
}
