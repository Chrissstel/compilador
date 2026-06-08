package vm_test

import (
	"bytes"
	"compilador/src/codegen"
	"compilador/src/lexer"
	"compilador/src/memory"
	"compilador/src/parser"
	"compilador/src/semantic"
	"compilador/src/vm"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func compilar(src string) ([]codegen.Quadruple, *memory.MemoryManager, error) {
	tokens := lexer.Tokenize(src)

	p := parser.New(tokens)
	programa := p.ParsePrograma()

	mem := memory.NewMemoryManager()
	analizador := semantic.NuevoAnalizador(mem)
	analizador.AnalizarPrograma(programa)

	if analizador.HayErrores() {
		return nil, nil, fmt.Errorf("errores semánticos")
	}

	return analizador.Generador.Quads, mem, nil
}

// capturarSalida ejecuta la VM y captura todo lo que se imprime a stdout.
func capturarSalida(quads []codegen.Quadruple, manager *memory.MemoryManager) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	machine := vm.NewVM(quads, manager)
	machine.Run()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return strings.TrimSpace(buf.String())
}

// correr compila y ejecuta, devuelve stdout. Llama t.Fatal si hay error.
func correr(t *testing.T, src string) string {
	t.Helper()
	quads, manager, err := compilar(src)
	if err != nil {
		t.Fatalf("error compilando: %v", err)
	}
	return capturarSalida(quads, manager)
}

// ---------------------------------------------------------------------------
// 1. Programa mínimo – imprime un literal entero
// ---------------------------------------------------------------------------

func TestHolaMundo(t *testing.T) {
	src := `
programa hola ;
inicio
{
    escribe( 42 ) ;
}
fin
`
	got := correr(t, src)
	if got != "42" {
		t.Errorf("esperaba '42', obtuve %q", got)
	}
}

// ---------------------------------------------------------------------------
// 2. Asignación y aritmética básica con enteros
// ---------------------------------------------------------------------------

func TestAritmeticaEntera(t *testing.T) {
	src := `
programa arit ;
vars
    a : entero ;
    b : entero ;
    c : entero ;
inicio
{
    a = 10 ;
    b = 3 ;
    c = a + b * 2 ;
    escribe( c ) ;
}
fin
`
	// 10 + (3*2) = 16
	got := correr(t, src)
	if got != "16" {
		t.Errorf("esperaba '16', obtuve %q", got)
	}
}

// ---------------------------------------------------------------------------
// 3. Aritmética con flotantes
// ---------------------------------------------------------------------------

func TestAritmeticaFlotante(t *testing.T) {
	src := `
programa floats ;
vars
    x : flotante ;
    y : flotante ;
    z : flotante ;
inicio
{
    x = 1.5 ;
    y = 2.5 ;
    z = x + y ;
    escribe( z ) ;
}
fin
`
	got := correr(t, src)
	if got != "4" && got != "4.0" {
		t.Errorf("esperaba '4' o '4.0', obtuve %q", got)
	}
}

// ---------------------------------------------------------------------------
// 4. Condicional – rama verdadera
// ---------------------------------------------------------------------------

func TestSiVerdadero(t *testing.T) {
	src := `
programa cond1 ;
vars
    x : entero ;
inicio
{
    x = 10 ;
    si ( x > 5 )
    {
        escribe( 1 ) ;
    }
    sino
    {
        escribe( 0 ) ;
    } ;
}
fin
`
	got := correr(t, src)
	if got != "1" {
		t.Errorf("esperaba '1', obtuve %q", got)
	}
}

// ---------------------------------------------------------------------------
// 5. Condicional – rama falsa
// ---------------------------------------------------------------------------

func TestSiFalso(t *testing.T) {
	src := `
programa cond2 ;
vars
    x : entero ;
inicio
{
    x = 3 ;
    si ( x > 5 )
    {
        escribe( 1 ) ;
    }
    sino
    {
        escribe( 0 ) ;
    } ;
}
fin
`
	got := correr(t, src)
	if got != "0" {
		t.Errorf("esperaba '0', obtuve %q", got)
	}
}

// ---------------------------------------------------------------------------
// 6. Ciclo mientras – cuenta regresiva
// ---------------------------------------------------------------------------

func TestMientras(t *testing.T) {
	src := `
programa loop1 ;
vars
    i : entero ;
inicio
{
    i = 3 ;
    mientras ( i > 0 ) haz
    {
        escribe( i ) ;
        i = i - 1 ;
    } ;
}
fin
`
	got := correr(t, src)
	lines := strings.Split(got, "\n")
	want := []string{"3", "2", "1"}
	if len(lines) != len(want) {
		t.Fatalf("esperaba %d líneas, obtuve %d: %q", len(want), len(lines), got)
	}
	for i, w := range want {
		if strings.TrimSpace(lines[i]) != w {
			t.Errorf("línea %d: esperaba %q, obtuve %q", i, w, lines[i])
		}
	}
}

