// Package reduxK implements a reduxK machine for EGG.
//
// "or r0, r0" (ebreak) performs a BREAK and "xor r0, r0" (ecall) performs a CALL.
package reduxK

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/gboncoffee/egg/assembler"
	"github.com/gboncoffee/egg/machine"
)

type ReduxK struct {
	registers    [4]uint8
	auxRegisters [2]uint8
	mem          [math.MaxUint8 + 1]uint8
	pc           uint8
}

func signExtend8(n uint8) uint8 {
	sign := n >> 3
	sign8 := uint8(^(sign - 1)) << 4
	return n | sign8
}

func (m *ReduxK) GetCurrentInstructionAddress() uint64 {
	return uint64(m.pc)
}

func (m *ReduxK) ArchitectureInfo() machine.ArchitectureInfo {
	return machine.ArchitectureInfo{
		Name:           "REDUX-K",
		RegistersNames: []string{"0", "1", "2", "3"},
		WordWidth:      8,
	}
}

//
// "Getters and setters" are literally copy-pasted from Sagui. Maybe we should
// factor them out in a library for creating the machines? That would be rather
// cool. Actually, imagine simply defining some stuff and BOOM you got a new
// architecture?
//

func (m *ReduxK) SetRegister(reg uint64, value uint64) error {
	if reg > 3 {
		return fmt.Errorf(machine.InterCtx.Get("no such register: %v"), reg)
	}
	if value > math.MaxUint8 {
		return fmt.Errorf(machine.InterCtx.Get("value %v is bigger than maximum 8 bit address %v"), value)
	}
	m.registers[reg] = uint8(value)
	return nil
}

func (m *ReduxK) GetRegister(reg uint64) (uint64, error) {
	if reg > 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("no such register: %v"), reg)
	}

	return uint64(m.registers[reg]), nil
}

func (m *ReduxK) SetAuxRegister(reg uint64, value uint64) error {
	if reg > 1 {
		return fmt.Errorf(machine.InterCtx.Get("no such register: %v"), reg)
	}
	if value > math.MaxUint8 {
		return fmt.Errorf(machine.InterCtx.Get("value %v is bigger than maximum 8 bit address %v"), value)
	}
	m.auxRegisters[reg] = uint8(value)
	return nil
}

func (m *ReduxK) GetAuxRegister(reg uint64) (uint64, error) {
	if reg > 1 {
		return 0, fmt.Errorf(machine.InterCtx.Get("no such register: %v"), reg)
	}

	return uint64(m.auxRegisters[reg]), nil
}

func (m *ReduxK) SetMemory(addr uint64, value uint8) error {
	if addr > math.MaxUint8 {
		return fmt.Errorf(machine.InterCtx.Get("value %v is bigger than maximum 8 bit address %v"), addr, math.MaxUint8)
	}
	m.mem[addr] = value
	return nil
}

func (m *ReduxK) GetMemory(addr uint64) (uint8, error) {
	if addr > math.MaxUint8 {
		return 0, fmt.Errorf(machine.InterCtx.Get("value %v is bigger than maximum 8 bit address %v"), addr, math.MaxUint8)
	}
	return m.mem[addr], nil
}

