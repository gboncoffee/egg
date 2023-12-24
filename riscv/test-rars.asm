#
# Test source file with RARS syntax for RISC-V IM 32 bits implementation at
# EGG.
#

_start:
	# Arithmetic instructions.
	add x3, x5, x7
	sub x3, x5, x7
	xor x3, x5, x7
	or x3, x5, x7
	and x3, x5, x7
	sll x3, x5, x7
	srl x3, x5, x7
	sra x3, x5, x7
	slt x3, x5, x7
	sltu x3, x5, x7

	# Arithmetic immediate instructions.
	addi x3, x5, 42
	xori x3, x5, 42
	ori x3, x5, 42
	andi x3, x5, 42
	slli x3, x5, 3
	srli x3, x5, 3
	srai x3, x5, 3
	slti x3, x5, 42
	sltiu x3, x5, 42

	# Load instructions.
	lb x3, 42(x5)
	lh x3, 42(x5)
	lw x3, 42(x5)
	lbu x3, 42(x5)
	lhu x3, 42(x5)

	# Store instructions.
	sb x3, 3(x5)
	sh x3, 3(x5)
	sw x3, 3(x5)
	sb x3, -3(x5)
	sh x3, -3(x5)
	sw x3, -3(x5)

	# Branch instructions.
	beq x3, x5, end
	bne x3, x5, end
	blt x3, x5, end
	bge x3, x5, end
	bltu x3, x5, end
	bgeu x3, x5, end
	beq x3, x5, _start
	bne x3, x5, _start
	blt x3, x5, _start
	bge x3, x5, _start
	bltu x3, x5, _start
	bgeu x3, x5, _start

	# Jump instructions.
	jal x3, end
	jal x3, _start
	jalr x3, x5, 3
	jalr x3, x5, -3

	# UI instructions.
	lui x3, 5
	auipc x3, 5
	lui x3, -5
	auipc x3, -5

	# Calls
	ecall
	ebreak

	# Multiplication extension.
	mul x3, x5, x7
	mulh x3, x5, x7
	mulhsu x3, x5, x7
	mulhu x3, x5, x7
	div x3, x5, x7
	divu x3, x5, x7
	rem x3, x5, x7
	remu x3, x5, x7
end:
