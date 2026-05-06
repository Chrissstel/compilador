package parser

import (
	"compilador/src/ast"
	"compilador/src/lexer"
	"os"
	"path/filepath"
	"testing"
)

// parsea un string y devuelve el ast o falla con panic
func parseStr(t *testing.T, input string) *ast.Programa {
	t.Helper()
	tokens := lexer.Tokenize(input)
	p := New(tokens) //el parser
	return p.ParsePrograma()
}

// parsea y espera que falle
func expectFail(t *testing.T, input string) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Error("se esperaba un error de parseo pero no falló")
		}
	}()
	tokens := lexer.Tokenize(input)
	New(tokens).ParsePrograma()
}

// ── Programa mínimo válido ─────────────────────────────────────

func TestParser_ProgramaMinimo(t *testing.T) {
	prog := parseStr(t, `
		programa minimo ;
		inicio
		{
		}
		fin
	`)
	if prog.ID != "minimo" {
		t.Errorf("ID esperado 'minimo', salió '%s'", prog.ID)
	}
	if prog.Vars != nil {
		t.Error("no debería haber vars")
	}
	if len(prog.Funcs) != 0 {
		t.Error("no debería haber funciones")
	}
}

// variables
func TestParser_VarsSimples(t *testing.T) {
	prog := parseStr(t, `
		programa test ;
		vars
			a, b : entero;
			c : flotante;
		inicio { } fin
	`)
	if prog.Vars == nil {
		t.Fatal("se esperaban vars")
	}
	if len(prog.Vars.Declaraciones) != 2 {
		t.Fatalf("se esperaban 2 declaraciones, salieron %d", len(prog.Vars.Declaraciones))
	}
}

// estatutos
func TestParser_Asignacion(t *testing.T) {
	prog := parseStr(t, `
		programa test ;
		inicio
		{
			x = 42 ;
		}
		fin
	`)
	if len(prog.Cuerpo.Estatutos) != 1 {
		t.Fatalf("se esperaba 1 estatuto, hay %d", len(prog.Cuerpo.Estatutos))
	}
	asigna, ok := prog.Cuerpo.Estatutos[0].(*ast.Asigna)
	if !ok {
		t.Fatal("se esperaba *ast.Asigna")
	}
	if asigna.ID != "x" {
		t.Errorf("ID esperado 'x', salió '%s'", asigna.ID)
	}
}

func TestParser_Condicion_SinSino(t *testing.T) {
	prog := parseStr(t, `
		programa test ;
		inicio
		{
			si ( x > 0 ) { } ;
		}
		fin
	`)
	cond, ok := prog.Cuerpo.Estatutos[0].(*ast.Condicion)
	if !ok {
		t.Fatal("se esperaba *ast.Condicion")
	}
	if cond.CuerpoSino != nil {
		t.Error("no debería haber cuerpo sino")
	}
}

func TestParser_Condicion_ConSino(t *testing.T) {
	prog := parseStr(t, `
		programa test ;
		inicio
		{
			si ( x > 0 ) { } sino { } ;
		}
		fin
	`)
	cond, ok := prog.Cuerpo.Estatutos[0].(*ast.Condicion)
	if !ok {
		t.Fatal("se esperaba *ast.Condicion")
	}
	if cond.CuerpoSino == nil {
		t.Error("se esperaba cuerpo sino")
	}
}

func TestParser_Ciclo(t *testing.T) {
	prog := parseStr(t, `
		programa test ;
		inicio
		{
			mientras ( x != 0 ) haz { } ;
		}
		fin
	`)
	_, ok := prog.Cuerpo.Estatutos[0].(*ast.Ciclo)
	if !ok {
		t.Fatal("se esperaba *ast.Ciclo")
	}
}

func TestParser_Escribe(t *testing.T) {
	prog := parseStr(t, `
		programa test ;
		inicio
		{
			escribe( "hola", x ) ;
		}
		fin
	`)
	imp, ok := prog.Cuerpo.Estatutos[0].(*ast.Imprime)
	if !ok {
		t.Fatal("se esperaba *ast.Imprime")
	}
	if len(imp.Items) != 2 {
		t.Fatalf("se esperaban 2 items, obtuve %d", len(imp.Items))
	}
	if !imp.Items[0].EsLetrero {
		t.Error("el primer item debería ser letrero")
	}
}

func TestParser_LlamadaFuncion(t *testing.T) {
	prog := parseStr(t, `
		programa test ;
		inicio
		{
			funcionA( 1, 2 ) ;
		}
		fin
	`)
	_, ok := prog.Cuerpo.Estatutos[0].(*ast.Llamada)
	if !ok {
		t.Fatal("se esperaba *ast.Llamada")
	}
}

// funciones
func TestParser_FuncionNula(t *testing.T) {
	prog := parseStr(t, `
		programa test ;
		nula saludar( nombre : entero ) {
			{
				escribe( nombre ) ;
			}
		} ;
		inicio { } fin
	`)
	if len(prog.Funcs) != 1 {
		t.Fatalf("se esperaba 1 función, salieron %d", len(prog.Funcs))
	}
	fn := prog.Funcs[0]
	if fn.ID != "saludar" {
		t.Errorf("nombre esperado 'saludar', obtuve '%s'", fn.ID)
	}
	if fn.TipoRetorno != "nula" {
		t.Errorf("tipo esperado 'nula', salió '%s'", fn.TipoRetorno)
	}
	if len(fn.Params) != 1 {
		t.Fatalf("se esperaba 1 param, obtuve %d", len(fn.Params))
	}
}

// orden de operaciones
func TestParser_PrecedenciaOperadores(t *testing.T) {
	parseStr(t, `
		programa test ;
		inicio
		{
			x = 2 + 3 * 4 ;
		}
		fin
	`)
}

// casos que deben fallar

func TestParser_FaltaFin(t *testing.T) {
	expectFail(t, `
		programa test ;
		inicio { }
	`)
}

func TestParser_FaltaSemicolonAsigna(t *testing.T) {
	expectFail(t, `
		programa test ;
		inicio
		{
			x = 5
		}
		fin
	`)
}

func TestParser_FaltaCierreParen(t *testing.T) {
	expectFail(t, `
		programa test ;
		inicio
		{
			si ( x > 0 { } ;
		}
		fin
	`)
}

// tests de archivos .patito

func TestParser_ArchivosValidos(t *testing.T) {
	archivos, err := filepath.Glob("../../testdata/valid/*.patito")
	if err != nil {
		t.Fatal(err)
	}
	if len(archivos) == 0 {
		t.Skip("no hay archivos en testdata/valid/")
	}
	for _, archivo := range archivos {
		t.Run(filepath.Base(archivo), func(t *testing.T) {
			src, err := os.ReadFile(archivo)
			if err != nil {
				t.Fatal(err)
			}
			tokens := lexer.Tokenize(string(src))
			p := New(tokens)
			p.ParsePrograma() // no debe hacer panic
		})
	}
}

func TestParser_ArchivosInvalidos(t *testing.T) {
	archivos, err := filepath.Glob("../../testdata/invalid/*.patito")
	if err != nil {
		t.Fatal(err)
	}
	if len(archivos) == 0 {
		t.Skip("no hay archivos en testdata/invalid/")
	}
	for _, archivo := range archivos {
		t.Run(filepath.Base(archivo), func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("se esperaba error en %s pero el parser no falló", archivo)
				}
			}()
			src, err := os.ReadFile(archivo)
			if err != nil {
				t.Fatal(err)
			}
			tokens := lexer.Tokenize(string(src))
			New(tokens).ParsePrograma()
		})
	}
}
