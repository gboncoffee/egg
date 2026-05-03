// Package egg/pia implements a PIÁ machine for the EGG emulator.
package pia

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gboncoffee/egg/assembler"
	"github.com/gboncoffee/egg/machine"
)

// The Pia struct implements the Machine interface for the PIÁ architecture.
type Pia struct {
	registers [16]uint32
	pc        uint32
	mem       [math.MaxUint32 + 1]uint8
}

// Helper function to sign-extend a value from n bits to 32 bits
func signExtend(val uint32, bits uint8) uint32 {
	if val&(1<<(bits-1)) != 0 {
		val |= ^((1 << bits) - 1)
	}
	return val
}

// Helper function to read a 32-bit word from memory (little-endian)
func (m *Pia) readWord(addr uint32) (uint32, error) {
	if addr > math.MaxUint32-3 {
		return 0, fmt.Errorf(machine.InterCtx.Get("value %v bigger than maximum 32 bit address %v"), addr)
	}
	return uint32(m.mem[addr]) |
		(uint32(m.mem[addr+1]) << 8) |
		(uint32(m.mem[addr+2]) << 16) |
		(uint32(m.mem[addr+3]) << 24), nil
}

// Helper function to write a 32-bit word to memory (little-endian)
func (m *Pia) writeWord(addr uint32, val uint32) error {
	if addr > math.MaxUint32-3 {
		return fmt.Errorf(machine.InterCtx.Get("value %v bigger than maximum 32 bit address %v"), addr)
	}
	m.mem[addr] = uint8(val)
	m.mem[addr+1] = uint8(val >> 8)
	m.mem[addr+2] = uint8(val >> 16)
	m.mem[addr+3] = uint8(val >> 24)
	return nil
}

// Helper function to read a 16-bit halfword from memory (little-endian)
func (m *Pia) readHalfword(addr uint32) (uint16, error) {
	if addr > math.MaxUint32-1 {
		return 0, fmt.Errorf(machine.InterCtx.Get("value %v bigger than maximum 32 bit address %v"), addr)
	}
	return uint16(m.mem[addr]) | (uint16(m.mem[addr+1]) << 8), nil
}

// Helper function to write a 16-bit halfword to memory (little-endian)
func (m *Pia) writeHalfword(addr uint32, val uint16) error {
	if addr > math.MaxUint32-1 {
		return fmt.Errorf(machine.InterCtx.Get("value %v bigger than maximum 32 bit address %v"), addr)
	}
	m.mem[addr] = uint8(val)
	m.mem[addr+1] = uint8(val >> 8)
	return nil
}

func (m *Pia) LoadProgram(program []uint8) error {
	m.pc = 0
	return m.SetMemoryChunk(0, program)
}

// Instruction type constants (based on opcode[3:0])
const (
	OpcodeC  = 0x0 // addsi (0000-0111 are C type)
	OpcodeR  = 0x8 // add, etc.
	OpcodeF  = 0x9 // fadd, etc.
	OpcodeLS = 0xA // lb, lw, etc.
	OpcodeS  = 0xF // dbgbrk, int, fwcall, syscall, etc.
	OpcodeI  = 0xB // addi, etc. (32-bit)
	OpcodeB  = 0xC // beq, etc. (32-bit)
	OpcodeJ  = 0xD // lj, ljl (32-bit)
	OpcodeJ2 = 0xE // lrj, lrjl (32-bit)
)

// Parse 16-bit instruction format C (compact)
// Format: imm[7:0] | rd[3:0] | opcode[3:0]
func parseC16(instr uint16) (uint8, uint8, uint8) {
	rd := uint8((instr >> 4) & 0xF)
	imm := uint8(instr >> 8)
	opcode := uint8(instr & 0xF)
	return rd, imm, opcode
}

// Parse 16-bit instruction format R/F/LS
// Format: rs[3:0] | func[3:0] | rd[3:0] | opcode[3:0]
func parseR16(instr uint16) (uint8, uint8, uint8, uint8) {
	rd := uint8((instr >> 4) & 0xF)
	func_ := uint8((instr >> 8) & 0xF)
	rs := uint8((instr >> 12) & 0xF)
	opcode := uint8(instr & 0xF)
	return rs, func_, rd, opcode
}

// Parse 16-bit instruction format S (special)
// Format: imm[7:0] | func[3:0] | opcode[3:0]
func parseS16(instr uint16) (uint8, uint8, uint8) {
	opcode := uint8(instr & 0xF)
	func_ := uint8((instr >> 4) & 0xF)
	imm := uint8(instr >> 8)
	return imm, func_, opcode
}

// Parse 32-bit instruction format I (immediate)
// Format: imm[15:0] at [31:16] | imm[19:16] at [15:12] | func at [11:8] | rd at [7:4] | opcode at [3:0]
func parseI32(instr uint32) (uint8, uint32) {
	rd := uint8((instr >> 4) & 0xF)
	immLow := (instr >> 16) & 0xFFFF
	immHigh := (instr >> 12) & 0xF
	imm := immLow | (immHigh << 16)
	return rd, imm
}

// Parse 32-bit instruction format B (branch)
// Format: imm[15:1] | imm[16] | rs[3:0] | func[3:0] | rd[3:0] | opcode[3:0]
func parseB32(instr uint32) (uint8, uint8, uint32) {
	rd := uint8((instr >> 4) & 0xF)
	rs := uint8((instr >> 12) & 0xF)
	imm := ((instr >> 17) | (instr & 0x10000)) & 0x1FFFF
	return rd, rs, imm
}

