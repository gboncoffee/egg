// Package egg/machine has the interface and syscall struct used by EGG backends.
package machine

import (
	"github.com/gboncoffee/egg/assembler"
	"github.com/gboncoffee/intergo"
)

type ArchitectureInfo struct {
	Name           string
	RegistersNames []string
	WordWidth      uint16
}

// Interface that machine structs are required to implement to work with EGG.
//
// It always use uint64 because if we used variable-lengh []bytes, they would
// need to be stored on the heap. Converting from any size less than and to 64
// bits can be done with a single x86 instruction.
type Machine interface {
	// Generic function for loading the program into memory and initializing
	// anything the machine may need. After this, the machine should be
	// ready to receive sequences of NextInstruction().
	LoadProgram([]uint8) error
	// Executes the next instruction. The execution should be handled in a
	// way that the instruction is executed and only them the instruction
	// pointer is incremented, thus the pointer always points to the
	// instruction to be performed next.
	NextInstruction() (*Call, error)
	// Retrieves a single byte from memory.
	GetMemory(uint64) (uint8, error)
	// Sets a single byte of memory.
	SetMemory(uint64, uint8) error
	// Retrieves multiple bytes from memory. First argument is the address
	// and second argument is the size. Should return an error if
	// address+size is not addressable.
	GetMemoryChunk(uint64, uint64) ([]uint8, error)
	// Same as GetMemoryChunk but for setting.
	SetMemoryChunk(uint64, []uint8) error
	// Gets the content of a register. When the ISA has registers with
	// ambiguous numbers, the backend should define unambiguous numbers for
	// them and only translate these numbers when needed on the instruction
	// execution. Example: a RISC-V backend with the floating-point
	// extension may define the floating-point registers as registers 32-63.
	GetRegister(uint64) (uint64, error)
	// Same as GetRegister but for setting. First argument is the register
	// number, and second is the content. This may fail silently if trying
	// to write to a forbidden register (as with x0 in RISC-V), so the
	// caller should not expected it to work always.
	SetRegister(uint64, uint64) error
	// Translates a register name (as used in Assembly) to it's number.
	// Registers may have more than a name (example: in the RISC-V backend,
	// both "zero" and "x0" translates to the register number 0)
	GetRegisterNumber(string) (uint64, error)
	// Assembles a program. Usually a bunch of calls to functions on the
	// egg/assembler package. If the second return value is nil, debugger
	// support is disabled.
	Assemble(string) ([]uint8, []assembler.DebuggerToken, error)
	// Self-explanatory. Usually just a "return m.pc" or something like
	// that.
	GetCurrentInstructionAddress() uint64
	// Self-explanatory.
	ArchitectureInfo() ArchitectureInfo
}

// Syscalls numbers. ISAs with specific calls for BREAK should send a BREAK on them.
//
// BREAK - 1 - Transfer control to debugger or stop machine.
// - No arguments.
//
// READ - 2 - Read input.
// - Arg1: Buffer address (will be put into SetMemoryChunk()).
// - Arg2: Size in bytes of input.
//
// WRITE - 3 - Write output.
// - Arg1: Buffer address (will be put into GetMemoryChunk()).
// - Arg2: Size in bytes of output.
const (
	SYS_BREAK = 1
	SYS_READ  = 2
	SYS_WRITE = 3
)

// Struct returned from NextInstruction() when a call is performed.
type Call struct {
	Number uint64
	Arg1   uint64
	Arg2   uint64
}

// Internationalization context available for architectures.
var InterCtx intergo.InterContext
