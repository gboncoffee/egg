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
//
// System calls are performed via the BRK instruction.

package mos6502

import (
	"fmt"
	"math"

	"github.com/gboncoffee/egg/machine"
)

const STACK_PAGE = 0x100
const TEXT_PAGE = 0x8000

const CARRY_FLAG = 0b1
const ZERO_FLAG = 0b10
const INTERRUPT_DISABLE_FLAG = 0b100
const DECIMAL_FLAG = 0b1000
const BREAK_FLAG = 0b10000
const OVERFLOW_FLAG = 0b100000
const NEGATIVE_FLAG = 0b1000000

type Mos6502 struct {
	registers struct {
		A  uint8
		P  uint8
		S  uint8
		X  uint8
		Y  uint8
		PC uint16
	}
	mem [math.MaxUint16 + 1]uint8
}

type AddressMode int

const (
	Immediate = iota
	ZeroPage
	ZeroPageX
	ZeroPageY
	Absolute
	AbsoluteX
	AbsoluteY
	IndirectIndexedX
	IndirectIndexedY
	IndexedIndirectX
	IndexedIndirectY
	Accumulator
)

func (m *Mos6502) isFlagSet(flag uint8) bool {
	return (m.registers.P & flag) != 0
}

func (m *Mos6502) setFlag(flag uint8, value bool) {
	if value {
		m.registers.P = m.registers.P | flag
	} else {
		m.registers.P = m.registers.P & (^flag)
	}
}

// Returns the operation then the addressing mode.
func parseOpcode(i uint8) (uint8, uint8) {
	return (i & 0b11100000 >> 3) | (i & 0b11), i & 0b11100
}

func signExtend(n uint8) uint16 {
	sign := n >> 7
	sign16 := uint16(^(sign - 1)) << 8
	return uint16(n) | sign16
}

func (m *Mos6502) getPointerByAddressMode(mode AddressMode) *uint8 {
	switch mode {
	case Immediate:
		addr := m.registers.PC
		m.registers.PC++
		return &m.mem[addr]
	case ZeroPage:
		addr, _ := m.GetMemory(uint64(m.registers.PC))
		m.registers.PC++
		return &m.mem[uint16(addr)]
	case ZeroPageX:
		addr, _ := m.GetMemory(uint64(m.registers.PC))
		m.registers.PC++
		return &m.mem[uint16(addr)+uint16(m.registers.X)]
	case ZeroPageY:
		addr, _ := m.GetMemory(uint64(m.registers.PC))
		m.registers.PC++
		return &m.mem[uint16(addr)+uint16(m.registers.Y)]
	case Absolute:
		addrSlice, _ := m.GetMemoryChunk(uint64(m.registers.PC), 2)
		addr := uint16(addrSlice[0]) | (uint16(addrSlice[1]) << 8)
		m.registers.PC += 2
		return &m.mem[addr]
	case AbsoluteX:
		addrSlice, _ := m.GetMemoryChunk(uint64(m.registers.PC), 2)
		addr := uint16(addrSlice[0]) | (uint16(addrSlice[1]) << 8)
		m.registers.PC += 2
		return &m.mem[addr+uint16(m.registers.X)]
	case AbsoluteY:
		addrSlice, _ := m.GetMemoryChunk(uint64(m.registers.PC), 2)
		addr := uint16(addrSlice[0]) | (uint16(addrSlice[1]) << 8)
		m.registers.PC += 2
		return &m.mem[addr+uint16(m.registers.Y)]
	}

	return nil
}

func (m *Mos6502) performSyscall() (*machine.Call, error) {
	syscall := uint64(m.registers.X)
	var arg1 uint64
	var arg2 uint64

	if syscall != machine.SYS_BREAK {
		memSlice, err := m.GetMemoryChunk(uint64(m.registers.S)|STACK_PAGE, 4)
		if err != nil {
			return nil, fmt.Errorf(machine.InterCtx.Get("couldn't get syscall arguments from stack: %v"), err)
		}
		arg2 = uint64(memSlice[0]) | (uint64(memSlice[1]) << 8)
		arg1 = uint64(memSlice[2]) | (uint64(memSlice[3]) << 8)
	}

	return &machine.Call{
		Number: syscall,
		Arg1:   arg1,
		Arg2:   arg2,
	}, nil
}

