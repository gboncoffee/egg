package main

type page struct {
	min  uint64
	next *page
	mem  []byte
}

type Memory struct {
	first    *page
	pagesize uint64
}

type Machine interface {
	CreateMachine() *Machine
	NextInstruction() error
	GetMemory(uint64) []byte
	SetMemory(uint64, []byte)
	GetRegister(uint64) []byte
	SetRegister(uint64, []byte)
}

func newPage(min uint64, size uint64, next *page) *page {
	p := page{min: min, next: next, mem: make([]byte, size)}

	return &p
}

func NewMemory(pagesize uint64) *Memory {
	mem := Memory{first: newPage(0, pagesize, nil), pagesize: pagesize}

	return &mem
}

// We need to convert everything to uint64 and it becomes a messy type
// confusion.
func fetchBytes(addr uint64, idx uint64, arr *[]byte, p *page) {
	if idx >= uint64(len(*arr)) {
		return
	}

	var i uint64 = 0
	for p != nil && i < uint64(len(*arr)) && p.min < addr+uint64(len(*arr)) {
		if (addr-p.min)+i < 0 {
			return
		}

		(*arr)[i] = p.mem[((addr + i) - p.min)]

		i++
		if addr+i > p.min+uint64(len(p.mem)) {
			p = p.next
			i += p.min - (addr + i)
		}
	}
}

func (m *Memory) Get(n_bytes uint64, addr uint64) []byte {
	arr := make([]byte, n_bytes)
	p := m.first
	for p != nil {
		if addr >= p.min && addr < uint64(len(p.mem))+p.min {
			fetchBytes(addr, 0, &arr, p)
			return arr
		}
		p = p.next
	}

	return arr
}

func (m *Memory) Set(n_bytes uint64, addr uint64, mem []byte) {
}
