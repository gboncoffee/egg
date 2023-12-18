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
		Token{Type: TOKEN_LABEL, Value: "_start"},
		Token{Type: TOKEN_INSTRUCTION, Value: "addi"},
		Token{Type: TOKEN_ARG, Value: "t1"},
		Token{Type: TOKEN_ARG, Value: "zero"},
		Token{Type: TOKEN_ARG, Value: "1"},
		Token{Type: TOKEN_INSTRUCTION, Value: "beq"},
		Token{Type: TOKEN_ARG, Value: "t1"},
		Token{Type: TOKEN_ARG, Value: "zero"},
		Token{Type: TOKEN_ARG, Value: "prob"},
		Token{Type: TOKEN_INSTRUCTION, Value: "ebreak"},
		Token{Type: TOKEN_LABEL, Value: "prob"},
		Token{Type: TOKEN_INSTRUCTION, Value: "addi"},
		Token{Type: TOKEN_ARG, Value: "a0"},
		Token{Type: TOKEN_ARG, Value: "zero"},
		Token{Type: TOKEN_ARG, Value: "msg"},
		Token{Type: TOKEN_INSTRUCTION, Value: "ecall"},
		Token{Type: TOKEN_LABEL, Value: "msg"},
		Token{Type: TOKEN_LITERAL, Value: "Your machine is broken\n"},
	}

	if !reflect.DeepEqual(tokens, expectedTokens) {
		t.Fatalf("Tokenizer doesnt work: %v", tokens)
	}
}