func getAdcAddressMode(mode uint8) AddressMode {
	switch mode {
	case 0b010:
		return Immediate
	case 0b001:
		return ZeroPage
	case 0b101:
		return ZeroPageX
	case 0b011:
		return Absolute
	case 0b111:
		return AbsoluteX
	case 0b110:
		return AbsoluteY
	case 0b000:
		return IndexedIndirectX
	case 0b100:
		return IndirectIndexedY
	}
	return 0
}

func (m *Mos6502) execAdc(addressMode uint8) error {
	mode := getAdcAddressMode(addressMode)
	operand := *m.getPointerByAddressMode(mode)

	if m.isFlagSet(CARRY_FLAG) {
		operand++
	}

	// Stores old signal for the overflow flag.
	signal := m.registers.A & 0b10000000

	m.setFlag(CARRY_FLAG, (uint16(m.registers.A)+uint16(operand)) > math.MaxUint8)

	m.registers.A += operand
	m.setFlag(OVERFLOW_FLAG, m.registers.A&0b10000000 != signal)
	m.setFlag(NEGATIVE_FLAG, m.registers.A&0b10000000 != 0)
	m.setFlag(ZERO_FLAG, m.registers.A == 0)

	return nil
}

func getAndAddressMode(mode uint8) AddressMode {
	switch mode {
	case 0b010:
		return Immediate
	case 0b001:
		return ZeroPage
	case 0b101:
		return ZeroPageX
	case 0b011:
		return Absolute
	case 0b111:
		return AbsoluteX
	case 0b110:
		return AbsoluteY
	case 0b000:
		return IndexedIndirectX
	case 0b100:
		return IndirectIndexedY
	}
	return 0
}

func (m *Mos6502) execAnd(addressMode uint8) {
	mode := getAndAddressMode(addressMode)
	operand := *m.getPointerByAddressMode(mode)

	m.registers.A = m.registers.A & operand
	m.setFlag(ZERO_FLAG, m.registers.A != 0)
	m.setFlag(NEGATIVE_FLAG, m.registers.A&0b10000000 != 0)
}

func getAslAddressMode(mode uint8) AddressMode {
	switch mode {
	case 0b010:
		return Accumulator
	case 0b001:
		return ZeroPage
	case 0b101:
		return ZeroPageX
	case 0b011:
		return Absolute
	case 0b111:
		return AbsoluteX
	}
	return 0
}

func (m *Mos6502) execAsl(addressMode uint8) {
	mode := getAslAddressMode(addressMode)
	if mode == Accumulator {
		m.setFlag(CARRY_FLAG, m.registers.A&0b10000000 != 0)
		m.registers.A = m.registers.A << 1
		m.setFlag(ZERO_FLAG, m.registers.A != 0)
		return
	}

	addr := m.getPointerByAddressMode(mode)

	m.setFlag(CARRY_FLAG, (*addr)&0b10000000 != 0)
	*addr = *addr << 1
	m.setFlag(ZERO_FLAG, (*addr) != 0)
}

func (m *Mos6502) execBit(addressMode uint8) {
	var content uint8
	if addressMode == 0b001 {
		content = *m.getPointerByAddressMode(ZeroPage)
	} else {
		content = *m.getPointerByAddressMode(Absolute)
	}

	res := int8(content & m.registers.A)
	m.setFlag(NEGATIVE_FLAG, res < 0)
	m.setFlag(OVERFLOW_FLAG, res & 0b01000000 != 0)
	m.setFlag(ZERO_FLAG, res == 0)
}

func (m *Mos6502) execCmp(addressMode uint8) {
	var content uint8
	switch addressMode {
	case 0b010:
		content = *m.getPointerByAddressMode(Immediate)
	case 0b001:
		content = *m.getPointerByAddressMode(ZeroPage)
	case 0b101:
		content = *m.getPointerByAddressMode(ZeroPageX)
	case 0b011:
		content = *m.getPointerByAddressMode(Absolute)
	case 0b111:
		content = *m.getPointerByAddressMode(AbsoluteX)
	case 0b110:
		content = *m.getPointerByAddressMode(AbsoluteY)
	case 0b000:
		content = *m.getPointerByAddressMode(IndirectIndexedX)
	case 0b100:
		content = *m.getPointerByAddressMode(IndexedIndirectY)
	}

	result := int8(content) - int8(m.registers.A)
	m.setFlag(CARRY_FLAG, result > 0)
	m.setFlag(NEGATIVE_FLAG, result < 0)
	m.setFlag(ZERO_FLAG, result == 0)
}

