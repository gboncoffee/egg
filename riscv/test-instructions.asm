;;
;; Test source file for RISC-V IM 32 bits implementation at EGG.
;;
;; This implements every instruction, the machine should change to a well
;; defined state after each.
;;

_start:
	;;
	;; Arithmetic immediate.
	;;

	;; t0 = 1
	addi t0, zero, 1

	;; t0 = t0 ^ 5 :: t0 == 4
	xori t0, t0, 5

	;; t0 = t0 | 2 :: t0 == 7
	ori t0, t0, 2

	;; t0 = t0 & 3 :: t0 == 2
	andi t0, t0, 3

	;; t0 = t0 << 2 :: t0 == 8
	slli t0, t0, 2

	;; t1 = 2147483658
	addi t1, zero, 1
	slli t1, t1, 31
	addi t1, t1, 10

	;; t2 = t1 >> 2 :: t2 == 536870914
	srli t2, t1, 2

	;; t2 = t1 !>> 2 :: t2 == 3758096386
	srai t2, t1, 2

	;; t3 = t2 < t0 (signed negative less than 4) :: t3 == 1
	slti t3, t2, t0

	;; Same but unsigned :: t3 == 0
	sltu t3, t2, t0

	;;
	;; Arithmetic.
	;;

	;; (Before testing set some constants so we can play).
	addi t0, zero, 1
	addi t1, zero, 2
	addi t2, zero, 3
	addi t3, zero, 5

	;; t4 = t0 + t1 :: t4 == 3
	add t4, t0, t1

	;; t4 = t2 - t1 :: t4 == 1
	sub t4, t2, t1

	;; t4 = t3 ^ t2 :: t4 == 6
	xor t4, t3, t2

	;; t4 = t0 | t1 :: t4 == 3
	or t4, t0, t1

	;; t4 = t2 & t3 :: t4 == 1
	and t4, t2, t3

	;; t4 = t0 << t1 :: t4 == 2
	sll t4, t0, t1

	;; Negative constant strikes again.
	addi t5, zero, 1
	slli t5, t5, 31
	addi t5, t5, 10

	;; t4 = t5 >> t1 :: t4 == 536870914
	srl t4, t5, t1

	;; t4 = t5 !>> t1 :: t4 == 3758096386
	sra t4, t5, t1

	;; t4 = t5 > t2 (signed negative less than 3) :: t4 == 1
	slt t4, t5, t2

	;; Same but unsigned :: t4 == 0
	sltu t4, t5, t2

	;;
	;; Loads and stores with the word 0xaaeeffff at address 0x1000.
	;; This is terrible because we didn't tested lui already.
	;;
	addi t0, zero, 1
	slli t0, t0, 12

	addi t2, zero, 0xaa
	slli t2, t2, 24
	or t1, zero, t2
	addi t2, zero, 0xee
	slli t2, t2, 16
	or t1, t1, t2
	addi t2, zero, 0xff
	slli t2, t2, 8
	or t1, t1, t2
	ori t1, t1, 0xff

	sw t0, t1, 0
	;; t2 == 0xaaeeffff
	lw t2, t0, 0

	sb t0, t1, 0
	;; t2 == 0xffffffff
	lb t2, t0, 0
	;; t2 == 0x000000ff
	lbu t2, t0, 0

	sh t0, t1, 0
	;; t2 == 0xffffffff
	lh t2, t0, 0
	;; t2 == 0x0000ffff
	lhu t2, t0, 0

	;;
	;; Jump.
	;;
	jal ra, func
	;; It should jump to func at the bottom them return.

	;;
	;; Calls.
	;;
	addi a7, zero, 2
	ecall
	ebreak

	;;
	;; Branches.
	;;
	addi t0, zero, -1
	addi t1, zero, 2

	beq t1, t0, func	;; Should not branch.
	beq zero, zero, beq1	;; Should branch.
	addi zero, zero, zero
beq1:

	bne zero, zero, func	;; Should not branch.
	bne t1, t0, bne1	;; Should branch.
	addi zero, zero, zero
bne1:

	blt t1, t0, func	;; Should not branch.
	blt zero, zero, func	;; Should not branch.
	blt t0, t1, blt1	;; Should branch.
	addi zero, zero, zero
blt1:

	bge t0, t1, func	;; Should not branch.
	bge zero, zero, bge1	;; Should branch.
	addi zero, zero, zero
bge1:
	bge t1, t0, bge2	;; Should branch.
	addi zero, zero, zero
bge2:

	bltu t0, t1, func	;; Should not branch.
	bltu zero, zero, func	;; Should not branch.
	bltu t1, t0, bltu1	;; Should branch.
	addi zero, zero, zero
bltu1:

	bgeu t1, t0, func	;; Should not branch.
	bgeu zero, zero, bgeu1	;; Should branch.
	addi zero, zero, zero
bgeu1:
	bgeu t0, t1, bgeu2	;; Should branch.
	addi zero, zero, zero
bgeu2:

	;;
	;; auipc/lui
	;;
	lui t0, 0xaaaaa
	auipc t0, 0xaaaaa

	;;
	;; Multiplication extension.
	;;

	;; 1610608640 (kinda convenient).
	lui t0, 0x5ffff
	addi t1, zero, 3

	;; t2 == 0x1fffd000
	mul t2, t0, t1
	;; t2 == 1
	mulh t2, t0, t1
	;; t2 == 0x1ffffaaa
	div t2, t0, t1

	addi t1, zero, -3

	;; t2 == 0xE0003000
	mul t2, t0, t1
	;; t2 == 0xFFFFFFFE
	mulh t2, t0, t1
	;; t2 == 0xF5556555
	div t2, t0, t1
func:
	jalr zero, ra, 0
