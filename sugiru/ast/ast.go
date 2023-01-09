package ast

import (
	"sugiru/token"
)

// Node the interface which all AST nodes will implement
type Node interface {
	TokenLiteral() string
}

// Statement nodes which evaluates to nothing
type Statement interface {
	Node
	statementNode()
}

// Expression nodes which evaluates to some value
type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

// LetStatement statement in the form: let `Name` = `Value`
type LetStatement struct {
	Token token.Token // token.LET
	Name  *Identifier
	Value Expression
}

// LetStatement implements Statement
func (ls *LetStatement) statementNode() {}
func (ls *LetStatement) TokenLiteral() string {
	return ls.Token.Literal
}

// Identifier a name that is used to identify some value, an EXPRESSION type
type Identifier struct {
	Token token.Token // token.IDENT
	Value string
}

// Identifier implements Expression
func (i *Identifier) expressionNode() {}
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}
