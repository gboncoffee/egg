[![Go](https://github.com/gboncoffee/egg/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/gboncoffee/egg/actions/workflows/go.yml)
[![CodeQL](https://github.com/gboncoffee/egg/actions/workflows/github-code-scanning/codeql/badge.svg?branch=master)](https://github.com/gboncoffee/egg/actions/workflows/github-code-scanning/codeql)

[Read this document in English](README.md)

# EGG, um emulador de processadores genérico

[Tutorial em Vídeo (legendado)](https://youtu.be/RtKvvXgdVak?si=OtjkNxSGmRGNou67)

[Documentação](https://egg.gboncoffee.dev.br/docs/pt)

[Contribuindo, bugs, requisição de features](CONTRIBUINDO.md)

EGG (Emulador Genérico do Gabriel) é um emulador modular de arquiteturas de
processador, criado para fins educacionais.

O pacote `egg` provém uma interface para interagir com máquinas, assim provendo
suporte a diferentes backends provendo arquiteturas. No momento, há backends de
RISC-V IM de 32 bits, MIPS32 e uma arquitetura fantasia de 8 bits, "Sagui".

O pacote `egg/assembler` provém uma pequena biblioteca para criação de
assemblers e o suporte ao debugger do EGG.

### Estudantes da UFPR

Quer ajuda, ou tem alguma dúvida? Me mande um email: `ggb23@inf.ufpr.br`. Ou
me procure no campus e nos laboratórios!

## Instalação

Baixe o binário estático para o seu sistema na página de
[releases](https://github.com/gboncoffee/egg/releases), ou, caso tenha o
compilador de Go instalado, baixe e compile o projeto.

Nota: o binário para Windows não foi testado. Não tenho acesso a nenhuma máquina
Windows atualmente. O teste do binário para Darwin (MacOS) depende da minha
paciência.

## Uso

O EGG é um aplicativo de linha de comando. Você deve utilizá-lo com um emulador
de terminal.

Rode o emulador com um arquivo de Assembly para montá-lo e iniciar uma máquina
rodando o programa. O backend utilizado por padrão é uma máquina de RISC-V 32
bits. Use a opção `-a` ou `-arch` para mudar a arquitetura. A opção `-h` mostra
todas as opções de linha de comando e a opção `-l` mostra todas as arquiteturas
suportadas.

A sintaxe de Assembly varia com a arquitetura, porém, como o projeto provém uma
biblioteca para tal, os backends podem usar uma sintaxe bem semelhante (RISC-V e
MIPS usam). Exemplo:

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

; Algumas diretivas são suportadas. As diretivas "bitsxx" criam números literais 
; no código, com o tamanho em bits especificado:
.bits8 0xca 0xfe 0xba 0xbe 0xde 0xad 0xbe 0xef
.bits16 0xcafe 0xbabe 0xdead 0xbeef
.bits32 0xcafebabe 0xdeadbeef
.bits64 0xcafebabedeadbeef
; A diretiva "space" adiciona alguns bytes de espaço vazio no código:
.space 16
; A diretiva "include" faz um "copia-cola" de outro arquivo no código:
.include other-asm.asm
```

Cada diretório de cada arquitetura possui programas Assembly de teste que podem
ser usados de exemplo.

## Chamadas

As chamadas (_environment calls_) respondidas pelo emulator são as
seguintes. Consulte a documentação da arquitetura para realizá-las:

- BREAK (Número 1): Transfere o controle para o debugger ou finaliza a máquina.  
  Nenhum argumento.  
- READ (Número 2): Lê uma entrada da linha de comando.  
  - Argumento 1: Endereço do buffer.  
  - Argumento 2: Tamanho da entrada em bytes.  
- WRITE (Número 3): Escreve uma saída.  
  - Argumento 1: Endereço do buffer.  
  - Argumento 2: Tamanho da saída em bytes.  

## Debugger

A interface do debugger é semelhante à do `gdb` porém bem enxuta. Use a opção
`-d` para entrar no debugger assim que iniciar o emulador. Não há comando `run`,
como no `gdb`, pois não há necessidade de iniciar um processo do sistema. Os
comandos `next` e `continue` podem ser utilizados para iniciar o programa
normalmente. Use o comando `help` para ver todos os comandos disponíveis.
