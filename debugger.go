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

// TODO: information on every sysbreak/breakpoint operation.

type Breakpoint struct {
	// nil if no file.
	File *string
	// Empty if no label
	Label string
	// This is only valid if File is non-nil, of course.
	Line int
	// This one always exists.
	Address uint64
}

func breakpoint2String(breakpoint Breakpoint) string {
	var s strings.Builder
	if breakpoint.File != nil {
		s.WriteString(*breakpoint.File)
		s.WriteString(fmt.Sprintf(":%v ", breakpoint.Line))
	}
	if breakpoint.Label != "" {
		s.WriteString(fmt.Sprintf(machine.InterCtx.Get("(Label %v) "), breakpoint.Label))
	}
	s.WriteString(fmt.Sprintf("0x%x", breakpoint.Address))

	return s.String()
}

func debuggerHelp() {
	help := machine.InterCtx.Get(MASSIVE_HELP_STRING)
	fmt.Print(help)
}

func tokensToString(sym []assembler.DebuggerToken, info *machine.ArchitectureInfo) []string {
	str := make([]string, len(sym))

	for i, tok := range sym {
		var build strings.Builder

		build.WriteString(fmt.Sprintf("%v:%v: ", *tok.File, tok.Line))
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
		for _, arg := range tok.Args {
			build.WriteString(arg)
			build.WriteRune(' ')
		}
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
			return nil, fmt.Errorf(machine.InterCtx.Get("cannot parse %v as unsigned number: %v"), sn, err)
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
			return nil, fmt.Errorf(machine.InterCtx.Get("%v is not an unsigned number"), sn)
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
		return nil, fmt.Errorf(machine.InterCtx.Get("%v is not an unsigned number"), length)
	}

	mem, err := m.GetMemoryChunk(a, l)
	if err != nil {
		return nil, err
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
			fmt.Printf("0x%02x (%d)\n", c, int8(c))
		case 16:
			fmt.Printf("0x%04x (%d)\n", c, int16(c))
		case 32:
			fmt.Printf("0x%08x (%d)\n", c, int32(c))
		default:
			fmt.Printf("0x%016x (%d)\n", c, int64(c))
		}
		return
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

func parseBreakpoint(arg string, sym []assembler.DebuggerToken) (Breakpoint, error) {
	// Try to parse it as address.
	value, err := strconv.ParseUint(arg, 0, 64)
	if err == nil {
		// In this case, the address has a symbol.
		for _, symb := range sym {
			if symb.Address == value {
				return Breakpoint{
					Address: value,
					File:    symb.File,
					Line:    symb.Line,
					Label:   symb.Label,
				}, nil
			}
		}
		// Here it does not.
		return Breakpoint{Address: value}, nil
	} else {
		// Ok, let's se if it's file:line instead.
		file, lineAsString, isFileLine := strings.Cut(arg, ":")
		line, err := strconv.ParseUint(lineAsString, 0, 64)
		if isFileLine && err == nil {
			// Search by line first as integer comparison is cheaper than
			// string.
			for _, symb := range sym {
				if uint64(symb.Line) == line && *symb.File == file {
					return Breakpoint{
						Address: symb.Address,
						File:    symb.File,
						Line:    symb.Line,
						Label:   symb.Label,
					}, nil
				}
			}
		} else {
			// Ok, so it must be a label.
			if len(arg) > 0 {
				for _, symb := range sym {
					if symb.Label == arg {
						return Breakpoint{
							Address: symb.Address,
							File:    symb.File,
							Line:    symb.Line,
							Label:   symb.Label,
						}, nil
					}
				}
			}
		}
	}

	return Breakpoint{}, fmt.Errorf(machine.InterCtx.Get("cannot parse %v as breakpoint"), arg)
}

func debuggerContinue(m machine.Machine, sym []assembler.DebuggerToken, breakpoints []Breakpoint, in *bufio.Reader, info *machine.ArchitectureInfo, regs []uint64) []uint64 {
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

		brk := false
		var breakpoint *Breakpoint
		for i := 0; i < len(breakpoints); i++ {
			if breakpoints[i].Address == pc {
				brk = true
				breakpoint = &breakpoints[i]
			}
		}

		if brk {
			fmt.Printf(machine.InterCtx.Get("Stopped at breakpoint: %v\n"), breakpoint2String(*breakpoint))
			debuggerPrint(m, sym, []string{"#3"}, info)
			return printRegisters(m, info, regs)
		}
	}
}

func printBreakpoints(breakpoints []Breakpoint) {
	fmt.Println(machine.InterCtx.Get("Breakpoints:"))
	for _, b := range breakpoints {
		fmt.Println(breakpoint2String(b))
	}
}

