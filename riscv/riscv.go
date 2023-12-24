// Package egg/riscv implements a RISC-V 32 IM machine for the EGG emulator.
package riscv

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/gboncoffee/egg/assembler"
	"github.com/gboncoffee/egg/machine"
)

// The RiscV struct implements the machine interface needed by the EGG emulator.
type RiscV struct {
	registers [32]uint32
	pc        uint32
	mem       [math.MaxUint32 + 1]uint8
}

// Sign extends the number n which has s bits. I hope gc inlines this function
// always. The drawback of not having macros is that you cannot ensure something
// is always inlined.
func signExtend(n uint32, s uint8) uint32 {
	sign := n >> (s - 1)
	sign = (^(sign - 1)) << s
	return n | sign
}

func signExtend64(n uint32) uint64 {
	sign := n >> 31
	sign64 := uint64(^(sign - 1)) << 32
	return uint64(n) | sign64
}

// Parses R-type instructions.
// Returns in order:
// - rd
// - rs1
// - rs2
// - func3
// - func7
func parseR(i uint32) (uint8, uint8, uint8, uint8, uint8) {
	rd := uint8((i & 0b00000000000000000000111110000000) >> 7)
	rs1 := uint8((i & 0b00000000000011111000000000000000) >> 15)
	rs2 := uint8((i & 0b00000001111100000000000000000000) >> 20)
	func3 := uint8((i & 0b00000000000000000111000000000000) >> 12)
	func7 := uint8((i & 0b11111110000000000000000000000000) >> 25)

	return rd, rs1, rs2, func3, func7
}

// Parses I-type instructions.
// Returns in order:
// - rd
// - rs1
// - imm
// - func3
func parseI(i uint32) (uint8, uint8, uint32, uint8) {
	rd := uint8((i & 0b00000000000000000000111110000000) >> 7)
	rs1 := uint8((i & 0b00000000000011111000000000000000) >> 15)
	imm := uint32((i & 0b11111111111100000000000000000000) >> 20)
	func3 := uint8((i & 0b00000000000000000111000000000000) >> 12)

	return rd, rs1, signExtend(imm, 12), func3
}

// Parses S-type instructions.
// Returns in order:
// - rs1
// - rs2
// - imm
// - func3
func parseS(i uint32) (uint8, uint8, uint32, uint8) {
	rs1 := uint8((i & 0b00000000000011111000000000000000) >> 15)
	rs2 := uint8((i & 0b00000001111100000000000000000000) >> 20)
	func3 := uint8((i & 0b00000000000000000111000000000000) >> 12)

	imm := (i & 0b111110000000) >> 7
	imm = imm | ((i & 0b11111110000000000000000000000000) >> 19)

	return rs1, rs2, signExtend(imm, 12), func3
}

// Parses B-type instructions.
// Returns in order:
// - rs1
// - rs2
// - imm
// - func3
func parseB(i uint32) (uint8, uint8, uint32, uint8) {
	rs1 := uint8((i & 0b00000000000011111000000000000000) >> 15)
	rs2 := uint8((i & 0b00000001111100000000000000000000) >> 20)
	func3 := uint8((i & 0b00000000000000000111000000000000) >> 12)

	imm := (i & 0b111100000000) >> 7
	imm = imm | ((i & 0b10000000) << 4)
	imm = imm | ((i & 0b01111110000000000000000000000000) >> 19)
	imm = imm | ((i & 0b10000000000000000000000000000000) >> 20)

	return rs1, rs2, signExtend(imm, 13), func3
}

// Parses U-type instructions.
// Returns in order:
// - rd
// - imm
func parseU(i uint32) (uint8, uint32) {
	rd := uint8((i & 0b00000000000000000000111110000000) >> 7)
	imm := uint32(i & 0b11111111111111111111000000000000)

	return rd, imm
}

// Parses J-type instructions.
// Returns in order:
// - rd
// - imm
func parseJ(i uint32) (uint8, uint32) {
	rd := uint8((i & 0b00000000000000000000111110000000) >> 7)

	imm := (i & 0b00000000000011111111000000000000)
	imm = imm | ((i & 0b00000000000100000000000000000000) >> 9)
	imm = imm | ((i & 0b01111111111000000000000000000000) >> 20)
	imm = imm | ((i & 0b10000000000000000000000000000000) >> 20)

	return rd, imm
}

