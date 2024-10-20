package mips

import (
	"reflect"
	"testing"

	"github.com/gboncoffee/egg/machine"
	"github.com/gboncoffee/egg/assembler"
)

func TestIsa(t *testing.T) {
	var m Mips

	// Assembling will fail if we don't initialize the InterCtx.
	machine.InterCtx.Init()
	machine.InterCtx.AutoSetPreferedLocale()
	assembler.InterCtx = &machine.InterCtx

	code, _, err := m.Assemble("mips-test.asm")
	if err != nil {
		t.Fatalf("Could not assemble file: %v", err)
	}

	m.LoadProgram(code)

	// add t0, t1, t2
	m.SetRegister(1+8, 3)
	m.SetRegister(2+8, 2)
	m.NextInstruction()
	v, _ := m.GetRegister(0 + 8)
	if v != 5 {
		t.Fatalf("add failed: %v", v)
	}

	// addi t0, t1, 2
	m.SetRegister(1+8, 3)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 5 {
		t.Fatalf("addi failed: %v", v)
	}

	// addiu t0, t1, 3
	m.SetRegister(1+8, 0xf0_00_00_00)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0xf0000003 {
		t.Fatalf("addiu failed: %v", v)
	}

	// addu t0, t1, t2
	m.SetRegister(1+8, 3)
	m.SetRegister(2+8, 0xf0_00_00_00)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0xf0000003 {
		t.Fatalf("addu failed: %v", v)
	}

	// clo t0, t1
	m.SetRegister(1+8, 0xf0_00_00_00)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 4 {
		t.Fatalf("clo failed: %v", v)
	}

	// clz t0, t1
	m.SetRegister(1+8, 3)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 30 {
		t.Fatalf("clz failed: %v", v)
	}

	// lui t0, 0xffff
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0xffff0000 {
		t.Fatalf("lui failed: %v", v)
	}

	// seb t0, t1
	m.SetRegister(1+8, 0xff)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0xffffffff {
		t.Fatalf("seb failed: %v", v)
	}

	// seh t0, t1
	m.SetRegister(1+8, 0xff_ff)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0xffffffff {
		t.Fatalf("seh failed: %v", v)
	}

	// sub t0, t1, t2
	m.SetRegister(1+8, 3)
	m.SetRegister(2+8, 2)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 1 {
		t.Fatalf("sub failed: %v", v)
	}

	// subu t0, t1, t2
	m.SetRegister(1+8, 0x80_00_00_03)
	m.SetRegister(2+8, 2)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0x80_00_00_01 {
		t.Fatalf("subu failed: %v", v)
	}

	// sll t0, t1, 2
	m.SetRegister(1+8, 1)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 4 {
		t.Fatalf("sll failed: %v", v)
	}

	// sllv t0, t1, t2
	m.SetRegister(1+8, 1)
	m.SetRegister(2+8, 2)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 4 {
		t.Fatalf("sllv failed: %v", v)
	}

	// sra t0, t1, 2
	m.SetRegister(1+8, 0x80_00_00_00)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0xe0_00_00_00 {
		t.Fatalf("sra failed: %v", v)
	}

	// srav t0, t1, t2
	m.SetRegister(1+8, 0x80_00_00_00)
	m.SetRegister(2+8, 2)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0xe0_00_00_00 {
		t.Fatalf("srav failed: %v", v)
	}

	// srl t0, t1, 2
	m.SetRegister(1+8, 0x80_00_00_00)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0x20_00_00_00 {
		t.Fatalf("srl failed: %v", v)
	}

	// srlv t0, t1, t2
	m.SetRegister(1+8, 0x80_00_00_00)
	m.SetRegister(2+8, 2)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0x20_00_00_00 {
		t.Fatalf("srlv failed: %v", v)
	}

	// and t0, t1, t2
	m.SetRegister(1+8, 0xffff)
	m.SetRegister(2+8, 0xff)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0xff {
		t.Fatalf("and failed: %v", v)
	}

	// andi t0, t1 0xff
	m.SetRegister(1+8, 0xffff)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0xff {
		t.Fatalf("andi failed: %v", v)
	}

	// nor t0, t1, t2
	m.SetRegister(1+8, 0xff_ff_00_00)
	m.SetRegister(2+8, 0x00_ff_ff_00)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0x00_00_00_ff {
		t.Fatalf("nor failed: %v", v)
	}

	// or t0, t1, t2
	m.SetRegister(1+8, 0xff_ff_00_00)
	m.SetRegister(2+8, 0x00_ff_ff_00)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0xff_ff_ff_00 {
		t.Fatalf("or failed: %v", v)
	}

	// ori t0, t1, 0xff
	m.SetRegister(1+8, 0xff_00)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0xff_ff {
		t.Fatalf("ori failed: %v", v)
	}

	// xor t0, t1, t2
	m.SetRegister(1+8, 0xff_ff_00_00)
	m.SetRegister(2+8, 0x00_ff_ff_00)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0xff_00_ff_00 {
		t.Fatalf("xor failed: %v", v)
	}

	// xori t0, t1, 0xff
	m.SetRegister(1+8, 0xff_01)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0xff_fe {
		t.Fatalf("xori failed: %v", v)
	}

	// movn t0, t1, t2
	m.SetRegister(0+8, 0)
	m.SetRegister(1+8, 0xff)
	m.SetRegister(2+8, 0)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0 {
		t.Fatalf("movn (not) failed: %v", v)
	}
	m.pc -= 4
	m.SetRegister(2+8, 1)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0xff {
		t.Fatalf("movn (yes) failed: %v", v)
	}

	// movz t0, t1, t2
	m.SetRegister(0+8, 0)
	m.SetRegister(1+8, 0xff)
	m.SetRegister(2+8, 1)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0 {
		t.Fatalf("movz (not) failed: %v", v)
	}
	m.pc -= 4
	m.SetRegister(2+8, 0)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0xff {
		t.Fatalf("movz (yes) failed: %v", v)
	}

	// slt t0, t1, t2
	m.SetRegister(1+8, 1)
	m.SetRegister(2+8, 2)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 1 {
		t.Fatalf("slt failed: %v", v)
	}
	m.pc -= 4
	m.SetRegister(1+8, 2)
	m.SetRegister(2+8, 1)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0 {
		t.Fatalf("slt failed: %v", v)
	}

	// slti t0, t1, 0xff
	m.SetRegister(1+8, 1)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 1 {
		t.Fatalf("slti failed: %v", v)
	}
	m.pc -= 4
	m.SetRegister(1+8, 0xff)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0 {
		t.Fatalf("slti failed: %v", v)
	}

	// sltiu t0, t1, 0xff
	m.SetRegister(1+8, 1)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 1 {
		t.Fatalf("sltiu failed: %v", v)
	}
	m.pc -= 4
	m.SetRegister(1+8, 0xff)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0 {
		t.Fatalf("sltiu failed: %v", v)
	}

	// sltu t0, t1, t2
	m.SetRegister(1+8, 1)
	m.SetRegister(2+8, 2)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 1 {
		t.Fatalf("sltu failed: %v", v)
	}
	m.pc -= 4
	m.SetRegister(1+8, 2)
	m.SetRegister(2+8, 1)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0 {
		t.Fatalf("sltu failed: %v", v)
	}

	// div t0, t1
	m.SetRegister(0+8, 10)
	m.SetRegister(1+8, 2)
	m.NextInstruction()
	v, _ = m.GetRegister(LO)
	if v != 5 {
		t.Fatalf("div failed (lo): %v", v)
	}
	v, _ = m.GetRegister(HI)
	if v != 0 {
		t.Fatalf("div failed (hi): %v", v)
	}

	// mult t0, t1
	m.SetRegister(0+8, 2)
	m.SetRegister(1+8, 3)
	m.NextInstruction()
	v, _ = m.GetRegister(LO)
	if v != 6 {
		t.Fatalf("mult failed: %v", v)
	}
	v, _ = m.GetRegister(HI)
	if v != 0 {
		t.Fatalf("mult failed: %v", v)
	}

	// mfhi t0
	m.SetRegister(HI, 0x12345678)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0x12345678 {
		t.Fatalf("mfhi failed: %v", v)
	}

	// mflo t0
	m.SetRegister(LO, 0x12345678)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0x12345678 {
		t.Fatalf("mflo failed: %v", v)
	}

	// mthi t0
	m.SetRegister(1+8, 0x12345678)
	m.NextInstruction()
	v, _ = m.GetRegister(HI)
	if v != 0x12345678 {
		t.Fatalf("mthi failed: %v", v)
	}

	// mtlo t0
	m.SetRegister(1+8, 0x12345678)
	m.NextInstruction()
	v, _ = m.GetRegister(LO)
	if v != 0x12345678 {
		t.Fatalf("mtlo failed: %v", v)
	}

	// beq t0, t1, 8
	m.SetRegister(0+8, 1)
	m.SetRegister(1+8, 1)
	savedPC := m.pc
	m.NextInstruction()
	if m.pc != 0 {
		t.Fatalf("beq failed: %v to %v", savedPC, m.pc)
	}
	m.pc = savedPC + 4

	// bgez t0, 8
	m.SetRegister(0+8, 1)
	savedPC = m.pc
	m.NextInstruction()
	if m.pc != 0 {
		t.Fatalf("bgez failed: %v to %v", savedPC, m.pc)
	}
	m.pc = savedPC + 4

	// bgtz t0, 8
	m.SetRegister(0+8, 1)
	savedPC = m.pc
	m.NextInstruction()
	if m.pc != 0 {
		t.Fatalf("bgtz failed: %v to %v", savedPC, m.pc)
	}
	m.pc = savedPC + 4

	// blez t0, 8
	m.SetRegister(0+8, 0)
	savedPC = m.pc
	m.NextInstruction()
	if m.pc != 0 {
		t.Fatalf("blez failed: %v to %v", savedPC, m.pc)
	}
	m.pc = savedPC + 4

	// bltz t0, 8
	m.SetRegister(0+8, 0xff_ff_ff_ff)
	savedPC = m.pc
	m.NextInstruction()
	if m.pc != 0 {
		t.Fatalf("bltz failed: %v to %v", savedPC, m.pc)
	}
	m.pc = savedPC + 4

	// bne t0, t1, 8
	m.SetRegister(0+8, 1)
	m.SetRegister(1+8, 2)
	savedPC = m.pc
	m.NextInstruction()
	if m.pc != 0 {
		t.Fatalf("bne failed: %v to %v", savedPC, m.pc)
	}
	m.pc = savedPC + 4

	// break
	call, _ := m.NextInstruction()
	if call.Number != machine.SYS_BREAK {
		t.Fatalf("break failed: %v", call.Number)
	}

	// syscall
	m.SetRegister(2, machine.SYS_READ)
	m.SetRegister(4, 1)
	m.SetRegister(5, 2)
	call, _ = m.NextInstruction()
	expectedCall := machine.Call{
		Number: machine.SYS_READ,
		Arg1:   1,
		Arg2:   2,
	}
	if *call != expectedCall {
		t.Fatalf("syscall failed: %v", call)
	}

	// j 8
	savedPC = m.pc
	m.NextInstruction()
	if m.pc != (savedPC&0xf0000000)|8 {
		t.Fatalf("j failed: %v", m.pc)
	}
	m.pc = savedPC + 4

	// jal 8
	savedPC = m.pc
	m.NextInstruction()
	if m.pc != (savedPC&0xf0000000)|8 {
		t.Fatalf("jal failed: %v", m.pc)
	}
	v, _ = m.GetRegister(31)
	if v != uint64(savedPC+4) {
		t.Fatalf("jal failed to set return address: %v", v)
	}
	m.pc = savedPC + 4

	// jalr t0
	m.SetRegister(0+8, 0x12345678)
	savedPC = m.pc
	m.NextInstruction()
	if m.pc != 0x12345678 {
		t.Fatalf("jalr failed: %v", m.pc)
	}
	v, _ = m.GetRegister(31)
	if v != uint64(savedPC+4) {
		t.Fatalf("jalr failed to set return address: %v", v)
	}
	m.pc = savedPC + 4

	// jr t0
	m.SetRegister(0+8, 0x12345678)
	savedPC = m.pc
	m.NextInstruction()
	if m.pc != 0x12345678 {
		t.Fatalf("jr failed: %v", m.pc)
	}
	m.pc = savedPC + 4

	// lb t0, t1, 0
	m.SetMemoryChunk(0xcafebabe, []uint8{0xff, 2, 3, 4})
	m.SetRegister(1 + 8, 0xcafebabe)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0xff_ff_ff_ff {
		t.Fatalf("lb failed: %v", v)
	}

	// lbu t0, t1, 0
	m.SetMemoryChunk(0xcafebabe, []uint8{0xff, 2, 3, 4})
	m.SetRegister(1 + 8, 0xcafebabe)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0xff {
		t.Fatalf("lbu failed: %v", v)
	}

	// lh t0, t1, 0
	m.SetMemoryChunk(0xcafebabe, []uint8{0xff, 0xff, 3, 4})
	m.SetRegister(1 + 8, 0xcafebabe)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0xff_ff_ff_ff {
		t.Fatalf("lh failed: %v", v)
	}

	// lhu t0, t1, 0
	m.SetMemoryChunk(0xcafebabe, []uint8{0xff, 0xff, 3, 4})
	m.SetRegister(1 + 8, 0xcafebabe)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0xff_ff {
		t.Fatalf("lhu failed: %v", v)
	}

	// lw t0, t1, 0
	m.SetMemoryChunk(0xcafebabe, []uint8{1, 2, 3, 4})
	m.SetRegister(1 + 8, 0xcafebabe)
	m.NextInstruction()
	v, _ = m.GetRegister(0 + 8)
	if v != 0x04_03_02_01 {
		t.Fatalf("lw failed: %v", v)
	}

	// sb t0, t1, 0
	m.SetMemoryChunk(0xcafebabe, []uint8{0, 0, 0, 0})
	m.SetRegister(0 + 8, 0xdeadbeef)
	m.SetRegister(1 + 8, 0xcafebabe)
	m.NextInstruction()
	slice, _ := m.GetMemoryChunk(0xcafebabe, 4)
	if !reflect.DeepEqual(slice, []uint8{0xef, 0, 0, 0}) {
		t.Fatalf("sb failed: %v", slice)
	}

	// sh t0, t1, 0
	m.SetMemoryChunk(0xcafebabe, []uint8{0, 0, 0, 0})
	m.SetRegister(0 + 8, 0xdeadbeef)
	m.SetRegister(1 + 8, 0xcafebabe)
	m.NextInstruction()
	slice, _ = m.GetMemoryChunk(0xcafebabe, 4)
	if !reflect.DeepEqual(slice, []uint8{0xef, 0xbe, 0, 0}) {
		t.Fatalf("sh failed: %v", slice)
	}

	// sw t0, t1, 0
	m.SetMemoryChunk(0xcafebabe, []uint8{0, 0, 0, 0})
	m.SetRegister(0 + 8, 0xdeadbeef)
	m.SetRegister(1 + 8, 0xcafebabe)
	m.NextInstruction()
	slice, _ = m.GetMemoryChunk(0xcafebabe, 4)
	if !reflect.DeepEqual(slice, []uint8{0xef, 0xbe, 0xad, 0xde}) {
		t.Fatalf("sw failed: %v", slice)
	}
}
