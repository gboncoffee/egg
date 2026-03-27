// Package reduxv implements a REDUX-V machine for EGG.
//
// Opcode 0101 (ebreak) performs a BREAK and 0110 (ecall) performs a CALL.
package reduxv

import (
	"fmt"
	"github.com/gboncoffee/egg/assembler"
	"github.com/gboncoffee/egg/machine"
	"github.com/gboncoffee/egg/reduxc"
)

func reduxVExecuteExtension(bin uint8, m *reduxc.ReduxC) (bool, *machine.Call, error) {
	op := bin >> 4
	switch op {
	case 0x5:
		*m.PC() = *m.PC() + 1
		return true, &machine.Call{
			Number: machine.SYS_BREAK,
		}, nil
	case 0x6:
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
	return false, nil, nil
}

func reduxVAssembleExtension(t assembler.ResolvedToken) (uint8, error) {
	switch string(t.Value) {
	case "ebreak":
		return 0x50, nil
	case "ecall":
		return 0x60, nil
	}

	return 0, fmt.Errorf(machine.InterCtx.Get("unknown instruction: %v"), string(t.Value))
}

func ReduxV() *reduxc.ReduxC {
	return reduxc.ReduxVExtension(
		"REDUX-V",
		reduxVExecuteExtension,
		reduxVAssembleExtension,
		nil,
	)
}
