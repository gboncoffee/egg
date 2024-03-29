package main

import (
	"fmt"
	"bufio"
	"os"
	"strings"
	"strconv"
	"errors"
	"sort"

	"github.com/gboncoffee/egg/machine"
	"github.com/gboncoffee/egg/assembler"
)

func debuggerHelp() {
	help := `Commands:

help
	Shows this help.
	Shortcut: h
print <expr>[@<length>]
	Prints the content of registers and memory.
	Shortcut: p
next
	Executes the next instruction, then pauses.
	Shortcut: n
continue
	Continue execution until a BREAK call or breakpoint.
	Shortcut: c
break [expr]
	With an argument, creates a new breakpoint. With no argument, shows all
	breakpoints. Accepts numbers and Assembly labels.
	Shortcut: b
remove <expr>
	Removes a breakpoint. Accepts numbers and Assembly labels.
	Shortcut: r
dump <address>@<length> <filename>
	Dumps the content of memory to a file.
	Shortcut: d
rewind
	Reloads the machine, i.e., asks it to return to it's original state.
	Shortcut: rew
set <expr>[@<length>] <content>
	Changes the content of a register or memory.
	Shortcut: s
exit
	Terminate debugging session.
	Shortcut: e
quit
	Alias to exit.
	Shortcut: q

The print command generally follows this rules:
- If the expression is only a register (e.g., x1, t1, zero, ra, etc), it prints
  it's contents;
- If the expression is a register with a length (e.g., t1@1, ra@7, etc), it
  prints the content of the memory addressed by the content of the register.
The set command works the same way.

The dump command also accepts registers, but always dereference them.

Both print and dump commands accepts the special expression #, which means the
program itself. For example, you may dump the assembled program to a file by
running 'dump # file'.

In the print command, <addr>#<length> means "length instructions after addr".
#<length> is a shortcut to use with the current instruction address.
`
	fmt.Print(help)
}

func tokensToString(sym []assembler.DebuggerToken) []string {
	str := make([]string, len(sym))

	for i, tok := range sym {
		var build strings.Builder
		build.WriteString(fmt.Sprintf("0x%016x: ", tok.Address))
		if tok.Label != "" {
			build.WriteString(tok.Label)
			build.WriteString(": ")
		} else {
			build.WriteRune('\t')
		}
		build.WriteString(tok.Instruction)
		build.WriteRune(' ')
		build.WriteString(tok.Args)
		str[i] = build.String()
	}

	return str
}

func getPrintHashExpr(m machine.Machine, sym []assembler.DebuggerToken, expr string) ([]string, error) {
	fn, sn, _ := strings.Cut(expr, "#")
	fn = strings.TrimSpace(fn)
	sn = strings.TrimSpace(sn)

	var faddr uint64
	var l uint64
	if fn == "" {
		var err error

		if sn == "" {
			return tokensToString(sym), nil
		}

		l, err = strconv.ParseUint(sn, 0, 64)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Cannot parse %v as number: %v", sn, err))
		}

		faddr = m.GetCurrentInstructionAddress()
	} else {
		var err error

		if sn == "" {
			return nil, errors.New("Length not supplied")
		}

		faddr, err = strconv.ParseUint(fn, 0, 64)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Cannot parse %v as address.", fn))
		}

		l, err = strconv.ParseUint(sn, 0, 64)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%v a number.", sn))
		}
	}

	i := uint64(sort.Search(len(sym), func(i int) bool {
		return sym[i].Address >= faddr
	}))

	if i == uint64(len(sym)) {
		return nil, errors.New(fmt.Sprintf("No instruction at address 0x%x", faddr))
	}

	l = l + i
	if l >= uint64(len(sym)) {
		l = uint64(len(sym))
	}

	return tokensToString(sym[i:l]), nil
}

func getMemoryContentPrint(m machine.Machine, addr string, length string) ([]uint64, error) {
	var a uint64
	reg, err := m.GetRegisterNumber(addr)
	if err != nil {
		a, err = strconv.ParseUint(addr, 0, 64)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%v is not a register or address", addr))
		}
	} else {
		a, _ = m.GetRegister(reg)
	}

	l, err := strconv.ParseUint(length, 0, 64)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v is not a number.", length))
	}

	mem, err := m.GetMemoryChunk(a, l)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Cannot get memory content: %v", err))
	}

	// I hope this is somehow "optimized out" to a simple padded copy.
	m64 := make([]uint64, l)
	for i, v := range mem {
		m64[i] = uint64(v)
	}

	return m64, nil
}

