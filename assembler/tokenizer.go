package assembler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Adds some specific bytes. Like a literal but easier with numbers. The size
// argument shall be 1, 2, 4 or 8.
func bitsDirective(fileName *string, lineNum int, args *string, size int, tokens *[]Token) error {
	*args = strings.TrimSpace(*args)
	if len(*args) == 0 {
		return fmt.Errorf(InterCtx.Get("%v:%v: Expected literal bytes after bits directive"), *fileName, lineNum)
	}

	bitsize := size * 8

	argsSlice := strings.Split(*args, " ")
	literal := make([]byte, size*len(argsSlice))
	for i, arg := range argsSlice {
		n, err := strconv.ParseUint(arg, 0, bitsize)
		if err != nil {
			return fmt.Errorf(InterCtx.Get("%v:%v: Cannot convert %v to a %v bits number"), *fileName, lineNum, arg, bitsize)
		}

		for j := 0; j < size; j++ {
			literal[i*size+j] = byte(n)
			n = n >> 8
		}
	}

	*tokens = append(*tokens, Token{
		Line:  lineNum,
		File:  fileName,
		Type:  TOKEN_LITERAL,
		Value: literal,
	})

	return nil
}

// Creates an empty literal with the number in *args as the size (in bytes).
func spaceDirective(fileName *string, lineNum int, args *string, tokens *[]Token) error {
	if len(*args) == 0 {
		return fmt.Errorf(InterCtx.Get("%v:%v: Expected a number of bytes after space directive"), *fileName, lineNum)
	}
	n, err := strconv.ParseUint(*args, 0, 64)
	if err != nil {
		return fmt.Errorf(InterCtx.Get("%v:%v: Cannot create space: Cannot parse %v to number: %v"), *fileName, lineNum, *args, err)
	}
	*tokens = append(*tokens, Token{
		Line:  lineNum,
		File:  fileName,
		Type:  TOKEN_LITERAL,
		Value: make([]byte, n),
	})

	return nil
}

func getHexNumber(r rune) (uint8, error) {
	if 0x30 <= r && r <= 0x39 {
		return uint8(r - 0x30), nil
	} else if 0x41 <= r && r <= 0x46 {
		return uint8(r - 0x37), nil
	} else if 0x61 <= r && r <= 0x66 {
		return uint8(r - 0x57), nil
	} else {
		return 0, errors.New("malformed hex number")
	}
}

// Line shall already be trimmed so the function works both with .literal and #.
// Parses everything cleanly excepts that any % followed by a two-digit hex
// number (e.g., %FA) in substituted by it's own value. Use %% for a literal %.
// Trailing % are also inserted.
func parseLiteral(fileName *string, line *string, lineNum int, tokens *[]Token) {
	var b strings.Builder
	lit := *line
	i := 0
	for i < len(lit) {
		if lit[i] == '%' {
			if i+1 < len(lit) && lit[i+1] == '%' {
				b.WriteRune('%')
				i += 2
			} else if i+2 < len(lit) {
				b2, err := getHexNumber(rune(lit[i+1]))
				if err != nil {
					b.WriteRune('%')
					i++
					continue
				}
				b1, err := getHexNumber(rune(lit[i+2]))
				if err != nil {
					b.WriteRune('%')
					i++
					continue
				}
				b.WriteRune(rune(b1 | (b2 << 4)))
				i += 3
			} else {
				b.WriteRune('%')
				i++
			}
		} else {
			b.WriteRune(rune(lit[i]))
			i++
		}
	}

	*tokens = append(*tokens, Token{
		File:  fileName,
		Line:  lineNum,
		Type:  TOKEN_LITERAL,
		Value: []byte(b.String()),
	})
}

// Line shall already be trimmed.
func parseDirective(fileName *string, line *string, lineNum int, tokens *[]Token) error {
	if len(*line) == 0 {
		return fmt.Errorf(InterCtx.Get("%v:%v: Expected a directive name"), *fileName, lineNum)
	}

	name, arg, _ := strings.Cut(*line, " ")
	switch name {
	case "include":
		file := strings.TrimSpace(arg)
		if len(file) == 0 {
			return fmt.Errorf(InterCtx.Get("%v:%v: Expected file name to include"), *fileName, lineNum)
		}
		return Tokenize(file, tokens)
	case "literal":
		lit := strings.TrimSpace(arg)
		if len(lit) == 0 {
			return fmt.Errorf(InterCtx.Get("%v:%v: Expected literal content"), *fileName, lineNum)
		}
		parseLiteral(fileName, &lit, lineNum, tokens)
		return nil
	case "bits8":
		return bitsDirective(fileName, lineNum, &arg, 1, tokens)
	case "bits16":
		return bitsDirective(fileName, lineNum, &arg, 2, tokens)
	case "bits32":
		return bitsDirective(fileName, lineNum, &arg, 4, tokens)
	case "bits64":
		return bitsDirective(fileName, lineNum, &arg, 8, tokens)
	case "space":
		return spaceDirective(fileName, lineNum, &arg, tokens)
	}

	return fmt.Errorf(InterCtx.Get("%v:%v: Unknown directive %v"), *fileName, lineNum, name)
}

func parseInstruction(fileName *string, line *string, lineNum int, tokens *[]Token) {
	*line = strings.TrimSpace(*line)
	mnemonic, args, hasMne := strings.Cut(*line, " ")
	*tokens = append(*tokens, Token{
		Line:  lineNum,
		File:  fileName,
		Type:  TOKEN_INSTRUCTION,
		Value: []byte(mnemonic),
	})

	if hasMne {
		for _, arg := range strings.Split(args, ",") {
			arg = strings.TrimSpace(arg)
			*tokens = append(*tokens, Token{
				Line:  lineNum,
				File:  fileName,
				Type:  TOKEN_ARG,
				Value: []byte(arg),
			})
		}
	}
}

func parseLine(fileName *string, line *string, lineNum int, tokens *[]Token) error {
	// This uncomments and trims the line.
	*line, _, _ = strings.Cut(*line, ";")
	*line = strings.TrimSpace(*line)
	if len(*line) == 0 {
		return nil
	}

	// If a literal.
	if (*line)[0] == '#' {
		if len(*line) <= 1 {
			return fmt.Errorf(InterCtx.Get("%v:%v: Expected literal content"), *fileName, lineNum)
		}
		*line = (*line)[1:]
		parseLiteral(fileName, line, lineNum, tokens)
		return nil
	}
	// If a directive.
	if (*line)[0] == '.' {
		*line = (*line)[1:]
		*line = strings.TrimSpace(*line)
		return parseDirective(fileName, line, lineNum, tokens)
	}

	// Now we check if there's a label declared there.
	beg, end, hasLabel := strings.Cut(*line, ":")
	if hasLabel {
		*tokens = append(*tokens, Token{Line: lineNum,
			File:  fileName,
			Type:  TOKEN_LABEL,
			Value: []byte(strings.Clone(beg)),
		})
		beg = end
	}

	beg = strings.TrimSpace(beg)
	if len(beg) != 0 {
		// Finally we put an instruction there.
		parseInstruction(fileName, &beg, lineNum, tokens)
	}

	return nil
}