func (m *ReduxK) SetMemoryChunk(addr uint64, content []uint8) error {
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

func (m *ReduxK) GetMemoryChunk(addr uint64, size uint64) ([]uint8, error) {
	end := addr + (size - 1)
	if end > math.MaxUint8 {
		return nil, fmt.Errorf(machine.InterCtx.Get("end address %v bigger than maximum 8 bit address %v"), end, math.MaxUint8)
	}
	return m.mem[addr:(end + 1)], nil
}

func (m *ReduxK) LoadProgram(program []uint8) error {
	m.pc = 0
	return m.SetMemoryChunk(0, program)
}

func (m *ReduxK) NextInstruction() (*machine.Call, error) {
	instr, err := m.GetMemory(m.GetCurrentInstructionAddress())
	if err != nil {
		return nil, fmt.Errorf(machine.InterCtx.Get("failed to fetch instruction from memory: %v"), err)
	}

	op := instr >> 4
	imm := instr & 0xf
	ra := (instr >> 2) & 0x3
	rb := instr & 0x3

	uimm := instr & 0x3
	sizev := instr & 0xf

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
		m.SetRegister(uint64(ra), uint64(mem))
	case 0x3:
		m.SetMemory(rbv, uint8(rav))
	case 0x4:
		m.SetRegister(0, uint64(uint8(r0v)+signExtend8(imm)))
	case 0x5:
		for x := int8(0); x < int8(sizev); x++ {
			r0v, _ := m.GetRegister(0)
			r1v, _ := m.GetRegister(1)
			r2v, _ := m.GetRegister(2)
			r3v, _ := m.GetRegister(3)
			
			m.SetMemory(r0v, uint8(r2v))
			m.SetRegister(uint64(2), uint64(uint8(int8(r2v)+int8(r3v))))
			m.SetRegister(uint64(0), uint64(uint8(int8(r0v)+int8(r1v))))
		}
	case 0x6:
		for x := int8(0); x < int8(sizev); x++ {
			r0v, _ := m.GetRegister(0)
			memr0, _ := m.GetMemory(r0v)
			m.SetAuxRegister(0, uint64(memr0))

			r1v, _ := m.GetRegister(1)
			memr1, _ := m.GetMemory(r1v)
			m.SetAuxRegister(1, uint64(memr1))

			rx, _ := m.GetAuxRegister(0)
			ry, _ := m.GetAuxRegister(1)

			m.SetRegister(uint64(3), uint64(uint8(int8(rx)+int8(ry))))

			r2v, _ := m.GetRegister(2)
			r3v, _ := m.GetRegister(3)

			m.SetMemory(r2v, uint8(r3v))

			for x := 0; x < 4; x++ {
				regv, _ := m.GetRegister(uint64(x))
				m.SetRegister(uint64(x), uint64(uint8(int8(regv) + int8(1))))
			}
		}
	case 0x7:
		if (ra == 0) {
			for x := 0; x < 4; x++ {
				regv, _ := m.GetRegister(uint64(x))
				m.SetRegister(uint64(x), uint64(uint8(int8(regv) + int8(uimm))))
			}
			break	
		}

		m.SetRegister(uint64(ra), uint64(uint8(int8(rav) + int8(uimm))))
	case 0x8:
		if rbv == 0 {
			m.SetRegister(uint64(ra), 1)
		} else {
			m.SetRegister(uint64(ra), 0)
		}
	case 0x9:
		m.SetRegister(uint64(ra), rav&rbv)
	case 0xa:
		if (rb == 0 && ra == 0) {
			m.pc++
			return &machine.Call{
				Number: machine.SYS_BREAK,
				Arg1:   0,
				Arg2:   0,
			}, nil
		}
		
		m.SetRegister(uint64(ra), rav|rbv)
	case 0xb:
		if (rb == 0 && ra == 0) {
				m.pc++
				r1v, _ := m.GetRegister(1)
				r2v, _ := m.GetRegister(2)
				return &machine.Call{
					Number: r0v,
					Arg1:   r1v,
					Arg2:   r2v,
			}, nil
		}

		m.SetRegister(uint64(ra), rav^rbv)
	case 0xc:
		m.SetRegister(uint64(ra), uint64(uint8(int8(rav)+int8(rbv))))
	case 0xd:
		m.SetRegister(uint64(ra), uint64(uint8(int8(rav)-int8(rbv))))
	case 0xe:
		m.SetRegister(uint64(ra), (rav<<rbv)&0xff)
	case 0xf:
		m.SetRegister(uint64(ra), rav>>rbv)
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

func (m *ReduxK) GetRegisterNumber(r string) (uint64, error) {
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
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
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

	t.Args[0] = uint64(int8(signExtend8(uint8(t.Args[0]))) - int8(t.Address))
	return assembleI(t)
}

func assembleInc(t assembler.ResolvedToken) (uint8, error) {
	if len(t.Args) != 2 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), t.Value)
	}

	ra := uint8(t.Args[0]) & 0x3
	uimm := uint8(t.Args[1]) & 0x3

	var op uint8
	switch string(t.Value) {
	case "inc":
		op = 0x7
	default:
		return 0, errors.New("unreachable")
	}

	return (op << 4) | (ra << 2) | uimm, nil
}

func assembleV(t assembler.ResolvedToken) (uint8, error) {
	if len(t.Args) != 1 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), t.Value)
	}

	sizev := uint8(t.Args[0]) & 0xf
	var op uint8
	switch string(t.Value) {
	case "loadv":
		op = 0x5
	case "addv":
		op = 0x6
	default:
		return 0, errors.New("unreachable")
	}

	return (op << 4) | sizev, nil
}

func assembleInstruction(code []uint8, addr int, t assembler.ResolvedToken) error {
	bin := uint8(0)
	var err error

	switch string(t.Value) {
	case "ji":
		bin, err = assembleJi(t)
	case "addi":
		bin, err = assembleI(t)
	case "brzr", "ld", "st", "not", "and", "or", "xor", "add", "sub", "slr", "srr":
		bin, err = assembleR(t)
	case "loadv", "addv":
		bin, err = assembleV(t)
	case "inc":
		bin, err = assembleInc(t)
	case "ebreak":
		bin = 0xa0
	case "ecall":
		bin = 0xb0
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

func (m *ReduxK) Assemble(file string) ([]uint8, []assembler.DebuggerToken, error) {
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
