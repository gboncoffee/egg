# RISC-V Cheatsheet - EGG

## Instruções

![instructions.png]

## Registradores

![registers.png]

## Sintaxe

```asm
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
	; Números menores que 12 bits
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
	lw sp, <reg>, 0
	
	; "pop"
	addi sp, sp, 4
	lw sp, <reg>, -4
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
