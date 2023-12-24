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

func TestArithmeticInstructions(t *testing.T) {
	// TODO: Maybe test more than only addi, add and sub?
	var m RiscV

	// addi x6, x0, 0x45
	m.execute(0x04500313)
	r, _ := m.GetRegister(6)
	if r != 0x45 {
		t.Fatalf("Failed addi: %x", r)
	}

	// addi x7, x0, 0xffffffd6 (-42)
	m.execute(0xfd600393)
	r, _ = m.GetRegister(7)
	if r != 0x00000000ffffffd6 {
		t.Fatalf("Failed addi: %x", r)
	}

	// add x6, x6, x7
	m.execute(0x00730333)
	r, _ = m.GetRegister(6)
	if r != 27 {
		t.Fatalf("Failed addi: %x", r)
	}

	// sub x6, x6, x7
	m.execute(0x40730333)
	r, _ = m.GetRegister(6)
	if r != 0x45 {
		t.Fatalf("Failed addi: %x", r)
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
		0x04518c63,
		0x04519a63,
		0x0451c863,
		0x0451d663,
		0x0451e463,
		0x0451f263,
		0xf65188e3,
		0xf65196e3,
		0xf651c4e3,
		0xf651d2e3,
		0xf651e0e3,
		0xf451fee3,
		0x028001ef,
		0xf55ff1ef,
		0x003281e7,
		0xffd281e7,
		0x000051b7,
		0x00005197,
		0xffffb1b7,
		0xffffb197,
		0x00000073,
		0x00100073,
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
