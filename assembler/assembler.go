// Package assembler implements a library for creating assemblers for the EGG
// emulator.
package assembler

import (
	"bufio"
	"fmt"
	"os"

	"github.com/gboncoffee/intergo"
)

const (
	TOKEN_LABEL = iota
	TOKEN_INSTRUCTION
	TOKEN_ARG
	TOKEN_LITERAL
)

// This variable is a workaround between circular imports: ideally, we would
// just use InterCtx, but we cannot import machine here as machine
// already imports us. The main function will point this to the machine one.
// Please don't touch.
var InterCtx *intergo.InterContext

// Can be a label, an instruction, an argument or a literal. All directives are
// resolved in the tokenizer stage.
type Token struct {
	File  *string
	Value []byte
	Line  int
	Type  int
}

// This token is specifically an instruction. An array of these is passed to the
// debugger.
type DebuggerToken struct {
	Instruction string
	Label       string
	File        *string
	Args        []string
	Address     uint64
	Line        int
}

// This kind of token can be only instructions and literals (only instructions
// have arguments). Labels and arguments are already "resolved". The Reserved
// field is reserved for architecture-specific use (example: storing information
// regarding the addressing mode in 6502).
type ResolvedToken struct {
	File     *string
	Args     []uint64
	Value    []byte
	Address  uint64
	Line     int
	Type     int
	Reserved uintptr
}

// For intermediate reasons.
type Instruction struct {
	Mnemonic string
	File     *string
	Args     []string
	Size     uint64
	Line     int
	Reserved uintptr
}

// If the token is not a label, uses the provided function to translate it.
func TranslateArgument(arg string, labels map[string]uint64, translateArg func(string) (uint64, error)) (uint64, error) {
	value, hasLabel := labels[arg]
	if !hasLabel {
		var err error
		value, err = translateArg(arg)
		if err != nil {
			return 0, err
		}
	}

	return value, nil
}

// "Resolve" tokens. The process callback can be used for basically anything,
// but it MUST set the Size field. For example, it may remove parenthesis from
// an argument when the architecture accepts them, and set something with the
// Reserved field informing that the addressing mode of the instruction is XYZ.
//
// translateArg is passed directly to TranslateArgument.
func ResolveTokens(tokens []Token, process func(*Instruction) error, translateArg func(string) (uint64, error)) ([]ResolvedToken, []DebuggerToken, error) {
	resolvedTokens := []ResolvedToken{}
	labels := make(map[string]uint64)
	reverseLabels := make(map[uint64]string)
	address := uint64(0)

	// We use this so we can process everything and only after translate
	// the arguments.
	arguments := make(map[uint64][]string)

	for i := 0; i < len(tokens); i++ {
		token := &tokens[i]
		switch token.Type {
		// Token before instruction.
		case TOKEN_ARG:
			panic(InterCtx.Get("If you're reading this, there's a bug in the emulator. Please fill an issue at https://github.com/gboncoffee/egg reporting the bug with the Assembly you're trying to run and command line arguments you used to run EGG."))
		case TOKEN_LABEL:
			labels[string(token.Value)] = address
			reverseLabels[address] = string(token.Value)
		case TOKEN_LITERAL:
			resolvedTokens = append(resolvedTokens, ResolvedToken{
				Line:    token.Line,
				File:    token.File,
				Type:    TOKEN_LITERAL,
				Address: address,
				Value:   token.Value,
			})
			address += uint64(len(token.Value))
		case TOKEN_INSTRUCTION:
			instruction := Instruction{
				Line:     token.Line,
				File:     token.File,
				Mnemonic: string(token.Value),
				Args:     []string{},
			}

			// Squeeze all arguments into the Intruction variable.
			i++
			for i < len(tokens) && tokens[i].Type == TOKEN_ARG {
				instruction.Args = append(instruction.Args, string(tokens[i].Value))
				i++
			}
			i--

			if err := process(&instruction); err != nil {
				return nil, nil, err
			}

			// Finally create a proper token and append it. The arguments are
			// going to be treated only after we finish with all labels, of
			// course.
			resolvedTokens = append(resolvedTokens, ResolvedToken{
				Line:     instruction.Line,
				File:     instruction.File,
				Type:     TOKEN_INSTRUCTION,
				Value:    []byte(instruction.Mnemonic),
				Address:  address,
				Reserved: instruction.Reserved,
			})

			arguments[address] = instruction.Args
			address += instruction.Size
		}
	}

	// Now that we have all labels, we can treat the arguments. We also create
	// the debugger tokens.
	debuggerTokens := []DebuggerToken{}
	for i := 0; i < len(resolvedTokens); i++ {
		token := &resolvedTokens[i]
		if token.Type == TOKEN_INSTRUCTION {
			args, ok := arguments[token.Address]
			if !ok {
				panic(InterCtx.Get("If you're reading this, there's a bug in the emulator. Please fill an issue at https://github.com/gboncoffee/egg reporting the bug with the Assembly you're trying to run and command line arguments you used to run EGG."))
			}

			debuggerTokens = append(debuggerTokens, DebuggerToken{
				Line:        token.Line,
				File:        token.File,
				Instruction: string(token.Value),
				Args:        args,
				Address:     token.Address,
				Label:       reverseLabels[token.Address],
			})

			for _, arg := range args {
				result, err := TranslateArgument(arg, labels, translateArg)
				if err != nil {
					return nil, nil, fmt.Errorf(InterCtx.Get("%v:%v: Error on argument translation: %v"), *token.File, token.Line, err)
				}
				token.Args = append(token.Args, result)
			}
		}
	}

	return resolvedTokens, debuggerTokens, nil
}

// Tokenize recursively (as of .include directives) creates a Token array from
// file names. I.e., it opens and reads the passed file, opening and reading
// other files when reaching a .include.
func Tokenize(fileName string, tokens *[]Token) error {
	// Private functions used here are defined in tokenizer.go for the sake
	// of organization.

	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf(InterCtx.Get("couldn't open file: %v"), err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// This reads line by line. The i is necessary so we actually knows where
	// we're in the file.
	i := 1
	for scanner.Scan() {
		line := scanner.Text()
		if err := scanner.Err(); err != nil {
			return fmt.Errorf(InterCtx.Get("error reading file %v: %v"), fileName, err)
		}

		err = parseLine(&fileName, &line, i, tokens)
		if err != nil {
			return err
		}

		i++
	}

	return nil
}
