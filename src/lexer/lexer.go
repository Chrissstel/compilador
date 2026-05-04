package lexer

import (
	"fmt"
	"regexp"
)

type regexHandler func(lex *lexer, regex *regexp.Regexp)

type regexPattern struct {
	regex   *regexp.Regexp
	handler regexHandler
}

type lexer struct {
	patterns []regexPattern
	Tokens   []Token
	source   string
	pos      int
}

func Tokenize(source string) []Token {
	lex := createLexer(source)

	//itera mientras haya tokens
	for !lex.at_eof() {
		matched := false

		for _, pattern := range lex.patterns {
			loc := pattern.regex.FindStringIndex(lex.remainder())

			if loc != nil && loc[0] == 0 {
				pattern.handler(lex, pattern.regex)
				matched = true
				break
			}
		}

		//si no encontró ningún match
		if !matched {
			panic(fmt.Sprintf("Lexer::Error -> no se reconoció el token en %s\n", lex.remainder()))
		}
	}

	lex.push(NewToken(EOF, "EOF"))
	return lex.Tokens
}

// avanza n posiciones
func (lex *lexer) advanceN(n int) {
	lex.pos += n
}

// añade el token que vio
func (lex *lexer) push(token Token) {
	lex.Tokens = append(lex.Tokens, token)
}

// regresa donde esta
func (lex *lexer) at() byte {
	return lex.source[lex.pos]
}

// regresa lo que falta
func (lex *lexer) remainder() string {
	return lex.source[lex.pos:]
}

// dice si está en eof
func (lex *lexer) at_eof() bool {
	return lex.pos >= len(lex.source)
}

func defaultHandler(kind TokenKind, value string) regexHandler {
	return func(lex *lexer, regex *regexp.Regexp) {
		//mueve la posicion del lexer al final del valor
		lex.advanceN(len(value))
		lex.push(NewToken(kind, value))
	}
}

func createLexer(source string) *lexer {
	return &lexer{
		pos:    0,
		source: source,
		Tokens: make([]Token, 0),
		patterns: []regexPattern{

			//Delimitadores
			{regexp.MustCompile(`\(`), defaultHandler(ABRE_PAREN, "(")},
			{regexp.MustCompile(`\)`), defaultHandler(CIERRA_PAREN, ")")},
			{regexp.MustCompile(`\{`), defaultHandler(ABRE_CORCH, "{")},
			{regexp.MustCompile(`\}`), defaultHandler(CIERRA_CORCH, "}")},
			{regexp.MustCompile(`;`), defaultHandler(SEMICOLON, ";")},
			{regexp.MustCompile(`,`), defaultHandler(COMA, ",")},
			{regexp.MustCompile(`:`), defaultHandler(DOS_PUNTOS, ":")},

			//Comentarios
			{regexp.MustCompile(`\/\/.*`), skipHandler()},

			//Operadores
			{regexp.MustCompile(`=`), defaultHandler(ASIGNACION, "=")},
			{regexp.MustCompile(`\+`), defaultHandler(MAS, "+")},
			{regexp.MustCompile(`-`), defaultHandler(MENOS, "-")},
			{regexp.MustCompile(`/`), defaultHandler(DIVISION, "/")},
			{regexp.MustCompile(`\*`), defaultHandler(MULTIPLICACION, "*")},
			{regexp.MustCompile(`>`), defaultHandler(MAYOR_QUE, ">")},
			{regexp.MustCompile(`<`), defaultHandler(MENOR_QUE, "<")},
			{regexp.MustCompile(`==`), defaultHandler(IGUAL, "==")},
			{regexp.MustCompile(`!=`), defaultHandler(DIFERENTE, "!=")},

			//Palabras reservadas
			{regexp.MustCompile(`\bprograma\b`), defaultHandler(P_PROGRAMA, "programa")},
			{regexp.MustCompile(`\binicio\b`), defaultHandler(P_INICIO, "inicio")},
			{regexp.MustCompile(`\bfin\b`), defaultHandler(P_FIN, "fin")},
			{regexp.MustCompile(`\bvars\b`), defaultHandler(P_VARS, "vars")},
			{regexp.MustCompile(`\bfuncs\b`), defaultHandler(P_FUNCS, "funcs")},
			{regexp.MustCompile(`\bentero\b`), defaultHandler(P_ENTERO, "entero")},
			{regexp.MustCompile(`\bflotante\b`), defaultHandler(P_FLOTANTE, "flotante")},
			{regexp.MustCompile(`\bmientras\b`), defaultHandler(P_MIENTRAS, "mientras")},
			{regexp.MustCompile(`\bhaz\b`), defaultHandler(P_HAZ, "haz")},
			{regexp.MustCompile(`\bsi\b`), defaultHandler(P_SI, "si")},
			{regexp.MustCompile(`\bsino\b`), defaultHandler(P_SINO, "sino")},
			{regexp.MustCompile(`\bescribe\b`), defaultHandler(P_ESCRIBE, "escribe")},
			{regexp.MustCompile(`\bnula\b`), defaultHandler(P_NULA, "nula")},

			//Constantes
			{regexp.MustCompile(`\d+\.\d+`), numberHandler(FLOTANTE)},
			{regexp.MustCompile(`\d+`), numberHandler(ENTERO)},

			//Letrero
			{regexp.MustCompile(`"[^"]*"`), stringHandler()},

			//Identificadores
			{regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`), identifierHandler()},

			//Espacios
			{regexp.MustCompile(`\s+`), skipHandler()},
		},
	}
}

func numberHandler(kind TokenKind) regexHandler {
	return func(lex *lexer, regex *regexp.Regexp) {
		match := regex.FindString(lex.remainder())
		lex.advanceN(len(match))
		lex.push(NewToken(kind, match))
	}
}

// en un futuro le puedo quitar las ""
func stringHandler() regexHandler {
	return func(lex *lexer, regex *regexp.Regexp) {
		match := regex.FindString(lex.remainder())
		lex.advanceN(len(match))
		lex.push(NewToken(LETRERO, match))
	}
}

func identifierHandler() regexHandler {
	return func(lex *lexer, regex *regexp.Regexp) {
		match := regex.FindString(lex.remainder())
		lex.advanceN(len(match))
		lex.push(NewToken(IDENTIFICADOR, match))
	}
}

func skipHandler() regexHandler {
	return func(lex *lexer, regex *regexp.Regexp) {
		match := regex.FindString(lex.remainder())
		lex.advanceN(len(match))
	}
}
