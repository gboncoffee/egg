// Package egg/mips implements a MIPS-I machine for EGG.
package mips

import (
	"errors"
	"fmt"
	"math"
	"math/bits"

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
	pc        uint32
	mem       [math.MaxUint32 + 1]uint8
}

//
// Instructions
//
// add DONE
// addi DONE
// addiu DONE
// addu DONE
// clo DONE
// clz DONE
// lui
// seb DONE
// seh DONE
// sub DONE
// subu DONE
// sll DONE
// sllv DONE
// sra DONE
// srav DONE
// srl DONE
// srlv DONE
// and DONE
// andi
// nor DONE
// or DONE
// ori
// xor DONE
// xori
// movn DONE
// movz DONE
// slt DONE
// slti
// sltiu
// sltu DONE
// div DONE
// mult DONE
// mfhi DONE
// mflo DONE
// mthi DONE
// mtlo DONE
// beq
// bgez
// bgtz
// blez
// bltz
// bne
// break
// syscall
// j
// jal
// jalr
// jr
// lb
// lbu
// lh
// lhu
// lw
// lwl
// lwr
// sb
// sh
// sw
// swl
// swr
//

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
func parseR(i uint32) (uint8, uint8, uint8, uint8, uint8) {
	rs := uint8(i & 0b11111000000000000000000000)
	rt := uint8(i & 0b111110000000000000000)
	rd := uint8(i & 0b1111100000000000)
	shamt := uint8(i & 0b11111000000)
	funct := uint8(i & 0b111111)

	return rs, rt, rd, shamt, funct
}

func (m *Mips) execSpecial(rs, rt, rd, shamt, funct uint8) {
	rsv64, _ := m.GetRegister(uint64(rs))
	rtv64, _ := m.GetRegister(uint64(rd))
	rsv := int32(rsv64)
	rtv := int32(rtv64)

	// A lot of instructions conditionally or do not change the RD, so we
	// init r with it's value.
	var r int32
	rdv64, _ := m.GetRegister(uint64(rd))
	r = int32(rdv64)
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
		r = int32(uint32(rsv) + uint32(rtv))
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
	// sll
	// sllv
	// sra
	// srav
	// srl
	// srlv
	case 0:
		r = rtv << shamt
	case 4:
		r = rtv << rsv
	case 3:
		r = rtv >> shamt
	case 7:
		r = rtv >> rsv
	case 2:
		r = int32(uint32(rtv) >> uint32(shamt))
	case 6:
		r = int32(uint32(rtv) >> uint32(rsv))
	// mult
	// div
	case 0x18:
		res := rsv64 * rtv64
		m.SetRegister(HI, uint64(res>>32))
		m.SetRegister(LO, uint64(res&0x00000000ffffffff))
	case 0x1a:
		m.SetRegister(HI, uint64(rsv%rtv))
		m.SetRegister(LO, uint64(rsv/rtv))
	// mfhi
	// mflo
	// mthi
	// mtlo
	case 0x10:
		hi, _ := m.GetRegister(HI)
		r = int32(hi)
	case 0x12:
		lo, _ := m.GetRegister(LO)
		r = int32(lo)
	case 0x11:
		m.SetRegister(HI, uint64(rsv))
	case 0x13:
		m.SetRegister(LO, uint64(rsv))
	// movz
	// movn
	case 0xa:
		if rtv == 0 {
			r = rsv
		}
	case 0xb:
		if rtv != 0 {
			r = rsv
		}
	}

	m.SetRegister(uint64(rd), uint64(r))
}

func (m *Mips) execSpecial2(rs, rt, rd, shamt, funct uint8) {
	rsv64, _ := m.GetRegister(uint64(rs))
	rsv := int32(rsv64)

	var r int32
	switch funct {
	// clz
	// clo
	case 16:
		r = int32(bits.LeadingZeros32(uint32(rsv)))
	case 17:
		r = int32(bits.LeadingZeros32(uint32(^rsv)))
	}

	m.SetRegister(uint64(rd), uint64(r))
}

func (m *Mips) execSpecial3(rs, rt, rd, shamt, funct uint8) {
	rsv64, _ := m.GetRegister(uint64(rs))
	rsv := int32(rsv64)

	var r int32
	switch funct {
	case 32:
		switch shamt {
		// seb
		// seh
		case 16:
			rsvb := rsv & 0xff
			sign := rsvb >> 7
			sign = (^(sign - 1)) << 8
			r = int32(rsvb | sign)
		case 24:
			rsvb := rsv & 0xff
			sign := rsvb >> 15
			sign = (^(sign - 1)) << 16
			r = int32(rsvb | sign)
		}
	}

	m.SetRegister(uint64(rd), uint64(r))
}

func (m *Mips) executeAddi(rs, rt uint8, imm uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))
	r := int32(rsv64) + int32(imm)
	m.SetRegister(uint64(rt), uint64(r))
}

func (m *Mips) executeAddiu(rs, rt uint8, imm uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))
	r := uint32(rsv64) + imm
	m.SetRegister(uint64(rt), uint64(r))
}

func (m *Mips) execute(i uint32) (*machine.Call, error) {
	opcode := i & 0b111111
	switch opcode {
	case 0:
		rs, rt, rd, shamt, funct := parseR(i)
		m.execSpecial(rs, rt, rd, shamt, funct)
	case 28:
		rs, rt, rd, shamt, funct := parseR(i)
		m.execSpecial2(rs, rt, rd, shamt, funct)
	case 31:
		rs, rt, rd, shamt, funct := parseR(i)
		m.execSpecial3(rs, rt, rd, shamt, funct)
	// addi
	case 8:
		rs, rt, imm := parseI(i)
		m.executeAddi(rs, rt, imm)
	// addiu
	case 9:
		rs, rt, imm := parseI(i)
		m.executeAddiu(rs, rt, imm)
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
		return errors.New(fmt.Sprintf("No such register: %d. MIPS-I has only 32 general purpouse registers and two special registers for multiplication and division (HI and LO, 32 and 33).", reg))
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
