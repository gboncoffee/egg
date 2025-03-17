// Package egg/mips implements a MIPS-I machine for EGG.
package mips

import (
	"errors"
	"fmt"
	"math"
	"math/bits"
	"strconv"

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
	pc        uint32
	mem       [math.MaxUint32 + 1]uint8
}

//
// Instructions
//
// add DONE ASM
// addi DONE ASM
// addiu DONE ASM
// addu DONE ASM
// clo DONE ASM
// clz DONE ASM
// lui DONE ASM
// seb DONE ASM
// seh DONE ASM
// sub DONE ASM
// subu DONE ASM
// sll DONE ASM
// sllv DONE ASM
// sra DONE ASM
// srav DONE ASM
// srl DONE ASM
// srlv DONE ASM
// and DONE ASM
// andi DONE ASM
// nor DONE ASM
// or DONE ASM
// ori DONE ASM
// xor DONE ASM
// xori DONE ASM
// movn DONE ASM
// movz DONE ASM
// slt DONE ASM
// slti DONE ASM
// sltiu DONE ASM
// sltu DONE ASM
// div DONE ASM
// mult DONE ASM
// mfhi DONE ASM
// mflo DONE ASM
// mthi DONE ASM
// mtlo DONE ASM
// beq DONE ASM
// bgez DONE ASM
// bgtz DONE ASM
// blez DONE ASM
// bltz DONE ASM
// bne DONE ASM
// break DONE ASM
// syscall DONE ASM
// j DONE ASM
// jal DONE ASM
// jalr DONE ASM
// jr DONE ASM
// lb DONE ASM
// lbu DONE ASM
// lh DONE ASM
// lhu DONE ASM
// lw DONE ASM
// lwl DONE ASM
// lwr DONE ASM
// sb DONE ASM
// sh DONE ASM
// sw DONE ASM
//

func signExtend16(n uint16) uint64 {
	sign := n >> 15
	sign64 := uint64(^(sign - 1)) << 16
	return uint64(n) | sign64
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
	rs := (i & 0b11111000000000000000000000) >> 21
	rt := (i & 0b111110000000000000000) >> 16
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
	rs := uint8((i & 0b11111000000000000000000000) >> 21)
	rt := uint8((i & 0b111110000000000000000) >> 16)
	rd := uint8((i & 0b1111100000000000) >> 11)
	shamt := uint8((i & 0b11111000000) >> 6)
	funct := uint8(i & 0b111111)

	return rs, rt, rd, shamt, funct
}

func (m *Mips) execSpecial(rs, rt, rd, shamt, funct uint8) {
	rsv64, _ := m.GetRegister(uint64(rs))
	rtv64, _ := m.GetRegister(uint64(rt))
	rsv := int32(rsv64)
	rtv := int32(rtv64)

	// A lot of instructions conditionally or do not change the RD, so we
	// init r with it's value.
	var r int32
	rdv64, _ := m.GetRegister(uint64(rd))
	r = int32(rdv64)
	switch funct {
	// jalr
	// jr
	case 9:
		r = int32(m.pc + 4)
		fallthrough
	case 8:
		m.pc = uint32(rsv) - 4
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
		r = rsv << shamt
	case 4:
		r = rsv << rtv
	case 3:
		r = rsv >> shamt
	case 7:
		r = rsv >> rtv
	case 2:
		r = int32(uint32(rsv) >> uint32(shamt))
	case 6:
		r = int32(uint32(rsv) >> uint32(rtv))
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

func (m *Mips) execSpecial2(rs, rd, funct uint8) {
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

func (m *Mips) execSpecial3(rs, rd, shamt, funct uint8) {
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
			rsvb := rsv & 0xffff
			sign := rsvb >> 15
			sign = (^(sign - 1)) << 16
			r = int32(rsvb | sign)
		}
	}

	m.SetRegister(uint64(rd), uint64(r))
}

func (m *Mips) execRegimm(rs, shamt uint8, imm uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))
	off := int32(signExtend16(uint16(imm)) << 2)

	switch shamt {
	// bltz
	case 0:
		if int32(rsv64) < 0 {
			m.pc = uint32(int32(m.pc) + off - 4)
		}
	// bgez
	case 1:
		if int32(rsv64) >= 0 {
			m.pc = uint32(int32(m.pc) + off - 4)
		}
	}
}