func (m *RiscV) execArithmetic(rd uint8, rs1 uint8, rs2 uint8, func3 uint8, func7 uint8) {
	rs1v64, _ := m.GetRegister(uint64(rs1))
	rs1v := int32(rs1v64)
	rs2v64, _ := m.GetRegister(uint64(rs2))
	rs2v := int32(rs2v64)
	var r int32

	if func7 == 0x1 {
		switch func3 {
		case 0x0:
			r = int32(uint64(rs1v) * uint64(rs2v))
		case 0x1:
			r = int32((int64(signExtend64(uint32(rs1v))) * int64(signExtend64(uint32(rs2v)))) >> 32)
		case 0x2:
			r = int32((int64(signExtend64(uint32(rs1v))) * int64(rs2v)) >> 32)
		case 0x3:
			r = int32(uint64(rs1v) * uint64(rs2v) >> 32)
		case 0x4:
			if rs2v == 0 {
				r = 0
			} else {
				r = rs1v / rs2v
			}
		case 0x5:
			if rs2v == 0 {
				r = 0
			} else {
				r = int32(uint32(rs1v) / uint32(rs2v))
			}
		case 0x6:
			if rs2v == 0 {
				r = 0
			} else {
				r = rs1v % rs2v
			}
		case 0x7:
			if rs2v == 0 {
				r = 0
			} else {
				r = int32(uint32(rs1v) % uint32(rs2v))
			}
		}
	} else {
		switch func3 {
		case 0x0:
			if func7 == 0x20 {
				r = rs1v - rs2v
			} else {
				r = rs1v + rs2v
			}
		case 0x1:
			r = rs1v << rs2v
		case 0x2:
			// In C this would look terrible but in Go bools and ints are different.
			if rs1v < rs2v {
				r = 1
			} else {
				r = 0
			}
		case 0x3:
			// Same.
			if uint32(rs1v) < uint32(rs2v) {
				r = 1
			} else {
				r = 0
			}
		case 0x4:
			r = rs1v ^ rs2v
		case 0x5:
			if func7 == 0x20 {
				r = rs1v >> rs2v
			} else {
				r = int32(uint32(rs1v) >> uint32(rs2v))
			}
		case 0x6:
			r = rs1v | rs2v
		case 0x7:
			r = rs1v & rs2v
		}
	}

	m.SetRegister(uint64(rd), uint64(r))

	m.pc += 4
}

func (m *RiscV) execImmArithmetic(rd uint8, rs1 uint8, imm uint32, func3 uint8) {
	rs1v64, _ := m.GetRegister(uint64(rs1))
	rs1v := int32(rs1v64)
	immv := int32(imm)
	var r int32

	switch func3 {
	case 0x0:
		r = rs1v + immv
	case 0x1:
		r = rs1v << (immv & 0b1111)
	case 0x2:
		if rs1v < int32(imm) {
			r = 1
		} else {
			r = 0
		}
	case 0x3:
		if uint32(rs1v) < uint32(imm) {
			r = 1
		} else {
			r = 0
		}
	case 0x4:
		r = rs1v ^ immv
	case 0x5:
		if (immv & 0b111111100000) == (0x20 << 5) {
			r = rs1v >> immv
		} else {
			r = int32(uint32(rs1v) >> uint32(immv))
		}
	case 0x6:
		r = rs1v | immv
	case 0x7:
		r = rs1v & immv
	}

	m.SetRegister(uint64(rd), uint64(r))

	m.pc += 4
}

func (m *RiscV) execLoad(rd uint8, rs1 uint8, imm uint32, func3 uint8) {
	rs1v, _ := m.GetRegister(uint64(rs1))
	addr32 := uint32(rs1v) + imm
	addr := uint64(addr32)
	var v uint32

	switch func3 {
	case 0x0:
		mem, _ := m.GetMemory(addr)
		rdv, _ := m.GetRegister(uint64(rd))
		v = (uint32(rdv) & 0xffffff00) | uint32(mem)
	case 0x1:
		mem, _ := m.GetMemoryChunk(addr, 2)
		rdv, _ := m.GetRegister(uint64(rd))
		if len(mem) != 2 {
			return
		}
		v = (uint32(rdv) & 0xffff0000) | uint32(mem[0]) | (uint32(mem[1]) << 8)
	case 0x2:
		mem, _ := m.GetMemoryChunk(addr, 4)
		if len(mem) != 4 {
			return
		}
		v = uint32(mem[0]) |
			(uint32(mem[1]) << 8) |
			(uint32(mem[2]) << 16) |
			(uint32(mem[3]) << 24)
	case 0x4:
		mem, _ := m.GetMemory(addr)
		v = uint32(mem)
	case 0x5:
		mem, _ := m.GetMemoryChunk(addr, 2)
		v = uint32(mem[0]) | (uint32(mem[1]) << 8)
	}

	m.SetRegister(uint64(rd), uint64(v))

	m.pc += 4
}