// Parse 32-bit instruction format J (jump)
// Format: imm[15:1] | imm[16] | imm[27:17] | opcode[3:0]
func parseJ32(instr uint32) (uint32, uint8) {
	opcode := uint8(instr & 0xF)
	imm := (instr >> 4) & 0xFFF
	imm |= ((instr >> 17) & 0x1FFFF) << 1
	return imm, opcode
}

func (m *Pia) NextInstruction() (*machine.Call, error) {
	// Check if PC is aligned (must be even for instructions)
	if m.pc%2 != 0 {
		return nil, errors.New(machine.InterCtx.Get("misaligned instruction"))
	}

	// Try to read as 16-bit instruction first
	hw, err := m.readHalfword(m.pc)
	if err != nil {
		return nil, err
	}

	opcode := uint8(hw & 0xF)

	// Determine if this is a 16-bit or 32-bit instruction
	if opcode == 0xB || opcode == 0xC || opcode == 0xD || opcode == 0xE {
		// 32-bit instruction
		word, err := m.readWord(m.pc)
		if err != nil {
			return nil, err
		}
		return m.execute32(word)
	} else {
		// 16-bit instruction
		return m.execute16(hw)
	}
}

// Execute 16-bit instructions
func (m *Pia) execute16(instr uint16) (*machine.Call, error) {
	opcode := uint8(instr & 0xF)

	switch opcode {
	case 0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7: // Format C
		return m.executeC(instr)
	case 0x8, 0x9: // Format R/F
		return m.executeRF(instr)
	case 0xA: // Format LS
		return m.executeLS(instr)
	case 0xF: // Format S
		return m.executeS(instr)
	}
	return nil, fmt.Errorf(machine.InterCtx.Get("unknown opcode: %b"), opcode)
}

// Execute 32-bit instructions
func (m *Pia) execute32(instr uint32) (*machine.Call, error) {
	opcode := uint8(instr & 0xF)

	switch opcode {
	case 0xB: // Format I
		return m.executeI(instr)
	case 0xC: // Format B
		return m.executeB(instr)
	case 0xD, 0xE: // Format J
		return m.executeJ(instr)
	}
	return nil, fmt.Errorf(machine.InterCtx.Get("unknown opcode: %b"), opcode)
}

// Execute format C instructions (16-bit compact arithmetic)
func (m *Pia) executeC(instr uint16) (*machine.Call, error) {
	rd, imm, opcode := parseC16(instr)
	simm := int32(int8(imm))

	val := int32(m.registers[rd])

	switch opcode {
	case 0x0: // addsi
		val += simm
	case 0x1: // sllsi
		val = int32(uint32(val) << uint32(imm&0x1F))
	case 0x2: // srlsi
		val = int32(uint32(val) >> uint32(imm&0x1F))
	case 0x3: // andsi
		val = int32(uint32(val) & uint32(imm))
	case 0x4: // orsi
		val = int32(uint32(val) | uint32(imm))
	case 0x5: // xorsi
		val = int32(uint32(val) ^ uint32(imm))
	case 0x6: // not
		val = ^val
	case 0x7: // inv
		val = int32(^uint32(val))
	}

	m.registers[rd] = uint32(val)
	m.pc += 2
	return nil, nil
}

// Execute format R/F instructions (16-bit register/float)
func (m *Pia) executeRF(instr uint16) (*machine.Call, error) {
	rs, func_, rd, opcode := parseR16(instr)

	rsv := m.registers[rs]
	rdv := m.registers[rd]

	switch opcode {
	case 0x8:
		return m.executeR(rd, rs, rsv, rdv, func_)
	case 0x9:
		return m.executeF(rd, rs, rsv, rdv, func_)
	default:
		return nil, fmt.Errorf("invalid R/F opcode: %x", opcode)
	}
}

// Execute format R instructions (16-bit register arithmetic)
func (m *Pia) executeR(rd uint8, rs uint8, rsv uint32, rdv uint32, func_ uint8) (*machine.Call, error) {
	var result uint32

	switch func_ {
	case 0x0: // add
		result = rdv + rsv
	case 0x1: // sll
		result = rdv << uint32(rsv&0x1F)
	case 0x2: // srl
		result = rdv >> uint32(rsv&0x1F)
	case 0x3: // and
		result = rdv & rsv
	case 0x4: // or
		result = rdv | rsv
	case 0x5: // xor
		result = rdv ^ rsv
	case 0x6: // sub
		result = rdv - rsv
	case 0x7: // mul
		result = uint32(int32(rdv) * int32(rsv))
	case 0x8: // mulu
		result = uint32(rdv) * uint32(rsv)
	case 0x9: // div
		if rsv == 0 {
			result = 0
		} else {
			result = uint32(int32(rdv) / int32(rsv))
		}
	case 0xA: // divu
		if rsv == 0 {
			result = 0
		} else {
			result = rdv / rsv
		}
	case 0xB: // mod
		if rsv == 0 {
			result = 0
		} else {
			result = rdv % rsv
		}
	case 0xC: // mov
		result = rsv
	case 0xD: // i2f
		// Integer to float conversion (store as uint32 bitpattern)
		result = math.Float32bits(float32(rsv))
	case 0xE: // ui2f
		// Unsigned integer to float conversion
		result = math.Float32bits(float32(uint32(rsv)))
	case 0xF: // Reserved
		return nil, fmt.Errorf("reserved instruction")
	}

	m.registers[rd] = result
	m.pc += 2
	return nil, nil
}

