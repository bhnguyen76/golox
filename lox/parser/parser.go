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

func (p *Parser) Parse() []ast.Stmt {
	var statements []ast.Stmt

	for !p.isAtEnd() {
		stmt := p.statement()
		if stmt != nil {
			statements = append(statements, p.declaration())
		}
	}

	return statements
}

func NewParser(tokens []scanner.Token) *Parser {
	return &Parser{
		tokens: tokens,
		current: 0,
	}
}

func (p *Parser) expression() ast.Expr {
	return p.assignment()
}

func (p *Parser) statement() ast.Stmt {
	if p.match(scanner.FOR) {
		return p.forStatement()
	}

	if p.match(scanner.IF) {
		return p.ifStatement()
	}

	if p.match(scanner.PRINT) {
		return p.printStatment()
	}

	if p.match(scanner.RETURN) {
		return p.returnStatement()
	}

	if p.match(scanner.WHILE) {
		return p.whileStatement()
	}

	if p.match(scanner.LEFT_BRACE) {
		return &ast.Block{
			Statements: p.block(),
		}	
	}

	return p.expressionStatement()
}

func (p *Parser) printStatment() ast.Stmt {
	value := p.expression()
	p.consume(scanner.SEMICOLON, "Expect ';' after value.")
	return &ast.Print{
		Expression: value,
	}
}

func (p *Parser) expressionStatement() ast.Stmt {
	expr := p.expression()
	p.consume(scanner.SEMICOLON, "Expect ';' after expression.")
	return &ast.Expression{
		Expression: expr,
	}
}

func (p *Parser) function(kind string) *ast.Function {
	name := p.consume(scanner.IDENTIFIER, "Expect " + kind + " name.")

	p.consume(scanner.LEFT_PAREN, "Expect '(' after " + kind + " name.")

	var parameters []scanner.Token
	if !p.check(scanner.RIGHT_PAREN) {
		for {
			if len(parameters) >= 255 {
				p.error(p.peek(), "Can't have more than 255 parameters.")
			}

			param := p.consume(scanner.IDENTIFIER, "Expect parameter name.")
			parameters = append(parameters, param)

			if !p.match(scanner.COMMA) {
				break
			}
		}
	}
	p.consume(scanner.RIGHT_PAREN, "Expect ')' after parameters.")

	p.consume(scanner.LEFT_BRACE, "Expect '{' before " + kind + " body.")
	body := p.block()

	return &ast.Function{
		Name: name,
		Params: parameters,
		Body: body,
	}
}

func (p *Parser) ifStatement() ast.Stmt {
	p.consume(scanner.LEFT_PAREN, "Expect '(' after 'if'.")
	condition := p.expression()
	p.consume(scanner.RIGHT_PAREN, "Expect ')' after if condition.")

	thenBranch := p.statement()

	var elseBranch ast.Stmt = nil
	if p.match(scanner.ELSE) {
		elseBranch = p.statement()
	}

	return &ast.If{
		Condition: condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}
}


func (p *Parser) assignment() ast.Expr {
	var expr ast.Expr = p.or()

	if p.match(scanner.EQUAL) {
		equals := p.previous()
		value := p.assignment()

		switch e := expr.(type) {
        case *ast.Variable:
            name := e.Name
            return &ast.Assign{
                Name:  name,
                Value: value,
            }

        case *ast.Get:
            return &ast.Set{
                Object: e.Object,
                Name:   e.Name,
                Value:  value,
            }

        default:
            p.error(equals, "Invalid assignment target.")
        }
	}

		return expr
}

