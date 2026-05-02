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
	ENTERO
	FLOTANTE

	//Palabras reservadas
	P_PROGRAMA
	P_INICIO
	P_FIN
	P_VARS
	P_FUNCS
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
	ABRE_CORCH
	CIERRA_CORCH
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
	case ENTERO:
		return "entero"
	case FLOTANTE:
		return "flotante"

	// Palabras reservadas
	case P_PROGRAMA:
		return "p_programa"
	case P_INICIO:
		return "p_inicio"
	case P_FIN:
		return "p_fin"
	case P_VARS:
		return "p_vars"
	case P_FUNCS:
		return "p_funcs"
	case P_ENTERO:
		return "p_entero"
	case P_FLOTANTE:
		return "p_flotante"
	case P_MIENTRAS:
		return "p_mientras"
	case P_HAZ:
		return "p_haz"
	case P_SI:
		return "p_si"
	case P_SINO:
		return "p_sino"
	case P_ESCRIBE:
		return "p_escribe"
	case P_NULA:
		return "p_nula"

	// Operadores
	case ASIGNACION:
		return "asignacion"
	case MAS:
		return "mas"
	case MENOS:
		return "menos"
	case DIVISION:
		return "division"
	case MULTIPLICACION:
		return "multiplicacion"
	case MAYOR_QUE:
		return "mayor_que"
	case MENOR_QUE:
		return "menor_que"
	case IGUAL:
		return "igual"
	case DIFERENTE:
		return "diferente"

	// Delimitadores
	case ABRE_PAREN:
		return "abre_paren"
	case CIERRA_PAREN:
		return "cierra_paren"
	case ABRE_CORCH:
		return "abre_corch"
	case CIERRA_CORCH:
		return "cierra_corch"
	case SEMICOLON:
		return "semicolon"
	case COMA:
		return "coma"
	case DOS_PUNTOS:
		return "dos_puntos"

	default:
		return "desconocido"
	}
}