// Execute format F instructions (16-bit floating-point)
func (m *Pia) executeF(rd uint8, rs uint8, rsv uint32, rdv uint32, func_ uint8) (*machine.Call, error) {
	// Convert register values to float32 for processing
	rsf := math.Float32frombits(rsv)
	rdf := math.Float32frombits(rdv)
	var result uint32

	switch func_ {
	case 0x0: // fadd
		result = math.Float32bits(rdf + rsf)
	case 0x1: // fabs
		if rsf < 0 {
			result = math.Float32bits(-rsf)
		} else {
			result = math.Float32bits(rsf)
		}
	case 0x2: // fsqrt
		result = math.Float32bits(float32(math.Sqrt(float64(rsf))))
	case 0x3: // fneg
		result = math.Float32bits(-rsf)
	case 0x6: // fsub
		result = math.Float32bits(rdf - rsf)
	case 0x7: // fmul
		result = math.Float32bits(rdf * rsf)
	case 0x9: // fdiv
		result = math.Float32bits(rdf / rsf)
	case 0xB: // fmin
		if rdf < rsf {
			result = math.Float32bits(rdf)
		} else {
			result = math.Float32bits(rsf)
		}
	case 0xC: // fmax
		if rdf > rsf {
			result = math.Float32bits(rdf)
		} else {
			result = math.Float32bits(rsf)
		}
	case 0xD: // f2i
		result = uint32(int32(rsf))
	case 0xE: // f2ui
		result = uint32(rsf)
	default:
		return nil, fmt.Errorf("reserved instruction")
	}

	m.registers[rd] = result
	m.pc += 2
	return nil, nil
}

// Execute format LS instructions (16-bit load/store)
func (m *Pia) executeLS(instr uint16) (*machine.Call, error) {
	rs, func_, rd, _ := parseR16(instr)
	rsv := m.registers[rs]

	switch func_ {
	case 0x0: // lb - Load byte
		val, err := m.GetMemory(uint64(rsv))
		if err != nil {
			return nil, err
		}
		m.registers[rd] = uint32(int32(int8(val)))
	case 0x1: // lh - Load halfword
		hw, err := m.readHalfword(rsv)
		if err != nil {
			return nil, err
		}
		m.registers[rd] = uint32(int32(int16(hw)))
	case 0x2: // lw - Load word
		w, err := m.readWord(rsv)
		if err != nil {
			return nil, err
		}
		m.registers[rd] = w
	case 0x4: // sb - Store byte
		err := m.SetMemory(uint64(rsv), uint8(m.registers[rd]))
		if err != nil {
			return nil, err
		}
	case 0x5: // sh - Store halfword
		err := m.writeHalfword(rsv, uint16(m.registers[rd]))
		if err != nil {
			return nil, err
		}
	case 0x6: // sw - Store word
		err := m.writeWord(rsv, m.registers[rd])
		if err != nil {
			return nil, err
		}
	case 0x8: // jr - Jump register
		m.pc = rsv
		return nil, nil
	case 0x9: // jlr - Jump and link register
		m.registers[0] = m.pc + 2 // ra = PC + 2
		m.pc = rsv
		return nil, nil
	case 0xF: // lear - Load Effective Address Register
		m.registers[rd] = rsv
	default:
		m.pc += 2
		return nil, nil
	}

	m.pc += 2
	return nil, nil
}

// Execute format S instructions (16-bit special)
func (m *Pia) executeS(instr uint16) (*machine.Call, error) {
	imm, func_, _ := parseS16(instr)

	switch func_ {
	case 0xF: // dbgbrk
		return &machine.Call{Number: machine.SYS_BREAK}, nil
	case 0x9: // int - Interrupt
		// Firmware not implemented yet
		m.pc += 2
		return nil, nil
	case 0xC: // fwcall - Firmware call
		// Firmware not implemented yet
		m.pc += 2
		return nil, nil
	case 0xA: // syscall - System call
		return &machine.Call{Number: uint64(imm)}, nil
	case 0x0: // sysrcall - System call (via ta)
		return &machine.Call{Number: uint64(m.registers[8])}, nil
	case 0x1: // movcr2ta - Move control register to ta
		// Control registers not implemented yet
		m.pc += 2
		return nil, nil
	case 0x2: // movta2cr - Move ta to control register
		// Control registers not implemented yet
		m.pc += 2
		return nil, nil
	case 0xD: // fwid - Get firmware ID
		m.registers[8] = 0xE99
		m.pc += 2
		return nil, nil
	}

	m.pc += 2
	return nil, nil
}

// Execute format I instructions (32-bit immediate)
func (m *Pia) executeI(instr uint32) (*machine.Call, error) {
	rd, imm := parseI32(instr)
	simm := int32(signExtend(imm, 20))
	rdv := int32(m.registers[rd])

	switch (instr >> 8) & 0xF {
	case 0x0: // addi
		m.registers[rd] = uint32(rdv + simm)
	case 0x1: // slli
		m.registers[rd] = uint32(uint32(rdv) << uint32(simm&0x1F))
	case 0x2: // srli
		m.registers[rd] = uint32(uint32(rdv) >> uint32(simm&0x1F))
	case 0x3: // andi
		m.registers[rd] = uint32(rdv) & imm
	case 0x4: // ori
		m.registers[rd] = uint32(rdv) | imm
	case 0x5: // xori
		m.registers[rd] = uint32(rdv) ^ imm
	case 0x6: // addiu
		m.registers[rd] = uint32(rdv) + imm
	case 0x7: // muli
		m.registers[rd] = uint32(rdv * simm)
	case 0x8: // mului
		m.registers[rd] = uint32(rdv) * imm
	case 0x9: // divi
		if simm == 0 {
			m.registers[rd] = 0
		} else {
			m.registers[rd] = uint32(rdv / simm)
		}
	case 0xA: // divui
		if imm == 0 {
			m.registers[rd] = 0
		} else {
			m.registers[rd] = uint32(rdv) / imm
		}
	case 0xB: // modi
		if simm == 0 {
			m.registers[rd] = 0
		} else {
			m.registers[rd] = uint32(rdv % simm)
		}
	case 0xC: // movi
		m.registers[rd] = uint32(simm)
	case 0xD: // lui - Load upper immediate
		m.registers[rd] = (uint32(m.registers[rd]) & 0xFFFF) | ((imm & 0xFFFF) << 16)
	case 0xE: // leai - Load effective address (PC + imm)
		m.registers[rd] = uint32(int32(m.pc) + simm)
	}

	m.pc += 4
	return nil, nil
}

