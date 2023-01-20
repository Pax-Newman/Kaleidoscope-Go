package main

import (
	"bytes"
	"fmt"
)

func main() {
	in := `def helloworld
	return 10`

	fmt.Println("Creating Lexer")
	lex := NewLexer(
		bytes.NewReader(
			[]byte(in),
		),
	)

	fmt.Println("Getting Channel")
	tokens := lex.Tokens()

	fmt.Println("Running Lexer")
	go lex.Run()

	for tok := range tokens {
		fmt.Printf("Got Token: %T%s\n", tok, tok)
	}

}
