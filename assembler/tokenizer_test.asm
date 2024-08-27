label1: instruction1
label2:	; Comment
	instruction2 arg1 ; Comment again
#literal %60 with hash%0a

.literal literal with directive
.bits32 0xcafebabe 0xdeadbeef

.bits8 0xca 0xfe 0xba 0xbe ; Comment
.bits64 0xcafebabedeadbeef
.space 12
label3:
	instruction3 arg1, arg2

.include included_tokenizer_test.asm
