package parser

import (
	"sugiru/ast"
	"sugiru/lexer"
	"sugiru/token"
)

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token // Pointer to the current token
	peekToken token.Token // Pointer to the next token
}

func (p *Parser) init() {
	// Read two tokens to properly set up curToken and peekToken
	p.nextToken()
	p.nextToken()
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken      // Retrieves the next current from the peek
	p.peekToken = p.l.NextToken() // Retrieves the next token from the lexer
}

func (p *Parser) ParseProgram() *ast.Program {
	// Construct the root node of the AST
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	// Keep iterating until we reach an EOF token
	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		// Advance to the next token
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	// Parse according to the current token
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	default:
		return nil
	}
}

// parseLetStatement parses the let statement, the expected form
// being: 'let' 'IDENT' '=' 'VALUE' ';'
func (p *Parser) parseLetStatement() ast.Statement {
	// Note: The current IS ALWAYS token.LET

	// Constructs a new AST node (*ast.LetStatement node)
	stmt := &ast.LetStatement{Token: p.curToken}

	// We expect to see an identifier after the 'let' keyword
	// example: let x
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	// Constructs the ast.Identifier node and attach it onto the statement
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// We expect an assignment operator next after the identifier
	// example: let x =
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	// Note: Currently the value is being skipped over, we're just jumping
	// straight to the semicolon which indicates the end of the statement
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// peekTokenIs returns whether the peek token is of specified type
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// curTokenIs returns whether the current token is of specified type
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// expectPeek returns true if the next token is as expected,
// the current token is also advanced as a side effect (if true)
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		return false
	}
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}

	// Set up current and peek token
	p.init()

	return p
}
