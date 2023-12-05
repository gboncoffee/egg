package main

// Interface that machine structs are required to implement to work with EGG.
//
// It always use uint64 because if we used variable-lengh []bytes, they would
// need to be stored on the heap. Converting from any size less than and to 64
// bits can be done with a single x86 instruction.
type Machine interface {
	LoadProgram([]bytes) error
	NextInstruction() error
	GetMemory(uint64) ([]bytes, error)
	SetMemory(uint64, []bytes) error
	GetRegister(uint64) (uint64, error)
	SetRegister(uint64, uint64) error
}