func (m *RiscV) execStore(rs1 uint8, rs2 uint8, imm uint32, func3 uint8) {
	rs1v, _ := m.GetRegister(uint64(rs1))
	addr32 := uint32(rs1v) + imm
	addr := uint64(addr32)
	rs2v, _ := m.GetRegister(uint64(rs2))

	switch func3 {
	case 0x0:
		m.SetMemory(addr, uint8(rs2v))
	case 0x1:
		v := []uint8{uint8(rs2v), uint8(rs2v >> 8)}
		m.SetMemoryChunk(addr, v)
	case 0x2:
		v := []uint8{
			uint8(rs2v),
			uint8(rs2v >> 8),
			uint8(rs2v >> 16),
			uint8(rs2v >> 24),
		}
		m.SetMemoryChunk(addr, v)
	}

	m.pc += 4
}

func (m *RiscV) execBranch(rs1 uint8, rs2 uint8, imm uint32, func3 uint8) {
	rs1v64, _ := m.GetRegister(uint64(rs1))
	rs1v := int32(rs1v64)
	rs2v64, _ := m.GetRegister(uint64(rs2))
	rs2v := int32(rs2v64)

	switch func3 {
	case 0x0:
		if rs1v == rs2v {
			m.pc = (m.pc + imm) - 4
		}
	case 0x1:
		if rs1v != rs2v {
			m.pc = (m.pc + imm) - 4
		}
	case 0x4:
		if rs1v < rs2v {
			m.pc = (m.pc + imm) - 4
		}
	case 0x5:
		if rs1v >= rs2v {
			m.pc = (m.pc + imm) - 4
		}
	case 0x6:
		if uint32(rs1v) < uint32(rs2v) {
			m.pc = (m.pc + imm) - 4
		}
	case 0x7:
		if uint32(rs1v) >= uint32(rs2v) {
			m.pc = (m.pc + imm) - 4
		}
	}

	m.pc += 4
}

func (m *RiscV) execJal(rd uint8, imm uint32) {
	m.SetRegister(uint64(rd), uint64(m.pc+4))
	m.pc += imm
}

func (m *RiscV) execJalr(rd uint8, rs1 uint8, imm uint32) {
	m.SetRegister(uint64(rd), uint64(m.pc+4))
	rs1v, _ := m.GetRegister(uint64(rs1))
	m.pc = uint32(rs1v) + imm
}

func (m *RiscV) execLui(rd uint8, imm uint32) {
	m.SetRegister(uint64(rd), uint64(imm)<<12)
}

func (m *RiscV) execAuipc(rd uint8, imm uint32) {
	m.SetRegister(uint64(rd), uint64(m.pc)+(uint64(imm)<<12))
}

func (m *RiscV) execute(i uint32) (*machine.Call, error) {
	opcode := i & 0b01111111
	switch opcode {
	case 0b0110011:
		rd, rs1, rs2, func3, func7 := parseR(i)
		m.execArithmetic(rd, rs1, rs2, func3, func7)
	case 0b0010011:
		rd, rs1, imm, func3 := parseI(i)
		m.execImmArithmetic(rd, rs1, imm, func3)
	case 0b0000011:
		rd, rs1, imm, func3 := parseI(i)
		m.execLoad(rd, rs1, imm, func3)
	case 0b0100011:
		rs1, rs2, imm, func3 := parseS(i)
		m.execStore(rs1, rs2, imm, func3)
	case 0b1100011:
		rs1, rs2, imm, func3 := parseB(i)
		m.execBranch(rs1, rs2, imm, func3)
	case 0b1101111:
		rd, imm := parseJ(i)
		m.execJal(rd, imm)
	case 0b1100111:
		rd, rs1, imm, _ := parseI(i)
		m.execJalr(rd, rs1, imm)
	case 0b0110111:
		rd, imm := parseU(i)
		m.execLui(rd, imm)
	case 0b0010111:
		rd, imm := parseU(i)
		m.execAuipc(rd, imm)
	case 0b1110011:
		_, _, imm, _ := parseI(i)
		num, _ := m.GetRegister(17)
		a1, _ := m.GetRegister(10)
		a2, _ := m.GetRegister(11)
		if imm == 1 {
			num = machine.SYS_BREAK
		}
		call := machine.Call{num, a1, a2}

		return &call, nil
	default:
		return nil, errors.New(fmt.Sprintf("Unknown opcode: %b", opcode))
	}

	return nil, nil
}

