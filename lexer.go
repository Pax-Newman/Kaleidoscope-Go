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

// TODO add line and maybe col fields to identify where an error occured
// Lexer Class
type Lexer struct {
	reader bufio.Reader
	tokens chan Token
	err    error
	line   int
	col    int
}

type stateFn func(l *Lexer) stateFn

// Inits a new Lexer Object
func NewLexer(rd io.Reader) *Lexer {
	return &Lexer{
		reader: *bufio.NewReader(rd),
		tokens: make(chan Token),
		line:   0,
		col:    0,
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
			log.Fatalf("Error encountered at line %d:%d -> %s", lex.line, lex.col, lex.err)
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
	lex.col += 1
	return char
}

func (lex *Lexer) back() {
	// TODO consider how to handle this error
	lex.reader.UnreadRune()
	if char := lex.peek(); char == '\n' || char == '\r' {
		lex.col = 0
		lex.line -= 1
	} else {
		lex.col -= 1
	}
}

func (lex *Lexer) peek() rune {
	char, err := lex.reader.Peek(0)
	if err != io.EOF {
		lex.err = err
	}
	if len(char) <= 0 {
		return 0
	}
	return rune(char[0])
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
		if nextChar == '\n' || nextChar == '\r' {
			lex.line += 1
			lex.col = 0
		}
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
