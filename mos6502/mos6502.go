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
	"unsafe"

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
	case Accumulator:
		return &m.registers.A
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

func (m *Mos6502) execAdc(addressMode uint8) error {
	var mode AddressMode
	switch addressMode {
	case 0b010:
		mode = Immediate
	case 0b001:
		mode = ZeroPage
	case 0b101:
		mode = ZeroPageX
	case 0b011:
		mode = Absolute
	case 0b111:
		mode = AbsoluteX
	case 0b110:
		mode = AbsoluteY
	case 0b000:
		mode = IndexedIndirectX
	case 0b100:
		mode = IndirectIndexedY
	}
	operand := *m.getPointerByAddressMode(mode)

	if m.isFlagSet(CARRY_FLAG) {
		operand++
	}

	// Stores old signal for the overflow flag.
	signal := m.registers.A & 0b10000000

	overflowedResult := uint16(m.registers.A) + uint16(operand)
	m.setFlag(CARRY_FLAG, overflowedResult > math.MaxUint8)

	m.registers.A += operand
	m.setFlag(OVERFLOW_FLAG, m.registers.A&0b10000000 != signal)
	m.setFlag(NEGATIVE_FLAG, m.registers.A&0b10000000 != 0)
	m.setFlag(ZERO_FLAG, m.registers.A == 0)

	return nil
}

func (m *Mos6502) execAnd(addressMode uint8) {
	var mode AddressMode
	switch addressMode {
	case 0b010:
		mode = Immediate
	case 0b001:
		mode = ZeroPage
	case 0b101:
		mode = ZeroPageX
	case 0b011:
		mode = Absolute
	case 0b111:
		mode = AbsoluteX
	case 0b110:
		mode = AbsoluteY
	case 0b000:
		mode = IndexedIndirectX
	case 0b100:
		mode = IndirectIndexedY
	}

	operand := *m.getPointerByAddressMode(mode)

	m.registers.A = m.registers.A & operand
	m.setFlag(ZERO_FLAG, m.registers.A != 0)
	m.setFlag(NEGATIVE_FLAG, m.registers.A&0b10000000 != 0)
}

