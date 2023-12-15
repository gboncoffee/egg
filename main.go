package main

import (
	"flag"
	"log"
	"fmt"
	"os"

	"github.com/gboncoffee/egg/riscv"
	"github.com/gboncoffee/egg/machine"
)

// Put new architetures here...
func listArchs() {
	fmt.Println(`Currently supported architetures:
'riscv' - RISC-V IM, 32 bits`)
}

func version() {
	fmt.Println("EGG - Emulador Genérico do Gabriel - version 0.0.1rc")
}

func runMachine(m machine.Machine) {
	for {
		call, err := m.NextInstruction()
		if err != nil {
			log.Println(fmt.Sprintf("Instruction execution failed: %v", err))
			return
		}

		if call != nil {
			if call.Number == machine.SYS_BREAK {
				return
			}
		}
	}
}

func main() {
	var architeture string
	var debug bool
	var list bool
	var m machine.Machine

	log.SetFlags(0)

	flag.StringVar(&architeture, "arch", "riscv", "Select architeture to use.")
	flag.StringVar(&architeture, "a", "riscv", "Select architeture to use (shorthand).")
	flag.BoolVar(&list, "list-archs", false, "Lists currently supported architetures and quit.")
	flag.BoolVar(&list, "l", false, "Lists currently supported architetures (shorthand).")
	flag.BoolVar(&debug, "debug", false, "Enter debugger upon startup.")
	flag.BoolVar(&debug, "d", false, "Enter debugger upon startup (shorthand).")

	flag.Parse()

	if list {
		version()
		listArchs()
		return
	}

	// ...and in the switch case!
	switch architeture {
	case "riscv":
		var r riscv.RiscV
		m = &r
	default:
		log.Println(fmt.Sprintf("Unknown architeture: %v", architeture))
		listArchs()
		os.Exit(1)
	}

	file := flag.Arg(0)
	if file == "" {
		log.Println("No Assembly file supplied.")
		os.Exit(1)
	}

	if debug {
		// Hello fellow Acme user. Plumb this: debugger.go:/debugMachine
		debugMachine(m)
	} else {
		runMachine(m)
	}
}
