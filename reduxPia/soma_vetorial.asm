;;
;; Soma vetorial, Assembly de REDUX-PIÁ.
;;

	xor r0, r0
	xor r1, r1
	xor r2, r2
	xor r3, r3

;; mov r0, 42
	addi 0xA
	ldui 0x2

;; mov r1, r0
	add r1, r0

;; mov r0, v
	xor r0, r0
	addi 7
	ldui 1

;; mov r2, r0
	add r2, r0

;; mov r0, 3
	xor r0, r0
	addi 3

;; mov r3, r0
	add r3, r0
;; mov r0, 1
	xor r0, r0
	addi 1
for:
	st r1, r2

	add r1, r0
	add r2, r0

	loop for

	xor r0, r0
	addi 1
	ecall
v:
