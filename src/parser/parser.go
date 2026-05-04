package parser

import (
	"compilador/src/ast"
	"compilador/src/lexer"
	"fmt"
	"strconv"
)

type Parser struct {
	//Aqui voy a manejar los errores
	// errors []error
	tokens []lexer.Token
	pos    int
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

// Funciones extras
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

func (p *Parser) advance() lexer.Token {
	t := p.tokens[p.pos]
	if p.pos < len(p.tokens)-1 {
		p.pos++
	}
	return t
}

func (p *Parser) check(kind lexer.TokenKind) bool {
	return p.current().Kind == kind
}

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
		ID:     id,
		Vars:   vars,
		Funcs:  funcs,
		Cuerpo: cuerpo,
	}
}

// <V> → <VARS> | ε
func (p *Parser) parseV() *ast.Vars {
	if p.check(lexer.P_VARS) {
		return p.parseV()
	}
	return nil
}

// <VARS> → vars <LOOP_VARS>
func (p *Parser) parseVars() *ast.Vars {
	p.expect(lexer.P_VARS)
	decls := p.parseLoopVars()
	return &ast.Vars{Declaraciones: decls}
}

// <LOOP_VARS> → <LOOP_ID> : <TIPO> ; <L>
func (p *Parser) parseLoopVars() []*ast.DeclaracionVar {
	ids := p.parseLoopID()
	p.expect(lexer.DOS_PUNTOS)
	tipo := p.parseTipo()
	p.expect(lexer.SEMICOLON)

	decl := &ast.DeclaracionVar{IDs: ids, Tipo: tipo}
	resto := p.parseL()
	return append([]*ast.DeclaracionVar{decl}, resto...)
}

// <L> → <LOOP_VARS> | ε
// FIRST(<LOOP_VARS>) = {id}
func (p *Parser) parseL() []*ast.DeclaracionVar {
	if p.check(lexer.IDENTIFICADOR) {
		return p.parseLoopVars()
	}
	return nil
}

// <LOOP_ID> → id <I>
func (p *Parser) parseLoopID() []string {
	id := p.expect(lexer.IDENTIFICADOR).Value
	resto := p.parseI()
	return append([]string{id}, resto...)
}

// <I> → , <LOOP_ID> | ε
func (p *Parser) parseI() []string {
	if p.check(lexer.COMA) {
		p.advance()
		return p.parseLoopID()
	}
	return nil
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

// <F> → <FUNCS> <F> | ε
func (p *Parser) parseF() []*ast.Func {
	if p.esInicioFunc() {
		f := p.parseFuncs()
		resto := p.parseF()
		return append([]*ast.Func{f}, resto...)
	}
	return nil
}

func (p *Parser) esInicioFunc() bool {
	k := p.current().Kind
	return k == lexer.P_NULA || k == lexer.P_ENTERO || k == lexer.P_FLOTANTE
}

// <FUNCS> → <DEF_FUNC> id ( <ID_FUNC> ) { <VARS_FUNC> <CUERPO> } ;
func (p *Parser) parseFuncs() *ast.Func {
	tipoRet := p.parseDefFunc()
	id := p.expect(lexer.IDENTIFICADOR).Value
	p.expect(lexer.ABRE_PAREN)
	params := p.parseIDFunc()
	p.expect(lexer.CIERRA_PAREN)
	p.expect(lexer.ABRE_LLAVE)
	vars := p.parseVarsFunc()
	cuerpo := p.parseCuerpo()
	p.expect(lexer.CIERRA_LLAVE)
	p.expect(lexer.SEMICOLON)

	return &ast.Func{
		TipoRetorno: tipoRet,
		ID:          id,
		Params:      params,
		Vars:        vars,
		Cuerpo:      cuerpo,
	}
}

// <DEF_FUNC> → nula | <TIPO>
func (p *Parser) parseDefFunc() string {
	if p.check(lexer.P_NULA) {
		p.advance()
		return "nula"
	}
	return p.parseTipo()
}

// <ID_FUNC> → <LOOP_FUNC> | ε
func (p *Parser) parseIDFunc() []*ast.Param {
	if p.check(lexer.IDENTIFICADOR) {
		return p.parseLoopFunc()
	}
	return nil
}

// <LOOP_FUNC> → id : <TIPO> <LF2>
func (p *Parser) parseLoopFunc() []*ast.Param {
	id := p.expect(lexer.IDENTIFICADOR).Value
	p.expect(lexer.DOS_PUNTOS)
	tipo := p.parseTipo()
	param := &ast.Param{ID: id, Tipo: tipo}
	resto := p.parseLF2()
	return append([]*ast.Param{param}, resto...)
}

// <LF2> → , <LOOP_FUNC> | ε
func (p *Parser) parseLF2() []*ast.Param {
	if p.check(lexer.COMA) {
		p.advance()
		return p.parseLoopFunc()
	}
	return nil
}

// <VARS_FUNC> → <VARS> | ε
func (p *Parser) parseVarsFunc() *ast.Vars {
	if p.check(lexer.P_VARS) {
		return p.parseVars()
	}
	return nil
}

//PARA EL CUERPO

// <CUERPO> → { <LOOP_CUERPO> }
func (p *Parser) parseCuerpo() *ast.Cuerpo {
	p.expect(lexer.ABRE_LLAVE)
	estatutos := p.parseLoopCuerpo()
	p.expect(lexer.CIERRA_LLAVE)
	return &ast.Cuerpo{Estatutos: estatutos}
}

// <LOOP_CUERPO> → <ESTATUTO> <LOOP_CUERPO> | ε
func (p *Parser) parseLoopCuerpo() []ast.Estatuto {
	if p.esInicioEstatuto() {
		e := p.parseEstatuto()
		resto := p.parseLoopCuerpo()
		return append([]ast.Estatuto{e}, resto...)
	}
	return nil
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
		estatutos := p.parseLoopEstatuto()
		p.expect(lexer.CIERRA_CORCHETE)
		return &ast.BloqueEstatutos{Estatutos: estatutos}
	}

	panic(fmt.Sprintf(
		"Parser::Error -> estatuto inesperado '%s'", p.current().Value,
	))
}

