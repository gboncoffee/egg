// Package reduxc implements the commonality between all REDUX-V dialects.
// To implement a REDUX-V dialect, one needs
package reduxc

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/gboncoffee/egg/assembler"
	"github.com/gboncoffee/egg/machine"
)

// Executes an extension instruction. It's called before trying to execute any
// other instruction. When an extension was executed, it should return true
// indicating ReduxC it shouldn't try to execute the instruction as a normal
// one. Thus, it's possible to encode extensions also as normal instructions
// (for instance, a ji 0 can be an extension, and ReduxC won't try to execute it
// as a jump).
type ExtensionExecutionFunction func(bin uint8, m *ReduxC) (bool, *machine.Call, error)

type ExtensionAssembleFunction func(t assembler.ResolvedToken) (uint8, error)

type ReduxC struct {
	mem               [math.MaxUint8 + 1]uint8
	name              string
	executeExtension  ExtensionExecutionFunction
	assembleExtension ExtensionAssembleFunction
	additionalState   any
	registers         [4]uint8
	pc                uint8
}

func ReduxVExtension(name string, e ExtensionExecutionFunction, a ExtensionAssembleFunction, additionalState any) *ReduxC {
	return &ReduxC{
		executeExtension:  e,
		assembleExtension: a,
		name:              name,
		additionalState:   additionalState,
	}
}

func (m *ReduxC) PC() *uint8 {
	return &m.pc
}

func (m *ReduxC) AdditionalState() any {
	return m.additionalState
}

func signExtend8(n uint8) uint8 {
	sign := n >> 3
	sign8 := uint8(^(sign - 1)) << 4
	return n | sign8
}

func (m *ReduxC) GetCurrentInstructionAddress() uint64 {
	return uint64(m.pc)
}

func (m *ReduxC) ArchitectureInfo() machine.ArchitectureInfo {
	return machine.ArchitectureInfo{
		Name:           m.name,
		RegistersNames: []string{"r0", "r1", "r2", "r3"},
		WordWidth:      8,
	}
}

// Historical comment:

//
// "Getters and setters" are literally copy-pasted from Sagui. Maybe we should
// factor them out in a library for creating the machines? That would be rather
// cool. Actually, imagine simply defining some stuff and BOOM you got a new
// architecture?
//

func (m *ReduxC) SetRegister(reg uint64, value uint64) error {
	if reg > 3 {
		return fmt.Errorf(machine.InterCtx.Get("no such register: %v"), reg)
	}
	if value > math.MaxUint8 {
		return fmt.Errorf(machine.InterCtx.Get("value %v is bigger than maximum 8 bit address %v"), value)
	}
	m.registers[reg] = uint8(value)
	return nil
}

func (m *ReduxC) GetRegister(reg uint64) (uint64, error) {
	if reg > 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("no such register: %v"), reg)
	}

	return uint64(m.registers[reg]), nil
}

func (m *ReduxC) SetMemory(addr uint64, value uint8) error {
	if addr > math.MaxUint8 {
		return fmt.Errorf(machine.InterCtx.Get("value %v is bigger than maximum 8 bit address %v"), addr, math.MaxUint8)
	}
	m.mem[addr] = value
	return nil
}

func (m *ReduxC) GetMemory(addr uint64) (uint8, error) {
	if addr > math.MaxUint8 {
		return 0, fmt.Errorf(machine.InterCtx.Get("value %v is bigger than maximum 8 bit address %v"), addr, math.MaxUint8)
	}
	return m.mem[addr], nil
}