func getPrintExpr(m machine.Machine, expr string) ([]string, error) {
	addr, length, has_at := strings.Cut(expr, "@")
	addr = strings.TrimSpace(addr)
	length = strings.TrimSpace(length)

	if !has_at {
		reg, err := m.GetRegisterNumber(addr)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Cannot get register content: %v", err))
		}
		c, _ := m.GetRegister(reg)

		return []string{fmt.Sprintf("0x%016x", c)}, nil
	}

	arr, err := getMemoryContentPrint(m, addr, length)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Cannot get memory content: %v", err))
	}

	c := make([]string, len(arr))
	for i, v := range arr {
		c[i] = fmt.Sprintf("0x%02x", v)
	}

	return c, nil
}

func debuggerPrint(m machine.Machine, sym []assembler.DebuggerToken, args []string) {
	if len(args) < 1 {
		fmt.Println("print expects one argument: <expr>[@<length>] or [<addr>]#[<length>]")
		return
	}

	expr := args[0]
	var res []string
	var err error
	if strings.IndexRune(expr, '#') >= 0 {
		res, err = getPrintHashExpr(m, sym, expr)
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			for _, v := range res {
				fmt.Println(v)
			}
		}
	} else {
		res, err = getPrintExpr(m, expr)
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			for i := 0; i < len(res)-1; i++ {
				if i % 8 == 7 {
					fmt.Printf("%s\n", res[i])
				} else {
					fmt.Printf("%s ", res[i])
				}
			}
			fmt.Printf("%s\n", res[len(res)-1])
		}
	}
}

func ioCall(m machine.Machine, call *machine.Call, in *bufio.Reader) {
	if call.Number == machine.SYS_READ {
		addr := call.Arg1
		size := call.Arg2
		fmt.Printf("READ call for address 0x%x with %d bytes:\n", addr, size)
		buf := make([]byte, size)
		_, err := in.Read(buf)
		if err != nil {
			fmt.Printf("Error reading stdin: %v\n", err)
			return
		}

		m.SetMemoryChunk(addr, []uint8(buf))
	} else {
		addr := call.Arg1
		size := call.Arg2
		buf, _ := m.GetMemoryChunk(addr, size)
		fmt.Print(string(buf))
	}
}

func debuggerNext(m machine.Machine, sym []assembler.DebuggerToken, in *bufio.Reader) {
	call, err := m.NextInstruction()
	if err != nil {
		fmt.Println(fmt.Sprintf("Instruction execution failed: %v", err))
		return
	}

	pc := m.GetCurrentInstructionAddress()
	if call != nil {
		if call.Number == machine.SYS_BREAK {
			fmt.Printf("BREAK call while stepping at address 0x%x\n", pc)
		} else {
			ioCall(m, call, in)
		}
	}

	debuggerPrint(m, sym, []string{"#3"})
}

func debuggerContinue(m machine.Machine, sym []assembler.DebuggerToken, breakpoints []uint64, in *bufio.Reader) {
	for {
		call, err := m.NextInstruction()
		if err != nil {
			fmt.Println(fmt.Sprintf("Instruction execution failed: %v", err))
			return
		}

		pc := m.GetCurrentInstructionAddress()
		if call != nil {
			if call.Number == machine.SYS_BREAK {
				fmt.Printf("Stopped at BREAK call at address 0x%x\n", pc)
				debuggerPrint(m, sym, []string{"#3"})
				return
			} else {
				ioCall(m, call, in)
			}
		}

		_, brk := sort.Find(len(breakpoints), func(i int) int {
			// This one MUST be optimized.
			if pc >= breakpoints[i] {
				if breakpoints[i] != pc {
					return 1
				} else {
					return 0
				}
			} else {
				return -1
			}
		})

		if brk {
			fmt.Printf("Stopped at breakpoint at address 0x%x\n", pc)
			debuggerPrint(m, sym, []string{"#3"})
			return
		}
	}
}

func printBreakpoints(breakpoints []uint64) {
	fmt.Println("Breakpoints:")
	for _, b := range breakpoints {
		fmt.Printf("0x%016x\n", b)
	}
}

