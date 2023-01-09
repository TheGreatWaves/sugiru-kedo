package lexer

import (
	"sugiru/token"
)

type Lexer struct {
	input        string
	position     int  // Current position in input (points to current character)
	readPosition int  // Current reading position in input (after current char)
	ch           byte // Current char under examination
}

// New creates a new lexer struct.
func New(input string) *Lexer {
	// Creates a new lexer
	l := &Lexer{input: input}

	// Initialize positional values etc.
	// (ch -> first character )
	// ( position -> 0 )
	// ( readPosition -> 1 )
	l.readChar()

	return l
}

func (l *Lexer) readChar() {
	// Check for EOF
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	// Advance to next character
	l.position = l.readPosition
	l.readPosition += 1
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{
		Type:    tokenType,
		Literal: string(ch),
	}
}

// NextToken returns the next token
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	switch l.ch {
	case '=':
		tok = newToken(token.ASSIGN, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	}

	// Advance to next character
	l.readChar()
	return tok
}