// Execute format B instructions (32-bit branch)
func (m *Pia) executeB(instr uint32) (*machine.Call, error) {
	rd, rs, imm := parseB32(instr)
	rdv := int32(m.registers[rd])
	rsv := int32(m.registers[rs])
	simm := int32(signExtend(imm<<1, 17))
	func_ := (instr >> 8) & 0xF

	branch := false
	switch func_ {
	case 0x0: // beq
		branch = rdv == rsv
	case 0x1: // bne
		branch = rdv != rsv
	case 0x2: // bge
		branch = rdv >= rsv
	case 0x3: // blt
		branch = rdv < rsv
	case 0x4: // bgeu
		branch = uint32(rdv) >= uint32(rsv)
	case 0x5: // bltu
		branch = uint32(rdv) < uint32(rsv)
	case 0x6: // bz
		branch = rdv == 0
	case 0x7: // bnz
		branch = rdv != 0
	}

	if branch {
		m.pc = uint32(int32(m.pc) + simm)
	} else {
		m.pc += 4
	}
	return nil, nil
}

// Execute format J instructions (32-bit jump)
func (m *Pia) executeJ(instr uint32) (*machine.Call, error) {
	imm, opcode := parseJ32(instr)
	simm := int32(signExtend(imm, 28))

	switch opcode {
	case 0xD:
		switch (instr >> 16) & 1 {
		case 0: // lj - Long jump (absolute)
			m.pc = uint32(simm)
		case 1: // ljl - Long jump and link
			m.registers[0] = m.pc + 4
			m.pc = uint32(simm)
		}
	case 0xE:
		switch (instr >> 16) & 1 {
		case 0: // lrj - Long relative jump
			m.pc = uint32(int32(m.pc) + simm)
		case 1: // lrjl - Long relative jump and link
			m.registers[0] = m.pc + 4
			m.pc = uint32(int32(m.pc) + simm)
		}
	}
	return nil, nil
}

func (m *Pia) GetMemory(addr uint64) (uint8, error) {
	if addr > math.MaxUint32 {
		return 0, fmt.Errorf(machine.InterCtx.Get("value %v bigger than maximum 32 bit address %v"), addr, math.MaxUint32)
	}
	return m.mem[addr], nil
}

func (m *Pia) SetMemory(addr uint64, content uint8) error {
	if addr > math.MaxUint32 {
		return fmt.Errorf(machine.InterCtx.Get("value %v bigger than maximum 32 bit address %v"), addr, math.MaxUint32)
	}
	m.mem[addr] = content
	return nil
}

func (m *Pia) GetMemoryChunk(addr uint64, size uint64) ([]uint8, error) {
	end := addr + (size - 1)
	if end > math.MaxUint32 {
		return nil, fmt.Errorf(machine.InterCtx.Get("value %v bigger than maximum 32 bit address %v"), addr, math.MaxUint32)
	}
	return m.mem[addr : end+1], nil
}

func (m *Pia) SetMemoryChunk(addr uint64, content []uint8) error {
	end := addr + uint64(len(content)) - 1
	if end > math.MaxUint32 {
		return fmt.Errorf(machine.InterCtx.Get("value %v bigger than maximum 32 bit address %v"), addr, math.MaxUint32)
	}
	copy(m.mem[addr:end+1], content)
	return nil
}

func (m *Pia) GetRegister(reg uint64) (uint64, error) {
	if reg >= 16 {
		return 0, fmt.Errorf(machine.InterCtx.Get("no such register: %d. PIÁ has only 16 registers"), reg)
	}
	return uint64(m.registers[reg]), nil
}

func (m *Pia) SetRegister(reg uint64, content uint64) error {
	if reg >= 16 {
		return fmt.Errorf(machine.InterCtx.Get("no such register: %d. PIÁ has only 16 registers"), reg)
	}
	m.registers[reg] = uint32(content)
	return nil
}

// parseRegisterArg translates a register name to its number
func (m *Pia) parseRegisterArg(r string) (uint8, error) {
	r = strings.ToLower(r)
	registerMap := map[string]uint8{
		"ra": 0, "sp": 1, "sa": 2, "sb": 3, "sc": 4, "sd": 5, "se": 6, "sf": 7,
		"ta": 8, "tb": 9, "tc": 10, "td": 11, "te": 12, "tf": 13, "tg": 14, "th": 15,
	}
	if val, ok := registerMap[r]; ok {
		return val, nil
	}
	// Try parsing as numeric register
	num, err := strconv.Atoi(r)
	if err == nil && num >= 0 && num < 16 {
		return uint8(num), nil
	}
	return 0, errors.New("invalid register name")
}

func (m *Pia) GetRegisterNumber(r string) (uint64, error) {
	if len(r) == 0 {
		return 0, fmt.Errorf(machine.InterCtx.Get("no such register: %v"), r)
	}
	reg, err := m.parseRegisterArg(r)
	if err != nil || reg >= 16 {
		return 0, fmt.Errorf(machine.InterCtx.Get("no such register: %v"), r)
	}
	return uint64(reg), nil
}