func (m *RiscV) LoadProgram(program []uint8) error {
	m.pc = 0
	return m.SetMemoryChunk(0, program)
}

func (m *RiscV) NextInstruction() (*machine.Call, error) {
	iarr, err := m.GetMemoryChunk(uint64(m.pc), 4)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not load 4 bytes from address at PC: %x", m.pc))
	}

	i := uint32(iarr[0]) | uint32(iarr[1]<<8) | uint32(iarr[2]<<16) | uint32(iarr[3]<<24)

	return m.execute(i)
}

func (m *RiscV) GetMemory(addr uint64) (uint8, error) {
	if addr > math.MaxUint32 {
		return 0, errors.New(fmt.Sprintf("Value %v bigger than maximum 32 bit address %v", addr, math.MaxUint32))
	}

	return m.mem[addr], nil
}

func (m *RiscV) SetMemory(addr uint64, content uint8) error {
	if addr > math.MaxUint32 {
		return errors.New(fmt.Sprintf("Value %v bigger than maximum 32 bit address %v", addr, math.MaxUint32))
	}

	m.mem[addr] = content

	return nil
}

func (m *RiscV) GetMemoryChunk(addr uint64, size uint64) ([]uint8, error) {
	end := addr + (size - 1)
	if end > math.MaxUint32 {
		return nil, errors.New(fmt.Sprintf("End address %v bigger than maximum 32 bit address %v", end, math.MaxUint32))
	}

	return m.mem[addr:(end + 1)], nil
}

func (m *RiscV) SetMemoryChunk(addr uint64, content []uint8) error {
	end := addr + (uint64(len(content)) - 1)
	if end > math.MaxUint32 {
		return errors.New(fmt.Sprintf("End address %v bigger than maximum 32 bit address %v", end, math.MaxUint32))
	}

	for _, b := range content {
		m.mem[addr] = b
		addr++
	}

	return nil
}

func (m *RiscV) GetRegister(reg uint64) (uint64, error) {
	if reg >= 32 {
		return 0, errors.New(fmt.Sprintf("No such register: %d. RISC-V has only 32 registers.", reg))
	}

	return uint64(m.registers[reg]), nil
}

func (m *RiscV) SetRegister(reg uint64, content uint64) error {
	if reg >= 32 {
		return errors.New(fmt.Sprintf("No such register: %d. RISC-V has only 32 registers.", reg))
	}

	if reg != 0 {
		m.registers[reg] = uint32(content) // Overflow is a feature.
	}

	return nil
}

func assembleArithmetic(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, errors.New(fmt.Sprintf("Wrong number of arguments for instruction '%s', expected 3 arguments", t.Value))
	}

	code := uint32(0b0110011)
	code = code | uint32(t.Args[0] << 7)
	code = code | uint32(t.Args[1] << 15)
	code = code | uint32(t.Args[2] << 20)

	func3 := uint32(0) // add and sub
	var func7 uint32
	switch t.Value {
	case "sub":
		func7 = 0x20
	// Avoid running the multiplication part.
	case "add":
		break
	case "sll":
		func3 = 1
	case "slt":
		func3 = 2
	case "sltu":
		func3 = 3
	case "xor":
		func3 = 4
	case "sra":
		func7 = 0x20
		fallthrough
	case "srl":
		func3 = 5
	case "or":
		func3 = 6
	case "and":
		func3 = 7
	// Multiplication extension.
	default:
		func7 = 1
		// 0 is mul.
		switch t.Value {
		case "mulh":
			func3 = 1
		case "mulsu":
			func3 = 2
		case "mulu":
			func3 = 3
		case "div":
			func3 = 4
		case "divu":
			func3 = 5
		case "rem":
			func3 = 6
		case "remu":
			func3 = 7
		}
	}

	code = code | (func3 << 12)
	code = code | (func7 << 25)

	return code, nil
}

