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

	l.skipWhiteSpace()

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
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()          // Read the identifier
			tok.Type = token.LookupIdent(tok.Literal) // Look up the identifier to get the appropriate token
			return tok
		} else if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = token.INT
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	// Advance to next character
	l.readChar()
	return tok
}

// isLetter returns whether a given byte represents a valid character
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// isDigit returns whether a given byte is a digit
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// Return the identifier string
func (l *Lexer) readIdentifier() string {

	// Mark the beginning position of the identifier
	position := l.position

	// White the current letter is a valid letter, we read and consume
	for isLetter(l.ch) {
		l.readChar()
	}

	// Return the slice from beginning to current position
	return l.fromPosToCurrent(position)
}

// Return the number as a string ( Note: fractions not currently supported )
func (l *Lexer) readNumber() string {

	// Mark the beginning of the lexeme
	position := l.position

	// Advance the character until we reach a non-digit
	for isDigit(l.ch) {
		l.readChar()
	}

	// Return the slice
	return l.fromPosToCurrent(position)
}

// skipWhiteSpace consumes characters as long as it is a white space character
func (l *Lexer) skipWhiteSpace() {
	for {
		switch l.ch {
		case ' ', '\t', '\n', '\r':
			l.readChar()
		default:
			return
		}
	}
}

// extractFromPosToCurrent retrieves a slice of the input, starting from position to current index
func (l *Lexer) fromPosToCurrent(position int) string {
	return l.input[position:l.position]
}
