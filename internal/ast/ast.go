package ast

import "vigil/internal/lexer"

// Node — базовый интерфейс для всех узлов дерева
type Node interface {
	TokenLiteral() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

// Program — корневой узел, содержит все инструкции
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// StreamStatement: stream chat_input [level: secure, tags: pii]
type StreamStatement struct {
	Token lexer.Token
	Name  *Identifier
	Level *Identifier
	Tags  []string // <-- НОВОЕ
}

func (s *StreamStatement) statementNode()       {}
func (s *StreamStatement) TokenLiteral() string { return s.Token.Literal }

// ReactorStatement: on chat_input as msg { ... }
type ReactorStatement struct {
	Token    lexer.Token
	Stream   *Identifier
	Variable *Identifier
	Body     *BlockStatement
}

func (r *ReactorStatement) statementNode()       {}
func (r *ReactorStatement) TokenLiteral() string { return r.Token.Literal }

// BlockStatement: { ... }
type BlockStatement struct {
	Token      lexer.Token
	Statements []Statement
}

func (b *BlockStatement) statementNode()       {}
func (b *BlockStatement) TokenLiteral() string { return b.Token.Literal }

// ExpressionStatement: alert("...") или mask(msg)
type ExpressionStatement struct {
	Token      lexer.Token
	Expression Expression
}

func (e *ExpressionStatement) statementNode()       {}
func (e *ExpressionStatement) TokenLiteral() string { return e.Token.Literal }

// CallExpression: mask(msg)
type CallExpression struct {
	Token     lexer.Token
	Function  Expression
	Arguments []Expression
}

func (c *CallExpression) expressionNode()      {}
func (c *CallExpression) TokenLiteral() string { return c.Token.Literal }

// Identifier: msg, chat_input, secure
type Identifier struct {
	Token lexer.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// TransmitStatement: transmit(client) to public [level: public, allowed_roles: client]
type TransmitStatement struct {
	Token        lexer.Token
	Data         Expression
	Channel      *Identifier
	Level        *Identifier
	AllowedRoles []string // <-- НОВОЕ
}

func (t *TransmitStatement) statementNode()       {}
func (t *TransmitStatement) TokenLiteral() string { return t.Token.Literal }
