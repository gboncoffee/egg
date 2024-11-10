[![Go](https://github.com/gboncoffee/egg/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/gboncoffee/egg/actions/workflows/go.yml)
[![CodeQL](https://github.com/gboncoffee/egg/actions/workflows/github-code-scanning/codeql/badge.svg?branch=master)](https://github.com/gboncoffee/egg/actions/workflows/github-code-scanning/codeql)

[Ler esse documento em português](README-pt.md)

# EGG, a generic processor emulator

[Video tutorial](https://youtu.be/RtKvvXgdVak?si=OtjkNxSGmRGNou67)

[Documentation](https://egg.gboncoffee.dev.br/docs/en)

[Contributing, bugs, feature requests](CONTRIBUTING.md)

EGG stands for "Emulador Genérico do Gabriel" ("Gabriel's Generic Emulator", in
portuguese). It's a modular emulator for processor architetures, made for
educational purpouses.

The `egg` package itself provides only an interface for interacting with
machines, thus supporting different architeture backends. Currently, there are
RISC-V IM 32 bits, MIPS32 and a fantasy 8 bit architecture "Sagui" backends
implemented.

`egg/assembler` also provides a small library for creating assemblers, and the
support for EGG's debugger.

### UFPR students

Have any questions, or want some help? Mail me: `ggb23@inf.ufpr.br`. Or find me
at the campus and the laboratories!

## Installation

Simply grab the static binary for your OS at the
[releases](https://github.com/gboncoffee/egg/releases) page. Or, if you have the
Go compiler, build the project from source.

Note: the Windows binary is untested, as I don't have access to any Windows
machine nowadays. The testing of the Darwin (MacOS) binary depends on my
patience.

## Quickstart

EGG is a command line application, you must run it from a terminal emulator.

Running the emulator with an Assembly file will assemble it and start a machine
to run it on. By default, the machine is a RISC-V IM 32 bits. Use the flag `-a`
or `-arch` to change the architeture. Run `egg -h` to see all command line
options and `egg -l` to see all supported architetures.

The Assembly syntax is architeture-dependent. Though, a library is provided for
creating assembler, so backends may use the same overall syntax (both RISC-V and
MIPS uses it).

An example follows:

```asm
; Semicolon makes a comment til the end of line.

; A label is defined with :.
label:
	; Instruction arguments starts with destination.
	addi t0, zero, 2

	; You may also put instructions after the labels.
label2:	add t0, t0, t0

	; There's no parenthesis as in RARS, store uses common immediates.
	sb t0, ra, 3

	; Hex, octal and binary immediates are supported.
	addi t1, zero, 0xff
	addi t1, zero, 0b010110
	addi t1, zero, 0o644
	addi t1, zero, 0755	; A leading 0 also defines an octal.

; A # defines a literal til the end of the line.
; Literals are inserted unchanged to the binary (as 'db' in other assemblers).
; If a % is followed by two hex digits, the hex value is inserted instead. Use
; %% to escape it.
msg:
#Hello, World!%0a

; Some directives are supported. "bitsxx" ones creates literal numbers in the
; code, with the bit length specified:
.bits8 0xca 0xfe 0xba 0xbe 0xde 0xad 0xbe 0xef
.bits16 0xcafe 0xbabe 0xdead 0xbeef
.bits32 0xcafebabe 0xdeadbeef
.bits64 0xcafebabedeadbeef
; The "space" directive adds some bytes of spacing in the code:
.space 16
; The "include" directive "copy-pastes" another file in the code:
.include other-asm.asm
```

Each architeture folder has test Assembly files you may use as examples.

## Calls

Standard calls handled by the emulator are as follows. Refer to the architeture
documentation on how to perform them:

- BREAK (Number 1): Transfer control to debugger or stop the machine.  
  No arguments.  
- READ (Number 2): Read input.  
  - Argument 1: Buffer address.  
  - Argument 2: Size of input in bytes.  
- WRITE (Number 3): Write output.  
  - Argument 1: Buffer address.  
  - Argument 2: Size of output in bytes.  

## Debugger

The debugger interface is kinda similar to `gdb`, though much smaller. Use the
`-d` flag to enter debugger uppon startup. There's no need of `run` command, as
there's no process. The `next` or the `continue` commands may be used to start
running the program normally. Use `help` to see all available commands.
