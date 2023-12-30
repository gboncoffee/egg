[![Go](https://github.com/gboncoffee/egg/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/gboncoffee/egg/actions/workflows/go.yml)
[![CodeQL](https://github.com/gboncoffee/egg/actions/workflows/github-code-scanning/codeql/badge.svg?branch=master)](https://github.com/gboncoffee/egg/actions/workflows/github-code-scanning/codeql)

[Read this document in English](README.md)

# EGG, um emulador de processadores genérico

EGG (Emulador Genérico do Gabriel) é um emulador modular de arquiteturas de
processador, criado para fins educacionais.

O pacote `egg` provém uma interface para interagir com máquinas, assim provendo
suporte a diferentes backends provendo arquiteturas. No momento, há somente o
pacote `egg/riscv` de backend, implementando uma máquina RISC-V IM de 32 bits.
Um backend de MIPS é planejado para o futuro.

O pacote `egg/assembler` provém uma pequena biblioteca para criação de
assemblers e o suporte ao debugger do EGG.

## Instalação

Baixe o binário estático para o seu sistema na página de
[releases](https://github.com/gboncoffee/egg/releases), ou, caso tenha o
compilador de Go instalado, baixe e compile o projeto.

Nota: o binário para Windows não foi testado. Não tenho acesso a nenhuma máquina
Windows atualmente.

## Uso

Rode o emulador com um arquivo de Assembly para montá-lo e iniciar uma máquina
rodando o programa. O backend utilizado por padrão é uma máquina de RISC-V 32
bits. Use a opção `-a` ou `-arch` para mudar a arquitetura (atualmente há
somente a opção RISC-V. Um backend de MIPS é planejado para o futuro). A opção
`-h` mostra todas as opções de linha de comando e a opção `-l` mostra todas as
arquiteturas suportadas.

A sintaxe de Assembly varia com a arquitetura, porém, como o projeto provém uma
biblioteca para tal, os backends podem usar uma sintaxe bem semelhante (RISC-V
usa). Exemplo:

```asm
; Ponto e vírgula define comentários até o fim da linha.

; Labels são definidos com dois pontos.
label:
	; Os argumentos começam sempre pelo destino.
	addi t0, zero, 2

	; Também é possível colocar instruções logo após os labels.
label2:	add t0, t0, t0

	; Não há parênteses como no RARS e stores usam imediatos normais.
	sb t0, ra, 3

	; Imediatos hexadecimais, octais e binários também são suportados.
	addi t1, zero, 0xff
	addi t1, zero, 0b010110
	addi t1, zero, 0o644
	addi t1, zero, 0755	; Um zero à esquerda também define um octal.

; # define um literal até o final da linha.
; Literais são inseridos no binário (assim como 'db' em outros assemblers).
; Se uma % é seguida de dois digitos hexadecimais, o valor hexadecimal é
; inserido. Use %% para inserir um % (ou %25).
msg:
#Hello, World!%0a
```

Cada diretório de cada arquitetura possui programas Assembly de teste que podem
ser usados de exemplo.

## Debugger

A interface do debugger é semelhante à do `gdb` porém bem enxuta. Use a opção
`-d` para entrar no debugger assim que iniciar o emulador. Não há comando `run`,
como no `gdb`, pois não há necessidade de iniciar um processo do sistema. Os
comandos `next` e `continue` podem ser utilizados para iniciar o programa
normalmente. Use o comando `help` para ver todos os comandos disponíveis.