// <LOOP_ESTATUTO> → <ESTATUTO> <LOOP_ESTATUTO> | ε
func (p *Parser) parseLoopEstatuto() []ast.Estatuto {
	if p.esInicioEstatuto() {
		e := p.parseEstatuto()
		resto := p.parseLoopEstatuto()
		return append([]ast.Estatuto{e}, resto...)
	}
	return nil
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
	cuerpoSino := p.parseElse()
	p.expect(lexer.SEMICOLON)
	return &ast.Condicion{
		Expresion:  expr,
		CuerpoSi:   cuerpoSi,
		CuerpoSino: cuerpoSino,
	}
}

// <ELSE> → sino <CUERPO> | ε
func (p *Parser) parseElse() *ast.Cuerpo {
	if p.check(lexer.P_SINO) {
		p.advance()
		return p.parseCuerpo()
	}
	return nil
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

// <LLAMADA> → id ( <AUX_LLAMADA> )
func (p *Parser) parseLlamada() *ast.Llamada {
	id := p.expect(lexer.IDENTIFICADOR).Value
	p.expect(lexer.ABRE_PAREN)
	args := p.parseAuxLlamada()
	p.expect(lexer.CIERRA_PAREN)
	return &ast.Llamada{ID: id, Args: args}
}

// <AUX_LLAMADA> → <EXPRESION> <LOOP_EXPRESION> | ε
func (p *Parser) parseAuxLlamada() []*ast.Expresion {
	if p.esInicioExpresion() {
		e := p.parseExpresion()
		resto := p.parseLoopExpresionArgs()
		return append([]*ast.Expresion{e}, resto...)
	}
	return nil
}

// <LOOP_EXPRESION> → , <EXPRESION> <LOOP_EXPRESION> | ε
func (p *Parser) parseLoopExpresionArgs() []*ast.Expresion {
	if p.check(lexer.COMA) {
		p.advance()
		e := p.parseExpresion()
		resto := p.parseLoopExpresionArgs()
		return append([]*ast.Expresion{e}, resto...)
	}
	return nil
}

// <IMPRIME> → escribe ( <AUX_IMPRIME> ) ;
func (p *Parser) parseImprime() *ast.Imprime {
	p.expect(lexer.P_ESCRIBE)
	p.expect(lexer.ABRE_PAREN)
	items := p.parseAuxImprime()
	p.expect(lexer.CIERRA_PAREN)
	p.expect(lexer.SEMICOLON)
	return &ast.Imprime{Items: items}
}

// <AUX_IMPRIME> → <EXPRESION> <LOOP_IMPRIME> | letrero <LOOP_IMPRIME>
func (p *Parser) parseAuxImprime() []ast.ImprimeItem {
	if p.check(lexer.LETRERO) {
		val := p.advance().Value
		item := ast.ImprimeItem{EsLetrero: true, Letrero: val}
		resto := p.parseLoopImprime()
		return append([]ast.ImprimeItem{item}, resto...)
	}
	expr := p.parseExpresion()
	item := ast.ImprimeItem{EsLetrero: false, Expr: expr}
	resto := p.parseLoopImprime()
	return append([]ast.ImprimeItem{item}, resto...)
}

// <LOOP_IMPRIME> → , <AUX_IMPRIME> | ε
func (p *Parser) parseLoopImprime() []ast.ImprimeItem {
	if p.check(lexer.COMA) {
		p.advance()
		return p.parseAuxImprime()
	}
	return nil
}

//PARA EXPRESIONES

// <EXPRESION> → <EXP> <ELSE_EXPRESION>
func (p *Parser) parseExpresion() *ast.Expresion {
	izq := p.parseExp()
	op, der := p.parseElseExpresion()
	return &ast.Expresion{Izq: izq, Op: op, Der: der}
}

// <ELSE_EXPRESION> → <COND> <EXP> | ε
// <COND> → > | < | != | ==
func (p *Parser) parseElseExpresion() (string, *ast.Exp) {
	switch p.current().Kind {
	case lexer.MAYOR_QUE, lexer.MENOR_QUE, lexer.DIFERENTE, lexer.IGUAL:
		op := p.advance().Value
		der := p.parseExp()
		return op, der
	}
	return "", nil
}

// <EXP> → <TERMINO> <LOOP_EXP>
func (p *Parser) parseExp() *ast.Exp {
	term := p.parseTermino()
	op, der := p.parseLoopExp()
	return &ast.Exp{Termino: term, Op: op, Der: der}
}

// <LOOP_EXP> → + <EXP> | - <EXP> | ε
func (p *Parser) parseLoopExp() (string, *ast.Exp) {
	if p.check(lexer.MAS) || p.check(lexer.MENOS) {
		op := p.advance().Value
		der := p.parseExp()
		return op, der
	}
	return "", nil
}

// <TERMINO> → <FACTOR> <LOOP_TERMINO>
func (p *Parser) parseTermino() *ast.Termino {
	factor := p.parseFactor()
	op, der := p.parseLoopTermino()
	return &ast.Termino{Factor: factor, Op: op, Der: der}
}

// <LOOP_TERMINO> → * <TERMINO> | / <TERMINO> | ε
func (p *Parser) parseLoopTermino() (string, *ast.Termino) {
	if p.check(lexer.MULTIPLICACION) || p.check(lexer.DIVISION) {
		op := p.advance().Value
		der := p.parseTermino()
		return op, der
	}
	return "", nil
}

// <FACTOR> → ( <EXPRESION> ) | <LLAMADA> | <SIGNO> <VALOR>
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

	// <SIGNO> <VALOR>
	signo := p.parseSigno()
	valor := p.parseValor()
	return &ast.Factor{Signo: signo, Valor: valor}
}

