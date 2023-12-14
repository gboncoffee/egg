// Package egg/machine has the interface used by EGG backends.
package machine

// Interface that machine structs are required to implement to work with EGG.
//
// It always use uint64 because if we used variable-lengh []bytes, they would
// need to be stored on the heap. Converting from any size less than and to 64
// bits can be done with a single x86 instruction.
type Machine interface {
	LoadProgram([]uint8) error
	NextInstruction() error
	GetMemory(uint64) (uint8, error)
	SetMemory(uint64, uint8) error
	// Address than size.
	GetMemoryChunk(uint64, uint64) ([]uint8, error)
	SetMemoryChunk(uint64, []uint8) error
	GetRegister(uint64) (uint64, error)
	// Address than content.
	SetRegister(uint64, uint64) error
}
