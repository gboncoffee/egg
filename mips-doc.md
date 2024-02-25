[Ler esse documento em português](mips-doc-pt.md)

# MIPS32 for EGG

The MIPS32 implementation for EGG uses the standard Assembly syntax described in
the [README](README.md).

Environment call numbers are placed in the `v0` register, and arguments are
placed in `a0` and `a1`.

An `break` instruction will perform a BREAK call.

The assembled program is loaded at the address `0` on the machine startup, and
the `pc` register is set to `0`.