// Helper function to encode 16-bit C format instruction
// Format: imm[7:0] | rd[3:0] | opcode[3:0]
func encodeC16(rd uint8, imm uint8, opcode uint8) uint16 {
	return (uint16(imm) << 8) | (uint16(rd) << 4) | uint16(opcode)
}

// Helper function to encode 16-bit R format instruction
// Format: rs[3:0] | func[3:0] | rd[3:0] | opcode[3:0]
func encodeR16(rs uint8, func_ uint8, rd uint8, opcode uint8) uint16 {
	return (uint16(rs) << 12) | (uint16(func_) << 8) | (uint16(rd) << 4) | uint16(opcode)
}

// Helper function to encode 16-bit S format instruction
// Format: imm[7:0] | func[3:0] | opcode[3:0]
func encodeS16(imm uint8, func_ uint8, opcode uint8) uint16 {
	return (uint16(imm) << 8) | (uint16(func_) << 4) | uint16(opcode)
}

// Helper function to encode 32-bit I format instruction
// Format: imm[15:0] at [31:16] | imm[19:16] at [15:12] | func at [11:8] | rd at [7:4] | opcode at [3:0]
func encodeI32(rd uint8, func_ uint8, imm uint32) uint32 {
	immLow := imm & 0xFFFF
	immHigh := (imm >> 16) & 0xF
	return (immLow << 16) | (immHigh << 12) | (uint32(func_) << 8) | (uint32(rd) << 4) | 0xB
}

// Helper function to encode 32-bit B format instruction
// Format: imm[15:1] | imm[16] | rs[3:0] | func[3:0] | rd[3:0] | opcode[3:0]
func encodeB32(rd uint8, rs uint8, func_ uint8, imm uint32) uint32 {
	immLow := (imm >> 1) & 0xFFFF
	immBit16 := (imm >> 16) & 1
	return (immLow << 17) | (immBit16 << 16) | (uint32(rs) << 12) | (uint32(func_) << 8) | (uint32(rd) << 4) | 0xC
}

// Helper function to encode 32-bit J format instruction
// Format: imm[15:1] | imm[16] | imm[27:17] | opcode[3:0]
func encodeJ32(imm uint32, bit16 uint8, opcode uint8) uint32 {
	immLow := (imm >> 1) & 0xFFFF
	immMid := (imm >> 17) & 0x7FF
	return (immLow << 17) | (uint32(bit16) << 16) | (immMid << 4) | uint32(opcode)
}

// Assemble 16-bit C format instructions
func assembleC16Instruction(mnemonic string, t assembler.ResolvedToken) (uint16, error) {
	if len(t.Args) != 2 {
		return 0, fmt.Errorf("instruction %s expects 2 arguments, got %d", mnemonic, len(t.Args))
	}

	rd := uint8(t.Args[0] & 0xF)
	imm := uint8(t.Args[1] & 0xFF)

	var opcode uint8
	switch mnemonic {
	case "addsi":
		opcode = 0x0
	case "sllsi":
		opcode = 0x1
	case "srlsi":
		opcode = 0x2
	case "andsi":
		opcode = 0x3
	case "orsi":
		opcode = 0x4
	case "xorsi":
		opcode = 0x5
	default:
		return 0, fmt.Errorf("unknown C16 instruction: %s", mnemonic)
	}

	return encodeC16(rd, imm, opcode), nil
}

// Assemble 16-bit C format instructions with no immediate
func assembleC16NoImmInstruction(mnemonic string, t assembler.ResolvedToken) (uint16, error) {
	if len(t.Args) != 1 {
		return 0, fmt.Errorf("instruction %s expects 1 argument, got %d", mnemonic, len(t.Args))
	}

	rd := uint8(t.Args[0] & 0xF)

	var opcode uint8
	switch mnemonic {
	case "not":
		opcode = 0x6
	case "inv":
		opcode = 0x7
	default:
		return 0, fmt.Errorf("unknown C16 no-imm instruction: %s", mnemonic)
	}

	return encodeC16(rd, 0, opcode), nil
}