func debuggerBreakpoint(sym []assembler.DebuggerToken, breakpoints *[]Breakpoint, args []string) {
	if len(args) < 1 {
		printBreakpoints(*breakpoints)
		return
	}

	new, err := parseBreakpoint(args[0], sym)
	if err != nil {
		fmt.Println(err)
		return
	}

	nBreakpoints := make([]Breakpoint, len(*breakpoints)+1)
	var breakpointIdx int
	for i, p := range *breakpoints {
		if p.Address < new.Address {
			nBreakpoints[i] = p
			breakpointIdx++
		} else if p.Address == new.Address {
			fmt.Println(machine.InterCtx.Get("Breakpoint already exists."))
			return
		} else {
			break
		}
	}

	nBreakpoints[breakpointIdx] = new
	if breakpointIdx < len(*breakpoints) {
		for i, p := range (*breakpoints)[breakpointIdx:] {
			nBreakpoints[i+breakpointIdx+1] = p
		}
	}

	*breakpoints = nBreakpoints

	fmt.Printf(machine.InterCtx.Get("New breakpoint %v\n"), breakpoint2String(new))
	printBreakpoints(*breakpoints)
}

func debuggerRemove(sym []assembler.DebuggerToken, breakpoints *[]Breakpoint, args []string) {
	if len(args) < 1 {
		fmt.Println(machine.InterCtx.Get("remove expects a breakpoint to remove: remove <address/label/file:line>"))
		return
	}

	breakpoint, err := parseBreakpoint(args[0], sym)
	if err != nil {
		fmt.Println(err)
		return
	}

	var idx int
	for i, b := range *breakpoints {
		if b.Address == breakpoint.Address {
			idx = i
			goto FOUND
		}
	}

	fmt.Printf(machine.InterCtx.Get("No breakpoint %v\n"), breakpoint2String(breakpoint))
	return

FOUND:
	nBreakpoints := make([]Breakpoint, len(*breakpoints)-1)

	for i := 0; i < idx; i++ {
		nBreakpoints[i] = (*breakpoints)[i]
	}

	for i := idx; i < len(nBreakpoints); i++ {
		nBreakpoints[i] = (*breakpoints)[i+1]
	}

	*breakpoints = nBreakpoints

	printBreakpoints(*breakpoints)
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
			return nil, fmt.Errorf(machine.InterCtx.Get("%v is not an unsigned number"), offset)
		}

		l, err := strconv.ParseUint(length, 0, 64)
		if err != nil {
			return nil, fmt.Errorf(machine.InterCtx.Get("%v is not an unsigned number"), length)
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
			return nil, fmt.Errorf(machine.InterCtx.Get("%v is not an unsigned number"), addr)
		}
		l, err := strconv.ParseUint(length, 0, 64)
		if err != nil {
			return nil, fmt.Errorf(machine.InterCtx.Get("%v is not an unsigned number"), length)
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

func debuggerReload(m machine.Machine, sym *[]assembler.DebuggerToken, breakpoints *[]Breakpoint, prog *[]uint8, fileName string) {
	newCode, newSym, err := m.Assemble(fileName)
	if err != nil {
		fmt.Println(machine.InterCtx.Get("Error assembling file:"))
		fmt.Println(err)
		fmt.Println(machine.InterCtx.Get("Keeping old program and program state."))
		return
	}

	err = m.LoadProgram(newCode)
	if err != nil {
		fmt.Println(machine.InterCtx.Get("Error loading new assembled code:"))
		fmt.Println(err)
		fmt.Println(machine.InterCtx.Get("Keeping old program and program state."))
	}

	*prog = newCode
	*sym = newSym
	*breakpoints = []Breakpoint{}

	fmt.Println(machine.InterCtx.Get("Rebuild Assembly."))
	fmt.Println(machine.InterCtx.Get("Reloaded machine."))
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
		return 0, 0, fmt.Errorf(machine.InterCtx.Get("%v is not an unsigned number"), length)
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

	value, err := strconv.ParseInt(args[1], 0, 64)
	if err != nil {
		fmt.Printf(machine.InterCtx.Get("Cannot parse %v as number: %v\n"), args[1], err)
		return
	}

	if length == 0 {
		err := m.SetRegister(addr, uint64(value))
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

func debugMachine(m machine.Machine, sym []assembler.DebuggerToken, prog []uint8, fileName string) {
	version()
	fmt.Println(machine.InterCtx.Get("Type 'help' for a list of commands."))

	info := m.ArchitectureInfo()
	fmt.Println(machine.InterCtx.Get("Debugging"), info.Name)

	var breakpoints []Breakpoint
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
				debuggerBreakpoint(sym, &breakpoints, wsl[1:])
			case "remove", "r":
				debuggerRemove(sym, &breakpoints, wsl[1:])
			case "dump", "d":
				debuggerDump(m, wsl[1:], prog)
			case "rewind", "rew":
				debuggerRewind(m, prog)
			case "reload", "rel":
				debuggerReload(m, &sym, &breakpoints, &prog, fileName)
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
