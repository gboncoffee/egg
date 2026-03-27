// Package reduxK implements a reduxK machine for EGG.
//
// "or r0, r0" (ebreak) performs a BREAK and "xor r0, r0" (ecall) performs a CALL.
package reduxK

import (
	"errors"
	"fmt"

	"github.com/gboncoffee/egg/assembler"
	"github.com/gboncoffee/egg/machine"
	"github.com/gboncoffee/egg/reduxc"
)

type ReduxKExtState struct {
	auxRegisters [2]uint8
}

func reduxKExecuteExtension(bin uint8, m *reduxc.ReduxC) (bool, *machine.Call, error) {
	op := bin >> 4
	ext := m.AdditionalState().(*ReduxKExtState)

	switch op {
	case 0x5:
		*m.PC() = *m.PC() + 1
		for x := int8(0); x < int8(bin&0xf); x++ {
			r0v, _ := m.GetRegister(0)
			r1v, _ := m.GetRegister(1)
			r2v, _ := m.GetRegister(2)
			r3v, _ := m.GetRegister(3)

			_ = m.SetMemory(r0v, uint8(r2v))
			_ = m.SetRegister(uint64(2), uint64(uint8(int8(r2v)+int8(r3v))))
			_ = m.SetRegister(uint64(0), uint64(uint8(int8(r0v)+int8(r1v))))
		}
		return true, nil, nil
	case 0x6:
		*m.PC() = *m.PC() + 1
		for x := int8(0); x < int8(bin&0xf); x++ {
			r0v, _ := m.GetRegister(0)
			memr0, _ := m.GetMemory(r0v)
			ext.auxRegisters[0] = memr0

			r1v, _ := m.GetRegister(1)
			memr1, _ := m.GetMemory(r1v)
			ext.auxRegisters[1] = memr1

			rx := ext.auxRegisters[0]
			ry := ext.auxRegisters[1]

			_ = m.SetRegister(uint64(3), uint64(uint8(int8(rx)+int8(ry))))

			r2v, _ := m.GetRegister(2)
			r3v, _ := m.GetRegister(3)

			_ = m.SetMemory(r2v, uint8(r3v))

			for x := 0; x < 4; x++ {
				regv, _ := m.GetRegister(uint64(x))
				_ = m.SetRegister(uint64(x), uint64(uint8(int8(regv)+int8(1))))
			}
		}
		return true, nil, nil
	case 0x7:
		ra := (bin >> 2) & 0x3
		if ra == 0 {
			for x := 0; x < 4; x++ {
				regv, _ := m.GetRegister(uint64(x))
				_ = m.SetRegister(uint64(x), uint64(uint8(int8(regv)+int8(bin&0x3))))
			}
			break
		}

		rav, _ := m.GetRegister(uint64(ra))
		_ = m.SetRegister(uint64(ra), uint64(uint8(int8(rav)+int8(bin&0x3))))
		*m.PC() = *m.PC() + 1
		return true, nil, nil
	case 0xa:
		if bin == 0xa0 {
			*m.PC() = *m.PC() + 1
			return true, &machine.Call{
				Number: machine.SYS_BREAK,
			}, nil
		}
	case 0xb:
		if bin == 0xb0 {
			*m.PC() = *m.PC() + 1
			r0v, _ := m.GetRegister(0)
			r1v, _ := m.GetRegister(1)
			r2v, _ := m.GetRegister(2)
			return true, &machine.Call{
				Number: r0v,
				Arg1:   r1v,
				Arg2:   r2v,
			}, nil
		}
	}

	return false, nil, nil
}

func reduxKAssembleExtension(t assembler.ResolvedToken) (uint8, error) {
	switch string(t.Value) {
	case "inc":
		return assembleInc(t)
	case "loadv", "addv":
		return assembleV(t)
	case "ebreak":
		return 0xa0, nil
	case "ecall":
		return 0xb0, nil
	}

	return 0, fmt.Errorf(machine.InterCtx.Get("Unknown instruction: %v"), string(t.Value))
}

func assembleInc(t assembler.ResolvedToken) (uint8, error) {
	if len(t.Args) != 2 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), t.Value)
	}

	ra := uint8(t.Args[0]) & 0x3
	uimm := uint8(t.Args[1]) & 0x3

	return (0x7 << 4) | (ra << 2) | uimm, nil
}

func assembleV(t assembler.ResolvedToken) (uint8, error) {
	if len(t.Args) != 1 {
		return 0, fmt.Errorf(machine.InterCtx.Get("wrong number of arguments for instruction '%s', expected 1 argument"), t.Value)
	}

	sizev := uint8(t.Args[0]) & 0xf
	var op uint8
	switch string(t.Value) {
	case "loadv":
		op = 0x5
	case "addv":
		op = 0x6
	default:
		return 0, errors.New("unreachable")
	}

	return (op << 4) | sizev, nil
}

func ReduxK() *reduxc.ReduxC {
	return reduxc.ReduxVExtension(
		"REDUX-K",
		reduxKExecuteExtension,
		reduxKAssembleExtension,
		&ReduxKExtState{},
	)
}
