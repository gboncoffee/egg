package riscv

import (
	"testing"
	"math"
	"reflect"
	_ "embed"
)

func TestRegister(t *testing.T) {
	var m RiscV
	m.SetRegister(1, 39)
	v, err := m.GetRegister(1)
	if v != 39 || err != nil {
		t.Fatalf("Error getting register 1: %d, %v", v, err)
	}
	m.SetRegister(0, 39)
	v, err = m.GetRegister(0)
	if v != 0 || err != nil {
		t.Fatalf("Error getting register 0: %d, %v", v, err)
	}
	m.SetRegister(31, 42)
	v, err = m.GetRegister(31)
	if v != 42 || err != nil {
		t.Fatalf("Error getting register 31: %d, %v", v, err)
	}
	v, err = m.GetRegister(32)
	if err == nil {
		t.Fatalf("No error getting register 32")
	}
}

func TestMem(t *testing.T) {
	var m RiscV

	err := m.SetMemory(0, 69)
	if err != nil {
		t.Fatalf("Error setting addr 0: %v", err)
	}

	v, err := m.GetMemory(0)
	if err != nil || v != 69 {
		t.Fatalf("Error getting addr 0: %v - %v", err, v)
	}

	err = m.SetMemory(math.MaxUint32, 42)
	if err != nil {
		t.Fatalf("Error setting addr %v: %v", math.MaxUint32, err)
	}

	v, err = m.GetMemory(math.MaxUint32)
	if err != nil || v != 42 {
		t.Fatalf("Error getting addr %v: %v - %v", math.MaxUint32, err, v)
	}

	err = m.SetMemory(math.MaxUint32 + 1, 39)
	if err == nil {
		t.Fatalf("Did not failed in trying to set more than %v", math.MaxUint32)
	}

	v, err = m.GetMemory(math.MaxUint32 + 1)
	if err == nil {
		t.Fatalf("Did not failed in trying to set more than %v: %v", math.MaxUint32, v)
	}
}

func TestChunkedMem(t *testing.T) {
	var m RiscV
	arr := []uint8{69, 42, 39}

	err := m.SetMemoryChunk(0, arr)
	if err != nil {
		t.Fatalf("Error setting addr 0: %v", err)
	}

	v, err := m.GetMemoryChunk(0, 3)
	if err != nil || !reflect.DeepEqual(v, arr) {
		t.Fatalf("Error getting addr 0: %v - %v", err, v)
	}

	err = m.SetMemoryChunk(math.MaxUint32 - 2, arr)
	if err != nil {
		t.Fatalf("Error setting addr %v: %v", math.MaxUint32, err)
	}

	v, err = m.GetMemoryChunk(math.MaxUint32 - 2, 3)
	if err != nil || !reflect.DeepEqual(v, arr) {
		t.Fatalf("Error getting addr %v: %v - %v", math.MaxUint32, err, v)
	}

	err = m.SetMemoryChunk(math.MaxUint32 - 1, arr)
	if err == nil {
		t.Fatalf("Did not failed in trying to set more than %v", math.MaxUint32)
	}

	v, err = m.GetMemoryChunk(math.MaxUint32 - 1, 3)
	if err == nil {
		t.Fatalf("Did not failed in trying to set more than %v: %v", math.MaxUint32, v)
	}
}

//go:embed test.asm
var asm string