func debuggerBreakpoint(m machine.Machine, sym []assembler.DebuggerToken, breakpoints *[]uint64, args []string) {
	if len(args) < 1 {
		printBreakpoints(*breakpoints)
		return
	}

	var addr uint64
	var found bool
	for _, s := range sym {
		if s.Label == args[0] {
			found = true
			addr = s.Address
			break
		}
	}

	if !found {
		var err error
		addr, err = strconv.ParseUint(args[0], 0, 64)
		if err != nil {
			fmt.Printf("%v is not a number.\n", args[0])
			return
		}
	}

	n_brks := make([]uint64, len(*breakpoints)+1)
	var brk_idx int
	for i, p := range *breakpoints {
		if p < addr {
			n_brks[i] = p
			brk_idx++
		} else if p == addr {
			fmt.Println("Breakpoint already exists")
			return
		} else {
			break
		}
	}

	n_brks[brk_idx] = addr
	if brk_idx < len(*breakpoints) {
		for i, p := range (*breakpoints)[brk_idx:] {
			n_brks[i+brk_idx+1] = p
		}
	}

	*breakpoints = n_brks

	fmt.Printf("New breakpoint at address 0x%x\n", addr)
	printBreakpoints(*breakpoints)
}

func debuggerRemove(m machine.Machine, sym []assembler.DebuggerToken, breakpoints *[]uint64, args []string) {
	if len(args) < 1 {
		fmt.Println("remove expects a breakpoint to remove: remove <address>")
		return
	}

	var addr uint64
	addr, err := strconv.ParseUint(args[0], 0, 64)
	if err != nil {
		for _, t := range sym {
			if t.Label == args[0] {
				addr = t.Address
				goto FIND
			}
		}

		fmt.Printf("Cannot parse %v as address.\n", args[0])
		return
	}
FIND:
	var idx int
	for i, b := range *breakpoints {
		if b == addr {
			idx = i
			goto FOUND
		}
	}

	fmt.Printf("No breakpoint at address 0x%x\n", addr)
	return
FOUND:
	n_brks := make([]uint64, len(*breakpoints)-1)

	for i := 0; i < idx; i++ {
		n_brks[i] = (*breakpoints)[i]
	}

	for i := idx; i < len(n_brks); i++ {
		n_brks[i] = (*breakpoints)[i+1]
	}

	*breakpoints = n_brks

	printBreakpoints(*breakpoints)
}

func getDumpExpr(m machine.Machine, expr string, prog []uint8) ([]uint8, error) {
	if strings.IndexRune(expr, '#') >= 0 {
		if expr == "#" {
			return prog, nil
		}

		offset, length, _ := strings.Cut(expr, "#")
		offset = strings.TrimSpace(offset)
		length = strings.TrimSpace(length)

		of, err := strconv.ParseUint(offset, 0, 64)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%v is not a number.", offset))
		}

		l, err := strconv.ParseUint(length, 0, 64)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%v is not a number.", length))
		}

		end := of + l
		if end > uint64(len(prog)) {
			end = uint64(len(prog))
		}

		return prog[of:end], nil
	} else {
		addr, length, has_at := strings.Cut(expr, "@")
		if !has_at {
			return nil, errors.New(fmt.Sprintf("Cannot parse %v as a dump argument", expr))
		}

		addr = strings.TrimSpace(addr)
		length = strings.TrimSpace(length)

		ad, err := strconv.ParseUint(addr, 0, 64)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%v is not a number.", addr))
		}
		l, err := strconv.ParseUint(length, 0, 64)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%v is not a number.", length))
		}

		mem, err := m.GetMemoryChunk(ad, l)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Error getting memory chunk: %v", err))
		}

		return mem, nil
	}
}

func debuggerDump(m machine.Machine, args []string, prog []uint8) {
	if len(args) < 2 {
		fmt.Println("dump expects two arguments: (<expr>@<length> or [<addr>]#[<length>]) <file>")
		return
	}

	expr := strings.TrimSpace(args[0])
	file := strings.TrimSpace(args[1])

	dump, err := getDumpExpr(m, expr, prog)
	if err != nil {
		fmt.Printf("Cannot get content to dump: %v\n", err)
		return
	}

	f, err := os.OpenFile(file, os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Cannot open %s for write: %v\n", file, err)
		return
	}

	_, err = f.Write(dump)
	if err != nil {
		fmt.Printf("Error while writing to %s: %v\n", file, err)
	}

	f.Close()
}

