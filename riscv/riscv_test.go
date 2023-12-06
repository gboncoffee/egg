package riscv

import (
	"testing"
	"math"
	"reflect"
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

func TestChukedMem(t *testing.T) {
	var m RiscV
	arr := []uint8{69, 42, 39}

	err := m.SetMemoryChunk(0, arr)
	if err != nil {
		t.Fatalf("Error setting addr 0: %v", err)
	}

	v, err := m.GetMemoryChunk(0, 3)
	if err != nil || reflect.DeepEqual(v, arr) {
		t.Fatalf("Error getting addr 0: %v - %v", err, v)
	}

	err = m.SetMemoryChunk(math.MaxUint32 - 2, arr)
	if err != nil {
		t.Fatalf("Error setting addr %v: %v", math.MaxUint32, err)
	}

	v, err = m.GetMemoryChunk(math.MaxUint32 - 2, 3)
	if err != nil || reflect.DeepEqual(v, arr) {
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