func TestAssembler(t *testing.T) {
	var m RiscV
	code, _, err := m.Assemble(asm)
	if err != nil {
		t.Fatalf("Assembling failed with '%v'", err)
	}

	// Build by RARS.
	correctCode := []uint32{
		0x007281b3,
		0x407281b3,
		0x0072c1b3,
		0x0072e1b3,
		0x0072f1b3,
		0x007291b3,
		0x0072d1b3,
		0x4072d1b3,
		0x0072a1b3,
		0x0072b1b3,
		0x02a28193,
		0x02a2c193,
		0x02a2e193,
		0x02a2f193,
		0x00329193,
		0x0032d193,
		0x4032d193,
		0x02a2a193,
		0x02a2b193,
		0x02a28183,
		0x02a29183,
		0x02a2a183,
		0x02a2c183,
		0x02a2d183,
		0x003281a3,
		0x003291a3,
		0x0032a1a3,
		0xfe328ea3,
		0xfe329ea3,
		0xfe32aea3,
		0x06518c63,
		0x06519a63,
		0x0651c863,
		0x0651d663,
		0x0651e463,
		0x0651f263,
		0xf65188e3,
		0xf65196e3,
		0xf651c4e3,
		0xf651d2e3,
		0xf651e0e3,
		0xf451fee3,
		0x048001ef,
		0xf55ff1ef,
		0x003281e7,
		0xffd281e7,
		0x000051b7,
		0x00005197,
		0xffffb1b7,
		0xffffb197,
		0x00000073,
		0x00100073,
		0x027281b3,
		0x027291b3,
		0x0272a1b3,
		0x0272b1b3,
		0x0272c1b3,
		0x0272d1b3,
		0x0272e1b3,
		0x0272f1b3,
	}

	for i := 0; i < len(correctCode); i++ {
		asmi := uint32(code[i * 4])
		asmi = asmi | uint32(code[i * 4 + 1]) << 8
		asmi = asmi | uint32(code[i * 4 + 2]) << 16
		asmi = asmi | uint32(code[i * 4 + 3]) << 24
		if asmi != correctCode[i] {
			t.Logf("Byte 0: %x Byte 1: %x Byte 2: %x Byte 3: %x",
				code[i * 4],
				code[i * 4 + 1],
				code[i * 4 + 2],
				code[i * 4 + 3])
			t.Fatalf("Incorrect instruction: %x (expected %x) at address %x", asmi, correctCode[i], i * 4)
		}
	}
}

//go:embed test-instructions.asm
var inst string

func (m *RiscV) getByName(n string) uint64 {
	r, _ := m.GetRegisterNumber(n)
	v, _ := m.GetRegister(r)
	return v
}

func (m *RiscV) assertRegister(t *testing.T, v string, r uint64, s string) {
	a := m.getByName(v)
	if a != r {
		t.Fatalf("Wrong result in '%v': %v", s, a)
	}
}

func (m *RiscV) ensureBranch(t *testing.T, i string) {
	pc := m.pc
	m.NextInstruction()
	if pc == m.pc - 4 {
		t.Fatalf("%s didn't branch correctly: %d", i, m.pc)
	}
}

func (m *RiscV) ensureDontBranch(t *testing.T, i string) {
	pc := m.pc
	m.NextInstruction()
	if pc != m.pc - 4 {
		t.Fatalf("%s didn't pass correctly: %x", i, m.pc)
	}
}

