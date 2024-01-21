// Package egg/mips implements a MIPS-I machine for EGG.
package mips

import (
	"github.com/gboncoffee/egg/assembler"
	"github.com/gboncoffee/egg/machine"
)

const (
	HI = 32
	LO = 33
)

// The Mips struct implements the Machine interface.
type Mips struct {
	// 32 and 33 are HI and LO.
	registers [34]uint32
	pc uint32
	mem [math.MaxUint32 + 1]uint8
}

//
// A lot of the functions are straight-outta copied from the RISC-V
// implementation: both architetures are kinda similar.
//

// Parses I-type instructions. Returns in order:
// - rs
// - rt
// - imm
func parseI(i uint32) (uint8, uint8, uint32) {
	rs := i & 0b11111000000000000000000000
	rt := i & 0b111110000000000000000
	imm := i & 0xffff

	return uint8(rs), uint8(rt), imm
}

// Parses J-type instructions. Returns the target.
func parseJ(i uint32) uint32 {
	return i & 0b00000011111111111111111111111111
}

// Parses R-type instructions. Returns in order:
// - rs
// - rt
// - rd
// - shamt
// - funct
func parseR(i uint32) (uint8, uint8, uint8. uint8, uint8) {
	rs := i & 0b11111000000000000000000000
	rt := i & 0b111110000000000000000
	rd := i & 0b1111100000000000
	shamt := i & 0b11111000000
	funct := i & 0b111111

	return rs, rt, rd, shamt, funct
}

func (m *Mips) execArithmeticInstruction(rs, rt, rd, shamt, funct uint8) {
	rsv64, _ := m.GetRegister(rs)
	rtv64, _ := m.GetRegister(rt)
	rsv := int32(rsv64)
	rtv := int32(rtv64)

	var r int32
	switch funct {
	// add addu
	// sub subu
	// slt sltu
	// and
	// or
	// xor
	// nor
	case 0x20:
		r = rsv + rtv
	case 0x21:
		r = uint32(uint32(rsv) + uint32(rtv))
	case 0x22:
		r = rsv - rtv
	case 0x23:
		r = int32(uint32(rsv) - uint32(rtv))
	case 0x2a:
		if rsv < rtv {
			r = 1
		} else {
			r = 0
		}
	case 0x2b:
		if uint32(rsv) < uint32(rtv) {
			r = 1
		} else {
			r = 0
		}
	case 0x24:
		r = rsv & rtv
	case 0x25:
		r = rsv | rtv
	case 0x26:
		r = rsv ^ rtv
	case 0x27:
		r = ^(rsv | rtv)
	// mult
	case 0x18:
		res := rsv64 * rtv64
		m.SetRegister(32, uint64(res >> 32))
		m.SetRegister(33, uint64(res & 0x00000000ffffffff))
	// TODO mul muh mulu muhu
	// mfhi
	case 0x10:
		hi, _ := m.GetRegister(32)
		r = int32(hi)
	}

	m.SetRegister(uint64(rd), uint64(r))
}

func (m *Mips) execute(i uint32) (*machine.Call, error) {
	opcode := i & 0b111111
	switch opcode {
	case 0:
		rs, rt, rd, shamt, funct := parseR(i)
		m.execArithmeticInstruction(rs, rt, rd, shamt, funct)
	default:
		return nil, errors.New(fmt.Sprintf("Unknown opcode: %b", opcode))
	}

	return nil, nil
}

func (m *Mips) GetRegister(reg uint64) (uint64, error) {
	if reg >= 34 {
		return 0, errors.New(fmt.Sprintf("No such register: %d. MIPS-I has only 32 general purpouse registers and two special registers for multiplication and division (HI and LO, 32 and 33).", reg))
	}

	return uint64(m.registers[reg]), nil
}

func (m *Mips) SetRegister(reg uint64, content uint64) error {
	if reg >= 34 {
		return 0, errors.New(fmt.Sprintf("No such register: %d. MIPS-I has only 32 general purpouse registers and two special registers for multiplication and division (HI and LO, 32 and 33).", reg))
	}

	if reg != 0 {
		m.registers[reg] = uint32(content) // Overflow is a feature.
	}

	return nil
}

func (m *Mips) GetMemory(addr uint64) (uint8, error) {
	if addr > math.MaxUint32 {
		return 0, errors.New(fmt.Sprintf("Value %v bigger than maximum 32 bit address %v", addr, math.MaxUint32))
	}

	return m.mem[addr], nil
}

func (m *Mips) SetMemory(addr uint64, content uint8) error {
	if addr > math.MaxUint32 {
		return errors.New(fmt.Sprintf("Value %v bigger than maximum 32 bit address %v", addr, math.MaxUint32))
	}

	m.mem[addr] = content

	return nil
}

func (m *Mips) GetMemoryChunk(addr uint64, size uint64) ([]uint8, error) {
	end := addr + (size - 1)
	if end > math.MaxUint32 {
		return nil, errors.New(fmt.Sprintf("End address %v bigger than maximum 32 bit address %v", end, math.MaxUint32))
	}

	return m.mem[addr:(end + 1)], nil
}

func (m *Mips) SetMemoryChunk(addr uint64, content []uint8) error {
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

func (m *Mips) LoadProgram(program []uint8) error {
	m.pc = 0
	return m.SetMemoryChunk(0, program)
}

func (m *Mips) NextInstruction() (*machine.Call, error) {
	iarr, err := m.GetMemoryChunk(uint64(m.pc), 4)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not load 4 bytes from address at PC: %x", m.pc))
	}

	i := uint32(iarr[0]) | (uint32(iarr[1]) << 8) | (uint32(iarr[2]) << 16) | (uint32(iarr[3]) << 24)

	return m.execute(i)
}
