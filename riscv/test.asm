;;
;; Test source file for RISC-V IM 32 bits implementation at EGG.
;;
;; This implements all instructions with arguments that allows correctness
;; checking in the assembler (endianess, etc) by comparison with a known-correct
;; assembler.
;;
;; The assembly file with RARS syntax is at test-rars.asm.
;;

	;; Arithmetic instructions.
	add 3, 5, 7
	sub 3, 5, 7
	xor 3, 5, 7
	or 3, 5, 7
	and 3, 5, 7
	sll 3, 5, 7
	srl 3, 5, 7
	sra 3, 5, 7
	slt 3, 5, 7
	sltu 3, 5, 7

	;; Arithmetic immediate instructions.
	addi 3, 5, 42
	xori 3, 5, 42
	ori 3, 5, 42
	andi 3, 5, 42
	slli 3, 5, 3
	srli 3, 5, 3
	srai 3, 5, 3
	slti 3, 5, 42
	sltiu 3, 5, 42

	;; Load instructions.
	lb 3, 5, 42
	lh 3, 5, 42
	lw 3, 5, 42
	lbu 3, 5, 42
	lhu 3, 5, 42

	;; Store instructions.
	sb 3, 5, 3
	sh 3, 5, 3
	sw 3, 5, 3
	sb 3, 5, -3
	sh 3, 5, -3
	sw 3, 5, -3

	;; Branch instructions.
	beq 3, 5, 3
	bne 3, 5, 3
	blt 3, 5, 3
	bge 3, 5, 3
	bltu 3, 5, 3
	bgeu 3, 5, 3
	beq 3, 5, -3
	bne 3, 5, -3
	blt 3, 5, -3
	bge 3, 5, -3
	bltu 3, 5, -3
	bgeu 3, 5, -3

	;; Jump instructions.
	jal 3, 5
	jal 3, -5
	jalr 3, 5, 7
	jalr 3, 5, -7

	;; UI instructions.
	lui 3, 5
	auipc 3, 5
	lui 3, -5
	auipc 3, -5

	;; Calls
	ecall
	ebreak
