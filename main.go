package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gboncoffee/egg/assembler"
	"github.com/gboncoffee/egg/machine"
	"github.com/gboncoffee/egg/mips"
	"github.com/gboncoffee/egg/riscv"
	"github.com/gboncoffee/egg/sagui"
)

const VERSION = "3.1.0"

// Put new architetures here... (main.go:/switch architeture)
func listArchs() {
	fmt.Println(machine.InterCtx.Get(`Currently supported architetures:
'riscv' - RISC-V IM, 32 bits
'mips'  - Subset of MIPS32
'sagui' - Fantasy 8 bit RISC`))
}

func version() {
	fmt.Println(machine.InterCtx.Get("EGG - Emulador Genérico do Gabriel - version ") + VERSION)
}

func runMachine(m machine.Machine) {
	for {
		call, err := m.NextInstruction()
		if err != nil {
			log.Printf(machine.InterCtx.Get("Instruction execution failed: %v\n"), err)
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

func main() {
	var architeture string
	var debug bool
	var list bool
	var ver bool
	var m machine.Machine

	machine.InterCtx.Init()
	machine.InterCtx.AddLocale("pt_BR", brazilian)
	machine.InterCtx.AutoSetPreferedLocale()
	assembler.InterCtx = &machine.InterCtx

	log.SetFlags(0)

	flag.StringVar(&architeture, "arch", "riscv", machine.InterCtx.Get("Select architeture to use."))
	flag.StringVar(&architeture, "a", "riscv", machine.InterCtx.Get("Select architeture to use (shorthand)."))
	flag.BoolVar(&list, "list-archs", false, machine.InterCtx.Get("Lists currently supported architetures and quit."))
	flag.BoolVar(&list, "l", false, machine.InterCtx.Get("Lists currently supported architetures (shorthand)."))
	flag.BoolVar(&ver, "version", false, machine.InterCtx.Get("Show current version and quit."))
	flag.BoolVar(&ver, "v", false, machine.InterCtx.Get("Show current version and quit (shorthand)."))
	flag.BoolVar(&debug, "debug", false, machine.InterCtx.Get("Enter debugger upon startup."))
	flag.BoolVar(&debug, "d", false, machine.InterCtx.Get("Enter debugger upon startup (shorthand)."))

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
		log.Printf(machine.InterCtx.Get("Unknown architeture: %v\n"), architeture)
		listArchs()
		os.Exit(1)
	}

	file := flag.Arg(0)
	if file == "" {
		log.Println(machine.InterCtx.Get("No Assembly file supplied."))
		os.Exit(1)
	}

	code, sym, err := m.Assemble(file)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	err = m.LoadProgram(code)
	if err != nil {
		log.Printf(machine.InterCtx.Get("Error loading assembled program: %v\n"), err)
		os.Exit(1)
	}

	if debug {
		if sym == nil {
			log.Println(machine.InterCtx.Get("Debugging is not supported for the selected backend."))
			os.Exit(1)
		}
		// Hello fellow Acme user. Plumb this: debugger.go:/debugMachine
		debugMachine(m, sym, code, file)
	} else {
		runMachine(m)
	}
}
