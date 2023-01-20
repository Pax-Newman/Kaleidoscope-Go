package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"reflect"
	"runtime"
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
	err    error
}

type stateFn func(l *Lexer) stateFn

// Inits a new Lexer Object
func NewLexer(rd io.Reader) *Lexer {
	return &Lexer{
		reader: *bufio.NewReader(rd),
		tokens: make(chan Token),
	}
}

// Returns a channel to read the lexer's tokens from
func (lex *Lexer) Tokens() chan Token {
	return lex.tokens
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func (lex *Lexer) Run() {
	fmt.Println("Running inside the goroutine")
	// Start the state transition loop
	for state := lexText; state != nil; {
		fname := GetFunctionName(state)
		fmt.Printf("Now running: %s\n", fname)
		state = state(lex)
		// Check for an error & report it
		if lex.err != nil {
			log.Fatalf("Error encountered: %s", lex.err)
		}
	}
	fmt.Println("Closing Channel")
	// Close out our tokens channel
	close(lex.tokens)
}

func (lex *Lexer) next() rune {
	char, _, err := lex.reader.ReadRune()
	if err != io.EOF {
		lex.err = err
	}
	return char
}

func (lex *Lexer) back() {
	// TODO consider how to handle this error
	lex.reader.UnreadRune()
}

func (lex *Lexer) emit(tok Token) {
	lex.tokens <- tok
}

// State func to lex the next token?
func lexText(lex *Lexer) stateFn {
	nextChar := lex.next()
	lex.back()

	switch {
	case unicode.IsNumber(nextChar):
		return lexNum
	case unicode.IsLetter(nextChar):
		return lexIdentifier
	case unicode.IsSpace(nextChar):
		return lexSpace
	case nextChar == 0:
		return lexEOF
	}
	return nil
}

func lexNum(lex *Lexer) stateFn {
	unconvertedNum := ""
	for nextChar := lex.next(); nextChar != 0 && (unicode.IsNumber(nextChar) || nextChar == '.'); {
		unconvertedNum += string(nextChar)
		nextChar = lex.next()
	}
	lex.back()

	num, err := strconv.ParseFloat(unconvertedNum, 64)
	if err != nil {
		lex.err = err
	}
	lex.emit(NumberToken{num})

	return lexText
}

func lexIdentifier(lex *Lexer) stateFn {
	id := []rune{}
	for nextChar := lex.next(); nextChar != 0 && (unicode.IsNumber(nextChar) || unicode.IsLetter(nextChar)); {
		id = append(id, nextChar)
		nextChar = lex.next()
	}
	lex.back()

	idString := string(id)
	var token Token
	switch idString {
	case "def":
		token = DefToken{}
	case "extern":
		token = ExternToken{}
	default:
		token = IdentifierToken{idString}
	}
	lex.emit(token)

	return lexText
}

func lexSpace(lex *Lexer) stateFn {
	// move through the whitespace until it's no longer whitespace
	for nextChar := lex.next(); nextChar != 0 && unicode.IsSpace(nextChar); {
		nextChar = lex.next()
	}
	// move back one rune to make up for using next() to look at it in the while loop
	lex.back()
	// go back to lexing the next piece of text
	return lexText
}

func lexEOF(lex *Lexer) stateFn {
	_, _, err := lex.reader.ReadRune()
	if err == io.EOF {
		lex.emit(EOFToken{})
	} else {
		// if it's not an EOF error then handle it appropriately
		lex.err = err
	}
	// Since we hit EOF, end the lexing session
	return nil
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
