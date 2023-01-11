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
	infixParseFns   map[token.TokenType]infixParserFn
}

func (p *Parser) init() {
	// Read two tokens to properly set up curToken and peekToken
	p.nextToken()
	p.nextToken()

	// Registering infix and prefix
	p.prefixParserFns = make(map[token.TokenType]prefixParserFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionExpression)

	p.infixParseFns = make(map[token.TokenType]infixParserFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)

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

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
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

	// Parse the expression
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
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

	// Parse the return expression
	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken() // Move onto the semicolon
	}

	// return statement
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

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	// Acquire the appropriate prefix function
	prefix := p.prefixParserFns[p.curToken.Type]

	// If none exist, we return error
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	// Otherwise we call it
	leftExp := prefix()

	// If end of statement isn't reached and the current
	// precedence is lower than the next, we evaluate the infix
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()

		// Effectively groups in the form: a -> (a op b)
		leftExp = infix(leftExp)
	}
	return leftExp
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParserFn) {
	p.prefixParserFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParserFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	il := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)

	// Error converting
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	il.Value = value

	return il
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()

	// Move to the right hand side expression
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	// Consume the LPAREN and move onto the expression
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	// Consume the RPAREN
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	// We expect to see left paren before condition,
	// if successful, we consume the left paren
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// Move onto the condition and parse it
	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	// We expect right paren to enclose the condition
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// We expect left brace to begin block stmt
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// Parse the true branch
	expression.Then = p.parseBlockStatement()

	// curToken is now }

	// <Optional> Check if there is an else branch ( peek token )
	if p.peekTokenIs(token.ELSE) {
		// Move to the Else token
		p.nextToken()

		// We expect else body to start with {
		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		// Parse the block statement
		expression.Else = p.parseBlockStatement()
	}

	// Note: When we are on the last closing
	// brace `}` at the time of our return
	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	// While we haven't reached the end of the block, and we're not at the EOF
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()

		// Nest the statement to the block
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}

		// Since after we parse each statement,
		// we'll be at the `;` we have to consume
		// it and move onto the next token
		p.nextToken()
	}
	return block
}

func (p *Parser) parseFunctionExpression() ast.Expression {
	expression := &ast.FunctionLiteral{Token: p.curToken}

	// Move to the expected `(`
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// Start parsing parameters ( curToken is `(` )
	expression.Parameters = p.parseFunctionParameters()
	// End, curToken at `)`

	// We expect the body to begin
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// curToken is '{'

	// Parse body
	expression.Body = p.parseBlockStatement()

	return expression
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	var identifiers []*ast.Identifier

	// Note curToken is at `(`

	// In the case of void param, next token is `)`
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	// Move to the first param token
	p.nextToken()

	// Manually parse the first identifier
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	// While there is a comma, we parse the next ident
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // Comma
		p.nextToken() // Next identifier
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	// Get the enclosing right paren,
	// move onto it if it exists
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// By the time we reach here, cur token is `)`
	return identifiers
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	expression := &ast.CallExpression{
		Token:    p.curToken,
		Function: function,
	}
	expression.Arguments = p.parseCallArguments()
	return expression
}

func (p *Parser) parseCallArguments() []ast.Expression {
	var args []ast.Expression

	// For void args
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken() // Move onto it
		return args
	}

	// Move onto first arg
	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // Comma
		p.nextToken() // Arg
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}
