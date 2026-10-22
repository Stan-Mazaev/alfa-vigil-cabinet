package parser

import (
	"fmt"
	"vigil/internal/ast"
	"vigil/internal/lexer"
)

type Parser struct {
	l         *lexer.Lexer
	curToken  lexer.Token
	peekToken lexer.Token
	errors    []string
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}
	// Читаем два токена, чтобы инициализировать curToken и peekToken
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Errors() []string { return p.errors }

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{Statements: []ast.Statement{}}

	for p.curToken.Type != lexer.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case lexer.STREAM:
		return p.parseStreamStatement()
	case lexer.ON:
		return p.parseReactorStatement()
	case lexer.TRANSMIT:
		return p.parseTransmitStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// Парсит: stream chat_input [level: secure, tags: pii, credit]
func (p *Parser) parseStreamStatement() *ast.StreamStatement {
	stmt := &ast.StreamStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.LBRACKET) {
		return nil
	}

	for p.curToken.Type != lexer.RBRACKET && p.curToken.Type != lexer.EOF {
		p.nextToken()
		switch p.curToken.Type {
		case lexer.LEVEL:
			p.nextToken() // :
			p.nextToken() // значение
			stmt.Level = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		case lexer.TAGS:
			p.nextToken() // :
			p.nextToken()
			stmt.Tags = append(stmt.Tags, p.curToken.Literal)
			for p.peekToken.Type == lexer.COMMA {
				p.nextToken() // ,
				p.nextToken() // следующий тег
				stmt.Tags = append(stmt.Tags, p.curToken.Literal)
			}
		}
		p.nextToken()
	}
	return stmt
}

// Парсит: on chat_input as msg { ... }
func (p *Parser) parseReactorStatement() *ast.ReactorStatement {
	stmt := &ast.ReactorStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Stream = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.AS) {
		return nil
	}
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Variable = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken, Statements: []ast.Statement{}}
	p.nextToken()

	for p.curToken.Type != lexer.RBRACE && p.curToken.Type != lexer.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression()

	if p.peekToken.Type == lexer.SEMICOLON {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpression() ast.Expression {
	switch p.curToken.Type {
	// В нашем лексере mask и alert - ключевые слова, обрабатываем их как функции
	case lexer.IDENT, lexer.MASK, lexer.ALERT, lexer.TRANSMIT:
		if p.peekToken.Type == lexer.LPAREN {
			return p.parseCallExpression()
		}
		return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	default:
		return nil
	}
}

func (p *Parser) parseCallExpression() ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken}
	exp.Function = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}
	p.nextToken() // Переходим к первому аргументу

	// Упрощенно читаем один аргумент (для MVP)
	if p.curToken.Type == lexer.IDENT || p.curToken.Type == lexer.STRING {
		exp.Arguments = append(exp.Arguments, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	}
	p.errors = append(p.errors, fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type))
	return false
}

// Парсит: transmit(client) to public [level: public, allowed_roles: scoring_engine, admin]
func (p *Parser) parseTransmitStatement() *ast.TransmitStatement {
	stmt := &ast.TransmitStatement{Token: p.curToken}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Data = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}
	if !p.expectPeek(lexer.TO) {
		return nil
	}
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Channel = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.LBRACKET) {
		return nil
	}

	for p.curToken.Type != lexer.RBRACKET && p.curToken.Type != lexer.EOF {
		p.nextToken()
		switch p.curToken.Type {
		case lexer.LEVEL:
			p.nextToken() // :
			p.nextToken()
			stmt.Level = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		case lexer.ALLOWED_ROLES:
			p.nextToken() // :
			p.nextToken()
			stmt.AllowedRoles = append(stmt.AllowedRoles, p.curToken.Literal)
			for p.peekToken.Type == lexer.COMMA {
				p.nextToken() // ,
				p.nextToken()
				stmt.AllowedRoles = append(stmt.AllowedRoles, p.curToken.Literal)
			}
		}
		p.nextToken()
	}
	return stmt
}
