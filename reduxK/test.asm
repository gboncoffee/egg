; Instructions test for EGG REDUX-V.

	; addi, prepare for jump.
	addi 5

	; Copy addr.
	add r1, r0
	; Zero r0 to test branching.
	sub r0, r0

	; The first addi should not be performed.
	brzr r0, r1
	addi 0xf
	addi 1
	; And this should not branch.
	brzr r0, r1

	; The addi should not be performed.
	ji ji_test
	addi 0xf

	; Arithmetic instructions.
ji_test:
	sub r0, r0 ; r0 = 0
	not r0, r0 ; r0 = 1
	add r0, r0 ; r0 = 2
	; For testing logical.
	sub r1, r1 ; r1 = 0
	not r1, r1 ; r1 = 1
	or r0, r1  ; r0 = 3
	and r0, r1 ; r0 = 1
	slr r0, r1 ; r0 = 2
	xor r0, r1 ; r0 = 3
	srr r0, r1 ; r0 = 1

	; Prepare address for load/store
	sub r0, r0
	sub r1, r1
	addi 0x0f
	add r1, r0
	sub r0, r0
	addi 0x0e
	st r0, r1
	sub r0, r0
	ld r0, r1

	; Calls
	ebreak

	sub r0, r0
	sub r1, r1
	sub r2, r2
	addi 2
	add r2, r0
	sub r0, r0
	addi 1
	add r1, r0
	sub r0, r0
	ecall

	sub r1, r1
	sub r2, r2
	sub r3, r3
	inc r0, 1
	loadv 3
	inc r2, 1
	loadv 3
	sub r0, r0
	sub r1, r1
	sub r2, r2
	addi 7
	add r2, r0
	addi -3
	add r1, r0
	addi -3
	addv 3