func assembleArithmeticImm(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, errors.New(fmt.Sprintf("Wrong number of arguments for instruction '%s', expected 3 arguments", t.Value))
	}

	code := uint32(0b0010011)
	code = code | uint32(t.Args[0] << 7)
	code = code | uint32(t.Args[1] << 15)
	code = code | uint32(t.Args[2] << 20)

	func3 := uint32(0) // addi
	switch t.Value {
	case "xori":
		func3 = 4
	case "ori":
		func3 = 6
	case "andi":
		func3 = 7
	case "slli":
		func3 = 1
	case "srai":
		code = code | uint32(0x20 << 25 )
		fallthrough
	case "srli":
		func3 = 5
	case "slti":
		func3 = 2
	case "sltiu":
		func3 = 3
	}

	code = code | (func3 << 12)

	return code, nil
}

func assembleLoad(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, errors.New(fmt.Sprintf("Wrong number of arguments for instruction '%s', expected 3 arguments", t.Value))
	}

	code := uint32(0b0000011)
	code = code | uint32(t.Args[0] << 7)
	code = code | uint32(t.Args[1] << 15)
	code = code | uint32(t.Args[2] << 20)

	func3 := uint32(0) // lb
	switch t.Value {
	case "lh":
		func3 = 1
	case "lw":
		func3 = 2
	case "lbu":
		func3 = 4
	case "lhu":
		func3 = 5
	}

	code = code | (func3 << 12)

	return code, nil
}

func assembleStore(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, errors.New(fmt.Sprintf("Wrong number of arguments for instruction '%s', expected 3 arguments", t.Value))
	}

	code := uint32(0b0100011)
	code = code | uint32(t.Args[0] << 15)
	code = code | uint32(t.Args[1] << 20)
	code = code | uint32((t.Args[2] & 0b11111) << 7)
	code = code | uint32((t.Args[2] & 0b111111100000 << 20))

	func3 := uint32(0) // sb
	switch t.Value {
	case "sh":
		func3 = 1
	case "sw":
		func3 = 2
	}

	code = code | (func3 << 12)

	return code, nil
}

func assembleBranch(t assembler.ResolvedToken, addr int) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, errors.New(fmt.Sprintf("Wrong number of arguments for instruction '%s', expected 3 arguments", t.Value))
	}

	t.Args[2] = uint64(signExtend64(uint32(t.Args[2])) - uint64(addr))

	code := uint32(0b1100011)
	code = code | uint32(t.Args[0] << 15)
	code = code | uint32(t.Args[1] << 20)
	code = code | uint32((t.Args[2] & 0b100000000000) >> 4)
	code = code | uint32((t.Args[2] & 0b11110) << 7)
	code = code | uint32((t.Args[2] & 0b1000000000000) << 19)
	code = code | uint32((t.Args[2] & 0b11111100000) << 20)

	func3 := uint32(0) // beq
	switch t.Value {
	case "bne":
		func3 = 1
	case "blt":
		func3 = 4
	case "bge":
		func3 = 5
	case "bltu":
		func3 = 6
	case "bgeu":
		func3 = 7
	}

	code = code | (func3 << 12)

	return code, nil
}

func assembleJal(t assembler.ResolvedToken, addr int) (uint32, error) {
	if len(t.Args) != 2 {
		return 0, errors.New(fmt.Sprintf("Wrong number of arguments for instruction '%s', expected 2 arguments", t.Value))
	}

	t.Args[1] = uint64(signExtend64(uint32(t.Args[1])) - uint64(addr))

	code := uint32(0b1101111)
	code = code | uint32(t.Args[0] << 7)
	code = code | uint32(t.Args[1] & 0b11111111000000000000)
	code = code | uint32((t.Args[1] & 0b100000000000) << 9)
	code = code | uint32((t.Args[1] & 0b11111111110) << 20)
	code = code | uint32((t.Args[1] & 0b100000000000000000000) << 11)

	return code, nil
}

func assembleJalr(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, errors.New(fmt.Sprintf("Wrong number of arguments for instruction '%s', expected 3 arguments", t.Value))
	}

	code := uint32(0b1100111)
	code = code | uint32(t.Args[0] << 7)
	code = code | uint32(t.Args[1] << 15)
	code = code | uint32(t.Args[2] << 20)

	return code, nil
}

