; Instructions test for EGG Sagui.

	; movl, prepare for jumps.
	movl 5

	; Copy addr.
	movr r1, r0
	; Zero r0 to test branching.
	sub r0, r0

	; The movr should not be performed.
	brzr r0, r1
	movr r0, r0

	movl 8
	jr r0
	; Again, movr shouldn't be performed.
	movr r0, r0

	; Again, movr shouldn't be performed.
	sub r0, r0
	brzi 2
	movr r0, r0
	; Now the branches shouldn't be performed.
	movl 1
	brzi 3
	brzr r0, r0
	; Lastly, test ji.
	ji 2
	movr r0, r0

	; Arithmetic instructions.
	sub r0, r0
	movl 0x1
	movr r1, r0
	sub r0, r0
	movh 0xf

	; r0 shall be 0xf1.
	add r0, r1
	; r0 shall be 0xf0.
	sub r0, r1

	; r0 shall be 0xf1.
	or r0, r1
	; r0 shall be 0x01.
	and r0, r1

	; r0 shall be 0x0.
	not r0, r0
	; And 0x1 again.
	not r0, r0

	movr r1, r0
	; r0 shall be 0x2.
	slr r0, r1
	; Now 0x1 again.
	srr r0, r1

	; Use a convenient address and value to test load and store.
	movl 0xa
	movr r1, r0
	movh 0xf
	st r1, r0
	; Zero r1 to be sure.
	sub r1, r1
	ld r1, r0
