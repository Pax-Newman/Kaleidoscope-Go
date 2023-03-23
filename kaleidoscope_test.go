package main

import (
	"bytes"
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
func matchTokens(in string, want Token, got Token, t *testing.T) {
	if want != got {
		t.Fatalf(`Lexer(%s) returned %T%s... expected %T%s`, in, got, got, want, want)
	}
}

func matchErrors(in string, want error, got error, t *testing.T) {
	if got == nil {
		t.Fatalf(`Lexer(%s) did not return an error... expected %s`, in, want)
	} else if want != got {
		t.Fatalf(`Lexer(%s) returned %s... expected %s`, in, got, want)
	}
}

// -------------------------- Keywords

// Test when a correct def keyword is passed to the lexer
func TestLexerDef(t *testing.T) {
	in := "def"

	lexer := newTestLexer(in)
	ch := lexer.Tokens()

	go lexer.Run()

	want := DefToken{}
	got := <-ch
	matchTokens(in, want, got, t)
}

// Test when a correct extern keyword is passed to the lexer
func TestLexerExtern(t *testing.T) {
	in := "extern"
	lexer := newTestLexer(in)
	ch := lexer.Tokens()

	go lexer.Run()

	want := ExternToken{}
	got := <-ch
	matchTokens(in, want, got, t)
}

// -------------------------- Primary

// Test when a correct identifier is passed to the lexer
func TestLexerIdentifier(t *testing.T) {
	in := "hello"
	lexer := newTestLexer(in)
	ch := lexer.Tokens()

	go lexer.Run()

	want := IdentifierToken{"hello"}
	got := <-ch
	matchTokens(in, want, got, t)
}

// Test when a correct float is passed to the lexer
func TestLexerNumber(t *testing.T) {
	in := "9.3"
	lexer := newTestLexer(in)
	ch := lexer.Tokens()

	go lexer.Run()

	want := NumberToken{in}
	got := <-ch

	matchTokens(in, want, got, t)
}

// Test when an incorrect float is passed to the lexer
func TestLexerBadNumber(t *testing.T) {
	in := "9.3.3.3.3"
	lexer := newTestLexer(in)
	ch := lexer.Tokens()

	go lexer.Run()

	// TODO decide what the goal for this test actually is
	want := NumberToken{"9.3"}
	
	// Consume the token to trigger the error
	got := <-ch
	
	matchTokens(in, want, got, t)
}

// -------------------------- EOF Marker

func TestLexerEOF(t *testing.T) {
	in := "def"
	lexer := newTestLexer(in)
	ch := lexer.Tokens()

	go lexer.Run()

	want := EOFToken{}
	var got Token

	got = <-ch // this should be a deftoken
	got = <-ch
	matchTokens(in, want, got, t)
}
