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

func TestGetRegister(t *testing.T) {
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