// Assemble 16-bit R/F format instructions
func assembleR16Instruction(mnemonic string, t assembler.ResolvedToken) (uint16, error) {
	// Special case for jr - single argument
	if mnemonic == "jr" {
		if len(t.Args) != 1 {
			return 0, fmt.Errorf("instruction %s expects 1 argument, got %d", mnemonic, len(t.Args))
		}
		rs := uint8(t.Args[0] & 0xF)
		return encodeR16(rs, 0x8, 0x0, 0xA), nil
	}

	if len(t.Args) != 2 {
		return 0, fmt.Errorf("instruction %s expects 2 arguments, got %d", mnemonic, len(t.Args))
	}

	rd := uint8(t.Args[0] & 0xF)
	rs := uint8(t.Args[1] & 0xF)

	var func_ uint8
	var opcode uint8

	switch mnemonic {
	case "add":
		opcode, func_ = 0x8, 0x0
	case "sll":
		opcode, func_ = 0x8, 0x1
	case "srl":
		opcode, func_ = 0x8, 0x2
	case "and":
		opcode, func_ = 0x8, 0x3
	case "or":
		opcode, func_ = 0x8, 0x4
	case "xor":
		opcode, func_ = 0x8, 0x5
	case "sub":
		opcode, func_ = 0x8, 0x6
	case "mul":
		opcode, func_ = 0x8, 0x7
	case "mulu":
		opcode, func_ = 0x8, 0x8
	case "div":
		opcode, func_ = 0x8, 0x9
	case "divu":
		opcode, func_ = 0x8, 0xA
	case "mod":
		opcode, func_ = 0x8, 0xB
	case "mov":
		opcode, func_ = 0x8, 0xC
	case "i2f":
		opcode, func_ = 0x8, 0xD
	case "ui2f":
		opcode, func_ = 0x8, 0xE
	case "fadd":
		opcode, func_ = 0x9, 0x0
	case "fabs":
		opcode, func_ = 0x9, 0x1
	case "fsqrt":
		opcode, func_ = 0x9, 0x2
	case "fneg":
		opcode, func_ = 0x9, 0x3
	case "fsub":
		opcode, func_ = 0x9, 0x6
	case "fmul":
		opcode, func_ = 0x9, 0x7
	case "fdiv":
		opcode, func_ = 0x9, 0x9
	case "fmin":
		opcode, func_ = 0x9, 0xB
	case "fmax":
		opcode, func_ = 0x9, 0xC
	case "f2i":
		opcode, func_ = 0x9, 0xD
	case "f2ui":
		opcode, func_ = 0x9, 0xE
	case "lb":
		opcode, func_ = 0xA, 0x0
	case "lh":
		opcode, func_ = 0xA, 0x1
	case "lw":
		opcode, func_ = 0xA, 0x2
	case "sb":
		opcode, func_ = 0xA, 0x4
	case "sh":
		opcode, func_ = 0xA, 0x5
	case "sw":
		opcode, func_ = 0xA, 0x6
	case "jlr":
		opcode, func_ = 0xA, 0x9
	case "lear":
		opcode, func_ = 0xA, 0xF
	default:
		return 0, fmt.Errorf("unknown R16 instruction: %s", mnemonic)
	}

	return encodeR16(rs, func_, rd, opcode), nil
}

// Assemble 16-bit S format instructions
func assembleS16Instruction(mnemonic string, t assembler.ResolvedToken) (uint16, error) {
	var func_ uint8
	var imm uint8

	switch mnemonic {
	case "dbgbrk":
		func_, imm = 0xF, 0xFF
	case "int":
		if len(t.Args) != 1 {
			return 0, fmt.Errorf("instruction %s expects 1 argument, got %d", mnemonic, len(t.Args))
		}
		func_, imm = 0x9, uint8(t.Args[0]&0xFF)
	case "fwcall":
		if len(t.Args) != 1 {
			return 0, fmt.Errorf("instruction %s expects 1 argument, got %d", mnemonic, len(t.Args))
		}
		func_, imm = 0xC, uint8(t.Args[0]&0xFF)
	case "syscall":
		if len(t.Args) != 1 {
			return 0, fmt.Errorf("instruction %s expects 1 argument, got %d", mnemonic, len(t.Args))
		}
		func_, imm = 0xA, uint8(t.Args[0]&0xFF)
	case "sysrcall":
		func_, imm = 0x0, 0x0F
	case "movcr2ta":
		if len(t.Args) != 1 {
			return 0, fmt.Errorf("instruction %s expects 1 argument, got %d", mnemonic, len(t.Args))
		}
		func_, imm = 0x1, uint8(t.Args[0]&0xFF)
	case "movta2cr":
		if len(t.Args) != 1 {
			return 0, fmt.Errorf("instruction %s expects 1 argument, got %d", mnemonic, len(t.Args))
		}
		func_, imm = 0x2, uint8(t.Args[0]&0xFF)
	case "fwid":
		func_, imm = 0xD, 0xDD
	default:
		return 0, fmt.Errorf("unknown S16 instruction: %s", mnemonic)
	}

	return encodeS16(imm, func_, 0xF), nil
}

// Assemble 32-bit I format instructions
func assembleI32Instruction(mnemonic string, t assembler.ResolvedToken) (uint32, error) {
	if len(t.Args) != 2 {
		return 0, fmt.Errorf("instruction %s expects 2 arguments, got %d", mnemonic, len(t.Args))
	}

	rd := uint8(t.Args[0] & 0xF)
	imm := uint32(int32(t.Args[1]))

	var func_ uint8
	switch mnemonic {
	case "addi":
		func_ = 0x0
	case "slli":
		func_ = 0x1
	case "srli":
		func_ = 0x2
	case "andi":
		func_ = 0x3
	case "ori":
		func_ = 0x4
	case "xori":
		func_ = 0x5
	case "addiu":
		func_ = 0x6
	case "muli":
		func_ = 0x7
	case "mului":
		func_ = 0x8
	case "divi":
		func_ = 0x9
	case "divui":
		func_ = 0xA
	case "modi":
		func_ = 0xB
	case "movi":
		func_ = 0xC
	case "lui":
		func_ = 0xD
	case "leai":
		func_ = 0xE
	default:
		return 0, fmt.Errorf("unknown I32 instruction: %s", mnemonic)
	}

	return encodeI32(rd, func_, imm), nil
}

