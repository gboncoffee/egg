// Package egg/sagui implements a Sagui machine for EGG.
// There's an extension in the original Sagui from Dr. Marco Zanata: an movr
// from 0 to 0 is a BREAK.
package sagui

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/gboncoffee/egg/assembler"
	"github.com/gboncoffee/egg/machine"
)

type Sagui struct {
	registers [4]uint8
	pc        uint8
	mem       [math.MaxUint8 + 1]uint8
}

func signExtend(n uint8) uint8 {
	sign := n >> 3
	sign = (^(sign - 1)) << 4
	return n | sign
}

func signExtend8(n uint8) uint8 {
	sign := n >> 3
	sign8 := uint8(^(sign - 1)) << 4
	return n | sign8
}

func (m *Sagui) SetRegister(reg uint64, value uint64) error {
	if reg > 3 {
		return fmt.Errorf(machine.InterCtx.Get("no such register: %v"), reg)
	}
	if value > math.MaxUint8 {
		return fmt.Errorf(machine.InterCtx.Get("value %v is bigger than maximum 8 bit address %v"), value)
	}
	m.registers[reg] = uint8(value)
	return nil
}

func (m *Sagui) GetRegister(reg uint64) (uint64, error) {
	if reg > 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("no such register: %v"), reg)
	}

	return uint64(m.registers[reg]), nil
}

func (m *Sagui) GetMemory(addr uint64) (uint8, error) {
	if addr > math.MaxUint8 {
		return 0, fmt.Errorf(machine.InterCtx.Get("value %v is bigger than maximum 8 bit address %v"), addr, math.MaxUint8)
	}
	return m.mem[addr], nil
}

func (m *Sagui) SetMemory(addr uint64, value uint8) error {
	if addr > math.MaxUint8 {
		return fmt.Errorf(machine.InterCtx.Get("value %v is bigger than maximum 8 bit address %v"), addr, math.MaxUint8)
	}
	m.mem[addr] = value
	return nil
}

func (m *Sagui) GetMemoryChunk(addr uint64, size uint64) ([]uint8, error) {
	end := addr + (size - 1)
	if end > math.MaxUint8 {
		return nil, fmt.Errorf(machine.InterCtx.Get("end address %v bigger than maximum 8 bit address %v"), end, math.MaxUint8)
	}
	return m.mem[addr:(end + 1)], nil
}

func (m *Sagui) SetMemoryChunk(addr uint64, content []uint8) error {
	end := addr + (uint64(len(content)) - 1)
	if end > math.MaxUint8 {
		return fmt.Errorf(machine.InterCtx.Get("end address %v bigger than maximum 8 bit address %v"), end, math.MaxUint8)
	}

	for _, b := range content {
		m.mem[addr] = b
		addr++
	}
	return nil
}

func (m *Sagui) LoadProgram(program []uint8) error {
	m.pc = 0
	return m.SetMemoryChunk(0, program)
}

func (m *Sagui) GetCurrentInstructionAddress() uint64 {
	return uint64(m.pc)
}

func (m *Sagui) NextInstruction() (*machine.Call, error) {
	instr, err := m.GetMemory(m.GetCurrentInstructionAddress())
	if err != nil {
		return nil, fmt.Errorf(machine.InterCtx.Get("failed to fetch instruction from memory: %v"), err)
	}

	if instr == 0x60 {
		m.pc++
		return &machine.Call{
			Number: machine.SYS_BREAK,
			Arg1:   0,
			Arg2:   0,
		}, nil
	}

	op := instr >> 4
	imm := instr & 0xf
	rb := instr & 0x3
	ra := (instr >> 2) & 0x3

	rav, _ := m.GetRegister(uint64(ra))
	rbv, _ := m.GetRegister(uint64(rb))
	r0v, _ := m.GetRegister(0)

	switch op {
	case 0x0:
		if rav == 0 {
			m.pc = uint8(rbv) - 1
		}
	case 0x1:
		if r0v == 0 {
			m.pc = m.pc + signExtend(uint8(imm)) - 1
		}
	case 0x2:
		m.pc = uint8(rbv) - 1
	case 0x3:
		m.pc = m.pc + signExtend(uint8(imm)) - 1
	case 0x4:
		mem, _ := m.GetMemory(rbv)
		m.SetRegister(uint64(ra), uint64(mem))
	case 0x5:
		m.SetMemory(rbv, uint8(rav))
	case 0x6:
		m.SetRegister(uint64(ra), rbv)
	case 0x7:
		m.SetRegister(0, uint64(imm<<4)|(r0v&0xf))
	case 0x8:
		m.SetRegister(0, uint64(imm)|r0v&0xf0)
	case 0x9:
		m.SetRegister(uint64(ra), rav+rbv)
	case 0xA:
		m.SetRegister(uint64(ra), rav-rbv)
	case 0xB:
		m.SetRegister(uint64(ra), rav&rbv)
	case 0xC:
		m.SetRegister(uint64(ra), rav|rbv)
	case 0xD:
		if rbv == 0 {
			m.SetRegister(uint64(ra), 1)
		} else {
			m.SetRegister(uint64(ra), 0)
		}
	case 0xE:
		m.SetRegister(uint64(ra), rav<<rbv)
	case 0xF:
		m.SetRegister(uint64(ra), rav>>rbv)
	}

	m.pc++

	return nil, nil
}

