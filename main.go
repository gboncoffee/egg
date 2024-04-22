package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gboncoffee/egg/machine"
	"github.com/gboncoffee/egg/mips"
	"github.com/gboncoffee/egg/riscv"
	"github.com/gboncoffee/egg/sagui"
)

// Put new architetures here... (main.go:/switch architeture)
func listArchs() {
	fmt.Println(`Currently supported architetures:
'riscv' - RISC-V IM, 32 bits
'mips'  - Subset of MIPS32 (experimental)
'sagui' - Fantasy 8 bit RISC`)
}

func version() {
	fmt.Println("EGG - Emulador Genérico do Gabriel - version 2.2.0")
}

func runMachine(m machine.Machine) {
	for {
		call, err := m.NextInstruction()
		if err != nil {
			log.Printf("Instruction execution failed: %v\n", err)
			return
		}

		if call != nil {
			switch call.Number {
			case machine.SYS_BREAK:
				return
			case machine.SYS_READ:
				// Lol I'm using 3 std modules just to read from stdin.
				addr := call.Arg1
				size := call.Arg2
				buf := make([]uint8, size)
				reader := bufio.NewReader(os.Stdin)
				io.ReadFull(reader, buf)
				m.SetMemoryChunk(addr, buf)
			case machine.SYS_WRITE:
				addr := call.Arg1
				size := call.Arg2
				buf, _ := m.GetMemoryChunk(addr, size)
				fmt.Print(string(buf))
			}
		}
	}
}

func readToString(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func main() {
	var architeture string
	var debug bool
	var list bool
	var ver bool
	var m machine.Machine

	log.SetFlags(0)

	flag.StringVar(&architeture, "arch", "riscv", "Select architeture to use.")
	flag.StringVar(&architeture, "a", "riscv", "Select architeture to use (shorthand).")
	flag.BoolVar(&list, "list-archs", false, "Lists currently supported architetures and quit.")
	flag.BoolVar(&list, "l", false, "Lists currently supported architetures (shorthand).")
	flag.BoolVar(&ver, "version", false, "Show current version and quit.")
	flag.BoolVar(&ver, "v", false, "Show current version and quit (shorthand).")
	flag.BoolVar(&debug, "debug", false, "Enter debugger upon startup.")
	flag.BoolVar(&debug, "d", false, "Enter debugger upon startup (shorthand).")

	flag.Parse()

	if list {
		version()
		listArchs()
		return
	}

	if ver {
		version()
		return
	}

	// ...and in the switch case! (main.go:/func listArchs)
	switch architeture {
	case "riscv":
		var r riscv.RiscV
		m = &r
	case "mips":
		var r mips.Mips
		m = &r
	case "sagui":
		var r sagui.Sagui
		m = &r
	default:
		log.Printf("Unknown architeture: %v", architeture)
		listArchs()
		os.Exit(1)
	}

	file := flag.Arg(0)
	if file == "" {
		log.Println("No Assembly file supplied.")
		os.Exit(1)
	}

	asm, err := readToString(file)
	if err != nil {
		log.Printf("Could not read supplied file %v", file)
		os.Exit(1)
	}

	code, sym, err := m.Assemble(asm)
	if err != nil {
		log.Printf("Error assembling file %v: %v\n", file, err)
		os.Exit(1)
	}

	err = m.LoadProgram(code)
	if err != nil {
		log.Printf("Error loading assembled program: %v\n", err)
		os.Exit(1)
	}

	if debug {
		if sym == nil {
			log.Println("Debugging is not supported for the selected backend.")
			os.Exit(1)
		}
		// Hello fellow Acme user. Plumb this: debugger.go:/debugMachine
		debugMachine(m, sym, code)
	} else {
		runMachine(m)
	}
}
