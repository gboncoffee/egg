	xor r0, r0
	xor r1, r1
	xor r2, r2
	xor r3, r3

	addi 5
	add r3, r0

	xor r0, r0
for:
;; Por conveniência, vamos usar a posição 0 da memória para guardar o
;; acumulador. Esse loop tem exatamente o tamanho limite para usar a instrução
;; loop.
	xor r1, r1
	st r0, r1

	xor r0, r0
;; Os vetores ficam em 0x20 para salvar uma instrução aqui e manter o loop no
;; tamanho para utilizar a instrução loop.
	ldui 2
	add r0, r3
	xor r2, r2
	add r1, r0
	add r2, r0
	xor r0, r0
	addi 5
	add r2, r0
	ld r1, r1
	ld r2, r2

	xor r0, r0
	ld r0, r0

	mac r1, r2

	loop for

;; ji 0
	ebreak

.space 8
a:
.bits8 1
.bits8 2
.bits8 3
.bits8 4
.bits8 5
b:
.bits8 5
.bits8 4
.bits8 3
.bits8 2
.bits8 1