func TestInstructions(t *testing.T) {

	var m RiscV
	code, _, err := m.Assemble(inst)
	if err != nil {
		t.Fatalf("Couldn't assemble: %v", err)
	}

	err = m.LoadProgram(code)
	if err != nil {
		t.Fatalf("Couldn't load program: %v", err)
	}

	//
	// Arithmetic immediate.
	//
	m.NextInstruction()
	m.assertRegister(t, "t0", 1, "addi t0, zero, 1")

	m.NextInstruction()
	m.assertRegister(t, "t0", 4, "xori t0, t0, 5")

	m.NextInstruction()
	m.assertRegister(t, "t0", 6, "ori t0, t0, 2")

	m.NextInstruction()
	m.assertRegister(t, "t0", 2, "andi t0, t0, 3")

	m.NextInstruction()
	m.assertRegister(t, "t0", 8, "slli t0, t0, 2")

	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.assertRegister(t, "t2", 536870914, "srli t2, t1, 2")

	m.NextInstruction()
	m.assertRegister(t, "t2", 3758096386, "srai t2, t1, 2")

	m.NextInstruction()
	m.assertRegister(t, "t3", 1, "slti t3, t2, t0")

	m.NextInstruction()
	m.assertRegister(t, "t3", 0, "sltu t3, t2, t0")

	//
	// Arithmetic.
	//
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()

	m.NextInstruction()
	m.assertRegister(t, "t4", 3, "add t4, t0, t1")

	m.NextInstruction()
	m.assertRegister(t, "t4", 1, "sub t4, t2, t1")

	m.NextInstruction()
	m.assertRegister(t, "t4", 6, "xor t4, t3, t2")

	m.NextInstruction()
	m.assertRegister(t, "t4", 3, "or t4, t0, t1")

	m.NextInstruction()
	m.assertRegister(t, "t4", 1, "and t4, t2, t3")

	m.NextInstruction()
	m.assertRegister(t, "t4", 4, "sll t4, t0, t1")

	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.assertRegister(t, "t4", 536870914, "srl t4, t5, t1")

	m.NextInstruction()
	m.assertRegister(t, "t4", 3758096386, "sra t4, t5, t1")

	m.NextInstruction()
	m.assertRegister(t, "t4", 1, "slt t4, t5, t2")

	m.NextInstruction()
	m.assertRegister(t, "t4", 0, "sltu t4, t5, t2")

	//
	// Loads and stores.
	//
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.assertRegister(t, "t1", 0xaaeeffff, "test value moving")

	m.NextInstruction()
	m.NextInstruction()
	m.assertRegister(t, "t2", 0xaaeeffff, "lw t2, t0, 0")

	m.NextInstruction()
	m.NextInstruction()
	m.assertRegister(t, "t2", 0xffffffff, "lb t2, t0, 0")
	m.NextInstruction()
	m.assertRegister(t, "t2", 0x000000ff, "lbu t2, t0, 0")

	m.NextInstruction()
	m.NextInstruction()
	m.assertRegister(t, "t2", 0xffffffff, "lh t2, t0, 0")
	m.NextInstruction()
	m.assertRegister(t, "t2", 0x0000ffff, "lhu t2, t0, 0")

	//
	// Jumps.
	//
	ret := m.pc
	m.NextInstruction()
	if m.pc == ret + 4 {
		t.Fatalf("jal does not jump correctly")
	}
	m.NextInstruction()
	if m.pc != ret + 4 {
		t.Fatalf("jalr does not return correctly")
	}

	//
	// Calls.
	//
	m.NextInstruction()
	call, _ := m.NextInstruction()
	if call == nil {
		t.Fatalf("call is nil")
	}
	if call.Number != 2 {
		t.Fatalf("call number is not 2")
	}
	call, _ = m.NextInstruction()
	if call == nil {
		t.Fatalf("break call is nil")
	}
	if call.Number != 1 {
		t.Fatalf("break is not 1")
	}

	//
	// Branches.
	//
	m.NextInstruction()
	m.NextInstruction()

	m.ensureDontBranch(t, "beq")
	m.ensureBranch(t, "beq")

	m.ensureDontBranch(t, "bne")
	m.ensureBranch(t, "bne")

	m.ensureDontBranch(t, "blt")
	m.ensureDontBranch(t, "blt")
	m.ensureBranch(t, "blt")

	m.ensureDontBranch(t, "bge")
	m.ensureBranch(t, "bge")
	m.ensureBranch(t, "bge")

	m.ensureDontBranch(t, "bltu")
	m.ensureDontBranch(t, "bltu")
	m.ensureBranch(t, "bltu")

	m.ensureDontBranch(t, "bgeu")
	m.ensureBranch(t, "bgeu")
	m.ensureBranch(t, "bgeu")

	//
	// auipc/lui
	//
	m.NextInstruction()
	m.assertRegister(t, "t0", 0xaaaaa000, "lui t0, 0xaaaaa")
	pc := m.pc
	m.NextInstruction()
	m.assertRegister(t, "t0", uint64(0xaaaaa000 + pc), "auipc t0, 0xaaaaa")

	//
	// Multiplication extension.
	//
	m.NextInstruction()
	m.NextInstruction()

	m.NextInstruction()
	m.assertRegister(t, "t2", uint64(0x1fffd000), "mul t2, t0, t1")
	m.NextInstruction()
	m.assertRegister(t, "t2", uint64(0x1), "mulh t2, t0, t1")
	m.NextInstruction()
	m.assertRegister(t, "t2", uint64(0x1ffffaaa), "div t2, t0, t1")

	m.NextInstruction()

	m.NextInstruction()
	m.assertRegister(t, "t2", uint64(0xE0003000), "mul t2, t0, t1 (neg)")
	m.NextInstruction()
	m.assertRegister(t, "t2", uint64(0xFFFFFFFE), "mulh t2, t0, t1 (neg)")
	m.NextInstruction()
	// -536869545 but without the 4 top bytes.
	m.assertRegister(t, "t2", uint64(3758097750), "div t2, t0, t1 (neg)")
}
