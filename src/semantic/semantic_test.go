package semantic

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"compilador/src/lexer"
	"compilador/src/parser"
)

// helper: analiza un string y devuelve los errores semánticos
func analizarStr(t *testing.T, input string) []string {
	t.Helper()
	tokens := lexer.Tokenize(input)
	p := parser.New(tokens)
	prog := p.ParsePrograma()

	a := NuevoAnalizador()
	a.AnalizarPrograma(prog)
	return a.errores
}

// helper: espera que no haya errores
func debeSerValido(t *testing.T, input string) {
	t.Helper()
	errores := analizarStr(t, input)
	if len(errores) > 0 {
		t.Errorf("no se esperaban errores pero se obtuvieron:\n%s",
			strings.Join(errores, "\n"))
	}
}

// helper: espera que haya al menos un error que contenga el substring dado
func debeContenerError(t *testing.T, input string, fragmento string) {
	t.Helper()
	errores := analizarStr(t, input)
	if len(errores) == 0 {
		t.Errorf("se esperaba error con '%s' pero no hubo errores", fragmento)
		return
	}
	for _, e := range errores {
		if strings.Contains(e, fragmento) {
			return
		}
	}
	t.Errorf("se esperaba error con '%s' pero los errores fueron:\n%s",
		fragmento, strings.Join(errores, "\n"))
}

// helper: espera exactamente N errores
func debeHaberNErrores(t *testing.T, input string, n int) {
	t.Helper()
	errores := analizarStr(t, input)
	if len(errores) != n {
		t.Errorf("se esperaban %d errores pero se obtuvieron %d:\n%s",
			n, len(errores), strings.Join(errores, "\n"))
	}
}

// ── Casos válidos ──────────────────────────────────────────────

func TestSemantico_ProgramaMinimo(t *testing.T) {
	debeSerValido(t, `
		programa test ;
		inicio { } fin
	`)
}

func TestSemantico_VarsYAsigna(t *testing.T) {
	debeSerValido(t, `
		programa test ;
		vars
			x : entero;
		inicio
		{
			x = 10 ;
		}
		fin
	`)
}

func TestSemantico_FuncionValida(t *testing.T) {
	debeSerValido(t, `
		programa test ;
		entero suma( a : entero, b : entero ) {
			{
				a = a + b ;
			}
		} ;
		inicio { } fin
	`)
}

func TestSemantico_LlamadaValida(t *testing.T) {
	debeSerValido(t, `
		programa test ;
		vars
			res : entero;
		entero suma( a : entero, b : entero ) {
			{
				a = a + b ;
			}
		} ;
		inicio
		{
			res = suma( 1, 2 ) ;
		}
		fin
	`)
}

// ── Variables ──────────────────────────────────────────────────

func TestSemantico_VarNoDec(t *testing.T) {
	debeContenerError(t, `
		programa test ;
		inicio
		{
			x = 10 ;
		}
		fin
	`, "variable 'x' no fue declarada")
}

func TestSemantico_VarDuplicada(t *testing.T) {
	debeContenerError(t, `
		programa test ;
		vars
			x : entero;
			x : flotante;
		inicio { } fin
	`, "variable 'x' ya fue declarada")
}

func TestSemantico_VarEnExpresion(t *testing.T) {
	debeContenerError(t, `
		programa test ;
		vars
			x : entero;
		inicio
		{
			x = x + z ;
		}
		fin
	`, "variable 'z' no fue declarada")
}

func TestSemantico_VarEnCondicion(t *testing.T) {
	debeContenerError(t, `
		programa test ;
		vars
			x : entero;
		inicio
		{
			si ( x > fantasma ) { } ;
		}
		fin
	`, "variable 'fantasma' no fue declarada")
}

func TestSemantico_VarEnCiclo(t *testing.T) {
	debeContenerError(t, `
		programa test ;
		vars
			x : entero;
		inicio
		{
			mientras ( x != limite ) haz { } ;
		}
		fin
	`, "variable 'limite' no fue declarada")
}

