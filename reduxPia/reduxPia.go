// Package reduxPia implements a REDUX-PIÁ machine for EGG.
// ji 0 performs an ebreak. or r0, r0 performs an ecall.
package reduxPia

import (
	"fmt"

	"github.com/gboncoffee/egg/assembler"
	"github.com/gboncoffee/egg/machine"
	"github.com/gboncoffee/egg/reduxc"
)

func reduxPiaExecuteExtension(bin uint8, m *reduxc.ReduxC) (bool, *machine.Call, error) {
	op := bin >> 4
	switch op {
	case 0x5:
		*m.PC() = *m.PC() + 1
		r0v, _ := m.GetRegister(0)
		imm := bin & 0xf
		_ = m.SetRegister(0, uint64((uint8(r0v)&0xf)|(imm<<4)))
		return true, nil, nil
	case 0x6:
		imm := bin & 0xf
		r3v, _ := m.GetRegister(3)
		r := uint64(r3v - 1)
		_ = m.SetRegister(3, r)
		if r != 0 {
			*m.PC() = *m.PC() - (imm + 1)
		} else {
			*m.PC() = *m.PC() + 1
		}
		return true, nil, nil
	case 0x7:
		*m.PC() = *m.PC() + 1
		ra := (bin >> 2) & 0x3
		rb := bin & 0x3
		r0v, _ := m.GetRegister(0)
		rav, _ := m.GetRegister(uint64(ra))
		rbv, _ := m.GetRegister(uint64(rb))
		_ = m.SetRegister(0, uint64(r0v+rav*rbv))
		return true, nil, nil
	case 0x1:
		if bin == 0x10 {
			*m.PC() = *m.PC() + 1
			return true, &machine.Call{
				Number: machine.SYS_BREAK,
			}, nil
		}
	case 0xa:
		if bin == 0xa0 {
			r0v, _ := m.GetRegister(0)
			r1v, _ := m.GetRegister(1)
			r2v, _ := m.GetRegister(2)
			*m.PC() = *m.PC() + 1
			return true, &machine.Call{
				Number: r0v,
				Arg1:   r1v,
				Arg2:   r2v,
			}, nil
		}
	}

	return false, nil, nil
}

func reduxPiaAssembleExtension(t assembler.ResolvedToken) (uint8, error) {
	switch string(t.Value) {
	case "mac":
		if len(t.Args) != 2 {
			return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), string(t.Value))
		}
		ra := uint8(t.Args[0]) & 0x3
		rb := uint8(t.Args[1]) & 0x3
		return 0x70 | (ra << 2) | rb, nil
	case "ldui":
		if len(t.Args) != 1 {
			return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), string(t.Value))
		}
		return 0x50 | (uint8(t.Args[0]) & 0xf), nil
	case "loop":
		if len(t.Args) != 1 {
			return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 2 arguments"), string(t.Value))
		}
		return 0x60 | (uint8(t.Address) - 1 - uint8(t.Args[0])), nil
	case "ebreak":
		return 0x10, nil
	case "ecall":
		return 0xa0, nil
	}

	return 0, fmt.Errorf(machine.InterCtx.Get("Unknown instruction: %v"), string(t.Value))
}

func ReduxPia() *reduxc.ReduxC {
	return reduxc.ReduxVExtension("REDUX-PIÁ", reduxPiaExecuteExtension, reduxPiaAssembleExtension, nil)
}
