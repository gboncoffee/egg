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
	// todo: mem
}

// Parses R-type instructions.
// Returns in order:
// - rd
// - rs1
// - rs2
// - func3
// - func7
func parseR(i uint32) (uint8, uint8, uint8, uint8, uint8) {
	return 0, 0, 0, 0, 0
}

// Parses I-type instructions.
// Returns in order:
// - rd
// - rs1
// - imm
// - func3
func parseI(i uint32) (uint8, uint8, uint32, uint8) {
	return 0, 0, 0, 0
}

// Parses S-type instructions.
// Returns in order:
// - rs1
// - rs2
// - imm
// - func3
func parseS(i uint32) (uint8, uint8, uint32, uint8) {
	return 0, 0, 0, 0
}

// Parses B-type instructions.
// Returns in order:
// - rs1
// - rs2
// - imm
// - func3
func parseB(i uint32) (uint8, uint8, uint32, uint8) {
	return 0, 0, 0, 0
}

// Parses U-type instructions.
// Returns in order:
// - rd
// - imm
func parseU(i uint32) (uint8, uint32) {
	return 0, 0
}

// Parses J-type instructions.
// Returns in order:
// - rd
// - imm
func parseJ(i uint32) (uint8, uint32) {
	return 0, 0
}

func (m *RiscV) execArithmetic(rd uint8, rs1 uint8, rs2 uint8, func3 uint8, func7 uint8) {
	panic("not implemented")
}

func (m *RiscV) execImmArithmetic(rd uint8, rs1 uint8, imm uint32, func3 uint8) {
	panic("not implemented")
}

func (m *RiscV) execLoad(rd uint8, rs1 uint8, imm uint32, func3 uint8) {
	panic("not implemented")
}

func (m *RiscV) execStore(rs1 uint8, rs2 uint8, imm uint32, func3 uint8) {
	panic("not implemented")
}

func (m *RiscV) execBranch(rs1 uint8, rs2 uint8, imm uint32, func3 uint8) {
	panic("not implemented")
}

func (m *RiscV) execJal(rd uint8, imm uint32) {
	panic("not implemented")
}

func (m *RiscV) execJalr(rd uint8, rs1 uint8, imm uint32) {
	panic("not implemented")
}

func (m *RiscV) execLui(rd uint8, imm uint32) {
	panic("not implemented")
}

func (m *RiscV) execAuipc(rd uint8, imm uint32) {
	panic("not implemented")
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

func (m *RiscV) LoadProgram(uint64) error {
	return errors.New("not implemented")
}

func (m *RiscV) NextInstruction() error {
	return errors.New("not implemented")
}

func (m *RiscV) GetMemory(uint64) (uint64, error) {
	return 0, errors.New("not implemented")
}

func (m *RiscV) SetMemory(uint64, uint64) error {
	return errors.New("not implemented")
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

	if content > math.MaxUint32 {
		return errors.New(fmt.Sprintf("Number beyond 32 limit: %d. RISC-V has only 32 bit registers.", content))
	}

	if reg != 0 {
		m.registers[reg] = uint32(content) // will not overflow, we already checked
	}

	return nil
}