func (p *Parser) or() ast.Expr {
	expr := p.and()

	for p.match(scanner.OR) {
		operator := p.previous()
		right := p.and()
		expr = &ast.Logical{
			Left: expr,
			Operator: operator,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) and() ast.Expr {
	expr := p.equality()

	for p.match(scanner.AND) {
		operator := p.previous()
		right := p.equality()
		expr = &ast.Logical{
			Left: expr,
			Operator: operator,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) whileStatement() ast.Stmt {
	p.consume(scanner.LEFT_PAREN, "Expect '(' after 'while'.")
	condition := p.expression()
	p.consume(scanner.RIGHT_PAREN, "Expect ')' after condition.")
	body := p.statement()

	return &ast.While{
		Condition: condition,
		Body: body,
	}
}

func (p *Parser) forStatement() ast.Stmt {
	p.consume(scanner.LEFT_PAREN, "Expect '(' after 'for'.")

	var initializer ast.Stmt 
	if p.match(scanner.SEMICOLON) {
		initializer = nil
	} else if p.match(scanner.VAR) {
		initializer = p.varDeclaration()
	} else {
		initializer = p.expressionStatement()
	}

	var condition ast.Expr 
	if !p.check(scanner.SEMICOLON) {
		condition = p.expression()
	}
	p.consume(scanner.SEMICOLON, "Expect ';' after loop condition.")

	var increment ast.Expr 
	if !p.check(scanner.RIGHT_PAREN) {
		increment = p.expression()
	}
	p.consume(scanner.RIGHT_PAREN, "Expect ')' after for clauses.")

	body := p.statement()

	if increment != nil {
		body = &ast.Block{
			Statements: []ast.Stmt{
				body,
				&ast.Expression{Expression: increment},
			},
		}
	}

	if condition == nil {
		condition = &ast.Literal{Value: true}
	}
	body = &ast.While{
		Condition: condition,
		Body: body,
	}

	if initializer != nil {
		body = &ast.Block{
			Statements: []ast.Stmt{
				initializer,
				body,
			},
		}
	}

	return body
}

func (p *Parser) returnStatement() ast.Stmt {
	keyword := p.previous()

	var value ast.Expr = nil
	if !p.check(scanner.SEMICOLON) {
		value = p.expression()
	}

	p.consume(scanner.SEMICOLON, "Expect ';' after return value.")
	
	return &ast.Return{
		Keyword: keyword,
		Value: value,
	}
}

func (p *Parser) declaration() (stmt ast.Stmt) {
	defer func() {
		if r := recover(); r != nil {
			// If it's a parseError, synchronize and return nil statement.
			if _, ok := r.(parseError); ok {
				p.synchronize()
				stmt = nil
				return
			}
			// Not a parseError → re-panic (real bug)
			panic(r)
		}
	}()
	
	if p.match(scanner.CLASS) {
		return p.classDeclaration()
	}
	if p.match(scanner.FUN) {
		return p.function("function")
	}
	if p.match(scanner.VAR) {
		return p.varDeclaration()
	}

	return p.statement()
}

func (p *Parser) varDeclaration() ast.Stmt {
	name := p.consume(scanner.IDENTIFIER, "Expect variable name.")

	var initializer ast.Expr = nil
	if p.match(scanner.EQUAL) {
		initializer = p.expression()
	}

	p.consume(scanner.SEMICOLON, "Expect ';' after variable declaration.")
	return &ast.Var{
		Name		: name,
		Initializer : initializer,
	}
}

func (p *Parser) classDeclaration() ast.Stmt {
    name := p.consume(scanner.IDENTIFIER, "Expect class name.")

	var superclass ast.Expr 
	if p.match(scanner.LESS) {
		p.consume(scanner.IDENTIFIER, "Expect superclass name.")
		superclass = &ast.Variable{
			Name: p.previous(),
		}
	}

    p.consume(scanner.LEFT_BRACE, "Expect '{' before class body.")

    var methods []*ast.Function
    for !p.check(scanner.RIGHT_BRACE) && !p.isAtEnd() {
        methods = append(methods, p.function("method"))
    }

    p.consume(scanner.RIGHT_BRACE, "Expect '}' after class body.")

    return &ast.Class{
        Name:    name,
		Superclass: superclass,
        Methods: methods,
    }
}

func (p *Parser) block() []ast.Stmt {
	var statements []ast.Stmt 

	for !p.check(scanner.RIGHT_BRACE) && !p.isAtEnd() {
		stmt := p.declaration()
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}

	p.consume(scanner.RIGHT_BRACE, "Expect '}' after block.")
	return statements
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

	return p.call()
}

func (p *Parser) call() ast.Expr {
	expr := p.primary()

	for true {
		if p.match(scanner.LEFT_PAREN) {
			expr = p.finishCall(expr)
		} else if p.match(scanner.DOT) {
			name := p.consume(scanner.IDENTIFIER, "Expect property name after '.'.")
			expr = &ast.Get{
				Object: expr,
				Name: name,
			}
		} else {
			break
		}
	}

	return expr
}

func (p *Parser) finishCall(callee ast.Expr) ast.Expr {
	var arguments []ast.Expr
	if !p.check(scanner.RIGHT_PAREN) {
		for {
			if len(arguments) >= 255 {
				p.error(p.peek(), "Can't have more than 255 arguments.")
			}

			arguments = append(arguments, p.expression())

			if !p.match(scanner.COMMA) {
				break
			}
		}
	}

	paren := p.consume(scanner.RIGHT_PAREN, "Expect ')' after arguments.")

	return &ast.Call{
		Callee: callee,
		Paren: paren,
		Arguments: arguments,
	}
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

	if p.match(scanner.SUPER) {
		keyword := p.previous()
		p.consume(scanner.DOT, "Expect '.' after 'super'.")
		method := p.consume(scanner.IDENTIFIER, "Expect superclass method name.")
		return &ast.Super{
			Keyword: keyword,
			Method: method,
		}
	}

	if p.match(scanner.THIS) {
		return &ast.This{
			Keyword: p.previous(),
		}
	}

	if p.match(scanner.IDENTIFIER) {
		return &ast.Variable{Name: p.previous()}
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

