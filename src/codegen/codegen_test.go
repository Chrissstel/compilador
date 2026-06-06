package codegen_test

// Tests de generación de cuádruplos para el compilador Patito
// Ejecutar con: go test ./src/codegen/ -run TestQuads -v
//
// MAPA DE MEMORIA (memory.go):
//   globalInt:   1000+    globalFloat: 2000+
//   localInt:    5000+    localFloat:  6000+
//   tempInt:     9000+    tempFloat:   10000+
//   constInt:    13000+   constFloat:  14000+   constString: 15000+
//
// Las constantes SE REUTILIZAN (mismo valor → misma dir).
// Los contadores NO se resetean entre funciones.

import (
	"compilador/src/codegen"
	"compilador/src/lexer"
	"compilador/src/memory"
	"compilador/src/parser"
	"compilador/src/semantic"
	"testing"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

func compilar(t *testing.T, src string) []codegen.Quadruple {
	t.Helper()
	tokens := lexer.Tokenize(src)
	p := parser.New(tokens)
	prog := p.ParsePrograma()
	mem := memory.NewMemoryManager()
	analizador := semantic.NuevoAnalizador(mem)
	analizador.AnalizarPrograma(prog)
	if analizador.HayErrores() {
		t.Fatal("errores semánticos inesperados:", analizador.Errores())
	}
	return analizador.Generador.Quads
}

func assertQuads(t *testing.T, got, want []codegen.Quadruple) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("cantidad de cuádruplos: got %d, want %d\ngot:  %v\nwant: %v",
			len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("quad[%d]: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 1. PROGRAMA MÍNIMO — solo END
// ─────────────────────────────────────────────────────────────────────────────
func TestQuads_ProgramaVacio(t *testing.T) {
	src := `
programa vacio ;
inicio
{
}
fin`
	quads := compilar(t, src)
	want := []codegen.Quadruple{
		{Op: "END"},
	}
	assertQuads(t, quads, want)
}

// ─────────────────────────────────────────────────────────────────────────────
//  2. ASIGNACIÓN SIMPLE — entero
//     x → globalInt 1000
//     5 → constInt  13000
//     = (13000, 0, 1000)
//
// ─────────────────────────────────────────────────────────────────────────────
func TestQuads_AsignacionEntera(t *testing.T) {
	src := `
programa p ;
vars
    x : entero;
inicio
{
    x = 5 ;
}
fin`
	quads := compilar(t, src)
	want := []codegen.Quadruple{
		{Op: "=", Left: 13000, Right: 0, Result: 1000},
		{Op: "END"},
	}
	assertQuads(t, quads, want)
}

// ─────────────────────────────────────────────────────────────────────────────
//  3. ASIGNACIÓN FLOTANTE
//     y → globalFloat 2000
//     3.14 → constFloat 14000
//
// ─────────────────────────────────────────────────────────────────────────────
func TestQuads_AsignacionFlotante(t *testing.T) {
	src := `
programa p ;
vars
    y : flotante;
inicio
{
    y = 3.14 ;
}
fin`
	quads := compilar(t, src)
	want := []codegen.Quadruple{
		{Op: "=", Left: 14000, Right: 0, Result: 2000},
		{Op: "END"},
	}
	assertQuads(t, quads, want)
}

// ─────────────────────────────────────────────────────────────────────────────
//  4. SUMA — x = 2 + 3
//     2→13000, 3→13001, temp→9000
//
// ─────────────────────────────────────────────────────────────────────────────
func TestQuads_Suma(t *testing.T) {
	src := `
programa p ;
vars
    x : entero;
inicio
{
    x = 2 + 3 ;
}
fin`
	quads := compilar(t, src)
	want := []codegen.Quadruple{
		{Op: "+", Left: 13000, Right: 13001, Result: 9000},
		{Op: "=", Left: 9000, Right: 0, Result: 1000},
		{Op: "END"},
	}
	assertQuads(t, quads, want)
}

// ─────────────────────────────────────────────────────────────────────────────
//  5. PRECEDENCIA — x = 2 + 3 * 4
//     2→13000, 3→13001, 4→13002
//     * (13001, 13002) → t9000
//     + (13000, t9000) → t9001
//
// ─────────────────────────────────────────────────────────────────────────────
func TestQuads_Precedencia(t *testing.T) {
	src := `
programa p ;
vars
    x : entero;
inicio
{
    x = 2 + 3 * 4 ;
}
fin`
	quads := compilar(t, src)
	want := []codegen.Quadruple{
		{Op: "*", Left: 13001, Right: 13002, Result: 9000},
		{Op: "+", Left: 13000, Right: 9000, Result: 9001},
		{Op: "=", Left: 9001, Right: 0, Result: 1000},
		{Op: "END"},
	}
	assertQuads(t, quads, want)
}

// ─────────────────────────────────────────────────────────────────────────────
//  6. OPERADOR RELACIONAL — x = 5 > 3
//     5→13000, 3→13001, resultado entero→t9000
//
// ─────────────────────────────────────────────────────────────────────────────
func TestQuads_Relacional(t *testing.T) {
	src := `
programa p ;
vars
    x : entero;
inicio
{
    x = 5 > 3 ;
}
fin`
	quads := compilar(t, src)
	want := []codegen.Quadruple{
		{Op: ">", Left: 13000, Right: 13001, Result: 9000},
		{Op: "=", Left: 9000, Right: 0, Result: 1000},
		{Op: "END"},
	}
	assertQuads(t, quads, want)
}

// ─────────────────────────────────────────────────────────────────────────────
//  7. CONSTANTE REUTILIZADA — x = 5 + 5
//     5 siempre es 13000 (se reutiliza del mapa)
//
// ─────────────────────────────────────────────────────────────────────────────
func TestQuads_ConstanteReutilizada(t *testing.T) {
	src := `
programa p ;
vars
    x : entero;
inicio
{
    x = 5 + 5 ;
}
fin`
	quads := compilar(t, src)
	want := []codegen.Quadruple{
		{Op: "+", Left: 13000, Right: 13000, Result: 9000},
		{Op: "=", Left: 9000, Right: 0, Result: 1000},
		{Op: "END"},
	}
	assertQuads(t, quads, want)
}

// ─────────────────────────────────────────────────────────────────────────────
//  8. IMPRIME — letrero y variable
//     "hola"→15000, x→1000
//
// ─────────────────────────────────────────────────────────────────────────────
func TestQuads_Imprime(t *testing.T) {
	src := `
programa p ;
vars
    x : entero;
inicio
{
    x = 1 ;
    escribe( "hola" ) ;
    escribe( x ) ;
}
fin`
	quads := compilar(t, src)
	want := []codegen.Quadruple{
		{Op: "=", Left: 13000, Right: 0, Result: 1000},
		{Op: "PRINT_STR", Left: 15000, Right: 0, Result: 0},
		{Op: "PRINT", Left: 1000, Right: 0, Result: 0},
		{Op: "END"},
	}
	assertQuads(t, quads, want)
}

// ─────────────────────────────────────────────────────────────────────────────
//  9. CONDICIONAL SIN SINO
//     x→1000, 1→13000, 0→13001, cond→t9000
//     quad 0: = 13000 0 1000       x = 1
//     quad 1: > 1000 13001 9000    x > 0
//     quad 2: GOTOF 9000 0 4       si falso salta al END
//     quad 3: PRINT 1000
//     quad 4: END
//
// ─────────────────────────────────────────────────────────────────────────────
func TestQuads_CondicionSinSino(t *testing.T) {
	src := `
programa p ;
vars
    x : entero;
inicio
{
    x = 1 ;
    si ( x > 0 )
    {
        escribe( x ) ;
    } ;
}
fin`
	quads := compilar(t, src)
	want := []codegen.Quadruple{
		{Op: "=", Left: 13000, Right: 0, Result: 1000},
		{Op: ">", Left: 1000, Right: 13001, Result: 9000},
		{Op: "GOTOF", Left: 9000, Right: 0, Result: 4},
		{Op: "PRINT", Left: 1000, Right: 0, Result: 0},
		{Op: "END"},
	}
	assertQuads(t, quads, want)
}

// ─────────────────────────────────────────────────────────────────────────────
//  10. CONDICIONAL CON SINO
//     quad 0: = 13000 0 1000
//     quad 1: > 1000 13001 9000
//     quad 2: GOTOF 9000 0 5       salta al sino
//     quad 3: PRINT 1000
//     quad 4: GOTO 0 0 6           salta al final
//     quad 5: PRINT 13001          escribe 0 (reutiliza 13001)
//     quad 6: END
//
// ─────────────────────────────────────────────────────────────────────────────
func TestQuads_CondicionConSino(t *testing.T) {
	src := `
programa p ;
vars
    x : entero;
inicio
{
    x = 1 ;
    si ( x > 0 )
    {
        escribe( x ) ;
    }
    sino
    {
        escribe( 0 ) ;
    } ;
}
fin`
	quads := compilar(t, src)
	want := []codegen.Quadruple{
		{Op: "=", Left: 13000, Right: 0, Result: 1000},
		{Op: ">", Left: 1000, Right: 13001, Result: 9000},
		{Op: "GOTOF", Left: 9000, Right: 0, Result: 5},
		{Op: "PRINT", Left: 1000, Right: 0, Result: 0},
		{Op: "GOTO", Left: 0, Right: 0, Result: 6},
		{Op: "PRINT", Left: 13001, Right: 0, Result: 0}, // 0 reutiliza 13001
		{Op: "END"},
	}
	assertQuads(t, quads, want)
}

// ─────────────────────────────────────────────────────────────────────────────
//  11. CICLO MIENTRAS
//     x→1000, 5→13000, 0→13001, 1→13002
//     quad 0: = 13000 0 1000        x = 5
//     quad 1: != 1000 13001 9000    x != 0
//     quad 2: GOTOF 9000 0 6
//     quad 3: - 1000 13002 9001     x - 1
//     quad 4: = 9001 0 1000
//     quad 5: GOTO 0 0 1
//     quad 6: END
//
// ─────────────────────────────────────────────────────────────────────────────
func TestQuads_Ciclo(t *testing.T) {
	src := `
programa p ;
vars
    x : entero;
inicio
{
    x = 5 ;
    mientras ( x != 0 ) haz
    {
        x = x - 1 ;
    } ;
}
fin`
	quads := compilar(t, src)
	want := []codegen.Quadruple{
		{Op: "=", Left: 13000, Right: 0, Result: 1000},
		{Op: "!=", Left: 1000, Right: 13001, Result: 9000},
		{Op: "GOTOF", Left: 9000, Right: 0, Result: 6},
		{Op: "-", Left: 1000, Right: 13002, Result: 9001},
		{Op: "=", Left: 9001, Right: 0, Result: 1000},
		{Op: "GOTO", Left: 0, Right: 0, Result: 1},
		{Op: "END"},
	}
	assertQuads(t, quads, want)
}

// ─────────────────────────────────────────────────────────────────────────────
//  12. FUNCIÓN NULA SIN PARÁMETROS
//     quad 0: GOTO 0 0 3
//     quad 1: PRINT 13000          escribe 1
//     quad 2: ENDFUNC
//     quad 3: END
//
// ─────────────────────────────────────────────────────────────────────────────
func TestQuads_FuncionNula(t *testing.T) {
	src := `
programa p ;

nula saluda() {
    {
        escribe( 1 ) ;
    }
} ;

inicio
{
}
fin`
	quads := compilar(t, src)
	want := []codegen.Quadruple{
		{Op: "GOTO", Left: 0, Right: 0, Result: 3},
		{Op: "PRINT", Left: 13000, Right: 0, Result: 0},
		{Op: "ENDFUNC"},
		{Op: "END"},
	}
	assertQuads(t, quads, want)
}

// ─────────────────────────────────────────────────────────────────────────────
//  13. LLAMADA A FUNCIÓN NULA
//     quad 0: GOTO 0 0 3
//     quad 1: PRINT 13000
//     quad 2: ENDFUNC
//     quad 3: ERA 0 0 0
//     quad 4: GOSUB 0 0 1
//     quad 5: END
//
// ─────────────────────────────────────────────────────────────────────────────
func TestQuads_LlamadaFuncionNula(t *testing.T) {
	src := `
programa p ;

nula saluda() {
    {
        escribe( 1 ) ;
    }
} ;

inicio
{
    saluda() ;
}
fin`
	quads := compilar(t, src)
	want := []codegen.Quadruple{
		{Op: "GOTO", Left: 0, Right: 0, Result: 3},
		{Op: "PRINT", Left: 13000, Right: 0, Result: 0},
		{Op: "ENDFUNC"},
		{Op: "ERA", Left: 0, Right: 0, Result: 0},
		{Op: "GOSUB", Left: 0, Right: 0, Result: 1},
		{Op: "END"},
	}
	assertQuads(t, quads, want)
}

// ─────────────────────────────────────────────────────────────────────────────
//  14. FUNCIÓN CON PARÁMETRO Y RETORNO
//     n → localInt 5000, r → localInt 5001
//     quad 0: GOTO 0 0 5
//     quad 1: + 5000 5000 9000     n + n
//     quad 2: = 9000 0 5001        r = ...
//     quad 3: RETURN 5001
//     quad 4: ENDFUNC
//     quad 5: END
//
// ─────────────────────────────────────────────────────────────────────────────
func TestQuads_FuncionConRetorno(t *testing.T) {
	src := `
programa p ;

entero doble( n : entero ) {
    vars
        r : entero;
    {
        r = n + n ;
    }
    retornar r ;
} ;

inicio
{
}
fin`
	quads := compilar(t, src)
	want := []codegen.Quadruple{
		{Op: "GOTO", Left: 0, Right: 0, Result: 5},
		{Op: "+", Left: 5000, Right: 5000, Result: 9000},
		{Op: "=", Left: 9000, Right: 0, Result: 5001},
		{Op: "RETURN", Left: 5001, Right: 0, Result: 0},
		{Op: "ENDFUNC"},
		{Op: "END"},
	}
	assertQuads(t, quads, want)
}

// ─────────────────────────────────────────────────────────────────────────────
//  15. LLAMADA A FUNCIÓN CON RETORNO
//     x→1000, n→5000, r→5001, 3→13000
//     quad 0: GOTO 0 0 5
//     quad 1: + 5000 5000 9000
//     quad 2: = 9000 0 5001
//     quad 3: RETURN 5001
//     quad 4: ENDFUNC
//     quad 5: ERA 0 0 0
//     quad 6: PARAM 13000 0 0      argumento 3
//     quad 7: GOSUB 0 0 1
//     quad 8: RETVAL 0 0 9001
//     quad 9: = 9001 0 1000
//     quad 10: END
//
// ─────────────────────────────────────────────────────────────────────────────
func TestQuads_LlamadaConRetorno(t *testing.T) {
	src := `
programa p ;
vars
    x : entero;

entero doble( n : entero ) {
    vars
        r : entero;
    {
        r = n + n ;
    }
    retornar r ;
} ;

inicio
{
    x = doble( 3 ) ;
}
fin`
	quads := compilar(t, src)
	want := []codegen.Quadruple{
		{Op: "GOTO", Left: 0, Right: 0, Result: 5},
		{Op: "+", Left: 5000, Right: 5000, Result: 9000},
		{Op: "=", Left: 9000, Right: 0, Result: 5001},
		{Op: "RETURN", Left: 5001, Right: 0, Result: 0},
		{Op: "ENDFUNC"},
		{Op: "ERA", Left: 2, Right: 0, Result: 0},
		{Op: "PARAM", Left: 13000, Right: 0, Result: 0},
		{Op: "GOSUB", Left: 0, Right: 0, Result: 1},
		{Op: "RETURN", Left: 0, Right: 0, Result: 9001},
		{Op: "=", Left: 9001, Right: 0, Result: 1000},
		{Op: "END"},
	}
	assertQuads(t, quads, want)
}

// ─────────────────────────────────────────────────────────────────────────────
//  16. NEGACIÓN UNARIA — x = -5
//     5→13000, NEG→t9000
//
// ─────────────────────────────────────────────────────────────────────────────
func TestQuads_Negacion(t *testing.T) {
	src := `
programa p ;
vars
    x : entero;
inicio
{
    x = -5 ;
}
fin`
	quads := compilar(t, src)
	want := []codegen.Quadruple{
		{Op: "NEG", Left: 13000, Right: 0, Result: 9000},
		{Op: "=", Left: 9000, Right: 0, Result: 1000},
		{Op: "END"},
	}
	assertQuads(t, quads, want)
}

// ─────────────────────────────────────────────────────────────────────────────
//  17. DOS VARIABLES GLOBALES — dirs consecutivas
//     a→1000, b→1001, 1→13000, 2→13001
//
// ─────────────────────────────────────────────────────────────────────────────
func TestQuads_DosVariablesGlobales(t *testing.T) {
	src := `
programa p ;
vars
    a : entero;
    b : entero;
inicio
{
    a = 1 ;
    b = 2 ;
}
fin`
	quads := compilar(t, src)
	want := []codegen.Quadruple{
		{Op: "=", Left: 13000, Right: 0, Result: 1000},
		{Op: "=", Left: 13001, Right: 0, Result: 1001},
		{Op: "END"},
	}
	assertQuads(t, quads, want)
}

// ─────────────────────────────────────────────────────────────────────────────
// 18. DOS LETREROS DIFERENTES — dirs de string consecutivas
// ─────────────────────────────────────────────────────────────────────────────
func TestQuads_ImprimeDosLetreros(t *testing.T) {
	src := `
programa p ;
inicio
{
    escribe( "hola" ) ;
    escribe( "mundo" ) ;
}
fin`
	quads := compilar(t, src)
	want := []codegen.Quadruple{
		{Op: "PRINT_STR", Left: 15000, Right: 0, Result: 0},
		{Op: "PRINT_STR", Left: 15001, Right: 0, Result: 0},
		{Op: "END"},
	}
	assertQuads(t, quads, want)
}

// ─────────────────────────────────────────────────────────────────────────────
// 19. LETRERO REUTILIZADO — misma string, misma dir
// ─────────────────────────────────────────────────────────────────────────────
func TestQuads_LetreroReutilizado(t *testing.T) {
	src := `
programa p ;
inicio
{
    escribe( "hola" ) ;
    escribe( "hola" ) ;
}
fin`
	quads := compilar(t, src)
	want := []codegen.Quadruple{
		{Op: "PRINT_STR", Left: 15000, Right: 0, Result: 0},
		{Op: "PRINT_STR", Left: 15000, Right: 0, Result: 0},
		{Op: "END"},
	}
	assertQuads(t, quads, want)
}

// ─────────────────────────────────────────────────────────────────────────────
//  20. CICLO ANIDADO EN CONDICIONAL
//     Valida estructura de saltos sin hardcodear todos los índices.
//     x→1000, 3→13000, 0→13001, 1→13002
//     quad 0: = 13000 0 1000
//     quad 1: > 1000 13001 9000    cond del si
//     quad 2: GOTOF 9000 0 9       salta al END
//     quad 3: != 1000 13001 9001   cond del mientras (reutiliza 13001)
//     quad 4: GOTOF 9001 0 8       salta al fin del ciclo
//     quad 5: - 1000 13002 9002
//     quad 6: = 9002 0 1000
//     quad 7: GOTO 0 0 3           vuelve a cond del mientras
//     quad 8: END                  ← GOTOF del si y GOTOF del ciclo apuntan aquí
//
// ─────────────────────────────────────────────────────────────────────────────
func TestQuads_CicloAnidadoEnCondicion(t *testing.T) {
	src := `
programa p ;
vars
    x : entero;
inicio
{
    x = 3 ;
    si ( x > 0 )
    {
        mientras ( x != 0 ) haz
        {
            x = x - 1 ;
        } ;
    } ;
}
fin`
	quads := compilar(t, src)

	if len(quads) == 0 {
		t.Fatal("no se generaron cuádruplos")
	}

	// Verificaciones estructurales sin asumir índices exactos de salto
	if quads[0].Op != "=" {
		t.Errorf("quad[0] debe ser =, got %s", quads[0].Op)
	}
	if quads[1].Op != ">" {
		t.Errorf("quad[1] debe ser > (cond del si), got %s", quads[1].Op)
	}
	gotofSi := quads[2]
	if gotofSi.Op != "GOTOF" {
		t.Errorf("quad[2] debe ser GOTOF (del si), got %s", gotofSi.Op)
	}
	if quads[3].Op != "!=" {
		t.Errorf("quad[3] debe ser != (cond del mientras), got %s", quads[3].Op)
	}
	gotofCiclo := quads[4]
	if gotofCiclo.Op != "GOTOF" {
		t.Errorf("quad[4] debe ser GOTOF (del ciclo), got %s", gotofCiclo.Op)
	}
	gotoVuelta := quads[7]
	if gotoVuelta.Op != "GOTO" || gotoVuelta.Result != 3 {
		t.Errorf("quad[7] debe ser GOTO→3, got %+v", gotoVuelta)
	}
	if quads[len(quads)-1].Op != "END" {
		t.Errorf("último quad debe ser END, got %s", quads[len(quads)-1].Op)
	}
	// El GOTOF del si debe saltar más lejos que el GOTOF del ciclo
	if gotofSi.Result < gotofCiclo.Result {
		t.Errorf("GOTOF del si (%d) debe apuntar más lejos que GOTOF del ciclo (%d)",
			gotofSi.Result, gotofCiclo.Result)
	}
}
