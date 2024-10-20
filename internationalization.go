package main

const MASSIVE_HELP_STRING = `Commands:

help
	Shows this help.
	Shortcut: h
print <expr>[@<length>]
	Prints the content of registers and memory.
	Shortcut: p
printall
	Prints the content of all registers.
	Shortcut: pall
next
	Executes the next instruction, then pauses.
	Shortcut: n
continue
	Continue execution until a BREAK call or breakpoint.
	Shortcut: c
break [expr]
	With an argument, creates a new breakpoint. With no argument, shows all
	breakpoints. Accepts numbers, Assembly labels and <file>:<line>.
	Shortcut: b
remove <expr>
	Removes a breakpoint. Accepts numbers and Assembly labels.
	Shortcut: r
dump <address>@<length> <filename>
	Dumps the content of memory to a file.
	Shortcut: d
rewind
	Reloads the machine, i.e., asks it to return to it's original state.
	Shortcut: rew
reload
	Reload the Assembly files and them reloads the machine.
	Shortcut: rel
set <expr>[@<length>] <content>
	Changes the content of a register or memory.
	Shortcut: s
exit
	Terminate debugging session.
	Shortcut: e
quit
	Alias to exit.
	Shortcut: q

The print command generally follows this rules:
- If the expression is only a register (e.g., x1, t1, zero, ra, etc), it prints
  it's contents;
- If the expression is a register with a length (e.g., t1@1, ra@7, etc), it
  prints the content of the memory addressed by the content of the register.
The set command works the same way.

The dump command also accepts registers, but always dereference them.

Both print and dump commands accepts the special expression #, which means the
program itself. For example, you may dump the assembled program to a file by
running 'dump # file'.

In the print command, <addr>#<length> means "length instructions after addr".
#<length> is a shortcut to use with the current instruction address.
`

