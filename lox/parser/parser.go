package parser

import (
	"example.com/golox/lox/ast"
	"example.com/golox/lox/scanner"
	"example.com/golox/lox/shared"
)

type parseError struct{}

func (parseError) Error() string { return "parse error" }

type Parser struct {
	tokens []scanner.Token
	current int
}

func (p *Parser) Parse() (expr ast.Expr) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(parseError); ok {
				expr = nil
				return
			}
			panic(r)
		}
	}()

	return p.expression()
}

func NewParser(tokens []scanner.Token) *Parser {
	return &Parser{
		tokens: tokens,
		current: 0,
	}
}

func (p *Parser) expression() ast.Expr {
	return p.equality()
}

func (p *Parser) equality() ast.Expr {
	expr := p.comparison()

	for p.match(scanner.BANG_EQUAL, scanner.EQUAL_EQUAL) {
		operator := p.previous()
		right := p.comparison()
		expr = &ast.Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) match(types ...scanner.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(t scanner.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == t
}

func (p *Parser) advance() scanner.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == scanner.EOF
}

func (p *Parser) peek() scanner.Token {
	return p.tokens[p.current]
}

func (p *Parser) previous() scanner.Token {
	return p.tokens[p.current-1]
}

func (p *Parser) comparison() ast.Expr {
	expr := p.term()

	for p.match(scanner.GREATER, scanner.EQUAL, scanner.LESS, scanner.LESS_EQUAL) {
		operator := p.previous()
		right := p.term()
		expr = &ast.Binary{
			Left: expr,
			Operator: operator,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) term() ast.Expr {
	expr := p.factor()

	for p.match(scanner.MINUS, scanner.PLUS) {
		operator := p.previous()
		right := p.factor()
		expr = &ast.Binary{
			Left: expr,
			Operator: operator,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) factor() ast.Expr {
	expr := p.unary()

	for p.match(scanner.SLASH, scanner.STAR) {
		operator := p.previous()
		right := p.unary()
		expr = &ast.Binary{
			Left: expr,
			Operator: operator,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) unary() ast.Expr {
	if p.match(scanner.BANG, scanner.MINUS) {
		operator := p.previous()
		right := p.unary()
		return &ast.Unary{
			Operator: operator,
			Right: right,
		}
	}

	return p.primary()
}

func (p *Parser) primary() ast.Expr {
	if p.match(scanner.FALSE) {
		return &ast.Literal{Value: false}
	}
	if p.match(scanner.TRUE) {
		return &ast.Literal{Value: true}
	}
	if p.match(scanner.NIL) {
		return &ast.Literal{Value: nil}
	}

	if p.match(scanner.NUMBER, scanner.STRING) {
		return &ast.Literal{Value: p.previous().Literal}
	}

	if p.match(scanner.LEFT_PAREN) {
		expr := p.expression()
		p.consume(scanner.RIGHT_PAREN, "Expect ')' after expression.")
		return &ast.Grouping{Expression: expr}
	}

	panic(p.error(p.peek(), "Expect expression."))
}

func (p *Parser) consume(t scanner.TokenType, message string) scanner.Token {
	if p.check(t) {
		return p.advance()
	}
	panic(p.error(p.peek(), message))
}

func (p *Parser) error(token scanner.Token, message string) error {
	if token.Type == scanner.EOF {
		shared.Report(token.Line, " at end", message)
	} else {
		shared.Report(token.Line, " at '" + token.Lexeme + "'", message)
	}
	
	return parseError{}
}

func (p *Parser) synchronize() {
	p.advance()

	for !p.isAtEnd() {
		if p.previous().Type == scanner.SEMICOLON {
			return
		}

		switch p.peek().Type {
		case scanner.CLASS,
			scanner.FUN,
			scanner.VAR,
			scanner.FOR,
			scanner.IF,
			scanner.WHILE,
			scanner.PRINT,
			scanner.RETURN:
			return
		}

		p.advance()
	}
}

