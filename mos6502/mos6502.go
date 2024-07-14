// Package mos6502 implements a MOS 6502 machine for the EGG emulator.
//
// The numbered functions for getting and setting registers really do not make
// sense in 6502 although they're required to work, so we'll just agree that the
// registers have the following numbers:
// - A: 0
// - P: 1
// - S: 2
// - X: 3
// - Y: 4
//
// When implementing, we found out that "LoadProgram" would be an actual
// problem: there's no point in loading the program to the 0 address because
// the architecture makes it easier to use the first page for readily-acessible
// memory. For the very same reason, the second page is the stack. Does it makes
// sense to put it at the third page so? No, the stack would be allowed to grow
// only 256 bytes.
//
// In the real 6502, the pointer at 0xfffc pointed to the text. As the upper
// half of the memory was usually ROM, the content was saved. When a reset
// happens, PC is loaded with the contents at that location.
//
// Thinking about the fact that most systems would have the upper half of memory
// ROM, I decided to load the program to exactly the 128 page.
//
// The program is loaded at mos6502.TEXT_PAGE (0x8000).

package mos6502

import (
	"fmt"
	"math"

	"github.com/gboncoffee/egg/machine"
)

const STACK_PAGE = 0x100
const TEXT_PAGE = 0x8000

type Mos6502 struct {
	registers struct {
		A uint8
		P uint8
		S uint8
		X uint8
		Y uint8
		PC uint16
	}
	mem [math.MaxUint16 + 1]uint8
}

// Returns the operation then the addressing mode.
func parseOpcode(i uint8) (uint8, uint8) {
	return (i & 0b11100000 >> 3) | (i & 0b11), i & 0b1100
}

func (m *Mos6502) LoadProgram(program []uint8) error {
	m.registers.PC = TEXT_PAGE
	return m.SetMemoryChunk(TEXT_PAGE, program)
}

func (m *Mos6502) GetMemory(addr uint64) (uint8, error) {
	if addr > math.MaxUint16 {
		return 0, fmt.Errorf(machine.InterCtx.Get("value %v bigger than maximum 16 bit address %v"), addr, math.MaxUint16)
	}

	return m.mem[addr], nil
}

func (m *Mos6502) SetMemory(addr uint64, content uint8) error {
	if addr > math.MaxUint16 {
		return fmt.Errorf(machine.InterCtx.Get("value %v bigger than maximum 16 bit address %v"), addr, math.MaxUint16)
	}

	m.mem[addr] = content
	return nil
}

func (m *Mos6502) GetMemoryChunk(addr uint64, size uint64) ([]uint8, error) {
	end := addr + (size - 1)
	if end > math.MaxUint16 {
		return nil, fmt.Errorf(machine.InterCtx.Get("end address %v bigger than maximum 16 bits address %v"), end, math.MaxUint16)
	}

	return m.mem[addr:(end + 1)], nil
}

func (m *Mos6502) SetMemoryChunk(addr uint64, content []uint8) error {
	end := addr + uint64(len(content))
	if end > math.MaxUint16 {
		return fmt.Errorf(machine.InterCtx.Get("end address %v greater than maximum 16 bits address %v"), end, math.MaxUint16)
	}

	for _, b := range content {
		m.mem[addr] = b
		addr++
	}

	return nil
}

func (m *Mos6502) GetRegister(r uint64) (uint64, error) {
	switch r {
	case 0:
		return uint64(m.registers.A), nil
	case 1:
		return uint64(m.registers.P), nil
	case 2:
		return uint64(m.registers.S), nil
	case 3:
		return uint64(m.registers.X), nil
	case 4:
		return uint64(m.registers.Y), nil
	}
	return 0, fmt.Errorf(machine.InterCtx.Get("no such register %v. 6502 has only 5 registers"), r)
}

func (m *Mos6502) SetRegister(r uint64, v uint64) error {
	switch r {
	case 0:
		m.registers.A = uint8(v)
		return nil
	case 1:
		m.registers.P = uint8(v)
		return nil
	case 2:
		m.registers.S = uint8(v)
		return nil
	case 3:
		m.registers.X = uint8(v)
		return nil
	case 4:
		m.registers.Y = uint8(v)
		return nil
	}

	return fmt.Errorf(machine.InterCtx.Get("no such register: %v. 6502 has only 5 registers"), r)
}

func (m *Mos6502) GetRegisterNumber(r string) (uint64, error) {
	switch r {
	case "A":
		return 0, nil
	case "P":
		return 1, nil
	case "S":
		return 2, nil
	case "X":
		return 3, nil
	case "Y":
		return 4, nil
	}
	return 0, fmt.Errorf(machine.InterCtx.Get("no such register: %s"), r)
}

func (m *Mos6502) GetCurrentInstructionAddress() uint64 {
	return uint64(m.registers.PC)
}

func (m *Mos6502) ArchitectureInfo() machine.ArchitectureInfo {
	return machine.ArchitectureInfo{
		Name: "MOS 6502",
		RegistersNames: []string{
			"A",
			"P",
			"PC",
			"S",
			"X",
			"Y",
		},
		// The 6502 word is actually 8 bits, we say it's 16 because the address
		// size is 16.
		WordWidth: 16,
	}
}
