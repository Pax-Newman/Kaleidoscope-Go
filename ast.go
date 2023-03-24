package main

import "fmt"

// -------------------- AST Types

// Base expression type
type ExprAST interface{}

// Expression type for a number
type NumberExprAST struct{ Val string }

// Expression type for a variable
type VariableExprAST struct{ Name string }

// Expression type for a binary operator
type BinaryExprAST struct {
	Op  rune
	LHS *ExprAST
	RHS *ExprAST
}

// Expression type for a function call
type CallExprAST struct {
	Callee string
	Args   []ExprAST
}

// Represent a function prototype
// Captures the name and args of a function
type PrototypeAST struct {
	Name string
	Args []string
}

// Represents a function itself
type FunctionAST struct {
	Prototype *PrototypeAST
	Body      *ExprAST
}

// -------------------- Parser Logic

// Parser handles all logic related to building/parsing the abstract syntax tree
type Parser struct {
	tokens <-chan Token
	curTok Token
}

func (p *Parser) logErrorF(format string, args ...interface{}) ExprAST {
	fmt.Printf(format, args...)

	return nil
}

// Retrieves the next token and assigns it to Parser.curTok
func (p *Parser) getNextToken() {
	p.curTok = <-p.tokens
}

func (p *Parser) parseNumberExpr() ExprAST {
	val := p.curTok.(NumberToken)
	p.getNextToken() // Consume the token
	return NumberExprAST{val.string}
}

func (p *Parser) parseParenExpr() ExprAST {
	p.getNextToken() // Consume '('

	// V := p.parseExpression()
	// if !V {
	// 	return nil
	// }

	val, ok := p.curTok.(IdentifierToken)
	if !ok || val.string != ")" {
		p.logErrorF("Expected )")
	}
	p.getNextToken() // Consume ')'

	return nil // TODO change this to return V
}