func (m *Mos6502) execCpx(addressMode uint8) {
	var content uint8
	switch addressMode {
	case 0b000:
		content = *m.getPointerByAddressMode(Immediate)
	case 0b001:
		content = *m.getPointerByAddressMode(ZeroPage)
	case 0b011:
		content = *m.getPointerByAddressMode(Absolute)
	}

	result := int8(content) - int8(m.registers.X)
	m.setFlag(CARRY_FLAG, result > 0)
	m.setFlag(NEGATIVE_FLAG, result < 0)
	m.setFlag(ZERO_FLAG, result == 0)
}

func (m *Mos6502) execCpy(addressMode uint8) {
	var content uint8
	switch addressMode {
	case 0b000:
		content = *m.getPointerByAddressMode(Immediate)
	case 0b001:
		content = *m.getPointerByAddressMode(ZeroPage)
	case 0b011:
		content = *m.getPointerByAddressMode(Absolute)
	}

	result := int8(content) - int8(m.registers.Y)
	m.setFlag(CARRY_FLAG, result > 0)
	m.setFlag(NEGATIVE_FLAG, result < 0)
	m.setFlag(ZERO_FLAG, result == 0)
}

func (m *Mos6502) execDec(addressMode uint8) {
	var content *uint8
	switch addressMode {
	case 0b001:
		content = m.getPointerByAddressMode(ZeroPage)
	case 0b101:
		content = m.getPointerByAddressMode(ZeroPageX)
	case 0b011:
		content = m.getPointerByAddressMode(Absolute)
	case 0b111:
		content = m.getPointerByAddressMode(AbsoluteX)
	}

	*content -= 1
	m.setFlag(NEGATIVE_FLAG, *content & 0b10000000 != 0)
	m.setFlag(ZERO_FLAG, *content == 0)
}

func (m *Mos6502) execEor(addressMode uint8) {
	var content uint8
	switch addressMode {
	case 0b010:
		content = *m.getPointerByAddressMode(Immediate)
	case 0b001:
		content = *m.getPointerByAddressMode(ZeroPage)
	case 0b101:
		content = *m.getPointerByAddressMode(ZeroPageX)
	case 0b011:
		content = *m.getPointerByAddressMode(Absolute)
	case 0b111:
		content = *m.getPointerByAddressMode(AbsoluteX)
	case 0b110:
		content = *m.getPointerByAddressMode(AbsoluteY)
	case 0b000:
		content = *m.getPointerByAddressMode(IndirectIndexedX)
	case 0b100:
		content = *m.getPointerByAddressMode(IndexedIndirectY)
	}

	m.registers.A ^= content
	m.setFlag(NEGATIVE_FLAG, m.registers.A & 0b1000000 != 0)
	m.setFlag(ZERO_FLAG, m.registers.A == 0)
}

func (m *Mos6502) execInc(addressMode uint8) {
	var content *uint8
	switch addressMode {
	case 0b001:
		content = m.getPointerByAddressMode(ZeroPage)
	case 0b101:
		content = m.getPointerByAddressMode(ZeroPageX)
	case 0b011:
		content = m.getPointerByAddressMode(Absolute)
	case 0b111:
		content = m.getPointerByAddressMode(AbsoluteX)
	}

	*content = *content + 1
	m.setFlag(NEGATIVE_FLAG, *content & 0b1000000 != 0)
	m.setFlag(ZERO_FLAG, *content == 0)
}

//
// Branches
//

func (m *Mos6502) execBcc() {
	addr := m.getPointerByAddressMode(Immediate)
	if !m.isFlagSet(CARRY_FLAG) {
		m.registers.PC = uint16(int16(signExtend(*addr)) + int16(m.registers.PC))
	}
}