// ---------------------------------------------------------------------------
// 7. Ciclo mientras que no ejecuta (condición falsa de inicio)
// ---------------------------------------------------------------------------

func TestMientrasNoEntra(t *testing.T) {
	src := `
programa loop2 ;
vars
    i : entero ;
inicio
{
    i = 0 ;
    mientras ( i > 0 ) haz
    {
        escribe( i ) ;
    } ;
    escribe( 99 ) ;
}
fin
`
	got := correr(t, src)
	if got != "99" {
		t.Errorf("esperaba '99', obtuve %q", got)
	}
}

// ---------------------------------------------------------------------------
// 8. Función nula – escribe su parámetro
// ---------------------------------------------------------------------------

func TestFuncionNula(t *testing.T) {
	src := `
programa fnula ;

nula saludar( n : entero ) {
    {
        escribe( n ) ;
    }
} ;

inicio
{
    saludar( 7 ) ;
}
fin
`
	got := correr(t, src)
	if got != "7" {
		t.Errorf("esperaba '7', obtuve %q", got)
	}
}

// ---------------------------------------------------------------------------
// 9. Función con retorno entero
// ---------------------------------------------------------------------------

func TestFuncionRetornoEntero(t *testing.T) {
	src := `
programa fret ;

entero doble( x : entero ) {
    vars
        r : entero ;
    {
        r = x * 2 ;
    }
    retornar r ;
} ;

inicio
{
    escribe( doble( 6 ) ) ;
}
fin
`
	got := correr(t, src)
	if got != "12" {
		t.Errorf("esperaba '12', obtuve %q", got)
	}
}

// ---------------------------------------------------------------------------
// 10. Retorno usado directamente en expresión aritmética
// ---------------------------------------------------------------------------

func TestFuncionEnExpresion(t *testing.T) {
	src := `
programa fexpr ;
vars
    res : entero ;

entero cuadrado( n : entero ) {
    vars
        r : entero ;
    {
        r = n * n ;
    }
    retornar r ;
} ;

inicio
{
    res = cuadrado( 4 ) + cuadrado( 3 ) ;
    escribe( res ) ;
}
fin
`
	// 16 + 9 = 25
	got := correr(t, src)
	if got != "25" {
		t.Errorf("esperaba '25', obtuve %q", got)
	}
}

// ---------------------------------------------------------------------------
// 11. Múltiples parámetros
// ---------------------------------------------------------------------------

func TestFuncionMultiplesParams(t *testing.T) {
	src := `
programa mparams ;

entero suma( a : entero, b : entero, c : entero ) {
    vars
        r : entero ;
    {
        r = a + b + c ;
    }
    retornar r ;
} ;

inicio
{
    escribe( suma( 1, 2, 3 ) ) ;
}
fin
`
	got := correr(t, src)
	if got != "6" {
		t.Errorf("esperaba '6', obtuve %q", got)
	}
}

// ---------------------------------------------------------------------------
// 12. Recursión – factorial
// ---------------------------------------------------------------------------

func TestFactorial(t *testing.T) {
	src := `
programa fact ;

entero factorial( n : entero ) {
    vars
        r : entero ;
    {
        si ( n <= 1 )
        {
            r = 1 ;
        }
        sino
        {
            r = n * factorial( n - 1 ) ;
        } ;
    }
    retornar r ;
} ;

inicio
{
    escribe( factorial( 5 ) ) ;
}
fin
`
	got := correr(t, src)
	if got != "120" {
		t.Errorf("esperaba '120', obtuve %q", got)
	}
}

// ---------------------------------------------------------------------------
// 13. Recursión – Fibonacci
// ---------------------------------------------------------------------------

func TestFibonacci(t *testing.T) {
	src := `
programa fib ;

entero fib( n : entero ) {
    vars
        r : entero ;
    {
        si ( n <= 1 )
        {
            r = n ;
        }
        sino
        {
            r = fib( n - 1 ) + fib( n - 2 ) ;
        } ;
    }
    retornar r ;
} ;

inicio
{
    escribe( fib( 10 ) ) ;
}
fin
`
	// fib(10) = 55
	got := correr(t, src)
	if got != "55" {
		t.Errorf("esperaba '55', obtuve %q", got)
	}
}

// ---------------------------------------------------------------------------
// 14. Los 6 operadores de comparación
// ---------------------------------------------------------------------------

func TestComparaciones(t *testing.T) {
	src := `
programa comp ;
vars
    a : entero ;
    b : entero ;
inicio
{
    a = 5 ;
    b = 10 ;
    si ( a < b )  { escribe( 1 ) ; } sino { escribe( 0 ) ; } ;
    si ( b > a )  { escribe( 1 ) ; } sino { escribe( 0 ) ; } ;
    si ( a == a ) { escribe( 1 ) ; } sino { escribe( 0 ) ; } ;
    si ( a != b ) { escribe( 1 ) ; } sino { escribe( 0 ) ; } ;
    si ( a <= a ) { escribe( 1 ) ; } sino { escribe( 0 ) ; } ;
    si ( b >= b ) { escribe( 1 ) ; } sino { escribe( 0 ) ; } ;
}
fin
`
	got := correr(t, src)
	lines := strings.Split(got, "\n")
	for i, l := range lines {
		if strings.TrimSpace(l) != "1" {
			t.Errorf("comparación %d falló: esperaba '1', obtuve %q", i+1, l)
		}
	}
}

