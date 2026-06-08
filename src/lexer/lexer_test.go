package lexer

import (
	"testing"
)

// helper: compara la lista de tokens esperados contra los producidos
func checkTokens(t *testing.T, input string, expected []Token) {
	t.Helper()
	got := Tokenize(input)

	// el lexer siempre añade EOF al final; lo incluimos en expected
	if len(got) != len(expected) {
		t.Fatalf("se esperaban %d tokens pero se obtuvieron %d\ninput: %q\ntokens: %v",
			len(expected), len(got), input, got)
	}
	for i, tok := range expected {
		if got[i].Kind != tok.Kind || got[i].Value != tok.Value {
			t.Errorf("token[%d]: esperaba {%v %q} pero obtuve {%v %q}",
				i, tok.Kind, tok.Value, got[i].Kind, got[i].Value)
		}
	}
}

// ── Casos básicos ──────────────────────────────────────────────

func TestTokens_PalabraReservada(t *testing.T) {
	checkTokens(t, "programa", []Token{
		{P_PROGRAMA, "programa"},
		{EOF, "EOF"},
	})
}

func TestTokens_Identificador(t *testing.T) {
	checkTokens(t, "miVariable", []Token{
		{IDENTIFICADOR, "miVariable"},
		{EOF, "EOF"},
	})
}

func TestTokens_ConstanteEntera(t *testing.T) {
	checkTokens(t, "42", []Token{
		{CTE_ENTERO, "42"},
		{EOF, "EOF"},
	})
}

func TestTokens_ConstanteFlotante(t *testing.T) {
	checkTokens(t, "3.14", []Token{
		{CTE_FLOTANTE, "3.14"},
		{EOF, "EOF"},
	})
}

func TestTokens_Letrero(t *testing.T) {
	checkTokens(t, `"hola mundo"`, []Token{
		{LETRERO, `"hola mundo"`},
		{EOF, "EOF"},
	})
}

// ── Operadores de dos caracteres van antes que los de uno ──────

func TestTokens_IgualVsAsignacion(t *testing.T) {
	checkTokens(t, "==", []Token{
		{IGUAL, "=="},
		{EOF, "EOF"},
	})
}

func TestTokens_Diferente(t *testing.T) {
	checkTokens(t, "!=", []Token{
		{DIFERENTE, "!="},
		{EOF, "EOF"},
	})
}

func TestTokens_Asignacion(t *testing.T) {
	checkTokens(t, "=", []Token{
		{ASIGNACION, "="},
		{EOF, "EOF"},
	})
}

// ── Comentarios y espacios se ignoran ─────────────────────────

func TestTokens_ComentarioIgnorado(t *testing.T) {
	checkTokens(t, "// esto es un comentario\n42", []Token{
		{CTE_ENTERO, "42"},
		{EOF, "EOF"},
	})
}

func TestTokens_EspaciosIgnorados(t *testing.T) {
	checkTokens(t, "   \t\n  42", []Token{
		{CTE_ENTERO, "42"},
		{EOF, "EOF"},
	})
}

// ── Secuencias completas ───────────────────────────────────────

func TestTokens_Declaracion(t *testing.T) {
	checkTokens(t, "vars x : entero ;", []Token{
		{P_VARS, "vars"},
		{IDENTIFICADOR, "x"},
		{DOS_PUNTOS, ":"},
		{P_ENTERO, "entero"},
		{SEMICOLON, ";"},
		{EOF, "EOF"},
	})
}

func TestTokens_Asigna(t *testing.T) {
	checkTokens(t, "x = 5 + 10 ;", []Token{
		{IDENTIFICADOR, "x"},
		{ASIGNACION, "="},
		{CTE_ENTERO, "5"},
		{MAS, "+"},
		{CTE_ENTERO, "10"},
		{SEMICOLON, ";"},
		{EOF, "EOF"},
	})
}

func TestTokens_Condicion(t *testing.T) {
	checkTokens(t, "si ( x > 0 )", []Token{
		{P_SI, "si"},
		{ABRE_PAREN, "("},
		{IDENTIFICADOR, "x"},
		{MAYOR_QUE, ">"},
		{CTE_ENTERO, "0"},
		{CIERRA_PAREN, ")"},
		{EOF, "EOF"},
	})
}

// ── Token desconocido debe hacer panic ────────────────────────

func TestTokens_TokenDesconocido(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("se esperaba un panic por token no reconocido pero no ocurrió")
		}
	}()
	Tokenize("@")
}
