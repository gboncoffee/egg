;;
;; Test source file for RISC-V IM 32 bits implementation at EGG.
;;
;; This implements all instructions with arguments that allows correctness
;; checking in the assembler (endianess, etc) by comparison with a known-correct
;; assembler.
;;
;; The assembly file with RARS syntax is at test-rars.asm.
;;
_start:
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
	sb 5, 3, 3
	sh 5, 3, 3
	sw 5, 3, 3
	sb 5, 3, -3
	sh 5, 3, -3
	sw 5, 3, -3

	;; Branch instructions.
	beq 3, 5, end
	bne 3, 5, end
	blt 3, 5, end
	bge 3, 5, end
	bltu 3, 5, end
	bgeu 3, 5, end
	beq 3, 5, _start
	bne 3, 5, _start
	blt 3, 5, _start
	bge 3, 5, _start
	bltu 3, 5, _start
	bgeu 3, 5, _start

	;; Jump instructions.
	jal 3, end
	jal 3, _start
	jalr 3, 5, 3
	jalr 3, 5, -3

	;; UI instructions.
	lui 3, 5
	auipc 3, 5
	lui 3, -5
	auipc 3, -5

	;; Calls
	ecall
	ebreak

	;; Multiplication extension.
	mul 3, 5, 7
	mulh 3, 5, 7
	mulsu 3, 5, 7
	mulu 3, 5, 7
	div 3, 5, 7
	divu 3, 5, 7
	rem 3, 5, 7
	remu 3, 5, 7
end:
