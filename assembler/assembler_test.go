package assembler

import (
	"testing"
	"reflect"
)

func TestTokenizer(t *testing.T) {
	asm := `
;; Test assembly string for the tokenizer test
_start:
	addi t1, zero, 1
	beq t1, zero, prob	;; Looks like your machine has some problem!
	ebreak
prob:	addi a0, zero, msg	;; Let's fixed it!
	ecall

msg:
#Your %6dachine is broken%0A
`
	tokens := Tokenize(asm)

	expectedTokens := []Token{
		{Type: TOKEN_LABEL, Value: "_start"},
		{Type: TOKEN_INSTRUCTION, Value: "addi"},
		{Type: TOKEN_ARG, Value: "t1"},
		{Type: TOKEN_ARG, Value: "zero"},
		{Type: TOKEN_ARG, Value: "1"},
		{Type: TOKEN_INSTRUCTION, Value: "beq"},
		{Type: TOKEN_ARG, Value: "t1"},
		{Type: TOKEN_ARG, Value: "zero"},
		{Type: TOKEN_ARG, Value: "prob"},
		{Type: TOKEN_INSTRUCTION, Value: "ebreak"},
		{Type: TOKEN_LABEL, Value: "prob"},
		{Type: TOKEN_INSTRUCTION, Value: "addi"},
		{Type: TOKEN_ARG, Value: "a0"},
		{Type: TOKEN_ARG, Value: "zero"},
		{Type: TOKEN_ARG, Value: "msg"},
		{Type: TOKEN_INSTRUCTION, Value: "ecall"},
		{Type: TOKEN_LABEL, Value: "msg"},
		{Type: TOKEN_LITERAL, Value: "Your machine is broken\n"},
	}

	if !reflect.DeepEqual(tokens, expectedTokens) {
		t.Fatalf("Tokenizer doesnt work: %v", tokens)
	}
}
