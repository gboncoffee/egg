package main

import (
	"flag"

	"github.com/gboncoffee/egg/riscv"
	"github.com/gboncoffee/egg/machine"
)

func runMachine(m machine.Machine) {
	panic("not yet implemented")
}

func debugMachine(m machine.Machine) {
	panic("not yet implemented")
}

func main() {
	var architeture string
	var debug bool
	var file string
	var outf string
	var m machine.Machine

	// Add new architetures to the help string!
	flag.StringVar(&architeture, "arch", "riscv", "Select architeture to use. Currently available: 'riscv'.")
	flag.StringVar(&architeture, "a", "riscv", "Select architeture to use (shorthand).")
	flag.BoolVar(&debug, "debug", false, "Enter debugger upon startup.")
	flag.BoolVar(&debug, "d", false, "Enter debugger upon startup (shorthand).")
	flag.StringVar(&file, "file", "", "Assembly source or ELF file to load.")
	flag.StringVar(&file, "f", "", "Assembly source or ELF file to load (shorthand).")
	flag.StringVar(&outf, "out", "", "ELF file to save assembled code.")
	flag.StringVar(&outf, "o", "", "ELF file to save assembled code (shorthand).")

	flag.Parse()

	// And the switch case!
	switch architeture {
	case "riscv":
		var r riscv.RiscV
		m = &r
	}

	if debug {
		debugMachine(m)
	} else {
		runMachine(m)
	}
}