func (m *Sagui) ArchitectureInfo() machine.ArchitectureInfo {
	return machine.ArchitectureInfo{
		Name: "Sagui",
		RegistersNames: []string{
			"r0",
			"r1",
			"r2",
			"r3",
		},
		WordWidth: 8,
	}
}

func getRegisterNumber(r string) (uint64, error) {
	switch r {
	case "r0", "0":
		return 0, nil
	case "r1", "1":
		return 1, nil
	case "r2", "2":
		return 2, nil
	case "r3", "3":
		return 3, nil
	default:
		return 0, fmt.Errorf(machine.InterCtx.Get("no such register: %v"), r)
	}
}

func (m *Sagui) GetRegisterNumber(r string) (uint64, error) {
	return getRegisterNumber(r)
}

func translateArgs(arg string) (uint64, error) {
	if len(arg) < 1 {
		return 0, errors.New(machine.InterCtx.Get("empty argument"))
	}
	reg, err := getRegisterNumber(arg)
	if err != nil {
		n, err := strconv.ParseInt(arg, 0, 64)
		if n > 0xf {
			return 0, fmt.Errorf(machine.InterCtx.Get("immediate bigger than immediate size: %v"), arg)
		}
		return uint64(n), err
	}
	return reg, nil
}

func assembleR(t assembler.ResolvedToken) (uint8, error) {
	if len(t.Args) != 2 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
	}

	ra := uint8(t.Args[0]) & 0x3
	rb := uint8(t.Args[1]) & 0x3
	var op uint8
	switch string(t.Value) {
	case "brzr":
		op = 0x0
	case "ld":
		op = 0x4
	case "st":
		op = 0x5
	case "movr":
		op = 0x6
	case "add":
		op = 0x9
	case "sub":
		op = 0xA
	case "and":
		op = 0xB
	case "or":
		op = 0xC
	case "not":
		op = 0xD
	case "slr":
		op = 0xE
	case "srr":
		op = 0xF
	default:
		return 0, errors.New("unreachable")
	}

	return (op << 4) | (ra << 2) | rb, nil
}

func assembleI(t assembler.ResolvedToken) (uint8, error) {
	if len(t.Args) != 1 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), t.Value)
	}

	imm := uint8(t.Args[0]) & 0xf
	var op uint8
	switch string(t.Value) {
	case "brzi":
		op = 0x1
	case "ji":
		op = 0x3
	case "movh":
		op = 0x7
	case "movl":
		op = 0x8
	default:
		return 0, errors.New("unreachable")
	}

	return (op << 4) | imm, nil
}

func assembleJumpImm(t assembler.ResolvedToken) (uint8, error) {
	if len(t.Args) != 1 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), t.Value)
	}

	// Ugly code in the name of reuse.
	t.Args[0] = uint64(int8(signExtend8(uint8(t.Args[0]))) - int8(t.Address))
	return assembleI(t)
}

func assembleJr(t assembler.ResolvedToken) (uint8, error) {
	if len(t.Args) != 1 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), "jr")
	}

	return 0x20 | uint8(t.Args[0]&0x3), nil
}

func assembleInstruction(code []uint8, addr int, t assembler.ResolvedToken) error {
	bin := uint8(0)
	var err error

	switch string(t.Value) {
	case "brzr", "ld", "st", "movr", "add", "sub", "and", "or", "not", "slr", "srr":
		bin, err = assembleR(t)
	case "brzi", "ji":
		bin, err = assembleJumpImm(t)
	case "movh", "movl":
		bin, err = assembleI(t)
	case "jr":
		bin, err = assembleJr(t)
	case "ebreak":
		bin = 0x60
	default:
		return fmt.Errorf(machine.InterCtx.Get("unknown instruction: %v"), string(t.Value))
	}

	if err != nil {
		return err
	}

	code[addr] = bin
	return nil
}

func assemble(t []assembler.ResolvedToken) ([]uint8, error) {
	size := uint64(0)
	for _, i := range t {
		if i.Type == assembler.TOKEN_INSTRUCTION {
			size += 4
		} else {
			size += uint64(len(i.Value))
		}
	}

	var err error

	code := make([]uint8, size)
	addr := 0
	for _, i := range t {
		if i.Type == assembler.TOKEN_INSTRUCTION {
			err = assembleInstruction(code, addr, i)
			if err != nil {
				return nil, fmt.Errorf(machine.InterCtx.Get("%v:%v: Error assembling: %v"), *i.File, i.Line, err)
			}
			addr++
		} else {
			for _, c := range []uint8(i.Value) {
				code[addr] = c
				addr++
			}
		}
	}

	return code, nil
}

func (m *Sagui) Assemble(file string) ([]uint8, []assembler.DebuggerToken, error) {
	tokens := []assembler.Token{}
	err := assembler.Tokenize(file, &tokens)
	if err != nil {
		return nil, nil, err
	}

	resolvedTokens, debuggerTokens, err := assembler.ResolveTokens(tokens, func(i *assembler.Instruction) error {
		i.Size = 1
		return nil
	}, translateArgs)

	if err != nil {
		return nil, nil, err
	}

	code, err := assemble(resolvedTokens)
	if err != nil {
		return nil, nil, err
	}

	return code, debuggerTokens, nil
}
