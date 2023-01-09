package parser

import (
	"fmt"
	"strconv"
	"sugiru/ast"
	"sugiru/lexer"
	"sugiru/token"
)

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token // Pointer to the current token
	peekToken token.Token // Pointer to the next token

	errors []string

	prefixParserFns map[token.TokenType]prefixParserFn
	infixParserFns  map[token.TokenType]infixParserFn
}

func (p *Parser) init() {
	// Read two tokens to properly set up curToken and peekToken
	p.nextToken()
	p.nextToken()

	// Registering infix and prefix
	p.prefixParserFns = make(map[token.TokenType]prefixParserFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// Set up current and peek token
	p.init()

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
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
	for !p.curTokenIs(token.EOF) {
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
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseLetStatement parses the let statement, the expected form
// being: 'let' 'IDENT' '=' 'VALUE' ';'
func (p *Parser) parseLetStatement() *ast.LetStatement {
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
		p.peekError(t)
		return false
	}
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	// Advance the token
	p.nextToken()

	// NOTE: Skipping expressions for now
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	// Construct ExpressionStatement node
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	// If next token is ; consume it ( optional )
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

///////////////////////////////
/// PRAT PARSER SECTION
///////////////////////////////

type (
	prefixParserFn func() ast.Expression
	infixParserFn  func(ast.Expression) ast.Expression
)

// Note: it is important that these values are in this order
// because we will need them in order to evaluate precedence
const (
	_ int = iota // Used to give the following constants incrementing numbers as values ( _ takes 0 )
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // fn()
)

func (p *Parser) parseExpression(precendence int) ast.Expression {
	// Acquire the appropriate prefix function
	prefix := p.prefixParserFns[p.curToken.Type]

	// If none exist, we return
	if prefix == nil {
		return nil
	}

	// Otherwise we call it
	leftExp := prefix()

	return leftExp
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParserFn) {
	p.prefixParserFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParserFn) {
	p.infixParserFns[tokenType] = fn
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)

	// Error converting
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}
