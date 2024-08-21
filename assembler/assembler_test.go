package assembler

import (
	"testing"
)

func TestTokenizer(t *testing.T) {
	var tokens []Token
	err := Tokenize("tokenizer_test.asm", &tokens)
	if err != nil {
		t.Fatalf("error tokenizing: %v", err)
	}

	// TODO: I manually guaranteed that the tokenizer is in fact working but it
	// would be great to add a proper test.
}
