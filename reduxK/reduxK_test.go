package reduxK

import (
	"testing"

	"github.com/gboncoffee/egg/assembler"
	"github.com/gboncoffee/egg/machine"
)

func TestReduxV(t *testing.T) {
	machine.InterCtx.Init()
	machine.InterCtx.AutoSetPreferedLocale()
	assembler.InterCtx = &machine.InterCtx

	var m ReduxV
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
	assertRegister(0, 5, "prepare for jump")

	m.NextInstruction()
	assertRegister(1, 5, "copy addr")

	m.NextInstruction()
	assertRegister(0, 0, "zero r0 for branching")

	assertBranch("brzr")

	m.NextInstruction()
	assertRegister(0, 1, "addi 1 for branching")

	assertDontBranch("brzr")

	assertBranch("ji")

	m.NextInstruction()
	assertRegister(0, 0, "sub test")

	m.NextInstruction()
	assertRegister(0, 1, "not test")

	m.NextInstruction()
	assertRegister(0, 2, "add test")

	m.NextInstruction()
	assertRegister(1, 0, "sub for logical test")

	m.NextInstruction()
	assertRegister(1, 1, "not for logical test")

	m.NextInstruction()
	assertRegister(0, 3, "or test")

	m.NextInstruction()
	assertRegister(0, 1, "and test")

	m.NextInstruction()
	assertRegister(0, 2, "sll test")

	m.NextInstruction()
	assertRegister(0, 3, "xor test")

	m.NextInstruction()
	assertRegister(0, 1, "srr test")

	m.NextInstruction()
	assertRegister(0, 0, "sub r0 for load/store")

	m.NextInstruction()
	assertRegister(1, 0, "sub r1 for load/store")

	m.NextInstruction()
	assertRegister(0, 0xff, "addi address")

	m.NextInstruction()
	assertRegister(1, 0xff, "add r1, r0 address")

	m.NextInstruction()
	assertRegister(0, 0, "zero r0 for value")

	m.NextInstruction()
	assertRegister(0, 0xfe, "addi value")

	m.NextInstruction()
	mem, _ := m.GetMemory(0xff)
	if mem != 0xfe {
		t.Fatalf("st failed")
	}

	m.NextInstruction()
	assertRegister(0, 0, "sub for load test")

	m.NextInstruction()
	assertRegister(0, 0xfe, "ld")

	call, _ := m.NextInstruction()
	if call == nil || call.Number != machine.SYS_BREAK {
		t.Fatalf("Break failed: %v", call)
	}

	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()
	m.NextInstruction()

	call, _ = m.NextInstruction()
	if call == nil || call.Number != 0 || call.Arg1 != 1 || call.Arg2 != 2 {
		t.Fatalf("Call failed: %v", call)
	}
}