func (m *Mips) executeBeq(rs, rt uint8, imm uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))
	rtv64, _ := m.GetRegister(uint64(rt))

	off := int32(signExtend16(uint16(imm)) << 2)
	if rsv64 == rtv64 {
		m.pc = uint32(int32(m.pc) + off - 4)
	}
}

func (m *Mips) executeBne(rs, rt uint8, imm uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))
	rtv64, _ := m.GetRegister(uint64(rt))
	off := int32(signExtend16(uint16(imm)) << 2)
	if rsv64 != rtv64 {
		m.pc = uint32(int32(m.pc) + off - 4)
	}
}

func (m *Mips) executeBgtz(rs uint8, imm uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))

	off := int32(signExtend16(uint16(imm)) << 2)
	if int32(rsv64) > 0 {
		m.pc = uint32(int32(m.pc) + off - 4)
	}
}

func (m *Mips) executeBlez(rs uint8, imm uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))

	off := int32(signExtend16(uint16(imm)) << 2)
	if int32(rsv64) <= 0 {
		m.pc = uint32(int32(m.pc) + off - 4)
	}
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

func (m *Mips) executeAndi(rs, rt uint8, imm uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))
	r := uint32(rsv64) & imm
	m.SetRegister(uint64(rt), uint64(r))
}

func (m *Mips) executeOri(rs, rt uint8, imm uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))
	r := uint32(rsv64) | imm
	m.SetRegister(uint64(rt), uint64(r))
}

func (m *Mips) executeXori(rs, rt uint8, imm uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))
	r := uint32(rsv64) ^ imm
	m.SetRegister(uint64(rt), uint64(r))
}

func (m *Mips) executeSlti(rs, rt uint8, imm uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))
	if int32(rsv64) < int32(imm) {
		m.SetRegister(uint64(rt), 1)
	} else {
		m.SetRegister(uint64(rt), 0)
	}
}

func (m *Mips) executeSltiu(rs, rt uint8, imm uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))
	if uint32(rsv64) < imm {
		m.SetRegister(uint64(rt), 1)
	} else {
		m.SetRegister(uint64(rt), 0)
	}
}

func (m *Mips) executeLui(rt uint8, imm uint32) {
	m.SetRegister(uint64(rt), uint64(imm<<16))
}

func (m *Mips) executeJ(imm uint32) {
	t := (m.pc + 4) & 0xf0000000
	t = t | (imm << 2)
	m.pc = t - 4
}

func (m *Mips) executeJal(imm uint32) {
	m.SetRegister(31, uint64(m.pc+4))
	t := (m.pc + 4) & 0xf0000000
	t = t | (imm << 2)
	m.pc = t - 4
}

func (m *Mips) executeLb(rs, rt uint8, off uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))
	mem, _ := m.GetMemory(uint64(uint32(rsv64) + off))

	memb := mem & 0xff
	sign := uint64(memb >> 7)
	sign = (^(sign - 1)) << 8
	r := uint64(memb) | sign

	m.SetRegister(uint64(rt), r)
}

func (m *Mips) executeLbu(rs, rt uint8, off uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))
	mem, _ := m.GetMemory(uint64(uint32(rsv64) + off))

	m.SetRegister(uint64(rt), uint64(mem))
}

func (m *Mips) executeLh(rs, rt uint8, off uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))
	memSlice, _ := m.GetMemoryChunk(uint64(uint32(rsv64)+off), 2)
	mem := memSlice[0]
	mem2 := memSlice[1]

	memb := uint64(mem) | (uint64(mem2) << 8)
	sign := memb >> 15
	sign = (^(sign - 1)) << 16
	r := uint64(memb | sign)

	m.SetRegister(uint64(rt), r)
}

func (m *Mips) executeLhu(rs, rt uint8, off uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))
	memSlice, _ := m.GetMemoryChunk(uint64(uint32(rsv64)+off), 2)
	mem := memSlice[0]
	mem2 := memSlice[1]

	m.SetRegister(uint64(rt), uint64(mem)|(uint64(mem2)<<8))
}

