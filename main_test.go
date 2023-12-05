package main

import "testing"

func TestCreateMemory(t *testing.T) {
	mem := NewMemory(1024)
	if mem.pagesize != 1024 {
		t.Fatalf("Incorrect pagesize")
	}

	if mem.first == nil {
		t.Fatalf("mem.first is nil")
	}

	if mem.first.next != nil {
		t.Fatalf("mem.first.next points somewhere")
	}

	if mem.first.min != 0 {
		t.Fatalf("Memory was not iniciated to the beggining")
	}

	if len(mem.first.mem) != 1024 {
		t.Fatalf("Memory buffer was not iniciated with 1024 positions")
	}
}

func TestMemoryAccess(t *testing.T) {
	// Small page size so I don't have to type a lot of random numbers.
	mem := NewMemory(4)

	mem.Set(4, 2, []byte{1, 2, 3, 4})

	arr := mem.Get(4, 1)
	for i, v := range []byte{1, 2, 3, 4} {
		if v != arr[i] {
			t.Fatalf("Wrong byte %d at position %d", arr[i], i+2)
		}
	}

	arr = mem.Get(4, 0)
	for i, v := range []byte{0, 0, 1, 2} {
		if v != arr[i] {
			t.Fatalf("Wrong byte %d at position %d", arr[i], i)
		}
	}

	arr = mem.Get(4, 3)
	for i, v := range []byte{3, 4, 0, 0} {
		if v != arr[i] {
			t.Fatalf("Wrong byte %d at position %d", arr[i], i+4)
		}
	}
}
