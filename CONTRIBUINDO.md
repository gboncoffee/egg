[Read this document in English](CONTRIBUTING.md)

# Achei um bug!

Por favor, [crie uma issue](https://github.com/gboncoffee/egg/issues)
descrevendo-o caso uma ainda não exista.

# Requisição de features

Por favor, [crie uma issue](https://github.com/gboncoffee/egg/issues)
descrevendo a feature desejada caso alguma ainda não exista. Uma nova feature
pode ser qualquer coisa desde uma nova _syscall_, um novo comando para o
debugger, ou um novo backend. Por favor note que o propósito desse projeto é ser
simples, então coisas como um backend de `x86_64` nunca vão ocorrer.

# Eu gostaria de contribuir com código

Bem, muito obrigado! A base de código é bem simples e navegável. Divirta-se, e
se tiver alguma dúvida,
[crie uma issue](https://github.com/gboncoffee/egg/issues).

Para contribuir com código, abra uma pull request!

## Adicionando backends

Crie um novo pacote Go no diretório do projeto. Por exemplo, vamos adicionar um
backend para a arquitetura fictícia R2D2:

```shell
$ cd egg
$ mkdir r2d2
$ cd r2d2
$ touch r2d2.go
$ touch r2d2_test.go
```

Implementaremos tudo no arquivo `r2d2.go`:

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

As funções necessárias estão na interface `Machine` importada pelo pacote
`egg/machine`. Elas devem ser implementadas como métodos em ponteiros e não na
struct em si.

A maioria das funções é bem simples. A documentação pode ser encontrada
[no site pkg.go.dev](https://pkg.go.dev/github.com/gboncoffee/egg).

O pacote `egg/assembler` também é importado porque possui definições necessárias
para o suporte ao debugger. A interface requer a implementação da função
`Assemble(string) ([]uint8, []assembler.DebuggerToken, error)`. Se o segundo
valor de retorno for `nil`, o suporte ao debugger é desativado. Além disso, esse
pacote facilita a criação de assemblers simples.