func (m *ReduxC) SetMemoryChunk(addr uint64, content []uint8) error {
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

func (m *ReduxC) GetMemoryChunk(addr uint64, size uint64) ([]uint8, error) {
	end := addr + (size - 1)
	if end > math.MaxUint8 {
		return nil, fmt.Errorf(machine.InterCtx.Get("end address %v bigger than maximum 8 bit address %v"), end, math.MaxUint8)
	}
	return m.mem[addr:(end + 1)], nil
}

func (m *ReduxC) LoadProgram(program []uint8) error {
	m.pc = 0
	return m.SetMemoryChunk(0, program)
}

func (m *ReduxC) NextInstruction() (*machine.Call, error) {
	instr, err := m.GetMemory(m.GetCurrentInstructionAddress())
	if err != nil {
		return nil, fmt.Errorf(machine.InterCtx.Get("failed to fetch instruction from memory: %v"), err)
	}

	wasExt, call, err := m.executeExtension(instr, m)
	if call != nil || err != nil {
		return call, err
	}
	if wasExt {
		return nil, nil
	}

	op := instr >> 4
	imm := instr & 0xf
	ra := (instr >> 2) & 0x3
	rb := instr & 0x3

	rav, _ := m.GetRegister(uint64(ra))
	rbv, _ := m.GetRegister(uint64(rb))
	r0v, _ := m.GetRegister(0)

	switch op {
	case 0x0:
		if rav == 0 {
			// Less one because the pc will be incremented.
			m.pc = uint8(rbv) - 1
		}
	case 0x1:
		m.pc += signExtend8(uint8(imm)) - 1
	case 0x2:
		mem, _ := m.GetMemory(rbv)
		_ = m.SetRegister(uint64(ra), uint64(mem))
	case 0x3:
		_ = m.SetMemory(rbv, uint8(rav))
	case 0x4:
		_ = m.SetRegister(0, uint64(uint8(r0v)+signExtend8(imm)))
		// 0x5, 0x6 and 0x7 are reserved to extensions.
	case 0x8:
		if rbv == 0 {
			_ = m.SetRegister(uint64(ra), 1)
		} else {
			_ = m.SetRegister(uint64(ra), 0)
		}
	case 0x9:
		_ = m.SetRegister(uint64(ra), rav&rbv)
	case 0xa:
		_ = m.SetRegister(uint64(ra), rav|rbv)
	case 0xb:
		_ = m.SetRegister(uint64(ra), rav^rbv)
	case 0xc:
		_ = m.SetRegister(uint64(ra), uint64(uint8(int8(rav)+int8(rbv))))
	case 0xd:
		_ = m.SetRegister(uint64(ra), uint64(uint8(int8(rav)-int8(rbv))))
	case 0xe:
		_ = m.SetRegister(uint64(ra), (rav<<rbv)&0xff)
	case 0xf:
		_ = m.SetRegister(uint64(ra), rav>>rbv)
	}

	m.pc++
	return nil, nil
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

func (m *ReduxC) GetRegisterNumber(r string) (uint64, error) {
	return getRegisterNumber(r)
}

// Also heavily-based on Sagui.
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
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), string(t.Value))
	}

	ra := uint8(t.Args[0]) & 0x3
	rb := uint8(t.Args[1]) & 0x3
	var op uint8
	switch string(t.Value) {
	case "brzr":
		op = 0x0
	case "ld":
		op = 0x2
	case "st":
		op = 0x3
	case "not":
		op = 0x8
	case "and":
		op = 0x9
	case "or":
		op = 0xa
	case "xor":
		op = 0xb
	case "add":
		op = 0xc
	case "sub":
		op = 0xd
	case "slr":
		op = 0xe
	case "srr":
		op = 0xf
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
	case "ji":
		op = 0x1
	case "addi":
		op = 0x4
	default:
		return 0, errors.New("unreachable")
	}

	return (op << 4) | imm, nil
}

func assembleJi(t assembler.ResolvedToken) (uint8, error) {
	if len(t.Args) != 1 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), t.Value)
	}

	// Ugly code in the name of reuse.
	t.Args[0] = uint64(int8(signExtend8(uint8(t.Args[0]))) - int8(t.Address))
	return assembleI(t)
}

func (m *ReduxC) assembleInstruction(code []uint8, addr int, t assembler.ResolvedToken) error {
	bin := uint8(0)
	var err error

	switch string(t.Value) {
	case "ji":
		bin, err = assembleJi(t)
	case "addi":
		bin, err = assembleI(t)
	case "brzr", "ld", "st", "not", "and", "or", "xor", "add", "sub", "slr", "srr":
		bin, err = assembleR(t)
	default:
		code, err := m.assembleExtension(t)
		if err != nil {
			return err
		}
		bin = code
	}

	if err != nil {
		return err
	}

	code[addr] = bin
	return nil
}

func (m *ReduxC) assemble(t []assembler.ResolvedToken) ([]uint8, error) {
	size := uint64(0)
	for _, i := range t {
		if i.Type == assembler.TOKEN_INSTRUCTION {
			size += 1
		} else {
			size += uint64(len(i.Value))
		}
	}

	var err error

	code := make([]uint8, size)
	addr := 0
	for _, i := range t {
		if i.Type == assembler.TOKEN_INSTRUCTION {
			err = m.assembleInstruction(code, addr, i)
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

func (m *ReduxC) Assemble(file string) ([]uint8, []assembler.DebuggerToken, error) {
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

	code, err := m.assemble(resolvedTokens)
	if err != nil {
		return nil, nil, err
	}

	return code, debuggerTokens, nil
}
