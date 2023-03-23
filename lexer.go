package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"reflect"
	"runtime"
	"strings"
	"unicode"
)

// Generic token interface
type Token interface{}

// Error token
type ErrorToken struct{ string }

// EOF marker
type EOFToken struct{}

// Keywords
type DefToken struct{}
type ExternToken struct{}

// Primary
type IdentifierToken struct{ string }
type NumberToken struct{ string }

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
	// fmt.Println("Running inside the goroutine")
	// Start the state transition loop
	for state := lexNext; state != nil; {
		// fname := GetFunctionName(state)
		// fmt.Printf("Now running: %s\n", fname)
		state = state(lex)
		// Check for an error & report it
		if lex.err != nil {
			log.Fatalf("Error encountered at line %d:%d -> %s", lex.line, lex.col, lex.err)
		}
	}
	// fmt.Println("Closing Channel")
	// Close out our tokens channel
	close(lex.tokens)
}

func (lex *Lexer) errorf(format string, args ...interface{}) stateFn {
	lex.tokens <- ErrorToken{ fmt.Sprintf(format, args...) }
	return nil
}

func (lex *Lexer) next() rune {
	char, _, err := lex.reader.ReadRune()
	if err != io.EOF {
		lex.err = err
	}
	return char
}

func (lex *Lexer) back() error {
	err := lex.reader.UnreadRune()
	return err
}

func (lex *Lexer) peek() (rune, error) {
	char, err := lex.reader.Peek(1)
	if err == io.EOF {
		return 0, nil
	}
	return rune(char[0]), err
}

// Consumes one rune if it's in a valid set of runes.
// Returns the consumed rune, or 0 if it wasn't valid
func (lex *Lexer) accept(valid string) rune {
	next := lex.next()
	if strings.IndexRune(valid, next) >= 0 {
		return next
	} else {
		lex.back()
		return 0
	}
}

// Consumes a series of runes that are all in a valid set of runes
func (lex *Lexer) acceptRun(valid string) []rune {
	var runes []rune
	for next := lex.next(); strings.IndexRune(valid, next) >= 0; next = lex.next() {
		runes = append(runes, next)
	}
	lex.back()

	return runes
}

func (lex *Lexer) emit(tok Token) {
	lex.tokens <- tok
}

// State func to lex the next token?
func lexNext(lex *Lexer) stateFn {
	nextChar, err := lex.peek()

	if err != nil {
		return lex.errorf("lexText: %s", err.Error())
	}
	
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
	rawNum := ""
	
	// Catch any leading positive/negative signs
	rawNum += string(lex.accept("+-"))

	digits := "0123456789"
	
	// Catch a hex specifier and augment our valid digits to include hex values
	hex := string(lex.accept("0") + lex.accept("xX"))
	if len(hex) >= 2 {
		digits += "abcdefABCDEF"
	}
	rawNum += hex
	
	// Consume the following valid digits
	rawNum += string(lex.acceptRun(digits))

	// Consume any decimal point and digits following it 
	if decPoint := lex.accept(".");	decPoint != 0 {
		rawNum += string(decPoint)
		rawNum += string(lex.acceptRun(digits))
	}

	cleanNum := strings.ReplaceAll(rawNum, string(rune(0)), "")

	lex.emit(NumberToken{cleanNum})

	return lexNext
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

	return lexNext
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
	err := lex.back()
	if err != nil {
		return lex.errorf("lexSpace: %s", err.Error())
	}
	// go back to lexing the next piece of text
	return lexNext
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