func (m *Mos6502) execBcs() {
	addr := m.getPointerByAddressMode(Immediate)
	if m.isFlagSet(CARRY_FLAG) {
		m.registers.PC = uint16(int16(signExtend(*addr)) + int16(m.registers.PC))
	}
}

func (m *Mos6502) execBranchOnFlag(flag uint8) {
	addr := m.getPointerByAddressMode(Immediate)
	if m.isFlagSet(flag) {
		m.registers.PC = uint16(int16(signExtend(*addr)) + int16(m.registers.PC))
	}
}

func (m *Mos6502) execBranchOnNotFlag(flag uint8) {
	addr := m.getPointerByAddressMode(Immediate)
	if !m.isFlagSet(flag) {
		m.registers.PC = uint16(int16(signExtend(*addr)) + int16(m.registers.PC))
	}
}

func (m *Mos6502) NextInstruction() (*machine.Call, error) {
	rawOpcode, _ := m.GetMemory(uint64(m.registers.PC))
	m.registers.PC++

	switch rawOpcode {
	// brk
	case 0:
		return m.performSyscall()
	//
	// Branches.
	//
	case 0b10010000:
		m.execBcc()
	case 0b10110000:
		m.execBcs()
	// beq
	case 0b11110000:
		m.execBranchOnFlag(ZERO_FLAG)
	// bmi
	case 0b00110000:
		m.execBranchOnFlag(NEGATIVE_FLAG)
	// bvs
	case 0b01110000:
		m.execBranchOnFlag(CARRY_FLAG)
	// bne
	case 0b11010000:
		m.execBranchOnNotFlag(ZERO_FLAG)
	// bpl
	case 0b00010000:
		m.execBranchOnNotFlag(NEGATIVE_FLAG)
	// bvc
	case 0b0101000:
		m.execBranchOnNotFlag(OVERFLOW_FLAG)
	//
	// Clears.
	//
	// clc
	case 0b00011000:
		m.setFlag(CARRY_FLAG, false)
	// cld
	case 0b11011000:
		m.setFlag(DECIMAL_FLAG, false)
	// cli
	case 0b01011000:
		m.setFlag(INTERRUPT_DISABLE_FLAG, false)
	// clv
	case 0b10111000:
		m.setFlag(OVERFLOW_FLAG, false)
	//
	// Others.
	//
	// dex
	case 0b11001010:
		m.registers.X--
		m.setFlag(NEGATIVE_FLAG, m.registers.X & 0b10000000 != 0)
		m.setFlag(ZERO_FLAG, m.registers.X == 0)
	// dey
	case 0b10001000:
		m.registers.Y--
		m.setFlag(NEGATIVE_FLAG, m.registers.Y & 0b10000000 != 0)
		m.setFlag(ZERO_FLAG, m.registers.Y == 0)
	// inx
	case 0b11101000:
		m.registers.X++
		m.setFlag(NEGATIVE_FLAG, m.registers.X & 0b10000000 != 0)
		m.setFlag(ZERO_FLAG, m.registers.X == 0)
	// iny
	case 0b11001000:
		m.registers.Y++
		m.setFlag(NEGATIVE_FLAG, m.registers.Y & 0b10000000 != 0)
		m.setFlag(ZERO_FLAG, m.registers.Y == 0)
default:
		opcode, addressMode := parseOpcode(rawOpcode)
		switch opcode {
		case 0b01101:
			m.execAdc(addressMode)
		case 0b00101:
			m.execAnd(addressMode)
		case 0b00010:
			m.execAsl(addressMode)
		case 0b00100:
			m.execBit(addressMode)
		case 0b11001:
			m.execCmp(addressMode)
		case 0b11100:
			m.execCpx(addressMode)
		case 0b11000:
			m.execCpy(addressMode)
		case 0b11010:
			m.execDec(addressMode)
		case 0b01001:
			m.execEor(addressMode)
		case 0b11110:
			m.execInc(addressMode)
		}
	}

	return nil, nil
}

func (m *Mos6502) LoadProgram(program []uint8) error {
	m.registers.PC = TEXT_PAGE
	m.registers.S = 0
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