func (m *Mips) executeLw(rs, rt uint8, off uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))
	memSlice, _ := m.GetMemoryChunk(uint64(uint32(rsv64)+off), 4)
	mem := memSlice[0]
	mem2 := memSlice[1]
	mem3 := memSlice[2]
	mem4 := memSlice[3]

	m.SetRegister(uint64(rt), uint64(mem)|(uint64(mem2)<<8)|(uint64(mem3)<<16)|(uint64(mem4)<<24))
}

func (m *Mips) executeLwl(rs, rt uint8, off uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))
	rtv64, _ := m.GetRegister(uint64(rt))
	memSlice, _ := m.GetMemoryChunk(uint64(uint32(rsv64)+off), 2)
	mem := memSlice[0]
	mem2 := memSlice[1]

	m.SetRegister(uint64(rt), ((uint64(mem)|(uint64(mem2)<<8))<<16)|(rtv64&0xffff))
}

func (m *Mips) executeLwr(rs, rt uint8, off uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))
	rtv64, _ := m.GetRegister(uint64(rt))
	memSlice, _ := m.GetMemoryChunk(uint64(uint32(rsv64)+off), 2)
	mem := memSlice[0]
	mem2 := memSlice[1]

	m.SetRegister(uint64(rt), (uint64(mem)|(uint64(mem2)<<8))|(rtv64&0xffff0000))
}

func (m *Mips) executeSb(rs, rt uint8, off uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))
	rtv64, _ := m.GetRegister(uint64(rt))

	m.SetMemory(uint64(uint32(rsv64)+off), uint8(rtv64&0xff))
}

func (m *Mips) executeSh(rs, rt uint8, off uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))
	rtv64, _ := m.GetRegister(uint64(rt))

	m.SetMemory(uint64(uint32(rsv64)+off), uint8(rtv64&0xff))
	m.SetMemory(uint64(uint32(rsv64)+off)+1, uint8((rtv64&0xff00)>>8))
}

func (m *Mips) executeSw(rs, rt uint8, off uint32) {
	rsv64, _ := m.GetRegister(uint64(rs))
	rtv64, _ := m.GetRegister(uint64(rt))

	m.SetMemory(uint64(uint32(rsv64)+off), uint8(rtv64&0xff))
	m.SetMemory(uint64(uint32(rsv64)+off)+1, uint8((rtv64&0xff00)>>8))
	m.SetMemory(uint64(uint32(rsv64)+off)+2, uint8((rtv64&0xff0000)>>16))
	m.SetMemory(uint64(uint32(rsv64)+off)+3, uint8((rtv64&0xff000000)>>24))
}