func TestSemantico_VarLocalFuera(t *testing.T) {
	debeContenerError(t, `
		programa test ;
		nula miFuncion( ) {
			vars
				local : entero;
			{
				local = 5 ;
			}
		} ;
		inicio
		{
			local = 10 ;
		}
		fin
	`, "variable 'local' no fue declarada")
}

// ── Funciones ──────────────────────────────────────────────────

func TestSemantico_FuncNoDec(t *testing.T) {
	debeContenerError(t, `
		programa test ;
		inicio
		{
			miFuncion( 1, 2 ) ;
		}
		fin
	`, "función 'miFuncion' no fue declarada")
}

func TestSemantico_FuncDuplicada(t *testing.T) {
	debeContenerError(t, `
		programa test ;
		nula saludar( ) { { } } ;
		nula saludar( ) { { } } ;
		inicio { } fin
	`, "función 'saludar' ya fue declarada")
}

func TestSemantico_ArgsDeMas(t *testing.T) {
	debeContenerError(t, `
		programa test ;
		entero suma( a : entero, b : entero ) {
			{ a = a + b ; }
		} ;
		inicio
		{
			suma( 1, 2, 3 ) ;
		}
		fin
	`, "espera 2 argumento(s) pero recibió 3")
}

func TestSemantico_ArgsDeMenos(t *testing.T) {
	debeContenerError(t, `
		programa test ;
		entero suma( a : entero, b : entero ) {
			{ a = a + b ; }
		} ;
		inicio
		{
			suma( 1 ) ;
		}
		fin
	`, "espera 2 argumento(s) pero recibió 1")
}

// ── Múltiples errores ──────────────────────────────────────────

func TestSemantico_MultipleErrores(t *testing.T) {
	debeHaberNErrores(t, `
		programa test ;
		inicio
		{
			a = 1 ;
			b = 2 ;
			funcionFantasma( a ) ;
		}
		fin
	`, 3)
}

// ── Archivos ───────────────────────────────────────────────────

func TestSemantico_ArchivosValidos(t *testing.T) {
	archivos, err := filepath.Glob("../../testdata/valid/semantico_*.patito")
	if err != nil {
		t.Fatal(err)
	}
	if len(archivos) == 0 {
		t.Skip("no hay archivos semantico_* en testdata/valid/")
	}
	for _, archivo := range archivos {
		t.Run(filepath.Base(archivo), func(t *testing.T) {
			src, _ := os.ReadFile(archivo)
			tokens := lexer.Tokenize(string(src))
			prog := parser.New(tokens).ParsePrograma()
			a := NuevoAnalizador()
			a.AnalizarPrograma(prog)
			if a.HayErrores() {
				t.Errorf("no se esperaban errores:\n%s",
					strings.Join(a.errores, "\n"))
			}
		})
	}
}

func TestSemantico_ArchivosInvalidos(t *testing.T) {
	archivos, err := filepath.Glob("../../testdata/invalid/*.patito")
	if err != nil {
		t.Fatal(err)
	}
	if len(archivos) == 0 {
		t.Skip("no hay archivos en testdata/invalid/")
	}
	for _, archivo := range archivos {
		t.Run(filepath.Base(archivo), func(t *testing.T) {
			src, _ := os.ReadFile(archivo)

			// algunos inválidos fallan en el parser (sintaxis rota)
			// los que fallan en el semántico deben tener errores
			func() {
				defer func() { recover() }() // atrapa panics del parser
				tokens := lexer.Tokenize(string(src))
				prog := parser.New(tokens).ParsePrograma()
				a := NuevoAnalizador()
				a.AnalizarPrograma(prog)
				if !a.HayErrores() {
					t.Errorf("%s debería tener errores pero pasó limpio",
						filepath.Base(archivo))
				}
			}()
		})
	}
}
