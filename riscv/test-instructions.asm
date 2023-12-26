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
	;; Branches.
	;;

	addi t0, zero, 1
	addi t1, zero, 2
bges:
	sub t1, t1, t0
	bge t1, t0, bges

	addi t0, zero, 1
	addi t1, zero, 2
beqs:
	sub t1, t1, t0
	beq t1, t0, beqs

	addi t0, zero, 1
	addi t1, zero, 3
blts:
	sub t1, t1, t0
	blt t0, t1, blts

	addi t0, zero, 1
	addi t1, zero, 3
bnes:
	sub t1, t1, t0
	blt t0, t1, bnes

	;; Unsigned
	addi t0, zero, 0xff000000
	addi t1, zero, 1
	bltu t1, t0, bltue
	addi zero, zero, 0	;; Shouldn't perform this.
bltue:
	bgeu t0, t1, bgeue
	addi zero, zero, 0	;; Shouldn't perform this.
bgeue:

	;;
	;; Jump.
	;;
	jal ra, func
	;; It should jump to func at the bottom them return.

	;; Calls.
	addi a7, zero, 2
	ecall
	ebreak

func:
	jalr zero, ra, 0