func (m *Mips) execute(i uint32) (*machine.Call, error) {
	opcode := i >> 26
	switch opcode {
	case 0:
		rs, rt, rd, shamt, funct := parseR(i)
		switch funct {
		case 0b1101:
			num := machine.SYS_BREAK
			call := machine.Call{Number: uint64(num), Arg1: 0, Arg2: 0}
			m.pc += 4
			return &call, nil
		case 0b1100:
			num, _ := m.GetRegister(2)
			a1, _ := m.GetRegister(4)
			a2, _ := m.GetRegister(5)
			call := machine.Call{Number: num, Arg1: a1, Arg2: a2}
			m.pc += 4
			return &call, nil
		default:
			m.execSpecial(rs, rt, rd, shamt, funct)
		}
	case 1:
		rs, shamt, imm := parseI(i)
		m.execRegimm(rs, shamt, imm)
	case 2:
		imm := parseJ(i)
		m.executeJ(imm)
	case 3:
		imm := parseJ(i)
		m.executeJal(imm)
	case 28:
		rs, _, rd, _, funct := parseR(i)
		m.execSpecial2(rs, rd, funct)
	case 31:
		_, rs, rd, shamt, funct := parseR(i)
		m.execSpecial3(rs, rd, shamt, funct)
	// beq
	case 4:
		rs, rt, imm := parseI(i)
		m.executeBeq(rs, rt, imm)
	// bne
	case 5:
		rs, rt, imm := parseI(i)
		m.executeBne(rs, rt, imm)
	// blez
	case 6:
		rs, _, imm := parseI(i)
		m.executeBlez(rs, imm)
	// bgtz
	case 7:
		rs, _, imm := parseI(i)
		m.executeBgtz(rs, imm)
	// addi
	case 8:
		rs, rt, imm := parseI(i)
		m.executeAddi(rs, rt, imm)
	// addiu
	case 9:
		rs, rt, imm := parseI(i)
		m.executeAddiu(rs, rt, imm)
	// slti
	case 10:
		rs, rt, imm := parseI(i)
		m.executeSlti(rs, rt, imm)
	// sltiu
	case 11:
		rs, rt, imm := parseI(i)
		m.executeSltiu(rs, rt, imm)
	// andi
	case 12:
		rs, rt, imm := parseI(i)
		m.executeAndi(rs, rt, imm)
	// ori
	case 13:
		rs, rt, imm := parseI(i)
		m.executeOri(rs, rt, imm)
	// xori
	case 14:
		rs, rt, imm := parseI(i)
		m.executeXori(rs, rt, imm)
	// lui
	case 15:
		_, rt, imm := parseI(i)
		m.executeLui(rt, imm)
	// lb
	// lbu
	// lh
	// lhu
	// lw
	// lwl
	// lwr
	case 32:
		rs, rt, off := parseI(i)
		m.executeLb(rs, rt, off)
	case 36:
		rs, rt, off := parseI(i)
		m.executeLbu(rs, rt, off)
	case 33:
		rs, rt, off := parseI(i)
		m.executeLh(rs, rt, off)
	case 37:
		rs, rt, off := parseI(i)
		m.executeLhu(rs, rt, off)
	case 35:
		rs, rt, off := parseI(i)
		m.executeLw(rs, rt, off)
	case 34:
		rs, rt, off := parseI(i)
		m.executeLwl(rs, rt, off)
	case 38:
		rs, rt, off := parseI(i)
		m.executeLwr(rs, rt, off)
	// sb
	// sh
	// sw
	case 40:
		rs, rt, off := parseI(i)
		m.executeSb(rs, rt, off)
	case 41:
		rs, rt, off := parseI(i)
		m.executeSh(rs, rt, off)
	case 43:
		rs, rt, off := parseI(i)
		m.executeSw(rs, rt, off)
	default:
		return nil, fmt.Errorf(machine.InterCtx.Get("unknown opcode: %b"), opcode)
	}

	m.pc += 4

	return nil, nil
}

func (m *Mips) GetRegister(reg uint64) (uint64, error) {
	if reg >= 34 {
		return 0, fmt.Errorf(machine.InterCtx.Get("no such register: %d. MIPS-I has only 32 general purpouse registers and two special registers for multiplication and division (HI and LO, 32 and 33)"), reg)
	}

	return uint64(m.registers[reg]), nil
}

func (m *Mips) SetRegister(reg uint64, content uint64) error {
	if reg >= 34 {
		return fmt.Errorf(machine.InterCtx.Get("no such register: %d. MIPS-I has only 32 general purpouse registers and two special registers for multiplication and division (HI and LO, 32 and 33)"), reg)
	}

	if reg != 0 {
		m.registers[reg] = uint32(content) // Overflow is a feature.
	}

	return nil
}

func (m *Mips) GetMemory(addr uint64) (uint8, error) {
	if addr > math.MaxUint32 {
		return 0, fmt.Errorf(machine.InterCtx.Get("value %v bigger than maximum 32 bit address %v"), addr, math.MaxUint32)
	}

	return m.mem[addr], nil
}

func (m *Mips) SetMemory(addr uint64, content uint8) error {
	if addr > math.MaxUint32 {
		return fmt.Errorf(machine.InterCtx.Get("value %v bigger than maximum 32 bit address %v"), addr, math.MaxUint32)
	}

	m.mem[addr] = content

	return nil
}

func (m *Mips) GetMemoryChunk(addr uint64, size uint64) ([]uint8, error) {
	end := addr + (size - 1)
	if end > math.MaxUint32 {
		return nil, fmt.Errorf(machine.InterCtx.Get("end address %v bigger than maximum 32 bit address %v"), end, math.MaxUint32)
	}

	return m.mem[addr:(end + 1)], nil
}

