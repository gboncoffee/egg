package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/gboncoffee/egg/assembler"
	"github.com/gboncoffee/egg/machine"
)

func debuggerHelp() {
	help := machine.InterCtx.Get(`Commands:

help
	Shows this help.
	Shortcut: h
print <expr>[@<length>]
	Prints the content of registers and memory.
	Shortcut: p
printall
	Prints the content of all registers.
	Shortcut: pall
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
`)
	fmt.Print(help)
}

func tokensToString(sym []assembler.DebuggerToken, info *machine.ArchitectureInfo) []string {
	str := make([]string, len(sym))

	for i, tok := range sym {
		var build strings.Builder

		switch info.WordWidth {
		case 8:
			build.WriteString(fmt.Sprintf("0x%02x: ", tok.Address))
		case 16:
			build.WriteString(fmt.Sprintf("0x%04x: ", tok.Address))
		case 32:
			build.WriteString(fmt.Sprintf("0x%08x: ", tok.Address))
		default:
			build.WriteString(fmt.Sprintf("0x%016x: ", tok.Address))
		}
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

func getPrintHashExpr(m machine.Machine, sym []assembler.DebuggerToken, expr string, info *machine.ArchitectureInfo) ([]string, error) {
	fn, sn, _ := strings.Cut(expr, "#")
	fn = strings.TrimSpace(fn)
	sn = strings.TrimSpace(sn)

	var faddr uint64
	var l uint64
	if fn == "" {
		var err error

		if sn == "" {
			return tokensToString(sym, info), nil
		}

		l, err = strconv.ParseUint(sn, 0, 64)
		if err != nil {
			return nil, fmt.Errorf(machine.InterCtx.Get("cannot parse %v as number: %v"), sn, err)
		}

		faddr = m.GetCurrentInstructionAddress()
	} else {
		var err error

		if sn == "" {
			return nil, errors.New(machine.InterCtx.Get("length not supplied"))
		}

		faddr, err = strconv.ParseUint(fn, 0, 64)
		if err != nil {
			return nil, fmt.Errorf(machine.InterCtx.Get("cannot parse %v as address"), fn)
		}

		l, err = strconv.ParseUint(sn, 0, 64)
		if err != nil {
			return nil, fmt.Errorf(machine.InterCtx.Get("%v is not a number"), sn)
		}
	}

	i := uint64(sort.Search(len(sym), func(i int) bool {
		return sym[i].Address >= faddr
	}))

	if i == uint64(len(sym)) {
		return nil, fmt.Errorf(machine.InterCtx.Get("no instruction at address 0x%x"), faddr)
	}

	l = l + i
	if l >= uint64(len(sym)) {
		l = uint64(len(sym))
	}

	return tokensToString(sym[i:l], info), nil
}

func getMemoryContentPrint(m machine.Machine, addr string, length string) ([]uint64, error) {
	var a uint64
	reg, err := m.GetRegisterNumber(addr)
	if err != nil {
		a, err = strconv.ParseUint(addr, 0, 64)
		if err != nil {
			return nil, fmt.Errorf(machine.InterCtx.Get("%v is not a register or address"), addr)
		}
	} else {
		a, _ = m.GetRegister(reg)
	}

	l, err := strconv.ParseUint(length, 0, 64)
	if err != nil {
		return nil, fmt.Errorf(machine.InterCtx.Get("%v is not a number"), length)
	}

	mem, err := m.GetMemoryChunk(a, l)
	if err != nil {
		return nil, fmt.Errorf(machine.InterCtx.Get("cannot get memory content: %v"), err)
	}

	// I hope this is somehow "optimized out" to a simple padded copy.
	m64 := make([]uint64, l)
	for i, v := range mem {
		m64[i] = uint64(v)
	}

	return m64, nil
}

func byte2ascii(b byte) rune {
	r := rune(b)
	if unicode.IsPrint(r) {
		return r
	}
	return '.'
}

func printMemory(s []uint64) {
	nLines := len(s) / 8
	for i := 0; i < nLines; i++ {
		for j := 0; j < 8; j++ {
			fmt.Printf("0x%02x ", s[i*8+j])
		}
		fmt.Print("  ")
		for j := 0; j < 8; j++ {
			fmt.Printf("%c", byte2ascii(byte(s[i*8+j])))
		}
		fmt.Println()
	}

	if len(s)%8 != 0 {
		i := 0
		for i < len(s)%8 {
			fmt.Printf("0x%02x ", s[nLines*8+i])
			i++
		}
		for i < 8 {
			fmt.Print("     ")
			i++
		}
		fmt.Print("  ")
		for i := 0; i < len(s)%8; i++ {
			fmt.Printf("%c", byte2ascii(byte(s[nLines*8+i])))
		}
		fmt.Println()
	}
}

func printExpr(m machine.Machine, expr string, info *machine.ArchitectureInfo) {
	addr, length, has_at := strings.Cut(expr, "@")
	addr = strings.TrimSpace(addr)
	length = strings.TrimSpace(length)

	if !has_at {
		reg, err := m.GetRegisterNumber(addr)
		if err != nil {
			fmt.Printf("%v\n", fmt.Errorf(machine.InterCtx.Get("cannot get register content: %v"), err))
			return
		}
		c, _ := m.GetRegister(reg)

		// Not very pretty but (mostly) does the job.
		switch info.WordWidth {
		case 8:
			fmt.Printf("0x%02x", c)
		case 16:
			fmt.Printf("0x%04x", c)
		case 32:
			fmt.Printf("0x%08x", c)
		default:
			fmt.Printf("0x%016x", c)
		}
	}

	arr, err := getMemoryContentPrint(m, addr, length)
	if err != nil {
		fmt.Printf("%v\n", fmt.Errorf(machine.InterCtx.Get("cannot get memory content: %v"), err))
		return
	}

	printMemory(arr)
}

func debuggerPrint(m machine.Machine, sym []assembler.DebuggerToken, args []string, info *machine.ArchitectureInfo) {
	if len(args) < 1 {
		fmt.Println(machine.InterCtx.Get("print expects one argument: <expr>[@<length>] or [<addr>]#[<length>]"))
		return
	}

	expr := args[0]
	var res []string
	var err error
	if strings.ContainsRune(expr, '#') {
		res, err = getPrintHashExpr(m, sym, expr, info)
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			for _, v := range res {
				fmt.Println(v)
			}
		}
	} else {
		printExpr(m, expr, info)
	}
}

func debuggerPrintAll(m machine.Machine, info *machine.ArchitectureInfo) {
	for i, r := range info.RegistersNames {
		v, _ := m.GetRegister(uint64(i))
		switch info.WordWidth {
		case 8:
			fmt.Printf("%8s: 0x%02x\n", r, v)
		case 16:
			fmt.Printf("%8s: 0x%04x\n", r, v)
		case 32:
			fmt.Printf("%8s: 0x%08x\n", r, v)
		default:
			fmt.Printf("%8s: 0x%016x\n", r, v)
		}
	}
}

func ioCall(m machine.Machine, call *machine.Call, in *bufio.Reader) {
	if call.Number == machine.SYS_READ {
		addr := call.Arg1
		size := call.Arg2
		fmt.Printf(machine.InterCtx.Get("READ call for address 0x%x with %d bytes:\n"), addr, size)
		buf := make([]byte, size)
		_, err := in.Read(buf)
		if err != nil {
			fmt.Printf(machine.InterCtx.Get("Error reading stdin: %v\n"), err)
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

func printRegisters(m machine.Machine, info *machine.ArchitectureInfo, regs []uint64) []uint64 {
	for i, r := range info.RegistersNames {
		v, _ := m.GetRegister(uint64(i))
		if regs[i] != v {
			switch info.WordWidth {
			case 8:
				fmt.Printf(machine.InterCtx.Get("Register %v: changed from 0x%02x to 0x%02x\n"), r, regs[i], v)
			case 16:
				fmt.Printf(machine.InterCtx.Get("Register %v: changed from 0x%04x to 0x%04x\n"), r, regs[i], v)
			case 32:
				fmt.Printf(machine.InterCtx.Get("Register %v: changed from 0x%08x to 0x%08x\n"), r, regs[i], v)
			default:
				fmt.Printf(machine.InterCtx.Get("Register %v: changed from 0x%016x to 0x%016x\n"), r, regs[i], v)
			}
			regs[i] = v
		}
	}
	return regs
}

func debuggerNext(m machine.Machine, sym []assembler.DebuggerToken, in *bufio.Reader, info *machine.ArchitectureInfo, regs []uint64) []uint64 {
	call, err := m.NextInstruction()
	if err != nil {
		fmt.Printf(machine.InterCtx.Get("Instruction execution failed: %v\n"), err)
		return regs
	}

	pc := m.GetCurrentInstructionAddress()
	if call != nil {
		if call.Number == machine.SYS_BREAK {
			fmt.Printf(machine.InterCtx.Get("BREAK call while stepping at address 0x%x\n"), pc)
		} else {
			ioCall(m, call, in)
		}
	}

	debuggerPrint(m, sym, []string{"#3"}, info)

	return printRegisters(m, info, regs)
}

func debuggerContinue(m machine.Machine, sym []assembler.DebuggerToken, breakpoints []uint64, in *bufio.Reader, info *machine.ArchitectureInfo, regs []uint64) []uint64 {
	for {
		call, err := m.NextInstruction()
		if err != nil {
			fmt.Printf(machine.InterCtx.Get("Instruction execution failed: %v\n"), err)
			return regs
		}

		pc := m.GetCurrentInstructionAddress()
		if call != nil {
			if call.Number == machine.SYS_BREAK {
				fmt.Printf(machine.InterCtx.Get("Stopped at BREAK call at address 0x%x\n"), pc)
				debuggerPrint(m, sym, []string{"#3"}, info)
				return printRegisters(m, info, regs)
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
			fmt.Printf(machine.InterCtx.Get("Stopped at breakpoint at address 0x%x\n"), pc)
			debuggerPrint(m, sym, []string{"#3"}, info)
			return printRegisters(m, info, regs)
		}
	}
}

func printBreakpoints(breakpoints []uint64, info *machine.ArchitectureInfo) {
	fmt.Println(machine.InterCtx.Get("Breakpoints:"))
	for _, b := range breakpoints {
		switch info.WordWidth {
		case 8:
			fmt.Printf("0x%02x\n", b)
		case 16:
			fmt.Printf("0x%04x\n", b)
		case 32:
			fmt.Printf("0x%08x\n", b)
		default:
			fmt.Printf("0x%016x\n", b)
		}
	}
}

func debuggerBreakpoint(sym []assembler.DebuggerToken, breakpoints *[]uint64, args []string, info *machine.ArchitectureInfo) {
	if len(args) < 1 {
		printBreakpoints(*breakpoints, info)
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
			fmt.Printf(machine.InterCtx.Get("%v is not a number.\n"), args[0])
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
			fmt.Println(machine.InterCtx.Get("Breakpoint already exists."))
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

	fmt.Printf(machine.InterCtx.Get("New breakpoint at address 0x%x\n"), addr)
	printBreakpoints(*breakpoints, info)
}

func debuggerRemove(sym []assembler.DebuggerToken, breakpoints *[]uint64, args []string, info *machine.ArchitectureInfo) {
	if len(args) < 1 {
		fmt.Println(machine.InterCtx.Get("remove expects a breakpoint to remove: remove <address>"))
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

		fmt.Printf(machine.InterCtx.Get("Cannot parse %v as address.\n"), args[0])
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

	fmt.Printf(machine.InterCtx.Get("No breakpoint at address 0x%x\n"), addr)
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

	printBreakpoints(*breakpoints, info)
}

func getDumpExpr(m machine.Machine, expr string, prog []uint8) ([]uint8, error) {
	if strings.ContainsRune(expr, '#') {
		if expr == "#" {
			return prog, nil
		}

		offset, length, _ := strings.Cut(expr, "#")
		offset = strings.TrimSpace(offset)
		length = strings.TrimSpace(length)

		of, err := strconv.ParseUint(offset, 0, 64)
		if err != nil {
			return nil, fmt.Errorf(machine.InterCtx.Get("%v is not a number"), offset)
		}

		l, err := strconv.ParseUint(length, 0, 64)
		if err != nil {
			return nil, fmt.Errorf(machine.InterCtx.Get("%v is not a number"), length)
		}

		end := of + l
		if end > uint64(len(prog)) {
			end = uint64(len(prog))
		}

		return prog[of:end], nil
	} else {
		addr, length, has_at := strings.Cut(expr, "@")
		if !has_at {
			return nil, fmt.Errorf(machine.InterCtx.Get("cannot parse %v as a dump argument"), expr)
		}

		addr = strings.TrimSpace(addr)
		length = strings.TrimSpace(length)

		ad, err := strconv.ParseUint(addr, 0, 64)
		if err != nil {
			return nil, fmt.Errorf(machine.InterCtx.Get("%v is not a number"), addr)
		}
		l, err := strconv.ParseUint(length, 0, 64)
		if err != nil {
			return nil, fmt.Errorf(machine.InterCtx.Get("%v is not a number"), length)
		}

		mem, err := m.GetMemoryChunk(ad, l)
		if err != nil {
			return nil, fmt.Errorf(machine.InterCtx.Get("error getting memory chunk: %v"), err)
		}

		return mem, nil
	}
}

func debuggerDump(m machine.Machine, args []string, prog []uint8) {
	if len(args) < 2 {
		fmt.Println(machine.InterCtx.Get("dump expects two arguments: (<expr>@<length> or [<addr>]#[<length>]) <file>"))
		return
	}

	expr := strings.TrimSpace(args[0])
	file := strings.TrimSpace(args[1])

	dump, err := getDumpExpr(m, expr, prog)
	if err != nil {
		fmt.Printf(machine.InterCtx.Get("Cannot get content to dump: %v\n"), err)
		return
	}

	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf(machine.InterCtx.Get("Cannot open %s for write: %v\n"), file, err)
		return
	}

	_, err = f.Write(dump)
	if err != nil {
		fmt.Printf(machine.InterCtx.Get("Error while writing to %s: %v\n"), file, err)
	}

	f.Close()
}

func debuggerRewind(m machine.Machine, prog []uint8) {
	err := m.LoadProgram(prog)
	if err != nil {
		fmt.Printf(machine.InterCtx.Get("Error while reloading machine: %v\n"), err)
	} else {
		fmt.Println(machine.InterCtx.Get("Reloaded machine."))
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
			return 0, 0, fmt.Errorf(machine.InterCtx.Get("cannot get register number: %v"), err)
		}
		return reg, 0, nil
	}

	var a uint64
	reg, err := m.GetRegisterNumber(addr)
	if err != nil {
		a, err = strconv.ParseUint(addr, 0, 64)
		if err != nil {
			return 0, 0, fmt.Errorf(machine.InterCtx.Get("%v is not a register or address"), addr)
		}
	} else {
		a, _ = m.GetRegister(reg)
	}

	l, err := strconv.ParseUint(length, 0, 64)
	if err != nil {
		return 0, 0, fmt.Errorf(machine.InterCtx.Get("%v is not a number"), length)
	}

	return a, l, nil
}

func debuggerSet(m machine.Machine, args []string) {
	if len(args) < 2 {
		fmt.Println(machine.InterCtx.Get("set expects two arguments: <expr>[@<length>] <value>"))
		return
	}

	addr, length, err := getSetExpr(m, args[0])
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	value, err := strconv.ParseUint(args[1], 0, 64)
	if err != nil {
		fmt.Printf(machine.InterCtx.Get("Cannot parse %v as number: %v"), args[1], err)
		return
	}

	if length == 0 {
		err := m.SetRegister(addr, value)
		if err != nil {
			fmt.Printf(machine.InterCtx.Get("Error while changing register content: %v\n"), err)
		}
	} else {
		arr := make([]uint8, length)
		// I hope this is somehow optimized out.
		shift := 0
		for i := range arr {
			arr[i] = uint8(value >> shift)
			shift += 8
		}

		err := m.SetMemoryChunk(addr, arr)
		if err != nil {
			fmt.Printf(machine.InterCtx.Get("Error while changing memory content: %v\n"), err)
		}
	}
}

func debugMachine(m machine.Machine, sym []assembler.DebuggerToken, prog []uint8) {
	version()
	fmt.Println(machine.InterCtx.Get("Type 'help' for a list of commands."))

	info := m.ArchitectureInfo()
	fmt.Println(machine.InterCtx.Get("Debugging"), info.Name)

	var breakpoints []uint64
	in := bufio.NewReader(os.Stdin)

	regs := make([]uint64, len(info.RegistersNames))
	for i := range info.RegistersNames {
		regs[i], _ = m.GetRegister(uint64(i))
	}

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
				debuggerPrint(m, sym, wsl[1:], &info)
			case "printall", "pall":
				debuggerPrintAll(m, &info)
			case "next", "n":
				regs = debuggerNext(m, sym, in, &info, regs)
			case "continue", "c":
				regs = debuggerContinue(m, sym, breakpoints, in, &info, regs)
			case "break", "b":
				debuggerBreakpoint(sym, &breakpoints, wsl[1:], &info)
			case "remove", "r":
				debuggerRemove(sym, &breakpoints, wsl[1:], &info)
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
				fmt.Printf(machine.InterCtx.Get("No such command: %v\n"), wsl[0])
			}
		} else {
			fmt.Println("")
		}

		fmt.Print("egg> ")
		line, err = in.ReadString('\n')
	}

	fmt.Println("")
EXIT:
	fmt.Println(machine.InterCtx.Get("bye!"))
}
