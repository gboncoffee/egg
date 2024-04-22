# RISC-V Cheatsheet - EGG

- [RISC-V Cheatsheet - EGG](#risc-v-cheatsheet---egg)
	- [Instruções](#instruções)
		- [Aritméticas](#aritméticas)
			- [Extensão de multiplicação](#extensão-de-multiplicação)
		- [Aritméticas com imediato](#aritméticas-com-imediato)
		- [Loads (carregar valores da memória)](#loads-carregar-valores-da-memória)
		- [Stores (salvar valores na memória)](#stores-salvar-valores-na-memória)
		- [Branches (saltos condicionais)](#branches-saltos-condicionais)
		- [Jumps (saltos incondicionais)](#jumps-saltos-incondicionais)
		- [Miscelânea](#miscelânea)
	- [Registradores](#registradores)
	- [Sintaxe](#sintaxe)
	- [Chamadas do emulador](#chamadas-do-emulador)
	- [Carregar imediatos](#carregar-imediatos)
	- [Manipular a stack](#manipular-a-stack)
	- [Chamar funções](#chamar-funções)
	- [Truques do debugger](#truques-do-debugger)
		- [Ver o PC](#ver-o-pc)
		- [Dump do programa](#dump-do-programa)

## Instruções

### Aritméticas

| Mnemônico | Argumentos         | Descrição                                                                                                        | Sinal |
| --------- | ------------------ | ---------------------------------------------------------------------------------------------------------------- | ----- |
| `add`     | `rd`, `rs1`, `rs2` | Soma os valores de `rs1` e `rs2`, guardando o resultado em `rd`.                                                 | Com.  |
| `sub`     | `rd`, `rs1`, `rs2` | Subtrai os valores de `rs1` e `rs2`, guardando o resultado em `rd`.                                              | Com.  |
| `xor`     | `rd`, `rs1`, `rs2` | Realiza o OU EXCLUSIVO entre os valores de `rs1` e `rs2`, guardando o resultado em `rd`.                         |       |
| `or`      | `rd`, `rs1`, `rs2` | Realiza o OU lógico entre os valores de `rs1` e `rs2`, guardando o resultado em `rd`.                            |       |
| `and`     | `rd`, `rs1`, `rs2` | Realiza o E lógico entre os valores de `rs1` e `rs2`, guardando o resultado em `rd`.                             |       |
| `sll`     | `rd`, `rs1`, `rs2` | Realiza um shift lógico para a esquerda do valor de `rs1` pelo valor de `rs2`, guardando o resultado em `rd`.    |       |
| `srl`     | `rd`, `rs1`, `rs2` | Realiza um shift lógico para a direita do valor de `rs1` pelo valor de `rs2`, guardando o resultado em `rd`.     |       |
| `sra`     | `rd`, `rs1`, `rs2` | Realiza um shift aritmético para a direita do valor de `rs1` pelo valor de `rs2`, guardando o resultado em `rd`. |       |
| `slt`     | `rd`, `rs1`, `rs2` | Compara os valores de `rs1` e `rs2`. O resultado guardado em `rd` é 1 caso `rs1` < `rs2`, e 0 caso contrário.    | Com.  |
| `sltu`    | `rd`, `rs1`, `rs2` | Igual à `slt`, porém considera os valores sem sinal.                                                             | Sem.  |

#### Extensão de multiplicação

| Mnemônico | Argumentos         | Descrição                                                                                | Sinal  |
| --------- | ------------------ | ---------------------------------------------------------------------------------------- | ------ |
| `mul`     | `rd`, `rs1`, `rs2` | Multiplica os valores de `rs1` e `rs2`, guardando o resultado em `rd`.                   | Com.   |
| `mulh`    | `rd`, `rs1`, `rs2` | Multiplica os valores de `rs1` e `rs2`, guardando a parte alta do resultado em `rd`.     | Com.   |
| `mulsu`   | `rd`, `rs1`, `rs2` | Igual à `mulh`, porém considera o valor de `rs1` COM SINAL e o valor de `rs2` SEM SINAL. | Misto. |
| `mulu`    | `rd`, `rs1`, `rs2` | Igual à `mulh`, porém considera os valores sem sinal.                                    | Sem.   |
| `div`     | `rd`, `rs1`, `rs2` | Divide o valor de `rs1` pelo valor de `rs2`, guardando o resultado em `rd`.              | Com.   |
| `divu`    | `rd`, `rs1`, `rs2` | Igual à `div`, porém considera os valores sem sinal.                                     | Sem.   |
| `rem`     | `rd`, `rs1`, `rs2` | Guarda em `rd` o resto da divisão do valor de `rs1` pelo valor de `rs2`.                 | Com.   |
| `remu`    | `rd`, `rs1`, `rs2` | Igual à `rem`, porém considera os valores sem sinal.                                     | Sem.   |

### Aritméticas com imediato

| Mnemônico | Argumentos         | Descrição                                                                                                    | Sinal |
| --------- | ------------------ | ------------------------------------------------------------------------------------------------------------ | ----- |
| `addi`    | `rd`, `rs1`, `imm` | Soma o valor de `rs1` com o `imm`, guardando o resultado em rd.                                              | Com.  |
| `xori`    | `rd`, `rs1`, `imm` | Realiza o OU EXCLUSIVO entre o valor de `rs1` e o `imm`, guardando o resultado em rd.                        |       |
| `ori`     | `rd`, `rs1`, `imm` | Realiza o OU lógico entre o valor de `rs1` e `imm`, guardando o resultado em rd.                             |       |
| `andi`    | `rd`, `rs1`, `imm` | Realiza o E lógico entre o valor de `rs1` e `imm`, guardando o resultado em rd.                              |       |
| `slli`    | `rd`, `rs1`, `imm` | Realiza um shift lógico para a esquerda do valor de `rs1` pelo `imm`, guardando o resultado em rd.           |       |
| `srli`    | `rd`, `rs1`, `imm` | Realiza um shift lógico para a direita do valor de `rs1` pelo `imm`, guardando o resultado em rd.            |       |
| `srai`    | `rd`, `rs1`, `imm` | Realiza um shift aritmético para a direita do valor de `rs1` pelo `imm`, guardando o resultado em rd.        |       |
| `slti`    | `rd`, `rs1`, `imm` | Compara o valor de `rs1` com o `imm`. O resultado guardado em rd é 1 caso `rs1` < `imm`, e 0 caso contrário. | Com.  |
| `sltiu`   | `rd`, `rs1`, `imm` | Igual à `slt`, porém considera os valores sem sinal.                                                         | Sem.  |

### Loads (carregar valores da memória)

| Mnemônico | Argumentos         | Descrição                                                                 | Sinal |
| --------- | ------------------ | ------------------------------------------------------------------------- | ----- |
| `lb`      | `rd`, `rs1`, `imm` | Carrega 1 byte do endereço dado por `rs1 + imm` na região baixa de `rd`.  | Com.  |
| `lh`      | `rd`, `rs1`, `imm` | Carrega 2 bytes do endereço dado por `rs1 + imm` na região baixa de `rd`. | Com.  |
| `lw`      | `rd`, `rs1`, `imm` | Carrega 4 bytes do endereço dado por `rs1 + imm` no `rd`.                 |       |
| `lbu`     | `rd`, `rs1`, `imm` | Igual à `lb` porém não extende o sinal do valor carregado.                | Sem.  |
| `lhu`     | `rd`, `rs1`, `imm` | Igual à `lh` porém não extende o sinal do valor carregado.                | Sem.  |

### Stores (salvar valores na memória)

| Mnemônico | Argumentos         | Descrição                                                               | Sinal |
| --------- | ------------------ | ----------------------------------------------------------------------- | ----- |
| `sb`      | `rd`, `rs1`, `imm` | Salva o byte da região baixa de `rs1` no endereço dado por `rd + imm`.  |       |
| `sh`      | `rd`, `rs1`, `imm` | Salva 2 bytes da região baixa de `rs1` no endereço dado por `rd + imm`. |       |
| `sw`      | `rd`, `rs1`, `imm` | Salva o valor de `rs1` no endereço dado por `rd + imm`.                 |       |

### Branches (saltos condicionais)

| Mnemônico | Argumentos          | Descrição                                           | Sinal |
| --------- | ------------------- | --------------------------------------------------- | ----- |
| `beq`     | `rs1`, `rs2`, `imm` | Salta para `PC + imm` caso `rs1 == rs2`.            |       |
| `bne`     | `rs1`, `rs2`, `imm` | Salta para `PC + imm` caso `rs1 != rs2`.            |       |
| `blt`     | `rs1`, `rs2`, `imm` | Salta para `PC + imm` caso `rs1 < rs2`.             | Com.  |
| `bge`     | `rs1`, `rs2`, `imm` | Salta para `PC + imm` caso `rs1 >= rs2`.            | Com.  |
| `bltu`    | `rs1`, `rs2`, `imm` | Igual à `blt` porém realiza a comparação sem sinal. | Sem.  |
| `bgeu`    | `rs1`, `rs2`, `imm` | Igual à `bge` porém realiza a comparação sem sinal. | Sem.  |

### Jumps (saltos incondicionais)

| Mnemônico | Argumentos         | Descrição                                                                | Sinal |
| --------- | ------------------ | ------------------------------------------------------------------------ | ----- |
| `jal`     | `rd`, `imm`        | Salta para `PC + imm`, salvando o valor de `PC + 4` em `rd`.             | Com.  |
| `jalr`    | `rd`, `rs1`, `imm` | Salta para o valor de `rs1 + imm`, salvando o valor de `PC + 4` em `rd`. | Com.  |

### Miscelânea

| Mnemônico | Argumentos  | Descrição                                                                               | Sinal |
| --------- | ----------- | --------------------------------------------------------------------------------------- | ----- |
| `lui`     | `rd`, `imm` | Carrega o valor `imm << 12` em `rd`.                                                    |       |
| `auipc`   | `rd`, `imm` | Carrega o valor `PC + (imm << 12)` em `rd`.                                             |       |
| `ecall`   |             | Realiza uma chamada ao emulador. (ver seção de chamadas)                                |       |
| `ebreak`  |             | Equivalente à `ecall`, porém não considera argumentos, realiza sempre uma chamada BREAK |       |

## Registradores

Registradores salvos devem ter seu valor original restaurado pela função que os utilizou.

| Registrador(es) | Nome(es)     | Função                                        | Salvo(s)? |
| --------------- | ------------ | --------------------------------------------- | --------- |
| `x0`            | `zero`       | Constante zero. Escrever nele não tem efeito. |           |
| `x1`            | `ra`         | Endereço de retorno. Utilizado em saltos.     |           |
| `x2`            | `sp`         | Endereço da pilha (ela cresce para baixo).    |           |
| `x3`            | `gp`         | Pointeiro global.                             |           |
| `x4`            | `tp`         | Pointeiro da thread.                          |           |
| `x5` à `x7`     | `t0` à `t2`  | Uso geral, temporários.                       |           |
| `x8` e `x9`     | `s0` e `s1`  | Uso geral, salvos.                            | Sim.      |
| `x10` e `x11`   | `a0` e `a1`  | Argumentos de função e valores de retorno.    |           |
| `x12` à `x17`   | `a2` à `a7`  | Argumentos de função.                         |           |
| `x18` à `x27`   | `s2` à `s11` | Uso geral, salvos.                            | Sim.      |
| `x28` à `x31`   | `t3` à `t6`  | Uso geral, temporários.                       |           |

## Sintaxe

```asm
	; Um ponto e vírgula introduz comentários
	addi a7, zero, 3
	addi a0, zero, msg
	addi a1, zero, 14
	ecall
	ebreak
msg:
#Hello, World!%0a
```

## Chamadas do emulador

Coloca-se o número em 'a7' e argumentos em 'a0' e 'a1'.

```asm
	; Sair/Parar
	addi a7, zero, 1
	ecall

	; Sair/Parar (mais simples)
	ebreak

	; Ler texto
	addi a7, zero, 2
	addi a0, zero, <endereço>
	addi a1, zero, <tamanho>
	ecall

	; Imprimir texto
	addi a7, zero, 3
	addi a0, zero, <endereço>
	addi a1, zero, <tamanho>
	ecall
```

## Carregar imediatos

```asm
	; Números menores que 12 bits (em complemento de 2).
	addi <reg>, zero, <num>
	; Exemplo: carregar o número 42 em t0
	addi t0, zero, 42
	
	; Números maiores que 12 bits
	lui <reg>, <parte alta>
	addi <reg>, <reg>, <parte baixa>
	; Exemplo: carregar o número 0x6747 em t0
	lui t0, 0x6
	addi t0, t0, 0x747

	; Números maiores que 12 bits sem extensão de sinal
	; (caso em que o bit 11 é 1 porém não se quer extender o sinal)
	; Para isso, adiciona-se 1 na parte alta para assim corrigir o valor com a extensão
	lui <reg>, <parte alta + 1>
	addi <reg>, <reg>, <parte baixa>
	; Exemplo: carregar o número 0x5a47 em t0
	lui t0, 0x6
	addi t0, t0, 0xa47
```

## Manipular a stack

```asm
	; Inicializar
	addi sp, zero, 0
	
	; "push"
	addi sp, sp, -4
	sw sp, <reg>, 0
	
	; "pop"
	addi sp, sp, 4
	lw <reg>, sp, -4
```

## Chamar funções

Os argumentos e retornos ficam nos registradores com letra 'a'.

```asm
	; Chamar função "func"
	jal ra, func
	
	; Retornar
	jalr zero, ra, 0
```

Para chamar uma função mais longe que 20 bits, precisa-se construir o
endereço e então usar 'jalr':

```asm
	jalr ra, <reg>, 0
```

## Truques do debugger

### Ver o PC

É possível utilizar o comando `print` com a expressão `#1` para ver o PC. `#1`
significa "ver uma instrução a partir do PC". Exemplo:

```
egg> print #1
0xcafebabe: func: add  t0 a0 a1
```

Nesse exemplo, o PC está no endereço `0xcafebabe`, onde há uma etiqueta `func`,
e a instrução é um `add`, guardando a soma de `a0` e `a1` em `t0`.

### Dump do programa

É possível utilizar o EGG para montar um programa em Assembly e guardar em um
arquivo. Para isso, abre-se o programa no emulador e utiliza-se o seguinte
comando:

```
egg> dump # <nome do arquivo>
```

Substituindo `<nome do arquivo>` pelo nome que queira salvar o arquivo do
programa. Note que esse arquivo não terá nenhum formato especial: ele conterá
apenas os bytes correspondentes às instruções do programa, junto com literais
que forem adicionados.