func (m *Mips) SetMemoryChunk(addr uint64, content []uint8) error {
	end := addr + (uint64(len(content)) - 1)
	if end > math.MaxUint32 {
		return fmt.Errorf(machine.InterCtx.Get("end address %v bigger than maximum 32 bit address %v"), end, math.MaxUint32)
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
		return nil, fmt.Errorf(machine.InterCtx.Get("could not load 4 bytes from address at PC: %x"), m.pc)
	}

	i := uint32(iarr[0]) | (uint32(iarr[1]) << 8) | (uint32(iarr[2]) << 16) | (uint32(iarr[3]) << 24)

	return m.execute(i)
}

func (m *Mips) GetCurrentInstructionAddress() uint64 {
	return uint64(m.pc)
}

//
// Most of the assembler is literally copied from the RISC-V.
//

func assembleSpecial(t assembler.ResolvedToken) (uint32, error) {
	var rs uint64
	var rt uint64
	var rd uint64
	var shamt uint8
	var funct uint8

	switch string(t.Value) {
	case "jalr":
		funct = 9
		if len(t.Args) == 1 {
			rs = t.Args[0]
			rd = 31
		} else if len(t.Args) == 2 {
			rs = t.Args[1]
			rd = t.Args[0]
		} else {
			return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected no argument"), t.Value)
		}
	case "jr":
		funct = 8
		if len(t.Args) != 1 {
			return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), t.Value)
		}
		rs = t.Args[0]
	case "mult":
		if len(t.Args) != 2 {
			return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
		}
		rs = t.Args[0]
		rt = t.Args[1]
		funct = 0x18
	case "div":
		if len(t.Args) != 2 {
			return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
		}
		rs = t.Args[0]
		rt = t.Args[1]
		funct = 0x1a
	case "mfhi":
		if len(t.Args) != 1 {
			return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
		}
		rd = t.Args[0]
		funct = 0x10
	case "mflo":
		if len(t.Args) != 1 {
			return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
		}
		rd = t.Args[0]
		funct = 0x12
	case "mthi":
		if len(t.Args) != 1 {
			return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
		}
		rs = t.Args[0]
		funct = 0x11
	case "mtlo":
		if len(t.Args) != 1 {
			return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
		}
		rs = t.Args[0]
		funct = 0x13
	default:
		if len(t.Args) != 3 {
			return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 3 arguments"), t.Value)
		}
		rd = t.Args[0]
		rs = t.Args[1]
		rt = t.Args[2]
		switch string(t.Value) {
		case "movz":
			funct = 0xa
		case "movn":
			funct = 0xb
		case "add":
			funct = 0x20
		case "addu":
			funct = 0x21
		case "sub":
			funct = 0x22
		case "subu":
			funct = 0x23
		case "slt":
			funct = 0x2a
		case "sltu":
			funct = 0x2b
		case "and":
			funct = 0x24
		case "or":
			funct = 0x25
		case "xor":
			funct = 0x26
		case "nor":
			funct = 0x27
		case "sllv":
			funct = 4
		case "srav":
			funct = 7
		case "srlv":
			funct = 6
		case "sll":
			funct = 0
			shamt = uint8(rt)
			rt = 0
		case "sra":
			funct = 3
			shamt = uint8(rt)
			rt = 0
		case "srl":
			funct = 2
			shamt = uint8(rt)
			rt = 0
		}
	}

	var code uint32
	code = code | (uint32(rs) << 21)
	code = code | (uint32(rt) << 16)
	code = code | (uint32(rd) << 11)
	code = code | (uint32(shamt) << 6)
	code = code | uint32(funct&0x3f)

	return code, nil
}

func assembleSpecial2(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 2 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
	}
	rd := t.Args[0]
	rs := t.Args[1]

	var funct uint8
	if string(t.Value) == "clz" {
		funct = 16
	} else {
		funct = 17
	}

	code := uint32(28 << 26)
	code = code | (uint32(rs) << 21)
	code = code | (uint32(rd) << 11)
	code = code | uint32(funct&0x3f)

	return code, nil
}

