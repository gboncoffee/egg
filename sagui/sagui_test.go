package sagui

import (
	"testing"

	"github.com/gboncoffee/egg/machine"
)

func TestSagui(t *testing.T) {
	var m Sagui
	code, _, err := m.Assemble("test.asm")
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
		_, _ = m.NextInstruction()
		if m.GetCurrentInstructionAddress() == pc+1 {
			t.Fatalf("Didn't branched on %v", name)
		}
	}

	assertDontBranch := func(name string) {
		pc := m.GetCurrentInstructionAddress()
		_, _ = m.NextInstruction()
		if m.GetCurrentInstructionAddress() != pc+1 {
			t.Fatalf("Branched on %v", name)
		}
	}

	_, _ = m.NextInstruction()
	assertRegister(0, 5, "first movl to branch")

	_, _ = m.NextInstruction()
	assertRegister(1, 5, "first movr to branch")
	_, _ = m.NextInstruction()
	assertRegister(0, 0, "zeroing r0 to branch")

	assertBranch("brzr")

	_, _ = m.NextInstruction()
	assertBranch("jr")

	_, _ = m.NextInstruction()
	assertBranch("brzi")

	_, _ = m.NextInstruction()
	assertDontBranch("brzi")
	assertDontBranch("brzr")
	assertBranch("ji")

	_, _ = m.NextInstruction()
	_, _ = m.NextInstruction()
	assertRegister(0, 0x1, "movl to arithmetic")
	_, _ = m.NextInstruction()
	assertRegister(1, 0x1, "movr to arithmetic")
	_, _ = m.NextInstruction()
	_, _ = m.NextInstruction()
	assertRegister(0, 0xf0, "movh to arithmetic")

	_, _ = m.NextInstruction()
	assertRegister(0, 0xf1, "add")
	_, _ = m.NextInstruction()
	assertRegister(0, 0xf0, "sub")

	_, _ = m.NextInstruction()
	assertRegister(0, 0xf1, "or")
	_, _ = m.NextInstruction()
	assertRegister(0, 0x1, "and")

	_, _ = m.NextInstruction()
	assertRegister(0, 0, "first not")
	_, _ = m.NextInstruction()
	assertRegister(0, 1, "second not")

	_, _ = m.NextInstruction()
	_, _ = m.NextInstruction()
	assertRegister(0, 0x2, "slr")
	_, _ = m.NextInstruction()
	assertRegister(0, 0x1, "srr")

	_, _ = m.NextInstruction()
	_, _ = m.NextInstruction()
	_, _ = m.NextInstruction()
	_, _ = m.NextInstruction()
	mem, _ := m.GetMemory(0xfa)
	if mem != 0xa {
		t.Fatalf("st failed")
	}
	_, _ = m.NextInstruction()
	assertRegister(1, 0, "zeroing r1 to be sure")
	_, _ = m.NextInstruction()
	assertRegister(1, 0xa, "ld")

	call, _ := m.NextInstruction()
	if call == nil || call.Number != machine.SYS_BREAK {
		t.Fatalf("Break failed: %v", call)
	}
}