// <SIGNO> → + | - | ε
func (p *Parser) parseSigno() string {
	if p.check(lexer.MAS) || p.check(lexer.MENOS) {
		return p.advance().Value
	}
	return ""
}

// <VALOR> → id | <CTE>
func (p *Parser) parseValor() *ast.Valor {
	if p.check(lexer.IDENTIFICADOR) {
		return &ast.Valor{ID: p.advance().Value}
	}
	return p.parseCte()
}

// <CTE> → cte_ent | cte_flot
func (p *Parser) parseCte() *ast.Valor {
	t := p.current()
	switch t.Kind {
	case lexer.CTE_ENTERO:
		p.advance()
		v, _ := strconv.Atoi(t.Value)
		return &ast.Valor{EsCte: true, CteEnt: &v}
	case lexer.CTE_FLOTANTE:
		p.advance()
		v, _ := strconv.ParseFloat(t.Value, 64)
		return &ast.Valor{EsCte: true, CteFlot: &v}
	}
	panic(fmt.Sprintf(
		"Parser::Error -> se esperaba un valor (id, entero o flotante) pero encontré '%s'", t.Value,
	))
}

// FIRST(<EXPRESION>) = {(, id, +, -, cte_ent, cte_flot}
func (p *Parser) esInicioExpresion() bool {
	switch p.current().Kind {
	case lexer.ABRE_PAREN, lexer.IDENTIFICADOR,
		lexer.MAS, lexer.MENOS,
		lexer.CTE_ENTERO, lexer.CTE_FLOTANTE:
		return true
	}
	return false
}
