;;
;; Test source file for RISC-V IM 32 bits implementation at EGG.
;;

_start:
	addi t1, zero, 5
	addi t2, zero, 1

test:	;; This is a loop.
	sub t1, t1, t2
	bne t1, zero, test	;; Go out of the loop when tq = 0.

	ebreak
