package sagui

import (
	_ "embed"
	"testing"
)

//go:embed test.asm
var asm string

func TestSagui(t *testing.T) {
	var m Sagui
	code, _, err := m.Assemble(asm)
	if err != nil {
		t.Fatalf("Couldn't assemble: %v", err)
	}

	err = m.LoadProgram(code)
	if err != nil {
		t.Fatalf("Couldn't load program: %v", err)
	}

	assertRegister := func(reg uint64, v uint64, name string) {
		r, _ := m.GetRegister(reg)
		if r != v {
			t.Fatalf("Testing %s failed: register %v expected to be %v, is %v", name, reg, v, r)
		}
	}

	assertBranch := func(name string) {
		pc := m.GetCurrentInstructionAddress()
		m.NextInstruction()
		if m.GetCurrentInstructionAddress() == pc+1 {
			t.Fatalf("Didn't branched on %v", name)
		}
	}

	assertDontBranch := func(name string) {
		pc := m.GetCurrentInstructionAddress()
		m.NextInstruction()
		if m.GetCurrentInstructionAddress() != pc+1 {
			t.Fatalf("Branched on %v", name)
		}
	}

	m.NextInstruction()
	assertRegister(0, 5, "first movl to branch")

	m.NextInstruction()
	assertRegister(1, 5, "first movr to branch")
	m.NextInstruction()
	assertRegister(0, 0, "zeroing r0 to branch")

	assertBranch("brzr")

	m.NextInstruction()
	assertBranch("jr")

	m.NextInstruction()
	assertBranch("brzi")

	m.NextInstruction()
	assertDontBranch("brzi")
	assertDontBranch("brzr")
	assertBranch("ji")

	m.NextInstruction()
	m.NextInstruction()
	assertRegister(0, 0x1, "movl to arithmetic")
	m.NextInstruction()
	assertRegister(1, 0x1, "movr to arithmetic")
	m.NextInstruction()
	m.NextInstruction()
	assertRegister(0, 0xf0, "movh to arithmetic")

	m.NextInstruction()
	assertRegister(0, 0xf1, "add")
	m.NextInstruction()
	assertRegister(0, 0xf0, "sub")

	m.NextInstruction()
	assertRegister(0, 0xf1, "or")
	m.NextInstruction()
	assertRegister(0, 0x1, "and")

	m.NextInstruction()
	assertRegister(0, 0, "first not")
	m.NextInstruction()
	assertRegister(0, 1, "second not")

	m.NextInstruction()
	m.NextInstruction()
	assertRegister(0, 0x2, "slr")
	m.NextInstruction()
	assertRegister(0, 0x1, "srr")

	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	mem, _ := m.GetMemory(0xfa)
	if mem != 0xa {
		t.Fatalf("st failed")
	}
	m.NextInstruction()
	assertRegister(1, 0, "zeroing r1 to be sure")
	m.NextInstruction()
	assertRegister(1, 0xa, "ld")
}
