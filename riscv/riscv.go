// Package egg/riscv implements a RISC-V 32 IM machine for the EGG emulator.
package riscv

import (
	"errors"
	"fmt"
	"math"
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
	m.SetRegister(uint64(rd), uint64(m.pc + 4))
	m.pc += imm
}

func (m *RiscV) execJalr(rd uint8, rs1 uint8, imm uint32) {
	m.SetRegister(uint64(rd), uint64(m.pc + 4))
	rs1v, _ := m.GetRegister(uint64(rs1))
	m.pc = uint32(rs1v) + imm
}

func (m *RiscV) execLui(rd uint8, imm uint32) {
	m.SetRegister(uint64(rd), uint64(imm) << 12)
}

func (m *RiscV) execAuipc(rd uint8, imm uint32) {
	m.SetRegister(uint64(rd), uint64(m.pc) + (uint64(imm) << 12))
}

func (m *RiscV) execute(i uint32) error {
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
		return errors.New("ecalls not implemented")
	default:
		return errors.New(fmt.Sprintf("Unknown opcode: %b", opcode))
	}

	return nil
}

func (m *RiscV) LoadProgram(program []uint8) error {
	m.pc = 0
	return m.SetMemoryChunk(0, program)
}

func (m *RiscV) NextInstruction() error {
	iarr, err := m.GetMemoryChunk(uint64(m.pc), 4)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not load 4 bytes from address at PC: %x", m.pc))
	}

	i := uint32(iarr[0]) | uint32(iarr[1] << 8) | uint32(iarr[2] << 16) | uint32(iarr[3] << 24)
	m.execute(i)

	return nil
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

	return m.mem[addr:end], nil
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
