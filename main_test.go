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