func assembleSpecial3(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 2 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
	}
	rd := t.Args[0]
	rt := t.Args[1]

	var shamt uint8
	if string(t.Value) == "seb" {
		shamt = 16
	} else {
		shamt = 24
	}

	code := uint32(31 << 26)
	code = code | (uint32(rt) << 16)
	code = code | (uint32(rd) << 11)
	code = code | (uint32(shamt) << 6)
	code = code | uint32(32)

	return code, nil
}

func assembleRegimm(t assembler.ResolvedToken, addr int) (uint32, error) {
	if len(t.Args) != 2 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
	}
	rs := t.Args[0]

	var funct uint8
	if string(t.Value) == "bltz" {
		funct = 0
	} else {
		funct = 1
	}

	br := (signExtend16(uint16(t.Args[1])) - uint64(addr)) >> 2
	code := uint32(1 << 26)
	code = code | (uint32(rs) << 21)
	code = code | (uint32(funct) << 16)
	code = code | uint32(br&0xffff)

	return code, nil
}

func assembleAddi(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
	}

	rt := t.Args[0] << 16
	rs := t.Args[1] << 21
	op := uint64(8 << 26)
	return uint32(op | rt | rs | t.Args[2]), nil
}

func assembleAddiu(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
	}

	rt := t.Args[0] << 16
	rs := t.Args[1] << 21
	op := uint64(9 << 26)
	return uint32(op | rt | rs | t.Args[2]), nil
}

func assembleAndi(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
	}

	rt := t.Args[0] << 16
	rs := t.Args[1] << 21
	op := uint64(12 << 26)
	return uint32(op | rt | rs | t.Args[2]), nil
}

func assembleOri(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
	}

	rt := t.Args[0] << 16
	rs := t.Args[1] << 21
	op := uint64(13 << 26)
	return uint32(op | rt | rs | t.Args[2]), nil
}

func assembleXori(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
	}

	rt := t.Args[0] << 16
	rs := t.Args[1] << 21
	op := uint64(14 << 26)
	return uint32(op | rt | rs | t.Args[2]), nil
}

func assembleSlti(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
	}

	rt := t.Args[0] << 16
	rs := t.Args[1] << 21
	op := uint64(10 << 26)
	return uint32(op | rt | rs | t.Args[2]), nil
}

func assembleSltiu(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
	}

	rt := t.Args[0] << 16
	rs := t.Args[1] << 21
	op := uint64(11 << 26)
	return uint32(op | rt | rs | t.Args[2]), nil
}

func assembleLui(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 2 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
	}

	rt := t.Args[0] << 16
	op := uint64(15 << 26)
	return uint32(op | rt | t.Args[1]), nil
}

func assembleBeq(t assembler.ResolvedToken, addr int) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
	}
	rs := t.Args[0]
	rt := t.Args[1]
	off := t.Args[2]

	br := (signExtend16(uint16(off)) - uint64(addr)) >> 2
	code := uint32(4 << 26)
	code = code | (uint32(rs) << 21)
	code = code | (uint32(rt) << 16)
	code = code | uint32(br&0xffff)

	return code, nil
}

func assembleBgtz(t assembler.ResolvedToken, addr int) (uint32, error) {
	if len(t.Args) != 2 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
	}
	rs := t.Args[0]
	off := t.Args[1]

	br := (signExtend16(uint16(off)) - uint64(addr)) >> 2
	code := uint32(7 << 26)
	code = code | (uint32(rs) << 21)
	code = code | uint32(br&0xffff)

	return code, nil
}

func assembleBlez(t assembler.ResolvedToken, addr int) (uint32, error) {
	if len(t.Args) != 2 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
	}
	rs := t.Args[0]
	off := t.Args[1]

	br := (signExtend16(uint16(off)) - uint64(addr)) >> 2
	code := uint32(6 << 26)
	code = code | (uint32(rs) << 21)
	code = code | uint32(br&0xffff)

	return code, nil
}

func assembleBne(t assembler.ResolvedToken, addr int) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), t.Value)
	}
	rs := t.Args[0]
	rt := t.Args[1]
	off := t.Args[2]

	br := (signExtend16(uint16(off)) - uint64(addr)) >> 2
	code := uint32(5 << 26)
	code = code | (uint32(rs) << 21)
	code = code | (uint32(rt) << 16)
	code = code | uint32(br&0xffff)

	return code, nil
}