func assembleU(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 2 {
		return 0, errors.New(fmt.Sprintf("Wrong number of arguments for instruction '%s', expected 2 arguments", t.Value))
	}

	var code uint32
	if t.Value == "lui" {
		code = 0b0110111
	} else {
		code = 0b0010111
	}
	code = code | uint32(t.Args[0] << 7)
	code = code | uint32(t.Args[1] << 12)

	return code, nil
}

func assembleCall(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 0 {
		return 0, errors.New(fmt.Sprintf("Wrong number of arguments for instruction '%s', expected no argument", t.Value))
	}

	code := uint32(0b1110011)
	if t.Value == "ebreak" {
		code = code | (1 << 20)
	}

	return code, nil
}

func assembleInstruction(code []uint8, addr int, t assembler.ResolvedToken) error {
	bin := uint32(0)
	var err error

	switch t.Value {
	case "add", "sub", "xor", "or", "and", "sll", "srl", "sra", "slt", "sltu", "mul", "mulh", "mulsu", "mulu", "div", "divu", "rem", "remu":
		bin, err = assembleArithmetic(t)
	case "addi", "xori", "ori", "andi", "slli", "srli", "srai", "slti", "sltiu":
		bin, err = assembleArithmeticImm(t)
	case "lb", "lh", "lw", "lbu", "lhu":
		bin, err = assembleLoad(t)
	case "sb", "sh", "sw":
		bin, err = assembleStore(t)
	case "beq", "bne", "blt", "bge", "bltu", "bgeu":
		bin, err = assembleBranch(t, addr)
	case "jal":
		bin, err = assembleJal(t, addr)
	case "jalr":
		bin, err = assembleJalr(t)
	case "lui", "auipc":
		bin, err = assembleU(t)
	case "ecall", "ebreak":
		bin, err = assembleCall(t)
	default:
		return errors.New(fmt.Sprintf("Unknown instruction: %v", t.Value))
	}

	if err != nil {
		return err
	}

	code[addr] = uint8(bin & 0xff)
	code[addr+1] = uint8((bin & 0xff00) >> 8)
	code[addr+2] = uint8((bin & 0xff0000) >> 16)
	code[addr+3] = uint8((bin & 0xff000000) >> 24)

	return nil
}

func assemble(t []assembler.ResolvedToken) ([]uint8, error) {
	// Pre calculate our size. Why not?
	size := uint64(0)
	for _, i := range t {
		size += i.Size
	}

	var err error

	code := make([]uint8, size)
	addr := 0
	for _, i := range t {
		if i.Type == assembler.TOKEN_INSTRUCTION {
			err = assembleInstruction(code, addr, i)
			if err != nil {
				return nil, err
			}
			addr += 4
		} else {
			for _, c := range []uint8(i.Value) {
				code[addr] = c
				addr++
			}
		}
	}

	return code, nil
}

func parseRegisterArg(arg string) (uint64, error) {
	n, err := strconv.Atoi(arg[1:])
	if err != nil {
		return 0, errors.New(fmt.Sprintf("No such register: %v", arg))
	}

	switch arg[0] {
	case 't':
		if n < 3 {
			return uint64(n + 5), nil
		}
		return uint64(n + 25), nil
	case 's':
		switch n {
		case 0:
			return 8, nil
		case 1:
			return 9, nil
		}
		return uint64(n + 16), nil
	case 'a':
		return uint64(n + 10), nil
	}

	return 0, errors.New(fmt.Sprintf("No such register: %v", arg))
}

func translateArgs(arg string) (uint64, error) {
	if len(arg) < 1 {
		return 0, errors.New("Empty argument")
	}

	if (0x30 <= arg[0] && arg[0] <= 0x39) || arg[0] == '-' {
		n, err := strconv.ParseInt(arg, 0, 64)
		return uint64(n), err
	}

	switch arg {
	case "zero":
		return 0, nil
	case "ra":
		return 1, nil
	case "sp":
		return 2, nil
	case "gp":
		return 3, nil
	case "tp":
		return 4, nil
	case "fp":
		return 8, nil
	}

	return parseRegisterArg(arg)
}

func (m *RiscV) Assemble(asm string) ([]uint8, []assembler.DebuggerToken, error) {
	tokens := assembler.Tokenize(asm)
	rt, err := assembler.ResolveTokensFixedSize(tokens, 4, translateArgs)
	if err != nil {
		return nil, nil, err
	}

	symbs := assembler.CreateDebugTokensFixedSize(tokens, 4)
	code, err := assemble(rt)
	if err != nil {
		return nil, nil, err
	}

	return code, symbs, nil
}
