[Ler esse documento em português](CONTRIBUINDO.md)

# I found a bug!

Please [fill an issue](https://github.com/gboncoffee/egg/issues) describing your
bug if none is already opened.

# Feature requests

Please [fill an issue](https://github.com/gboncoffee/egg/issues) describing your
feature request if none is already opened. Features can be anything from a new
syscall, a new debugger command, or a new architeture backend. Please remind
that the purpouse of this project is to be simple, so stuff like a `x86_64`
backend will never happen.

# I would like to write code for it

Well, first thanks! The codebase is very simple and navigatable. Enjoy your
time, and if you have any questions,
[fill an issue](https://github.com/gboncoffee/egg/issues).

To contribute with code, open a pull request!

## Adding backends

Create a new Go package at the project directory. For example, let's add a
backend for the fictional R2D2 instruction set architeture:

```shell
$ cd egg
$ mkdir r2d2
$ cd r2d2
$ touch r2d2.go
$ touch r2d2_test.go
```

We'll implement everything in `r2d2.go`:

```go
// Package egg/r2d2 implements a R2D2 machine for the EGG emulator.
package r2d2

import (
	"github.com/gboncoffee/egg/assembler"
	"github.com/gboncoffee/egg/machine"
)

// The R2D2 struct implements the machine interface needed by the EGG emulator.
type R2D2 struct {
	registers [32]uint32
	pc        uint32
	mem       [math.MaxUint32 + 1]uint8
}
```

The required functions are in the interface `Machine` imported via the
`egg/machine` package. They should be implemented as methods on pointers rather
than on the struct itself.

Most functions there are rather simple. The documentation can be found
[at pkg.go.dev](https://pkg.go.dev/github.com/gboncoffee/egg).

The `egg/assembler` package is also imported because it has the definitions
required for the debugger support. The interface requires the implementation of
the assembler function `Assemble(string) ([]uint8, []assembler.DebuggerToken,
error)`. If it's second return value is `nil`, debugger support is
disabled. Furthermore, this package helps with the boilerplate of creating a
simple assembler.