func assembleJ(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 1 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), t.Value)
	}

	return uint32(2<<26) | uint32((t.Args[0]>>2)&0xfffffff), nil
}

func assembleJal(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 1 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), t.Value)
	}

	return uint32(3<<26) | uint32((t.Args[0]>>2)&0xfffffff), nil
}

func assembleLb(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), t.Value)
	}

	code := uint32(t.Args[2] & 0xffff)
	code = code | uint32(t.Args[0]<<16)
	code = code | uint32(t.Args[1]<<21)

	return code | (0x20 << 26), nil
}

func assembleLbu(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), t.Value)
	}

	code := uint32(t.Args[2] & 0xffff)
	code = code | uint32(t.Args[0]<<16)
	code = code | uint32(t.Args[1]<<21)

	return code | (0x24 << 26), nil
}

func assembleLh(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), t.Value)
	}

	code := uint32(t.Args[2] & 0xffff)
	code = code | uint32(t.Args[0]<<16)
	code = code | uint32(t.Args[1]<<21)

	return code | (0x21 << 26), nil
}

func assembleLhu(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), t.Value)
	}

	code := uint32(t.Args[2] & 0xffff)
	code = code | uint32(t.Args[0]<<16)
	code = code | uint32(t.Args[1]<<21)

	return code | (0x25 << 26), nil
}

func assembleLw(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), t.Value)
	}

	code := uint32(t.Args[2] & 0xffff)
	code = code | uint32(t.Args[0]<<16)
	code = code | uint32(t.Args[1]<<21)

	return code | (0x23 << 26), nil
}

func assembleLwl(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), t.Value)
	}

	code := uint32(t.Args[2] & 0xffff)
	code = code | uint32(t.Args[0]<<16)
	code = code | uint32(t.Args[1]<<21)

	return code | (0x22 << 26), nil
}

func assembleLwr(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), t.Value)
	}

	code := uint32(t.Args[2] & 0xffff)
	code = code | uint32(t.Args[0]<<16)
	code = code | uint32(t.Args[1]<<21)

	return code | (0x26 << 26), nil
}

func assembleSb(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), t.Value)
	}

	code := uint32(t.Args[2] & 0xffff)
	code = code | uint32(t.Args[0]<<16)
	code = code | uint32(t.Args[1]<<21)

	return code | (0x28 << 26), nil
}

func assembleSh(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), t.Value)
	}

	code := uint32(t.Args[2] & 0xffff)
	code = code | uint32(t.Args[0]<<16)
	code = code | uint32(t.Args[1]<<21)

	return code | (0x29 << 26), nil
}

func assembleSw(t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), t.Value)
	}

	code := uint32(t.Args[2] & 0xffff)
	code = code | uint32(t.Args[0]<<16)
	code = code | uint32(t.Args[1]<<21)

	return code | (0x2b << 26), nil
}