var brazilian = map[string]string{

	//
	// main.go.
	//

	// Infos.
	`Currently supported architetures:
'riscv' - RISC-V IM, 32 bits
'mips'  - Subset of MIPS32
'sagui' - Fantasy 8 bit RISC`: `Arquiteturas suportadas:
'riscv' - RISC-V IM, 32 bits
'mips'  - Subconjunto de MIPS32
'sagui' - RISC fantasia de 8 bits`,
	"EGG - Emulador Genérico do Gabriel - version ": "EGG - Emulador Genérico do Gabriel - versão ",
	// Main execution loop.
	"Instruction execution failed: %v\n": "Falha na execução da instrução: %v\n",
	// Args.
	"Select architeture to use.":                          "Seleciona a arquitetura a ser utilizada.",
	"Select architeture to use (shorthand).":              "Seleciona a arquitetura a ser utilizada (abrev.).",
	"Lists currently supported architetures and quit.":    "Lista arquiteturas suportadas e sai.",
	"Lists currently supported architetures (shorthand).": "Lista arquiteturas suportadas e sai (abrev.).",
	"Show current version and quit.":                      "Mostra a versão atual e sai.",
	"Show current version and quit (shorthand).":          "Mostra a versão atual e sai (abrev.).",
	"Enter debugger upon startup.":                        "Entra no debugger após inicialização.",
	"Enter debugger upon startup (shorthand).":            "Entra no debugger após inicialização (abrev).",
	// main().
	"Unknown architeture: %v\n":                            "Arquitetura desconhecida: %v\n",
	"No Assembly file supplied.":                           "Nenhum arquivo Assembly providenciado.",
	"Could not read supplied file %v\n":                    "Erro lendo o arquivo %v providenciado",
	"Error assembling file %v: %v\n":                       "Erro montando o arquivo %v: %v\n",
	"Error loading assembled program: %v\n":                "Erro carregando o programa montado: %v\n",
	"Debugging is not supported for the selected backend.": "Debugging não é suportado pelo backend selecionado.",

	//
	// debugger.go.
	//

	// Massive help string.
	MASSIVE_HELP_STRING: `Comandos:

help
	Mostra esse texto de ajuda.
	Abreviação: h
print <expr>[@<tamanho>]
	Imprime o conteúdo de registradores e da memória.
	Abreviação: p
printall
	Imprime o conteúdo de todos os registradores.
	Abreviação: pall
next
	Executa a próxima instrução, então pausa.
	Abreviação: n
continue
	Continua a execução até uma chamada BREAK ou um ponto de parada
	(breakpoint).
break [expr]
	Com um argumento, cria um novo ponto de parada (breakpoint). Sem o
	argumento, mostra todos os pontos de parada. Aceita números, etiquetas Assembly e <arquivo>:<linha>.
	Abreviação: b
remove <expr>
	Remove um ponto de parada (breakpoint). Aceita números e etiquetas Assembly.
	Abreviação: r
dump <endereço>@<tamanho> <arquivo>
	Salva conteúdo da memória em um arquivo.
	Abreviação: d
rewind
	Recarrega a máquina, isso é, pede para que ela retorne ao estado original.
	Abreviação: rew
reload
	Recarrega os arquivos Assembly e então recarrega a máquina.
	Abreviação: rel
set <expr>[@tamanho] <conteúdo>
	Muda o conteúdo de registradores ou da memória.
	Abreviação: s
exit
	Termina a sessão de debugging.
	Abreviação: e
quit
	Igual à exit.
	Abreviação: q

O comando print, de maneira geral, segue as seguintes regras:
- Se a expressão é somente um registrador (por exemplo, x1, t1, zero, ra, etc),
  imprime seu conteúdo;
- Se a expressão é um registrador com um tamanho (por exemplo, x1@1, ra@7, etc),
  imprime o conteúdo da memória endereçada pelo conteúdo do registrador.
O comando set funciona da mesma maneira.

O comando dump também aceita registradores, porém sempre os utiliza como
endereços.

Tanto print quanto dump também aceitam a expressão especial #, que representa o
programa em si. Por exemplo, é possível salvar o programa montado em um arquivo
com o comando 'dump # arquivo'.

No comando print, <addr>#tamanho significa "tantas instruções após addr".
#<tamanho> é uma abreviação para usar o endereço da instrução atual.
`,
	// Debugger functions.
	"cannot parse %v as number: %v":                                        "impossível converter %v para número",
	"cannot parse %v as number: %v\n":                                      "impossível converter %v para número\n",
	"length not supplied":                                                  "tamanho não providenciado",
	"cannot parse %v as address":                                           "impossível converter %v para endereço",
	"%v is not a number":                                                   "%v não é um número",
	"%v is not an unsigned number":                                         "%v não é um número sem sinal",
	"no instruction at address 0x%x":                                       "nenhuma instrução no endereço 0x%x",
	"%v is not a register or address":                                      "%v não é um registrador ou endereço",
	"cannot get memory content: %v":                                        "não foi possível ler o conteúdo da memória: %v",
	"cannot get register content: %v":                                      "não foi possível ler o conteúdo do registrador: %v",
	"print expects one argument: <expr>[@<length>] or [<addr>]#[<length>]": "print necessita de um argumento: <expr>[@<tamanho>] ou [<addr>]#[<tamanho>]",
	"READ call for address 0x%x with %d bytes:\n":                          "Chamada READ para o endereço 0x%x com %d bytes:\n",
	"Error reading stdin: %v\n":                                            "Erro lendo input padrão: %v\n",
	"Register %v: changed from 0x%02x to 0x%02x\n":                         "Registrador %v: mudou de 0x%02x para 0x%02x\n",
	"Register %v: changed from 0x%04x to 0x%04x\n":                         "Registrador %v: mudou de 0x%04x para 0x%04x\n",
	"Register %v: changed from 0x%08x to 0x%08x\n":                         "Registrador %v: mudou de 0x%08x para 0x%08x\n",
	"Register %v: changed from 0x%016x to 0x%016x\n":                       "Registrador %v: mudou de 0x%016x para 0x%016x\n",
	"BREAK call while stepping at address 0x%x\n":                          "Chamada BREAK enquanto executando o endereço 0x%x\n",
	"Stopped at BREAK call at address 0x%x\n":                              "Parado na chamada BREAK no endereço 0x%x\n",
	"Breakpoints:":                       "Pontos de parada (breakpoints):",
	"%v is not a number.\n":              "%v não é um número.\n",
	"Breakpoint already exists":          "Ponto de parada (breakpoint) já existe.",
	"Cannot parse %v as address.\n":      "Impossível converter %v para endereço.\n",
	"cannot parse %v as a dump argument": "impossível converter %v para um argumento de dump",
	"error getting memory chunk: %v":     "erro lendo região da memória: %v",
	"dump expects two arguments: (<expr>@<length> or [<addr>]#[<length>]) <file>": "dump necessita de dois argumentos: (<expr>@<tamanho> ou [<addr>]#[<tamanho>]) <arquivo>",
	"Cannot get content to dump: %v\n":                                            "Não foi possível ler o conteúdo para o dump: %v\n",
	"Cannot open %s for write: %v\n":                                              "Não foi possível abrir %s para escrita: %v\n",
	"Error while writing to %s: %v\n":                                             "Erro enquanto escrevendo para %s: %v\n",
	"Error while reloading machine: %v\n":                                         "Erro enquanto recarregando a máquina: %v\n",
	"Reloaded machine.":                                                           "Máquina recarregada.",
	"cannot get register number: %v":                                              "não foi possível encontrar o número do registrador: %v",
	"set expects two arguments: <expr>[@<length>] <value>":                        "set necessita de dois argumentos: <expr>[@<tamanho>] <valor>",
	"Cannot parse %v as number: %v":                                               "Impossível converter %v para número: %v",
	"Error while changing register content: %v\n":                                 "Erro modificando o conteúdo do registrador: %v\n",
	"Error while changing memory content: %v\n":                                   "Erro modificando o conteúdo da memória",
	"Type 'help' for a list of commands.":                                         "Entre 'help' para ver uma lista de comandos",
	"Debugging":                                                                   "Debuggando",
	"No such command: %v\n":                                                       "Comando inexistente: %v\n",
	"bye!":                                                                        "até mais!",
	// Version 3 updates
	"(Label %v) ":                   "(Etiqueta %v) ",
	"cannot parse %v as breakpoint": "impossível converter %v para ponto de parada (breakpoint)",
	"Stopped at breakpoint: %v\n":   "Parado no ponto (breakpoint): %v\n",
	"New breakpoint %v\n":           "Novo ponto de parada (breakpoint) %v\n",
	"remove expects a breakpoint to remove: remove <address/label/file:line>": "remove necessita de um ponto (breakpoint) para remover: remove <endereço/etiqueta/arquivo:linha>",
	"No breakpoint %v\n":                     "Nenhum ponto de parada (breakpoint) %v\n",
	"Error assembling file:":                 "Erro montando arquivo:",
	"Keeping old program and program state.": "Mantendo o programa e estado anteriores",
	"Error loading new assembled code:":      "Erro carregando o novo código montado:",
	"Rebuild Assembly.":                      "Assembly remontado.",

	//
	// assembler.go and tokenizer.go
	//
	"If you're reading this, there's a bug in the emulator. Please fill an issue at https://github.com/gboncoffee/egg reporting the bug with the Assembly you're trying to run and command line arguments you used to run EGG.": "Se você está lendo isso, há um bug no emulador. Por favor preencha uma issue em https://github.com/gboncoffee/egg reportando o bug com o Assembly que você está tentando rodar e os argumentos de linha de comando que você usou para rodar o EGG.",
	"%v:%v: Error on argument translation: %v":                  "%v:%v: Erro na tradução de argumento: %v",
	"couldn't open file: %v":                                    "não foi possível abrir o arquivo: %v",
	"error reading file %v: %v":                                 "erro lendo o arquivo %v: %v",
	"%v:%v: Expected literal bytes after bits directive":        "%v:%v: Bytes literais esperados após a diretiva de bits",
	"%v:%v: Cannot convert %v to a %v bits number":              "%v:%v: Não foi possível converter %v para um número de %v bits",
	"%v:%v: Expected a number of bytes after space directive":   "%v:%v: Número de bytes esperado após a diretiva de espaço",
	"%v:%v: Cannot create space: Cannot parse %v to number: %v": "%v:%v: Impossível criar espaço: Impossível traduzir %v para número: %v",
	"%v:%v: Expected a directive name":                          "%v:%v: Nome de diretiva esperado",
	"%v:%v: Expected file name to include":                      "%v:%v: Nome de arquivo para inclusão esperado.",
	"%v:%v: Unknown directive %v":                               "%v:%v: Diretiva desconhecida %v",
	"%v:%v: Expected literal content":                           "%v:%v: Conteúdo literal esperado",

	//
	// riscv.go and others.
	//
	"unknown opcode: %b":                                                   "opcode desconhecido: %b",
	"could not load 4 bytes from address at PC: %x":                        "não foi possível carregar 4 bytes do endereço do PC: %x",
	"value %v bigger than maximum 32 bit address %v":                       "valor %v maior que o máximo endereço de 32 bits %v",
	"end address %v bigger than maximum 32 bit address %v":                 "endereço final %v maior que o máximo endereço de 32 bits %v",
	"no such register: %d. RISC-V has only 32 registers":                   "registrador %d inexistente. RISC-V possui somente 32 registradores",
	"wrong number of arguments for instruction '%s', expected 3 arguments": "número de argumentos para instrução '%s' errado: 3 argumentos esperados",
	"wrong number of arguments for instruction '%s', expected 2 arguments": "número de argumentos para instrução '%s' errado: 2 argumentos esperados",
	"wrong number of arguments for instruction '%s', expected no argument": "número de argumentos para instrução '%s' errado: nenhum argumento esperado",
	"unknown instruction: %v":                                              "instrução desconhecida: %v",
	"no such register: %v":                                                 "registrador inexistente: %v",
	"empty argument":                                                       "argumento vazio",
	"error executing sra: negative shift amount":                           "erro executando sra: tamanho de shift negativo.",
	// Version 3 updates.
	"%v:%v: Error assembling: %v": "%v:%v: Erro montando: %v",

	//
	// mips.go specific.
	//
	"no such register: %d. MIPS-I has only 32 general purpouse registers and two special registers for multiplication and division (HI and LO, 32 and 33)": "registrador inexistente: %d. MIPS-I possui somente 32 registradores de propósito geral e dois registradores especiais para multiplicação e divisão (HI e LO, 32 e 33)",
	// This one is used by sagui also.
	"wrong number of arguments for instruction '%s', expected 1 argument": "número de argumentos para instrução '%s' errado: 1 argumento esperado",

	//
	// sagui.go specific.
	//
	"value %v is bigger than maximum 8 bit address %v":    "valor %v maior que o máximo endereço de 8 bits %v",
	"end address %v bigger than maximum 8 bit address %v": "endereço final %v maior que o máximo endereço de 8 bits %v",
	"failed to fetch instruction from memory: %v":         "falha lendo a instrução da memória: %v",
	"immediate bigger than immediate size: %v":            "imediato maior que o tamanho do imediato: %v",
}
