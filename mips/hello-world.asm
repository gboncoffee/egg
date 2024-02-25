;;
;; MIPS Hello World for the EGG emulator.
;;

_start:
	addi v0, zero, 3
	addi a0, zero, msg
	addi a1, zero, 14
	syscall
	break

msg:
#Hello, World!%0a
