package main

import (
	"bytes"
	"fmt"
	"testing"
)

// -------------------------- Utilities

// Creates a new lexer that reads from a given string
func newTestLexer(in string) *Lexer {
	return NewLexer(
		bytes.NewReader(
			[]byte(in),
		),
	)
}

// Tests if two tokens match eachother, reports a test failure if they do not
func matchTokens(in string, want Token, got Token, err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
	} else if want != got {
		t.Fatalf(`GetTok(%s) = %T%s, %s want %T%s, nil`, in, got, got, err, want, want)
	}
}

// -------------------------- Keywords

// Test when a correct def keyword is passed to the lexer
func TestLexerDef(t *testing.T) {
	in := "def"

	fmt.Println("Making Lexer")
	lexer := newTestLexer(in)

	fmt.Println("Getting Channel")
	ch := lexer.Tokens()

	fmt.Println("Running Lexer")
	go lexer.Run()

	want := DefToken{}

	fmt.Println("Grabbing token")
	got := <-ch
	fmt.Println("Got Token")
	fmt.Println(got)
	err := lexer.err
	matchTokens(in, want, got, err, t)
}

// Test when a correct extern keyword is passed to the lexer
func TestLexerExtern(t *testing.T) {
	in := "extern"
	lexer := newTestLexer(in)

	want := ExternToken{}
	got, err := lexer.GetTok()
	matchTokens(in, want, got, err, t)
}

// -------------------------- Primary

// Test when a correct identifier is passed to the lexer
func TestLexerIdentifier(t *testing.T) {
	in := "hello"
	lexer := newTestLexer(in)

	want := IdentifierToken{"hello"}
	got, err := lexer.GetTok()
	matchTokens(in, want, got, err, t)
}

// Test when a correct float is passed to the lexer
func TestLexerNumber(t *testing.T) {
	in := "9.3"
	lexer := newTestLexer(in)

	want := NumberToken{9.3}
	got, err := lexer.GetTok()
	matchTokens(in, want, got, err, t)
}

// Test when an incorrect float is passed to the lexer
func TestLexerBadNumber(t *testing.T) {
	in := "9.3.3.3.3"
	lexer := newTestLexer(in)

	want := NumberToken{9.3}
	got, err := lexer.GetTok()
	matchTokens(in, want, got, err, t)
}

// -------------------------- EOF Marker

func TestLexerEOF(t *testing.T) {
	in := "def"
	lexer := newTestLexer(in)

	// consume def, the next gettok should receive an EOF error and correctly return an EOFToken
	lexer.GetTok()

	want := EOFToken{}
	got, err := lexer.GetTok()
	matchTokens(in, want, got, err, t)
}