// ---------------------------------------------------------------------------
// 15. Función nula modifica variable global
// ---------------------------------------------------------------------------

func TestGlobalModificadaPorFuncion(t *testing.T) {
	src := `
programa gmod ;
vars
    contador : entero ;

nula incrementar() {
    {
        contador = contador + 1 ;
    }
} ;

inicio
{
    contador = 0 ;
    incrementar() ;
    incrementar() ;
    incrementar() ;
    escribe( contador ) ;
}
fin
`
	got := correr(t, src)
	if got != "3" {
		t.Errorf("esperaba '3', obtuve %q", got)
	}
}

// ---------------------------------------------------------------------------
// 16. Función que llama a otra función (no recursiva)
// ---------------------------------------------------------------------------

func TestFuncionLlamaFuncion(t *testing.T) {
	src := `
programa fcall ;

entero triple( x : entero ) {
    vars
        r : entero ;
    {
        r = x * 3 ;
    }
    retornar r ;
} ;

entero seis_veces( x : entero ) {
    vars
        r : entero ;
    {
        r = triple( x ) * 2 ;
    }
    retornar r ;
} ;

inicio
{
    escribe( seis_veces( 4 ) ) ;
}
fin
`
	// triple(4)=12, 12*2=24
	got := correr(t, src)
	if got != "24" {
		t.Errorf("esperaba '24', obtuve %q", got)
	}
}

// ---------------------------------------------------------------------------
// 17. Ciclo con condicional anidado dentro
// ---------------------------------------------------------------------------

func TestLoopConSi(t *testing.T) {
	src := `
programa loopsi ;
vars
    i : entero ;
    s : entero ;
inicio
{
    i = 1 ;
    s = 0 ;
    mientras ( i <= 5 ) haz
    {
        si ( i == 3 )
        {
            escribe( i ) ;
        }
        sino
        {
            s = s + i ;
        } ;
        i = i + 1 ;
    } ;
    escribe( s ) ;
}
fin
`
	// i=3 se imprime; resto acumula: 1+2+4+5=12
	got := correr(t, src)
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("esperaba 2 líneas, obtuve %d: %q", len(lines), got)
	}
	if strings.TrimSpace(lines[0]) != "3" {
		t.Errorf("línea 1: esperaba '3', obtuve %q", lines[0])
	}
	if strings.TrimSpace(lines[1]) != "12" {
		t.Errorf("línea 2: esperaba '12', obtuve %q", lines[1])
	}
}

// ---------------------------------------------------------------------------
// 18. Impresión de string literal
// ---------------------------------------------------------------------------

func TestEscribeString(t *testing.T) {
	src := `
programa str1 ;
inicio
{
    escribe( "hola mundo" ) ;
}
fin
`
	got := correr(t, src)
	if got != "\"hola mundo\"" {
		t.Errorf("esperaba \"hola mundo\", obtuve %q", got)
	}
}

// ---------------------------------------------------------------------------
// 19. Suma acumulada con ciclo (1 a 10)
// ---------------------------------------------------------------------------

func TestSumaAcumulada(t *testing.T) {
	src := `
programa sumaloop ;
vars
    i : entero ;
    s : entero ;
    n : entero ;
inicio
{
    n = 10 ;
    s = 0 ;
    i = 1 ;
    mientras ( i <= n ) haz
    {
        s = s + i ;
        i = i + 1 ;
    } ;
    escribe( s ) ;
}
fin
`
	// 1+2+...+10 = 55
	got := correr(t, src)
	if got != "55" {
		t.Errorf("esperaba '55', obtuve %q", got)
	}
}

// ---------------------------------------------------------------------------
// 20. Variables locales aisladas entre funciones distintas
// ---------------------------------------------------------------------------

func TestVariablesLocalesAisladas(t *testing.T) {
	src := `
programa aislar ;

entero f1( x : entero ) {
    vars
        r : entero ;
    {
        r = x + 100 ;
    }
    retornar r ;
} ;

entero f2( x : entero ) {
    vars
        r : entero ;
    {
        r = x + 200 ;
    }
    retornar r ;
} ;

inicio
{
    escribe( f1( 1 ) ) ;
    escribe( f2( 1 ) ) ;
}
fin
`
	got := correr(t, src)
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("esperaba 2 líneas, obtuve %d: %q", len(lines), got)
	}
	if strings.TrimSpace(lines[0]) != "101" {
		t.Errorf("f1: esperaba '101', obtuve %q", lines[0])
	}
	if strings.TrimSpace(lines[1]) != "201" {
		t.Errorf("f2: esperaba '201', obtuve %q", lines[1])
	}
}
