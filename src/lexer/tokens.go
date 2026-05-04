package lexer

import (
	"fmt"
)

type TokenKind int

// todos los tipos de tokens
const (
	EOF TokenKind = iota
	IDENTIFICADOR
	LETRERO //=string

	//Constantes
	CTE_ENTERO
	CTE_FLOTANTE

	//Palabras reservadas
	P_PROGRAMA
	P_INICIO
	P_FIN
	P_VARS
	//P_FUNCS
	P_ENTERO
	P_FLOTANTE
	P_MIENTRAS
	P_HAZ
	P_SI
	P_SINO
	P_ESCRIBE
	P_NULA

	//Operadores
	ASIGNACION
	MAS
	MENOS
	DIVISION
	MULTIPLICACION

	MAYOR_QUE
	MENOR_QUE
	IGUAL
	DIFERENTE

	//Delimitadores
	ABRE_PAREN
	CIERRA_PAREN
	ABRE_LLAVE
	CIERRA_LLAVE
	ABRE_CORCHETE
	CIERRA_CORCHETE
	SEMICOLON
	COMA
	DOS_PUNTOS
)

type Token struct {
	Kind  TokenKind
	Value string
}

func (token Token) Debug() {
	fmt.Printf("%s (%s)\n", TokenKindString(token.Kind), token.Value)
}

// función para crear los tokens
func NewToken(kind TokenKind, value string) Token {
	return Token{
		kind, value,
	}
}

func TokenKindString(kind TokenKind) string {
	switch kind {
	case EOF:
		return "eof"
	case IDENTIFICADOR:
		return "identificador"
	case LETRERO:
		return "letrero"

	// Constantes
	case CTE_ENTERO:
		return "cte_entero"
	case CTE_FLOTANTE:
		return "cte_flotante"

	// Palabras reservadas
	case P_PROGRAMA:
		return "programa"
	case P_INICIO:
		return "inicio"
	case P_FIN:
		return "fin"
	case P_VARS:
		return "vars"
	case P_ENTERO:
		return "entero"
	case P_FLOTANTE:
		return "flotante"
	case P_MIENTRAS:
		return "mientras"
	case P_HAZ:
		return "haz"
	case P_SI:
		return "si"
	case P_SINO:
		return "sino"
	case P_ESCRIBE:
		return "escribe"
	case P_NULA:
		return "nula"

	// Operadores
	case MAYOR_QUE:
		return "mayor que"
	case MENOR_QUE:
		return "menor que"
	case IGUAL:
		return "igual que"
	case DIFERENTE:
		return "diferente de"

	case ASIGNACION:
		return "asignación"
	case MAS:
		return "mas"
	case MENOS:
		return "menos"
	case DIVISION:
		return "division"
	case MULTIPLICACION:
		return "multiplicación"

	// Delimitadores
	case ABRE_PAREN:
		return "abre_paren"
	case CIERRA_PAREN:
		return "cierra_paren"
	case ABRE_LLAVE:
		return "abre_llave"
	case CIERRA_LLAVE:
		return "cierra_llave"
	case ABRE_CORCHETE:
		return "abre_corch"
	case CIERRA_CORCHETE:
		return "cierra_corch"
	case SEMICOLON:
		return "semicolon"
	case COMA:
		return "coma"
	case DOS_PUNTOS:
		return "dos puntos"
	default:
		return "desconocido"
	}
}