func (m *Mos6502) execAsl(addressMode uint8) {
	var mode AddressMode
	switch addressMode {
	// Accumulator
	case 0b010:
		mode = Accumulator
	case 0b001:
		mode = ZeroPage
	case 0b101:
		mode = ZeroPageX
	case 0b011:
		mode = Absolute
	case 0b111:
		mode = AbsoluteX
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
	m.setFlag(OVERFLOW_FLAG, res&0b01000000 != 0)
	m.setFlag(ZERO_FLAG, res == 0)
}

func (m *Mos6502) execCmp(addressMode uint8) {
	var mode AddressMode
	switch addressMode {
	case 0b010:
		mode = Immediate
	case 0b001:
		mode = ZeroPage
	case 0b101:
		mode = ZeroPageX
	case 0b011:
		mode = Absolute
	case 0b111:
		mode = AbsoluteX
	case 0b110:
		mode = AbsoluteY
	case 0b000:
		mode = IndirectIndexedX
	case 0b100:
		mode = IndexedIndirectY
	}
	content := *m.getPointerByAddressMode(mode)

	result := int8(content) - int8(m.registers.A)
	m.setFlag(CARRY_FLAG, result > 0)
	m.setFlag(NEGATIVE_FLAG, result < 0)
	m.setFlag(ZERO_FLAG, result == 0)
}

func (m *Mos6502) execCpx(addressMode uint8) {
	var mode AddressMode
	switch addressMode {
	case 0b000:
		mode = Immediate
	case 0b001:
		mode = ZeroPage
	case 0b011:
		mode = Absolute
	}
	content := *m.getPointerByAddressMode(mode)

	result := int8(content) - int8(m.registers.X)
	m.setFlag(CARRY_FLAG, result > 0)
	m.setFlag(NEGATIVE_FLAG, result < 0)
	m.setFlag(ZERO_FLAG, result == 0)
}

func (m *Mos6502) execCpy(addressMode uint8) {
	var mode AddressMode
	switch addressMode {
	case 0b000:
		mode = Immediate
	case 0b001:
		mode = ZeroPage
	case 0b011:
		mode = Absolute
	}
	content := *m.getPointerByAddressMode(mode)

	result := int8(content) - int8(m.registers.Y)
	m.setFlag(CARRY_FLAG, result > 0)
	m.setFlag(NEGATIVE_FLAG, result < 0)
	m.setFlag(ZERO_FLAG, result == 0)
}

func (m *Mos6502) execDec(addressMode uint8) {
	var mode AddressMode
	switch addressMode {
	case 0b001:
		mode = ZeroPage
	case 0b101:
		mode = ZeroPageX
	case 0b011:
		mode = Absolute
	case 0b111:
		mode = AbsoluteX
	}
	content := m.getPointerByAddressMode(mode)

	*content -= 1
	m.setFlag(NEGATIVE_FLAG, *content&0b10000000 != 0)
	m.setFlag(ZERO_FLAG, *content == 0)
}

func (m *Mos6502) execEor(addressMode uint8) {
	var mode AddressMode
	switch addressMode {
	case 0b010:
		mode = Immediate
	case 0b001:
		mode = ZeroPage
	case 0b101:
		mode = ZeroPageX
	case 0b011:
		mode = Absolute
	case 0b111:
		mode = AbsoluteX
	case 0b110:
		mode = AbsoluteY
	case 0b000:
		mode = IndirectIndexedX
	case 0b100:
		mode = IndexedIndirectY
	}
	content := *m.getPointerByAddressMode(mode)

	m.registers.A ^= content
	m.setFlag(NEGATIVE_FLAG, m.registers.A&0b1000000 != 0)
	m.setFlag(ZERO_FLAG, m.registers.A == 0)
}

func (m *Mos6502) execInc(addressMode uint8) {
	var mode AddressMode
	switch addressMode {
	case 0b001:
		mode = ZeroPage
	case 0b101:
		mode = ZeroPageX
	case 0b011:
		mode = Absolute
	case 0b111:
		mode = AbsoluteX
	}
	content := m.getPointerByAddressMode(mode)

	*content = *content + 1
	m.setFlag(NEGATIVE_FLAG, *content&0b1000000 != 0)
	m.setFlag(ZERO_FLAG, *content == 0)
}

func (m *Mos6502) execLda(addressMode uint8) {
	var mode AddressMode
	switch addressMode {
	case 0b010:
		mode = Immediate
	case 0b001:
		mode = ZeroPage
	case 0b101:
		mode = ZeroPageX
	case 0b011:
		mode = Absolute
	case 0b111:
		mode = AbsoluteX
	case 0b110:
		mode = AbsoluteY
	case 0b000:
		mode = IndirectIndexedX
	case 0b100:
		mode = IndexedIndirectY
	}
	content := *m.getPointerByAddressMode(mode)

	m.registers.A = content
	m.setFlag(NEGATIVE_FLAG, m.registers.A&0b1000000 != 0)
	m.setFlag(ZERO_FLAG, m.registers.A == 0)
}

func (m *Mos6502) execLdxOrLdy(addressMode uint8, register *uint8) {
	var mode AddressMode
	switch addressMode {
	case 0b000:
		mode = Immediate
	case 0b001:
		mode = ZeroPage
	case 0b101:
		mode = ZeroPageX
	case 0b011:
		mode = Absolute
	case 0b111:
		mode = AbsoluteX
	}
	content := *m.getPointerByAddressMode(mode)

	*register = content
	m.setFlag(NEGATIVE_FLAG, *register&0b1000000 != 0)
	m.setFlag(ZERO_FLAG, *register == 0)
}

func (m *Mos6502) execLsr(addressMode uint8) {
	var mode AddressMode
	switch addressMode {
	case 0b010:
		mode = Accumulator
	case 0b001:
		mode = ZeroPage
	case 0b101:
		mode = ZeroPageX
	case 0b011:
		mode = Absolute
	case 0b111:
		mode = AbsoluteX
	}
	content := m.getPointerByAddressMode(mode)

	m.setFlag(NEGATIVE_FLAG, false)
	m.setFlag(CARRY_FLAG, *content&0b1 != 0)

	*content = *content >> 1

	m.setFlag(ZERO_FLAG, *content == 0)
}

func (m *Mos6502) execOra(addressMode uint8) {
	var mode AddressMode
	switch addressMode {
	case 0b010:
		mode = Immediate
	case 0b001:
		mode = ZeroPage
	case 0b101:
		mode = ZeroPageX
	case 0b011:
		mode = Absolute
	case 0b111:
		mode = AbsoluteX
	case 0b110:
		mode = AbsoluteY
	case 0b000:
		mode = IndirectIndexedX
	case 0b100:
		mode = IndexedIndirectY
	}
	content := *m.getPointerByAddressMode(mode)

	m.registers.A |= content

	m.setFlag(NEGATIVE_FLAG, m.registers.A & 0b10000000 != 0)
	m.setFlag(ZERO_FLAG, m.registers.A == 0)
}

func (m *Mos6502) execRol(addressMode uint8) {
	var mode AddressMode
	switch addressMode {
	case 0b010:
		mode = Accumulator
	case 0b001:
		mode = ZeroPage
	case 0b101:
		mode = ZeroPageX
	case 0b011:
		mode = Absolute
	case 0b111:
		mode = AbsoluteX
	}
	addr := m.getPointerByAddressMode(mode)

	// Afaik, this is the idiomatic way to get the byte value of a boolean. I
	// could just ignore the existence of isFlagSet and just grab the flag
	// straight from m.registers.P but I prefer not to do it for the sake of
	// separation of concerns, as the flags were abstracted away with the
	// function. I still hope this is optimized-out tho.
	var carry uint8
	if m.isFlagSet(CARRY_FLAG) {
		carry = 1
	} else {
		carry = 0
	}

	m.setFlag(CARRY_FLAG, *addr & 0b10000000 != 0)
	*addr = *addr << 1
	*addr |= carry

	m.setFlag(NEGATIVE_FLAG, *addr & 0b10000000 != 0)
	m.setFlag(ZERO_FLAG, *addr != 0)
}

//
// Branches and jumps.
//

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

func (m *Mos6502) execJmpAbsolute() {
	addr := m.getPointerByAddressMode(Absolute)

	// Unsafe magic to convert the address to an index.
	// Works because Mos6502.mem is an array of bytes (uint8) so we don't have
	// to call sizeof().
	jmpLoc := uintptr(unsafe.Pointer(addr))
	memBase := uintptr(unsafe.Pointer(&m.mem[0]))
	m.registers.PC = uint16(jmpLoc - memBase)
}

func (m *Mos6502) execJmpIndirect() {
	addr := m.getPointerByAddressMode(Absolute)

	// Unsafe magic to read two bytes and jump to them.
	m.registers.PC = *(*uint16)(unsafe.Pointer(addr))
}

func (m *Mos6502) execJsr() {
	// Same as the previous. We need to get it before computing `retAddrSlice`
	// as the target is get with the Absolute addressing mode.
	target := m.getPointerByAddressMode(Absolute)

	m.registers.S -= 2
	stackAddr := uint64(m.registers.S) | STACK_PAGE
	retAddrSlice := []uint8{uint8(m.registers.PC), uint8(m.registers.PC >> 8)}
	m.SetMemoryChunk(stackAddr, retAddrSlice)

	m.registers.PC = *(*uint16)(unsafe.Pointer(target))
}

func (m *Mos6502) NextInstruction() (*machine.Call, error) {
	rawOpcode, _ := m.GetMemory(uint64(m.registers.PC))
	m.registers.PC++

	switch rawOpcode {
	// brk
	case 0:
		return m.performSyscall()
	// nop
	case 0b11101010:
		return nil, nil
	//
	// Branches and jumps
	//
	// bcc
	case 0b10010000:
		m.execBranchOnNotFlag(CARRY_FLAG)
	// bcs
	case 0b10110000:
		m.execBranchOnFlag(CARRY_FLAG)
	// beq
	case 0b11110000:
		m.execBranchOnFlag(ZERO_FLAG)
	// bmi
	case 0b00110000:
		m.execBranchOnFlag(NEGATIVE_FLAG)
	// bvs
	case 0b01110000:
		m.execBranchOnFlag(OVERFLOW_FLAG)
	// bne
	case 0b11010000:
		m.execBranchOnNotFlag(ZERO_FLAG)
	// bpl
	case 0b00010000:
		m.execBranchOnNotFlag(NEGATIVE_FLAG)
	// bvc
	case 0b01010000:
		m.execBranchOnNotFlag(OVERFLOW_FLAG)
	case 0b01001100:
		m.execJmpAbsolute()
	case 0b01101100:
		m.execJmpIndirect()
	case 0b00100000:
		m.execJsr()
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
		m.setFlag(NEGATIVE_FLAG, m.registers.X&0b10000000 != 0)
		m.setFlag(ZERO_FLAG, m.registers.X == 0)
	// dey
	case 0b10001000:
		m.registers.Y--
		m.setFlag(NEGATIVE_FLAG, m.registers.Y&0b10000000 != 0)
		m.setFlag(ZERO_FLAG, m.registers.Y == 0)
	// inx
	case 0b11101000:
		m.registers.X++
		m.setFlag(NEGATIVE_FLAG, m.registers.X&0b10000000 != 0)
		m.setFlag(ZERO_FLAG, m.registers.X == 0)
	// iny
	case 0b11001000:
		m.registers.Y++
		m.setFlag(NEGATIVE_FLAG, m.registers.Y&0b10000000 != 0)
		m.setFlag(ZERO_FLAG, m.registers.Y == 0)
	// pha
	case 0b01001000:
		m.registers.S--
		m.SetMemory(uint64(m.registers.S), m.registers.A)
	// pla
	case 0b01101000:
		m.registers.A, _ = m.GetMemory(uint64(m.registers.S))
		m.registers.S++
		m.setFlag(NEGATIVE_FLAG, m.registers.A & 0b10000000 != 0)
		m.setFlag(ZERO_FLAG, m.registers.A != 0)
	// php
	case 0b00001000:
		m.registers.S--
		m.SetMemory(uint64(m.registers.S), m.registers.P)
	// plp
	case 0b00101000:
		m.registers.P, _ = m.GetMemory(uint64(m.registers.S))
		m.registers.S++
		m.setFlag(NEGATIVE_FLAG, m.registers.P & 0b10000000 != 0)
		m.setFlag(ZERO_FLAG, m.registers.P != 0)
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
		case 0b10101:
			m.execLda(addressMode)
		case 0b10110:
			// Ldx
			m.execLdxOrLdy(addressMode, &m.registers.X)
		case 0b10100:
			// Ldy
			m.execLdxOrLdy(addressMode, &m.registers.Y)
		case 0b01010:
			m.execLsr(addressMode)
		case 0b00001:
			m.execOra(addressMode)
		case 0b00110:
			m.execRol(addressMode)
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