// Assemble 32-bit B format instructions (requires address for branch calculation)
func assembleB32Instruction(mnemonic string, t assembler.ResolvedToken, addr uint64) (uint32, error) {
	// Special case for bz and bnz - single register argument
	if mnemonic == "bz" || mnemonic == "bnz" {
		if len(t.Args) != 2 {
			return 0, fmt.Errorf("instruction %s expects 2 arguments, got %d", mnemonic, len(t.Args))
		}

		rd := uint8(t.Args[0] & 0xF)
		target := int32(t.Args[1])

		offset := target - int32(addr) - 4
		offset = offset >> 1

		imm := uint32(offset) & 0x1FFFF

		var func_ uint8
		if mnemonic == "bz" {
			func_ = 0x6
		} else {
			func_ = 0x7
		}

		return encodeB32(rd, 0, func_, imm), nil
	}

	if len(t.Args) != 3 {
		return 0, fmt.Errorf("instruction %s expects 3 arguments, got %d", mnemonic, len(t.Args))
	}

	rd := uint8(t.Args[0] & 0xF)
	rs := uint8(t.Args[1] & 0xF)
	target := int32(t.Args[2])

	imm := uint32(target - int32(addr))

	var func_ uint8
	switch mnemonic {
	case "beq":
		func_ = 0x0
	case "bne":
		func_ = 0x1
	case "bge":
		func_ = 0x2
	case "blt":
		func_ = 0x3
	case "bgeu":
		func_ = 0x4
	case "bltu":
		func_ = 0x5
	default:
		return 0, fmt.Errorf("unknown B32 instruction: %s", mnemonic)
	}

	return encodeB32(rd, rs, func_, imm), nil
}

// Assemble 32-bit J format instructions (requires address for jump calculation)
func assembleJ32Instruction(mnemonic string, t assembler.ResolvedToken, addr uint64) (uint32, error) {
	if len(t.Args) != 1 {
		return 0, fmt.Errorf("instruction %s expects 1 argument, got %d", mnemonic, len(t.Args))
	}

	target := int32(t.Args[0])

	var bit16 uint8

	switch mnemonic {
	case "lj":
		imm := uint32(target)
		bit16 = 0
		return encodeJ32(imm, bit16, 0xD), nil
	case "ljl":
		imm := uint32(target)
		bit16 = 1
		return encodeJ32(imm, bit16, 0xD), nil
	case "lrj":
		offset := target - int32(addr) - 4
		offset = offset >> 1
		imm := uint32(offset) & 0xFFFFFFF
		bit16 = 0
		return encodeJ32(imm, bit16, 0xE), nil
	case "lrjl":
		offset := target - int32(addr) - 4
		offset = offset >> 1
		imm := uint32(offset) & 0xFFFFFFF
		bit16 = 1
		return encodeJ32(imm, bit16, 0xE), nil
	default:
		return 0, fmt.Errorf("unknown J32 instruction: %s", mnemonic)
	}
}

// Process function for ResolveTokens - determines instruction size
func processPiaInstruction(i *assembler.Instruction) error {
	mnemonic := strings.ToLower(i.Mnemonic)

	switch mnemonic {
	case "addsi", "sllsi", "srlsi", "andsi", "orsi", "xorsi", "not", "inv",
		"add", "sll", "srl", "and", "or", "xor", "sub", "mul", "mulu", "div", "divu", "mod", "mov", "i2f", "ui2f",
		"fadd", "fabs", "fsqrt", "fneg", "fsub", "fmul", "fdiv", "fmin", "fmax", "f2i", "f2ui",
		"lb", "lh", "lw", "sb", "sh", "sw", "jr", "jlr", "lear",
		"dbgbrk", "int", "fwcall", "syscall", "sysrcall", "movcr2ta", "movta2cr", "fwid":
		i.Size = 2
	case "addi", "slli", "srli", "andi", "ori", "xori", "addiu", "muli", "mului", "divi", "divui", "modi", "movi", "lui", "leai",
		"beq", "bne", "bge", "blt", "bgeu", "bltu", "bz", "bnz",
		"lj", "ljl", "lrj", "lrjl":
		i.Size = 4
	default:
		return fmt.Errorf("unknown instruction: %s", mnemonic)
	}

	return nil
}

// Assemble one instruction and write to code buffer
func assemblePiaInstruction(code []uint8, addr int, t assembler.ResolvedToken) error {
	mnemonic := strings.ToLower(string(t.Value))
	var bin uint32
	var err error
	var size int

	switch mnemonic {
	case "addsi", "sllsi", "srlsi", "andsi", "orsi", "xorsi":
		bin16, err := assembleC16Instruction(mnemonic, t)
		if err != nil {
			return err
		}
		bin = uint32(bin16)
		size = 2
	case "not", "inv":
		bin16, err := assembleC16NoImmInstruction(mnemonic, t)
		if err != nil {
			return err
		}
		bin = uint32(bin16)
		size = 2
	case "add", "sll", "srl", "and", "or", "xor", "sub", "mul", "mulu", "div", "divu", "mod", "mov", "i2f", "ui2f",
		"fadd", "fabs", "fsqrt", "fneg", "fsub", "fmul", "fdiv", "fmin", "fmax", "f2i", "f2ui",
		"lb", "lh", "lw", "sb", "sh", "sw", "jr", "jlr", "lear":
		bin16, err := assembleR16Instruction(mnemonic, t)
		if err != nil {
			return err
		}
		bin = uint32(bin16)
		size = 2
	case "dbgbrk", "int", "fwcall", "syscall", "sysrcall", "movcr2ta", "movta2cr", "fwid":
		bin16, err := assembleS16Instruction(mnemonic, t)
		if err != nil {
			return err
		}
		bin = uint32(bin16)
		size = 2
	case "addi", "slli", "srli", "andi", "ori", "xori", "addiu", "muli", "mului", "divi", "divui", "modi", "movi", "lui", "leai":
		bin, err = assembleI32Instruction(mnemonic, t)
		if err != nil {
			return err
		}
		size = 4
	case "beq", "bne", "bge", "blt", "bgeu", "bltu", "bz", "bnz":
		bin, err = assembleB32Instruction(mnemonic, t, t.Address)
		if err != nil {
			return err
		}
		size = 4
	case "lj", "ljl", "lrj", "lrjl":
		bin, err = assembleJ32Instruction(mnemonic, t, t.Address)
		if err != nil {
			return err
		}
		size = 4
	default:
		return fmt.Errorf("unknown instruction: %v", mnemonic)
	}

	if size == 2 {
		code[addr] = uint8(bin & 0xff)
		code[addr+1] = uint8((bin & 0xff00) >> 8)
	} else {
		code[addr] = uint8(bin & 0xff)
		code[addr+1] = uint8((bin & 0xff00) >> 8)
		code[addr+2] = uint8((bin & 0xff0000) >> 16)
		code[addr+3] = uint8((bin & 0xff000000) >> 24)
	}

	return nil
}

