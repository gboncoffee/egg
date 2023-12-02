package riscv

import (
	"errors"
	"fmt"
	"math"
)

type RiscV struct {
	registers [32]uint32
	pc        uint32
	// todo: mem
}

func (m *RiscV) LoadProgram(uint64) error {
	panic("not implemented")
}

func (m *RiscV) NextInstruction() error {
	panic("not implemented")
}

func (m *RiscV) GetMemory(uint64) (uint64, error) {
	panic("not implemented")
}

func (m *RiscV) SetMemory(uint64, uint64) error {
	panic("not implemented")
}

func (m *RiscV) GetRegister(reg uint64) (uint64, error) {
	if reg >= 32 {
		return 0, errors.New(fmt.Sprintf("No such register: %d. RISC-V has only 32 registers.", reg))
	}

	return uint64(m.registers[reg]), nil
}

func (m *RiscV) SetRegister(reg uint64, content uint64) error {
	if reg >= 32 {
		return errors.New(fmt.Sprintf("No such register: %d. RISC-V has only 32 registers.", reg))
	}

	if content > math.MaxUint32 {
		return errors.New(fmt.Sprintf("Number beyond 32 limit: %d. RISC-V has only 32 bit registers.", content))
	}

	if reg != 0 {
		m.registers[reg] = uint32(content) // will not overflow, we already checked
	}

	return nil
}
