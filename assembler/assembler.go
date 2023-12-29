// Package egg/assembler implements a set of usefull functions and structs for
// creating assemblers.
package assembler

import (
	"fmt"
	"errors"
	"strings"
)

const (
	TOKEN_LABEL = iota
	TOKEN_INSTRUCTION
	TOKEN_ARG
	TOKEN_LITERAL
)

// Can be label, instruction, arg or literal.
type Token struct {
	Type uint8
	Value string
}

// Can be instruction or literal. Only instructions have Args.
type ResolvedToken struct {
	Type uint8
	Size uint64
	Value string
	Args []uint64
}

// Only instructions.
type DebuggerToken struct {
	Instruction string
	Args string
	Address uint64
	Label string
}

func CreateDebugTokensFixedSize(tokens []Token, size uint64) []DebuggerToken {
	var dt []DebuggerToken
	addr := uint64(0)
	var last_label string
	var last_instruction *DebuggerToken

	for _, t := range tokens {
		switch t.Type {

		// I'm pretty sure the tokenizer cannot create an argument
		// before an instruction so we just ignore it and let have an
		// undefined behaviour if that happens.
		case TOKEN_INSTRUCTION:
			dt = append(dt, DebuggerToken{
				Instruction: t.Value,
				Address: addr,
				Label: last_label,
			})
			addr += size
			last_label = ""
			last_instruction = &dt[len(dt)-1]
		case TOKEN_ARG:
			last_instruction.Args = last_instruction.Args + " " + t.Value
		case TOKEN_LABEL:
			last_label = t.Value
		case TOKEN_LITERAL:
			addr += uint64(len(t.Value))
			last_label = ""
		}
	}

	return dt
}

// Function to pass to ResolveTokensFixedSize.
type ArgTranslateFunction func(string) (uint64, error)

// Resolves labels and arguments for fixed-size instructions with size bytes.
// ArgTranslateFunction translates arguments into numbers.
func ResolveTokensFixedSize(tokens []Token, size uint64, arg ArgTranslateFunction) ([]ResolvedToken, error) {
	labels := make(map[string]uint64)
	addr := uint64(0)
	for _, t := range tokens {
		switch (t.Type) {
		case TOKEN_INSTRUCTION:
			addr += size
		case TOKEN_LITERAL:
			addr += uint64(len(t.Value))
		case TOKEN_LABEL:
			labels[t.Value] = addr
		}
	}

	var rt []ResolvedToken
	i := 0

	// I'm pretty sure the tokenizer cannot create an argument before an
	// instruction so we just ignore it and let have an undefined behaviour
	// if that happens.
	for _, t := range tokens {
		switch t.Type {
		case TOKEN_INSTRUCTION:
			rt = append(rt, ResolvedToken{
				Type: TOKEN_INSTRUCTION,
				Size: size,
				Value: t.Value,
			})
			i++
		case TOKEN_LITERAL:
			rt = append(rt, ResolvedToken{
				Type: TOKEN_LITERAL,
				Size: uint64(len(t.Value)),
				Value: t.Value,
			})
			i++
		case TOKEN_ARG:
			av, islabel := labels[t.Value]
			if !islabel {
				var err error
				av, err = arg(t.Value)
				if err != nil {
					return nil, errors.New(fmt.Sprintf("Cannot resolve argument: %v", err))
				}
			}

			rt[i - 1].Args = append(rt[i - 1].Args, av)
		}
	}

	return rt, nil
}

func getHexNumber(r rune) (uint8, error) {
	if 0x30 <= r && r <= 0x39 {
		return uint8(r - 0x30), nil
	} else if 0x41 <= r && r <= 0x46 {
		return uint8(r - 0x37), nil
	} else if 0x61 <= r && r <= 0x66 {
		return uint8(r - 0x57), nil
	} else {
		return 0, errors.New("Malformed hex number")
	}
}

// Parses everything cleanly excepts that any % followed by a two-digit hex
// number (e.g., %FA) in substituted by it's own value. Use %% for a literal %.
func ParseLiteral(lit string) string {
	var b strings.Builder
	i := 0
	for i < len(lit) {
		if lit[i] == '%' {
			if i + 1 < len(lit) && lit[i + 1] == '%' {
				b.WriteRune('%')
			} else if i + 2 < len(lit) {
				b2, err := getHexNumber(rune(lit[i + 1]))
				if err != nil {
					b.WriteRune('%')
					i++
					continue
				}
				b1, err := getHexNumber(rune(lit[i + 2]))
				if err != nil {
					b.WriteRune('%')
					i++
					continue
				}
				b.WriteRune(rune(b1 | (b2 << 4)))
				i += 3
			}
		} else {
			b.WriteRune(rune(lit[i]))
			i++
		}
	}
	return b.String()
}

// Tokenize uses a very standard Assembly syntax to create a token array:
// - Define labels with :
// - Arguments are separated by ,
// - Comments start with ; and go to the end of the line
// - Literals must be placed in lines beggining with # and are parsed by ParseLiteral
func Tokenize(asm string) []Token {
	var tokens []Token
	// This handles trailing newlines and carriage returns ;).
	for _, line := range strings.Split(strings.TrimRight(asm, "\n"), "\n") {
		if len(line) == 0 {
			continue
		}
		if line[0] == '#' {
			tokens = append(tokens, Token{Type: TOKEN_LITERAL, Value: ParseLiteral(line[1:])})
			continue
		}
		uncommented, _, _ := strings.Cut(line, ";")
		uncommented = strings.TrimSpace(uncommented)
		if len(uncommented) == 0 {
			continue
		}

		label, ins, has_label := strings.Cut(uncommented, ":")
		if !has_label {
			ins = label
		} else {
			tokens = append(tokens, Token{Type: TOKEN_LABEL, Value: label})
		}

		ins = strings.TrimSpace(ins)
		if len(ins) == 0 {
			continue
		}

		mne, args, has_mne := strings.Cut(ins, " ")
		// We already trimmed it so there's no way mne == "".
		tokens = append(tokens, Token{Type: TOKEN_INSTRUCTION, Value: mne})
		if has_mne {
			for _, arg := range strings.Split(args, ",") {
				arg = strings.TrimSpace(arg)
				tokens = append(tokens, Token{Type: TOKEN_ARG, Value: arg})
			}
		}
	}

	return tokens
}
