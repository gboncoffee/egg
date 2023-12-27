;;
;; RISC-V Hello World for the EGG emulator.
;;

_start:
	addi a7, zero, 3
	addi a0, zero, msg
	addi a1, zero, 14
	ecall
	ebreak

msg:
#Hello, World!%0a