func assembleInstruction(code []uint8, addr int, t assembler.ResolvedToken) error {
	bin := uint32(0)
	var err error

	switch string(t.Value) {
	case "jalr", "jr", "add", "addu", "sub", "subu", "slt", "sltu", "and", "or", "xor", "nor", "sll", "sllv", "sra", "srav", "srl", "srlv", "mult", "div", "mfhi", "mflo", "mthi", "mtlo", "movz", "movn":
		bin, err = assembleSpecial(t)
	case "clz", "clo":
		bin, err = assembleSpecial2(t)
	case "seb", "seh":
		bin, err = assembleSpecial3(t)
	case "bltz", "bgez":
		bin, err = assembleRegimm(t, addr)
	case "addi":
		bin, err = assembleAddi(t)
	case "addiu":
		bin, err = assembleAddiu(t)
	case "lui":
		bin, err = assembleLui(t)
	case "andi":
		bin, err = assembleAndi(t)
	case "ori":
		bin, err = assembleOri(t)
	case "xori":
		bin, err = assembleXori(t)
	case "slti":
		bin, err = assembleSlti(t)
	case "sltiu":
		bin, err = assembleSltiu(t)
	case "beq":
		bin, err = assembleBeq(t, addr)
	case "bgtz":
		bin, err = assembleBgtz(t, addr)
	case "blez":
		bin, err = assembleBlez(t, addr)
	case "bne":
		bin, err = assembleBne(t, addr)
	case "break":
		bin = 13
	case "syscall":
		bin = 12
	case "j":
		bin, err = assembleJ(t)
	case "jal":
		bin, err = assembleJal(t)
	case "lb":
		bin, err = assembleLb(t)
	case "lbu":
		bin, err = assembleLbu(t)
	case "lh":
		bin, err = assembleLh(t)
	case "lhu":
		bin, err = assembleLhu(t)
	case "lw":
		bin, err = assembleLw(t)
	case "lwl":
		bin, err = assembleLwl(t)
	case "lwr":
		bin, err = assembleLwr(t)
	case "sb":
		bin, err = assembleSb(t)
	case "sh":
		bin, err = assembleSh(t)
	case "sw":
		bin, err = assembleSw(t)
	default:
		return fmt.Errorf(machine.InterCtx.Get("unknown instruction: %v"), string(t.Value))
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
		if i.Type == assembler.TOKEN_INSTRUCTION {
			size += 4
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
	switch arg {
	case "zero":
		return 0, nil
	case "ap":
		return 1, nil
	case "gp":
		return 28, nil
	case "sp":
		return 29, nil
	case "fp", "s8":
		return 30, nil
	case "ra":
		return 31, nil
	case "k0":
		return 26, nil
	case "k1":
		return 27, nil
	case "v0":
		return 2, nil
	case "v1":
		return 3, nil
	}

	n, err := strconv.Atoi(arg[1:])
	if err != nil {
		return 0, fmt.Errorf(machine.InterCtx.Get("no such register: %v"), arg)
	}

	switch arg[0] {
	case 't':
		if n < 8 {
			return uint64(n + 8), nil
		}
		return uint64(n + 16), nil
	case 's':
		return uint64(n + 16), nil
	case 'a':
		return uint64(n + 4), nil
	case 'x':
		return uint64(n), nil
	}

	return 0, fmt.Errorf(machine.InterCtx.Get("no such register: %v"), arg)
}

func translateArgs(arg string) (uint64, error) {
	if len(arg) < 1 {
		return 0, errors.New(machine.InterCtx.Get("empty argument"))
	}

	if (0x30 <= arg[0] && arg[0] <= 0x39) || arg[0] == '-' {
		n, err := strconv.ParseInt(arg, 0, 64)
		return uint64(n), err
	}

	return parseRegisterArg(arg)
}

func (m *Mips) GetRegisterNumber(r string) (uint64, error) {
	if len(r) < 2 {
		return 0, fmt.Errorf(machine.InterCtx.Get("no such register: %v"), r)
	}
	reg, err := parseRegisterArg(r)
	if err != nil || reg >= 32 {
		return 0, fmt.Errorf(machine.InterCtx.Get("no such register: %v"), r)
	}

	return reg, nil
}

func (m *Mips) Assemble(file string) ([]uint8, []assembler.DebuggerToken, error) {
	tokens := []assembler.Token{}
	err := assembler.Tokenize(file, &tokens)
	if err != nil {
		return nil, nil, err
	}

	resolvedTokens, debuggerTokens, err := assembler.ResolveTokens(tokens, func(i *assembler.Instruction) error {
		i.Size = 4
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

func (m *Mips) ArchitectureInfo() machine.ArchitectureInfo {
	var info machine.ArchitectureInfo
	info.Name = "MIPS32"
	info.WordWidth = 32
	info.RegistersNames = []string{
		"zero",
		"at",
		"v0",
		"v1",
		"a0",
		"a1",
		"a2",
		"a3",
		"t0",
		"t1",
		"t2",
		"t3",
		"t4",
		"t5",
		"t6",
		"t7",
		"s0",
		"s1",
		"s2",
		"s3",
		"s4",
		"s5",
		"s6",
		"s7",
		"t8",
		"t9",
		"k0",
		"k1",
		"gp",
		"sp",
		"s8",
		"ra",
	}

	return info
}
