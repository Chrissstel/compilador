package parser

import (
	"compilador/src/ast"
	"compilador/src/lexer"
)

// no se expone la estructura
type parser struct {
	//Aqui voy a manejar los errores
	// errors []error
	tokens []lexer.Token
	pos    int
}

func createParser(tokens []lexer.Token) *parser {
	createTokenLookups()
	return &parser{
		tokens: tokens,
		pos:    0,
	}
}
func Parse(tokens []lexer.Token) ast.CuerpoStmt {
	Body := make([]ast.Stmt, 0)
	p := createParser(tokens) //nuestro parser

	for p.hasTokens() {
		Body = append(Body, parse_stmt(p))
	}

	return ast.CuerpoStmt{
		Body: Body,
	}
}

//métodos para hacer más bonito el código

func (p *parser) currentToken() lexer.Token {
	return p.tokens[p.pos]
}

func (p *parser) currentTokenKind() lexer.TokenKind {
	return p.currentToken().Kind
}

func (p *parser) advance() lexer.Token {
	tk := p.currentToken()
	p.pos++
	return tk
}

func (p *parser) hasTokens() bool {
	return p.pos < len(p.tokens) && p.currentTokenKind() != lexer.EOF
}
