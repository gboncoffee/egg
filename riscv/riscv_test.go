package riscv

import (
	"testing"
	"math"
)

func TestSetRegister(t *testing.T) {
	var m RiscV
	err := m.SetRegister(1, 42)
	if err != nil {
		t.Fatalf("Failed to set register 1")
	}

	err = m.SetRegister(31, 69)
	if err != nil {
		t.Fatalf("Failed to set register 31")
	}

	err = m.SetRegister(0, 39)
	if err != nil {
		t.Fatalf("Failed to set register 0")
	}

	// fails

	err = m.SetRegister(32, 42)
	if err == nil {
		t.Fatalf("Setting register 32 did not failed")
	}

	err = m.SetRegister(1, math.MaxUint32 + 1)
	if err == nil {
		t.Fatalf("Setting register to more than MaxUint32 did not failed")
	}
}
