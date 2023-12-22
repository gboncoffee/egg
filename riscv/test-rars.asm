#
# Test source file with RARS syntax for RISC-V IM 32 bits implementation at
# EGG.
#

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
	beq x3, x5, 3
	bne x3, x5, 3
	blt x3, x5, 3
	bge x3, x5, 3
	bltu x3, x5, 3
	bgeu x3, x5, 3
	beq x3, x5, 3
	bne x3, x5, -3
	blt x3, x5, -3
	bge x3, x5, -3
	bltu x3, x5, -3
	bgeu x3, x5, -3

	# Jump instructions.
	jal x3, 5
	jal x3, -5
	jalr x3, x5, 7
	jalr x3, x5, -7

	# UI instructions.
	lui x3, 5
	auipc x3, 5
	lui x3, -5
	auipc x3, -5

	# Calls
	ecall
	ebreak
