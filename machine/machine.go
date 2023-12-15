// Package egg/machine has the interface and syscall struct used by EGG backends.
package machine

// Interface that machine structs are required to implement to work with EGG.
//
// It always use uint64 because if we used variable-lengh []bytes, they would
// need to be stored on the heap. Converting from any size less than and to 64
// bits can be done with a single x86 instruction.
type Machine interface {
	LoadProgram([]uint8) error
	NextInstruction() (*Call, error)
	GetMemory(uint64) (uint8, error)
	SetMemory(uint64, uint8) error
	// Address than size.
	GetMemoryChunk(uint64, uint64) ([]uint8, error)
	SetMemoryChunk(uint64, []uint8) error
	GetRegister(uint64) (uint64, error)
	// Address than content.
	SetRegister(uint64, uint64) error
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

type Call struct {
	Number uint64
	Arg1 uint64
	Arg2 uint64
}