func debuggerRewind(m machine.Machine, prog []uint8) {
	err := m.LoadProgram(prog)
	if err != nil {
		fmt.Printf("Error while reloading machine: %v\n", err)
	} else {
		fmt.Printf("Reloaded machine.\n")
	}
}

// Returns length (second value) as 0 if it's a register.
func getSetExpr(m machine.Machine, expr string) (uint64, uint64, error) {
	addr, length, has_at := strings.Cut(expr, "@")
	addr = strings.TrimSpace(addr)
	length = strings.TrimSpace(length)

	if !has_at {
		reg, err := m.GetRegisterNumber(addr)
		if err != nil {
			return 0, 0, errors.New(fmt.Sprintf("Cannot get register number: %v", err))
		}
		return reg, 0, nil
	}

	var a uint64
	reg, err := m.GetRegisterNumber(addr)
	if err != nil {
		a, err = strconv.ParseUint(addr, 0, 64)
		if err != nil {
			return 0, 0, errors.New(fmt.Sprintf("%v is not a register or address", addr))
		}
	} else {
		a, _ = m.GetRegister(reg)
	}

	l, err := strconv.ParseUint(length, 0, 64)
	if err != nil {
		return 0, 0, errors.New(fmt.Sprintf("%v is not a number.", length))
	}

	return a, l, nil
}

func debuggerSet(m machine.Machine, args []string) {
	if len(args) < 2 {
		fmt.Printf("set expects two arguments: <expr>[@<length>] <value>")
		return
	}

	addr, length, err := getSetExpr(m, args[0])
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	value, err := strconv.ParseUint(args[1], 0, 64)
	if err != nil {
		fmt.Printf("Cannot parse %v as number: %v", args[1], err)
		return
	}

	if length == 0 {
		err := m.SetRegister(addr, value)
		if err != nil {
			fmt.Printf("Error while changing register content: %v\n", err)
		}
	} else {
		arr := make([]uint8, length)
		// I hope this is somehow optimized out.
		shift := 0
		for i, _ := range arr {
			arr[i] = uint8(value >> shift)
			shift += 8
		}

		err := m.SetMemoryChunk(addr, arr)
		if err != nil {
			fmt.Printf("Error while changing memory content: %v\n", err)
		}
	}
}

func debugMachine(m machine.Machine, sym []assembler.DebuggerToken, prog []uint8) {
	version()
	fmt.Println("Type 'help' for a list of commands.")
	fmt.Println("Debugging", m.ArchitetureName())

	var breakpoints []uint64
	in := bufio.NewReader(os.Stdin)

	fmt.Print("egg> ")
	line, err := in.ReadString('\n')
	for err == nil {
		wsldirty := strings.Split(strings.TrimRight(line, "\n"), " ")
		if len(wsldirty) == 0 {
			continue
		}
		wsl := []string{}

		for i := range wsldirty {
			if len(wsldirty[i]) != 0 {
				wsl = append(wsl, wsldirty[i])
			}
		}

		if len(wsl) > 0 {
			switch wsl[0] {
			case "help", "h":
				debuggerHelp()
			case "print", "p":
				debuggerPrint(m, sym, wsl[1:])
			case "next", "n":
				debuggerNext(m, sym, in)
			case "continue", "c":
				debuggerContinue(m, sym, breakpoints, in)
			case "break", "b":
				debuggerBreakpoint(m, sym, &breakpoints, wsl[1:])
			case "remove", "r":
				debuggerRemove(m, sym, &breakpoints, wsl[1:])
			case "dump", "d":
				debuggerDump(m, wsl[1:], prog)
			case "rewind", "rew":
				debuggerRewind(m, prog)
			case "set", "s":
				debuggerSet(m, wsl[1:])
			case "exit", "e", "quit", "q":
				goto EXIT
			case "ping":
				fmt.Println("pong!")
			default:
				fmt.Printf("no such command: %v\n", wsl[0])
			}
		} else {
			fmt.Println("")
		}

		fmt.Print("egg> ")
		line, err = in.ReadString('\n')
	}

	fmt.Println("")
EXIT:
	fmt.Println("bye!")
}
