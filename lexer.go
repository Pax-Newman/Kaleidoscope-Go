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

// Begins lexing the input and outputting tokens through the tokens channel
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

// Prints an error message and halts the lexer
func (lex *Lexer) errorf(format string, args ...interface{}) stateFn {
	lex.tokens <- ErrorToken{ fmt.Sprintf(format, args...) }
	return nil
}

// Consumes and returns the next rune
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

// Emits a single token through the lexer's token channel
func (lex *Lexer) emit(tok Token) {
	lex.tokens <- tok
}

// Determines the appropriate lexing stateFn to use next
func lexNext(lex *Lexer) stateFn {
	nextChar, err := lex.peek()

	if err != nil {
		return lex.errorf("lexText: %s", err.Error())
	}
	
	// Move to the next state based on what the next rune in the input is
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

// Lexes the next number
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
	
	// accept and acceptRun will insert 0-value runes into our number, here we remove them to clean the number
	cleanNum := strings.ReplaceAll(rawNum, string(rune(0)), "")

	lex.emit(NumberToken{cleanNum})

	return lexNext
}

// Lexes the next identifier
func lexIdentifier(lex *Lexer) stateFn {
	// Consume runes until they aren't letters or numbers
	// TODO Change this to possibly blacklist bad characters like whitespace, rather than whitelist numbers & letters
	id := []rune{}
	for nextChar := lex.next(); nextChar != 0 && (unicode.IsNumber(nextChar) || unicode.IsLetter(nextChar)); {
		id = append(id, nextChar)
		nextChar = lex.next()
	}
	lex.back()

	// Identify the token and emit it
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

// Lexes the next whitespace
func lexSpace(lex *Lexer) stateFn {
	// move through the whitespace until it's no longer whitespace
	lex.acceptRun(" \n\r")

	// go back to lexing the next piece of text
	return lexNext
}

// Lexes the EOF
func lexEOF(lex *Lexer) stateFn {
	_, _, err := lex.reader.ReadRune()
	if err == io.EOF {
		lex.emit(EOFToken{})
	} else {
		// if it's not an EOF error then handle it appropriately
		lex.errorf("lexEOF: %s", err)
	}
	// Since we hit EOF, end the lexing session
	return nil
}