// Assemble all tokens
func assemblePia(t []assembler.ResolvedToken) ([]uint8, error) {
	size := uint64(0)
	for _, i := range t {
		if i.Type == assembler.TOKEN_INSTRUCTION {
			// Size must be set during token resolution
			if i.Value != nil {
				mnemonic := strings.ToLower(string(i.Value))
				switch mnemonic {
				case "addsi", "sllsi", "srlsi", "andsi", "orsi", "xorsi", "not", "inv",
					"add", "sll", "srl", "and", "or", "xor", "sub", "mul", "mulu", "div", "divu", "mod", "mov", "i2f", "ui2f",
					"fadd", "fabs", "fsqrt", "fneg", "fsub", "fmul", "fdiv", "fmin", "fmax", "f2i", "f2ui",
					"lb", "lh", "lw", "sb", "sh", "sw", "jr", "jlr", "lear",
					"dbgbrk", "int", "fwcall", "syscall", "sysrcall", "movcr2ta", "movta2cr", "fwid":
					size += 2
				case "addi", "slli", "srli", "andi", "ori", "xori", "addiu", "muli", "mului", "divi", "divui", "modi", "movi", "lui", "leai",
					"beq", "bne", "bge", "blt", "bgeu", "bltu", "bz", "bnz",
					"lj", "ljl", "lrj", "lrjl":
					size += 4
				}
			}
		} else {
			size += uint64(len(i.Value))
		}
	}

	code := make([]uint8, size)
	addr := 0
	for _, i := range t {
		if i.Type == assembler.TOKEN_INSTRUCTION {
			err := assemblePiaInstruction(code, addr, i)
			if err != nil {
				return nil, err
			}

			mnemonic := strings.ToLower(string(i.Value))
			switch mnemonic {
			case "addsi", "sllsi", "srlsi", "andsi", "orsi", "xorsi", "not", "inv",
				"add", "sll", "srl", "and", "or", "xor", "sub", "mul", "mulu", "div", "divu", "mod", "mov", "i2f", "ui2f",
				"fadd", "fabs", "fsqrt", "fneg", "fsub", "fmul", "fdiv", "fmin", "fmax", "f2i", "f2ui",
				"lb", "lh", "lw", "sb", "sh", "sw", "jr", "jlr", "lear",
				"dbgbrk", "int", "fwcall", "syscall", "sysrcall", "movcr2ta", "movta2cr", "fwid":
				addr += 2
			case "addi", "slli", "srli", "andi", "ori", "xori", "addiu", "muli", "mului", "divi", "divui", "modi", "movi", "lui", "leai",
				"beq", "bne", "bge", "blt", "bgeu", "bltu", "bz", "bnz",
				"lj", "ljl", "lrj", "lrjl":
				addr += 4
			}
		} else {
			for _, c := range []uint8(i.Value) {
				code[addr] = c
				addr++
			}
		}
	}

	return code, nil
}

// Translate argument - can be a register name or a number
func translatePiaArgs(arg string) (uint64, error) {
	if len(arg) < 1 {
		return 0, errors.New(machine.InterCtx.Get("empty argument"))
	}

	if (0x30 <= arg[0] && arg[0] <= 0x39) || arg[0] == '-' {
		n, err := strconv.ParseInt(arg, 0, 64)
		return uint64(n), err
	}

	arg = strings.ToLower(arg)
	registerMap := map[string]uint8{
		"ra": 0, "sp": 1, "sa": 2, "sb": 3, "sc": 4, "sd": 5, "se": 6, "sf": 7,
		"ta": 8, "tb": 9, "tc": 10, "td": 11, "te": 12, "tf": 13, "tg": 14, "th": 15,
	}
	if val, ok := registerMap[arg]; ok {
		return uint64(val), nil
	}

	// Try parsing as numeric register
	num, err := strconv.Atoi(arg)
	if err == nil && num >= 0 && num < 16 {
		return uint64(num), nil
	}

	return 0, fmt.Errorf(machine.InterCtx.Get("no such register: %v"), arg)
}

func (m *Pia) Assemble(file string) ([]uint8, []assembler.DebuggerToken, error) {
	tokens := []assembler.Token{}
	err := assembler.Tokenize(file, &tokens)
	if err != nil {
		return nil, nil, err
	}

	resolvedTokens, debuggerTokens, err := assembler.ResolveTokens(tokens, processPiaInstruction, translatePiaArgs)
	if err != nil {
		return nil, nil, err
	}

	code, err := assemblePia(resolvedTokens)
	if err != nil {
		return nil, nil, err
	}

	return code, debuggerTokens, nil
}

func (m *Pia) GetCurrentInstructionAddress() uint64 {
	return uint64(m.pc)
}

func (m *Pia) ArchitectureInfo() machine.ArchitectureInfo {
	return machine.ArchitectureInfo{
		Name:      "PIÁ - Processador de Informação Avançado",
		WordWidth: 32,
		RegistersNames: []string{
			"ra", "sp", "sa", "sb", "sc", "sd", "se", "sf",
			"ta", "tb", "tc", "td", "te", "tf", "tg", "th",
		},
	}
}
