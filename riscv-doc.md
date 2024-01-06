[Ler esse documento em português](riscv-doc-pt.md)

# RISC-V IM 32 bits for EGG

The RISC-V implementation for EGG uses the standard Assembly syntax described in
the [README](README.md). It implements the registers and instructions from the
base integer set and the multiplication extension described in
[this](riscv/riscv.pdf) document.

Environment call numbers are placed in the `a7` registers, and arguments are
placed in `a0` and `a1`.

An `ebreak` instruction will perform a BREAK call.

Example program (writes "Hello, World!", breaks a line and exits):

```asm
	addi a7, zero, 3
	addi a0, zero, msg
	addi a1, zero, 14
	ecall
	ebreak

msg:
#Hello, World!%0a
```

The assembled program is loaded at the address `0` on the machine startup, and
the `pc` register is set to `0`. The stack pointer register is not initialized:
the program should initialize it if it wants to use the stack.
