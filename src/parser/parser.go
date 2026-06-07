package parser

import (
	"compilador/src/ast"
	"compilador/src/lexer"
	"fmt"
	"strconv"
)

type Parser struct {
	tokens []lexer.Token
	pos    int
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

// Funciones extras
// devuelve el token actual sin consumirlo
func (p *Parser) current() lexer.Token {
	return p.tokens[p.pos]
}

// ve el siguiente token pero no lo consume
func (p *Parser) peek() lexer.Token {
	if p.pos+1 < len(p.tokens) {
		return p.tokens[p.pos+1]
	}
	return p.tokens[p.pos]
}

// avanza al siguiente token y lo devuelve
func (p *Parser) advance() lexer.Token {
	t := p.tokens[p.pos]
	if p.pos < len(p.tokens)-1 {
		p.pos++
	}
	return t
}

// verifica que el token actual sea del tipo esperado, si es así lo consume y devuelve
func (p *Parser) check(kind lexer.TokenKind) bool {
	return p.current().Kind == kind
}

// verifica que el token actual sea del tipo esperado, si es así lo consume y devuelve
func (p *Parser) expect(kind lexer.TokenKind) lexer.Token {
	t := p.current()
	if t.Kind != kind {
		panic(fmt.Sprintf(
			"Parser::Error -> esperaba '%s' pero encontré '%s' (valor: '%s')",
			lexer.TokenKindString(kind),
			lexer.TokenKindString(t.Kind),
			t.Value,
		))
	}
	return p.advance()
}

// <PROGRAMA> → programa id ; <V> <F> inicio <CUERPO> fin
func (p *Parser) ParsePrograma() *ast.Programa {
	p.expect(lexer.P_PROGRAMA)
	id := p.expect(lexer.IDENTIFICADOR).Value
	p.expect(lexer.SEMICOLON)

	vars := p.parseV()
	funcs := p.parseF()

	p.expect(lexer.P_INICIO)
	cuerpo := p.parseCuerpo()
	p.expect(lexer.P_FIN)

	return &ast.Programa{
		ID:        id,
		DeclsVars: vars,
		Funcs:     funcs,
		Cuerpo:    cuerpo,
	}
}

// <V> → <VARS> | ε
func (p *Parser) parseV() *ast.DeclsVars {
	if p.check(lexer.P_VARS) {
		return p.parseVars()
	}
	return nil
}

// <VARS> → vars (<LOOP_ID> : <TIPO> ;)+
func (p *Parser) parseVars() *ast.DeclsVars {
	p.expect(lexer.P_VARS)
	var decls []*ast.Vars //vars es varios IDs y un tipo

	for p.check(lexer.IDENTIFICADOR) {
		ids := p.parseLoopID()
		p.expect(lexer.DOS_PUNTOS)
		tipo := p.parseTipo()
		p.expect(lexer.SEMICOLON)
		decls = append(decls, &ast.Vars{IDs: ids, Tipo: tipo})
	}
	return &ast.DeclsVars{Decls: decls}
}

// id (, id)*
func (p *Parser) parseLoopID() []string {
	var ids []string
	ids = append(ids, p.expect(lexer.IDENTIFICADOR).Value)

	for p.check(lexer.COMA) {
		p.advance()
		ids = append(ids, p.expect(lexer.IDENTIFICADOR).Value)
	}
	return ids
}

// <TIPO> → entero | flotante
func (p *Parser) parseTipo() string {
	t := p.current()
	if t.Kind == lexer.P_ENTERO || t.Kind == lexer.P_FLOTANTE {
		p.advance()
		return t.Value
	}
	panic(fmt.Sprintf(
		"Parser::Error -> esperaba tipo (entero|flotante) pero encontré '%s'", t.Value,
	))
}

//PARA FUNCIONES

// <F> → <FUNCS>*
func (p *Parser) parseF() []*ast.Func {
	var funcs []*ast.Func
	for p.esInicioFunc() {
		funcs = append(funcs, p.parseFuncs())
	}
	return funcs
}

func (p *Parser) esInicioFunc() bool {
	k := p.current().Kind
	return k == lexer.P_NULO || k == lexer.P_ENTERO || k == lexer.P_FLOTANTE
}

// <FUNCS> → <DEF_FUNC> id ( (<PARAM> (, <PARAM>)*)? ) { <VARS>? <CUERPO> <RETORNO>? } ;
func (p *Parser) parseFuncs() *ast.Func {
	tipoRet := p.parseDefFunc()
	id := p.expect(lexer.IDENTIFICADOR).Value
	p.expect(lexer.ABRE_PAREN)
	params := p.parseParams()
	p.expect(lexer.CIERRA_PAREN)
	p.expect(lexer.ABRE_LLAVE)
	vars := p.parseV()
	cuerpo := p.parseCuerpo()
	retorno := p.parseRetorno()
	p.expect(lexer.CIERRA_LLAVE)
	p.expect(lexer.SEMICOLON)

	return &ast.Func{
		TipoRetorno: tipoRet,
		ID:          id,
		Params:      params,
		Vars:        vars,
		Cuerpo:      cuerpo,
		Retorno:     retorno,
	}
}

// <RETORNO> → retornar id ; | ε
func (p *Parser) parseRetorno() *ast.Retorno {
	if p.check(lexer.P_RETORNAR) {
		p.advance()
		id := p.expect(lexer.IDENTIFICADOR).Value
		p.expect(lexer.SEMICOLON)
		return &ast.Retorno{ID: id}
	}
	return nil
}

// <DEF_FUNC> → nulo | entero | flotante
func (p *Parser) parseDefFunc() string {
	if p.check(lexer.P_NULO) {
		p.advance()
		return "nulo"
	}
	return p.parseTipo()
}

// <ID_FUNC> → <LOOP_FUNC> | ε
func (p *Parser) parseParams() []*ast.Param {
	var params []*ast.Param
	if !p.check(lexer.IDENTIFICADOR) {
		return params
	}
	params = append(params, p.parseParam())

	for p.check(lexer.COMA) {
		p.advance()
		params = append(params, p.parseParam())
	}
	return params
}

// <PARAM> → id : <TIPO>
func (p *Parser) parseParam() *ast.Param {
	id := p.expect(lexer.IDENTIFICADOR).Value
	p.expect(lexer.DOS_PUNTOS)
	tipo := p.parseTipo()
	return &ast.Param{ID: id, Tipo: tipo}
}

//PARA EL CUERPO

// <CUERPO> → { <ESTATUTO>* }
func (p *Parser) parseCuerpo() *ast.Cuerpo {
	p.expect(lexer.ABRE_LLAVE)

	var estatutos []ast.Estatuto
	for p.esInicioEstatuto() {
		estatutos = append(estatutos, p.parseEstatuto())
	}
	p.expect(lexer.CIERRA_LLAVE)
	return &ast.Cuerpo{Estatutos: estatutos}
}

// FIRST(<ESTATUTO>) = {id, si, mientras, escribe, [}
func (p *Parser) esInicioEstatuto() bool {
	switch p.current().Kind {
	case lexer.IDENTIFICADOR, lexer.P_SI, lexer.P_MIENTRAS,
		lexer.P_ESCRIBE, lexer.ABRE_CORCHETE:
		return true
	}
	return false
}

//PARA LOS ESTATITOS

// <ESTATUTO> → <ASIGNA> | <CONDICION> | <CICLO> | <LLAMADA> ; | <IMPRIME> | [ <LOOP_ESTATUTO> ]
func (p *Parser) parseEstatuto() ast.Estatuto {
	switch p.current().Kind {
	case lexer.IDENTIFICADOR:
		//ve el siguiente porque puede ser un id de asignación o uno de una llamada
		if p.peek().Kind == lexer.ASIGNACION {
			return p.parseAsigna()
		}
		llamada := p.parseLlamada()
		p.expect(lexer.SEMICOLON)
		return llamada

	case lexer.P_SI:
		return p.parseCondicion()

	case lexer.P_MIENTRAS:
		return p.parseCiclo()

	case lexer.P_ESCRIBE:
		return p.parseImprime()

	case lexer.ABRE_CORCHETE:
		p.advance()
		var estatutos []ast.Estatuto
		for p.esInicioEstatuto() {
			estatutos = append(estatutos, p.parseEstatuto())
		}
		p.expect(lexer.CIERRA_CORCHETE)
		return &ast.BloqueEstatutos{Estatutos: estatutos}
	}

	panic(fmt.Sprintf(
		"Parser::Error -> estatuto inesperado '%s'", p.current().Value,
	))
}

// <ASIGNA> → id = <EXPRESION> ;
func (p *Parser) parseAsigna() *ast.Asigna {
	id := p.expect(lexer.IDENTIFICADOR).Value
	p.expect(lexer.ASIGNACION)
	expr := p.parseExpresion()
	p.expect(lexer.SEMICOLON)
	return &ast.Asigna{ID: id, Expresion: expr}
}

// <CONDICION> → si ( <EXPRESION> ) <CUERPO> <ELSE> ;
func (p *Parser) parseCondicion() *ast.Condicion {
	p.expect(lexer.P_SI)
	p.expect(lexer.ABRE_PAREN)
	expr := p.parseExpresion()
	p.expect(lexer.CIERRA_PAREN)
	cuerpoSi := p.parseCuerpo()

	var cuerpoSino *ast.Cuerpo
	if p.check(lexer.P_SINO) {
		p.advance()
		cuerpoSino = p.parseCuerpo()
	}

	p.expect(lexer.SEMICOLON)
	return &ast.Condicion{
		Expresion:  expr,
		CuerpoSi:   cuerpoSi,
		CuerpoSino: cuerpoSino,
	}
}

// <CICLO> → mientras ( <EXPRESION> ) haz <CUERPO> ;
func (p *Parser) parseCiclo() *ast.Ciclo {
	p.expect(lexer.P_MIENTRAS)
	p.expect(lexer.ABRE_PAREN)
	expr := p.parseExpresion()
	p.expect(lexer.CIERRA_PAREN)
	p.expect(lexer.P_HAZ)
	cuerpo := p.parseCuerpo()
	p.expect(lexer.SEMICOLON)
	return &ast.Ciclo{Expresion: expr, Cuerpo: cuerpo}
}

// <LLAMADA> → id ( (<EXPRESION> (, <EXPRESION>)*)? )
func (p *Parser) parseLlamada() *ast.Llamada {
	id := p.expect(lexer.IDENTIFICADOR).Value
	p.expect(lexer.ABRE_PAREN)
	var args []*ast.Expresion
	if p.esInicioExpresion() {
		args = append(args, p.parseExpresion())
		for p.check(lexer.COMA) {
			p.advance()
			args = append(args, p.parseExpresion())
		}
	}
	p.expect(lexer.CIERRA_PAREN)
	return &ast.Llamada{ID: id, Args: args}
}

// <IMPRIME> → escribe ( <ITEM> (, <ITEM>)* ) ;
func (p *Parser) parseImprime() *ast.Imprime {
	p.expect(lexer.P_ESCRIBE)
	p.expect(lexer.ABRE_PAREN)

	var items []ast.ImprimeItem
	items = append(items, p.parseItem())
	for p.check(lexer.COMA) {
		p.advance()
		items = append(items, p.parseItem())
	}

	p.expect(lexer.CIERRA_PAREN)
	p.expect(lexer.SEMICOLON)
	return &ast.Imprime{Items: items}
}

// <ITEM> → <EXPRESION> | letrero
func (p *Parser) parseItem() ast.ImprimeItem {
	if p.check(lexer.LETRERO) {
		return ast.ImprimeItem{EsLetrero: true, Letrero: p.advance().Value}
	}
	return ast.ImprimeItem{EsLetrero: false, Expr: p.parseExpresion()}
}

//PARA EXPRESIONES

// <EXPRESION> → <EXP> (<COND> <EXP>)?
func (p *Parser) parseExpresion() *ast.Expresion {
	izq := p.parseExp()
	var op string
	var der *ast.Exp
	switch p.current().Kind {
	case lexer.MAYOR_QUE, lexer.MENOR_QUE, lexer.DIFERENTE, lexer.IGUAL, lexer.MAYOR_IGUAL, lexer.MENOR_IGUAL:
		op = p.advance().Value
		der = p.parseExp()
	}
	return &ast.Expresion{Izq: izq, Op: op, Der: der}
}

// <EXP> → <TERMINO> ((+ | -) <TERMINO>)*
func (p *Parser) parseExp() *ast.Exp {
	raiz := &ast.Exp{Termino: p.parseTermino()}
	actual := raiz
	for p.check(lexer.MAS) || p.check(lexer.MENOS) {
		op := p.advance().Value
		der := &ast.Exp{Termino: p.parseTermino()}
		actual.Op = op
		actual.Der = der
		actual = der
	}
	return raiz
}

// <TERMINO> → <FACTOR> ((* | /) <FACTOR>)*
func (p *Parser) parseTermino() *ast.Termino {
	raiz := &ast.Termino{Factor: p.parseFactor()}
	actual := raiz
	for p.check(lexer.MULTIPLICACION) || p.check(lexer.DIVISION) {
		op := p.advance().Value
		der := &ast.Termino{Factor: p.parseFactor()}
		actual.Op = op
		actual.Der = der
		actual = der
	}
	return raiz
}

// <FACTOR> → ( <EXPRESION> ) | <LLAMADA> | <VALOR>
func (p *Parser) parseFactor() *ast.Factor {
	// ( <EXPRESION> )
	if p.check(lexer.ABRE_PAREN) {
		p.advance()
		expr := p.parseExpresion()
		p.expect(lexer.CIERRA_PAREN)
		return &ast.Factor{Expr: expr}
	}

	// <LLAMADA>: id seguido de '('
	if p.check(lexer.IDENTIFICADOR) && p.peek().Kind == lexer.ABRE_PAREN {
		return &ast.Factor{Llamada: p.parseLlamada()}
	}

	// <VALOR>: puede ser id o constante
	return &ast.Factor{Valor: p.parseValor()}
}

// <VALOR> → (+ | -)? (id | cte_ent | cte_flot)
func (p *Parser) parseValor() *ast.Valor {
	signo := ""
	if p.check(lexer.MAS) || p.check(lexer.MENOS) {
		signo = p.advance().Value
	}

	t := p.current()
	switch t.Kind {
	case lexer.IDENTIFICADOR:
		p.advance()
		return &ast.Valor{ID: t.Value, Signo: signo, EsCte: false}

	case lexer.CTE_ENTERO:
		p.advance()
		v, _ := strconv.Atoi(t.Value)
		return &ast.Valor{EsCte: true, CteEnt: &v, Signo: signo}
	case lexer.CTE_FLOTANTE:
		p.advance()
		v, _ := strconv.ParseFloat(t.Value, 64)
		return &ast.Valor{EsCte: true, CteFlot: &v, Signo: signo}
	}

	panic(fmt.Sprintf(
		"Parser::Error -> se esperaba id o constante pero se encontró '%s'", t.Value,
	))
}

// FIRST(<EXPRESION>) = { (, id, +, -, cte_ent, cte_flot }
func (p *Parser) esInicioExpresion() bool {
	switch p.current().Kind {
	case lexer.ABRE_PAREN, lexer.IDENTIFICADOR,
		lexer.MAS, lexer.MENOS,
		lexer.CTE_ENTERO, lexer.CTE_FLOTANTE:
		return true
	}
	return false
}
