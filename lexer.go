package kaleidoscope

import (
	"bufio"
	"io"
	"strconv"
	"unicode"
)

// Generic token interface
type Token interface{}

// EOF marker
type EOFToken struct{}

// Keywords
type DefToken struct{}
type ExternToken struct{}

// Primary
type IdentifierToken struct{ string }
type NumberToken struct{ float64 }

// Lexer Class
type Lexer struct {
	reader bufio.Reader
	tokens chan Token
}

type stateFn func(l *Lexer) stateFn

// Inits a new Lexer Object
func NewLexer(rd io.Reader) *Lexer {
	return &Lexer{
		reader: *bufio.NewReader(rd),
		tokens: make(chan Token),
	}
}

func (lex Lexer) Run() {
	// Start the state transition loop
	for state := lexText; state != nil; {
		state = state(&lex)
	}
	// Close out our tokens channel
	close(lex.tokens)
}

// State func to lex the next token?
// func lexText(lex *Lexer) stateFn {
// 	...
// }

func (lex *Lexer) next() (rune, error) {
	char, _, err := lex.reader.ReadRune()
	return char, err
}

// Gets the next token from the reader
func (lex Lexer) GetTok() (Token, error) {
	// Skip whitespace runes
	char, _, err := lex.reader.ReadRune()
	if err == io.EOF {
		return EOFToken{}, nil
	} else if err != nil {
		return nil, err
	}
	for unicode.IsSpace(char) {
		char, _, err = lex.reader.ReadRune()
		if err != nil {
			return nil, err
		}
	}

	// if the first char is a letter
	if unicode.IsLetter(char) {
		// load entire identifier string into a var
		identifier := ""
		for unicode.IsLetter(char) || unicode.IsNumber(char) {
			identifier += string(char)
			char, _, err = lex.reader.ReadRune()
			// if we encounter an EOF token, place it back into the buffer and break the loop
			if err == io.EOF {
				lex.reader.UnreadRune()
				break
			} else if err != nil {
				return nil, err
			}
		}
		// check if the identifier is a keyword
		switch identifier {
		case "def":
			return DefToken{}, nil
		case "extern":
			return ExternToken{}, nil
		// if it's not a keyword, simply return the id string
		default:
			return IdentifierToken{identifier}, nil
		}

	}

	// If the first char is instead a number
	if unicode.IsDigit(char) {
		numStr := ""
		// load the rest of the number into our string representation
		// FIXME this currently has a bug where a user can write "1.2.3.4.5" and the lexer will try to turn it into a float
		for unicode.IsDigit(char) || char == '.' {
			numStr += string(char)
			char, _, err = lex.reader.ReadRune()
			// if we encounter an EOF token, place it back into the buffer and break the loop
			if err == io.EOF {
				lex.reader.UnreadRune()
				break
			} else if err != nil {
				return nil, err
			}
		}
		// turn our string into a float64
		num, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return nil, err
		}
		return NumberToken{num}, nil
	}

	// If this is a comment read through the whole line to remove it from the reader
	if char == '#' {
		// Skip the line
		for {
			char, _, err = lex.reader.ReadRune()
			// stop when we hit eof or end of the line
			if err == io.EOF || char == '\n' || char == '\r' {
				break
			}
			if err != nil {
				return nil, err
			}
		}
		// since this line was a comment, get whatever token is on the next line
		// provided we haven't hit eof yet
		if err != io.EOF {
			return lex.GetTok()
		}
	}

	if err == io.EOF {
		return EOFToken{}, nil
	}
	return nil, nil
}
