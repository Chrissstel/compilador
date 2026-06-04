package codegen

import (
	"compilador/src/ast"
	"compilador/src/memory"
	"compilador/src/semantic"
	"testing"
)

// helper: crea un generador con un programa ya analizado semánticamente
func setupGenerator(prog *ast.Programa) *Generator {
	mem := memory.NewMemoryManager()
	ana := semantic.NuevoAnalizador(mem)
	ana.AnalizarPrograma(prog)
	tabla := ana.ObtenerTabla()
	gen := NewGenerator(mem, tabla)
	gen.Visit(prog)
	return gen
}

// helper: construye un Valor entero constante
func cteEnt(v int) *ast.Valor {
	return &ast.Valor{EsCte: true, CteEnt: &v}
}

// helper: construye un Valor flotante constante
func cteFlot(v float64) *ast.Valor {
	return &ast.Valor{EsCte: true, CteFlot: &v}
}

// helper: construye una expresión simple sin operador relacional
func exprSimple(izq *ast.Exp) *ast.Expresion {
	return &ast.Expresion{Izq: izq}
}

// helper: construye un Exp de un solo factor valor
func expValor(v *ast.Valor) *ast.Exp {
	return &ast.Exp{Termino: &ast.Termino{Factor: &ast.Factor{Valor: v}}}
}

// helper: construye un Exp de un solo factor ID
func expID(id string) *ast.Exp {
	return &ast.Exp{Termino: &ast.Termino{Factor: &ast.Factor{Valor: &ast.Valor{ID: id}}}}
}

// ─────────────────────────────────────────────────────────────────────────────
// 1. Programa vacío — solo END
// ─────────────────────────────────────────────────────────────────────────────
func TestProgramaVacio(t *testing.T) {
	prog := &ast.Programa{
		ID:     "vacio",
		Cuerpo: &ast.Cuerpo{},
	}
	gen := setupGenerator(prog)

	if len(gen.Quads) != 1 {
		t.Fatalf("esperaba 1 cuádruplo (END), got %d", len(gen.Quads))
	}
	if gen.Quads[0].Op != "END" {
		t.Errorf("esperaba END, got %s", gen.Quads[0].Op)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 2. Asignación de constante entera: x = 5
// ─────────────────────────────────────────────────────────────────────────────
func TestAsignaCteEntera(t *testing.T) {
	prog := &ast.Programa{
		ID: "asigna",
		Vars: &ast.Vars{
			Declaraciones: []*ast.DeclaracionVar{
				{IDs: []string{"x"}, Tipo: "entero"},
			},
		},
		Cuerpo: &ast.Cuerpo{
			Estatutos: []ast.Estatuto{
				&ast.Asigna{
					ID:        "x",
					Expresion: exprSimple(expValor(cteEnt(5))),
				},
			},
		},
	}
	gen := setupGenerator(prog)

	// esperamos: (=  addr_cte  0  addr_x) y (END)
	if len(gen.Quads) != 2 {
		t.Fatalf("esperaba 2 cuádruplos, got %d", len(gen.Quads))
	}
	asigna := gen.Quads[0]
	if asigna.Op != "=" {
		t.Errorf("op esperada '=', got '%s'", asigna.Op)
	}
	if asigna.Right != 0 {
		t.Errorf("Right debe ser 0 en asignación, got %d", asigna.Right)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 3. Expresión aritmética: x = 2 + 3
// ─────────────────────────────────────────────────────────────────────────────
func TestExpresionSuma(t *testing.T) {
	prog := &ast.Programa{
		ID: "suma",
		Vars: &ast.Vars{
			Declaraciones: []*ast.DeclaracionVar{
				{IDs: []string{"x"}, Tipo: "entero"},
			},
		},
		Cuerpo: &ast.Cuerpo{
			Estatutos: []ast.Estatuto{
				&ast.Asigna{
					ID: "x",
					Expresion: exprSimple(&ast.Exp{
						Termino: &ast.Termino{Factor: &ast.Factor{Valor: cteEnt(2)}},
						Op:      "+",
						Der:     expValor(cteEnt(3)),
					}),
				},
			},
		},
	}
	gen := setupGenerator(prog)

	// esperamos: (+  addr2  addr3  temp) y (=  temp  0  addrX) y END
	if len(gen.Quads) != 3 {
		t.Fatalf("esperaba 3 cuádruplos, got %d", len(gen.Quads))
	}
	if gen.Quads[0].Op != "+" {
		t.Errorf("primer quad debe ser '+', got '%s'", gen.Quads[0].Op)
	}
	if gen.Quads[1].Op != "=" {
		t.Errorf("segundo quad debe ser '=', got '%s'", gen.Quads[1].Op)
	}
	// el resultado del + debe ser el Left del =
	if gen.Quads[0].Result != gen.Quads[1].Left {
		t.Errorf("el temp de '+' debe fluir al '=': %d != %d",
			gen.Quads[0].Result, gen.Quads[1].Left)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 4. Cubo semántico: entero + flotante produce flotante
// ─────────────────────────────────────────────────────────────────────────────
func TestSumaMixta(t *testing.T) {
	prog := &ast.Programa{
		ID: "mixta",
		Vars: &ast.Vars{
			Declaraciones: []*ast.DeclaracionVar{
				{IDs: []string{"f"}, Tipo: "flotante"},
			},
		},
		Cuerpo: &ast.Cuerpo{
			Estatutos: []ast.Estatuto{
				&ast.Asigna{
					ID: "f",
					Expresion: exprSimple(&ast.Exp{
						Termino: &ast.Termino{Factor: &ast.Factor{Valor: cteEnt(1)}},
						Op:      "+",
						Der:     expValor(cteFlot(2.5)),
					}),
				},
			},
		},
	}
	gen := setupGenerator(prog)

	// el temp generado por '+' debe estar en rango flotante (10000+)
	sumQuad := gen.Quads[0]
	if sumQuad.Op != "+" {
		t.Fatalf("esperaba '+', got '%s'", sumQuad.Op)
	}
	if sumQuad.Result < 10000 {
		t.Errorf("resultado de suma mixta debe ser dirección flotante (>=10000), got %d", sumQuad.Result)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 5. Condicional sin sino: si (x > 0) { y = 1; }
// ─────────────────────────────────────────────────────────────────────────────
func TestCondicionSinSino(t *testing.T) {
	prog := &ast.Programa{
		ID: "cond",
		Vars: &ast.Vars{
			Declaraciones: []*ast.DeclaracionVar{
				{IDs: []string{"x", "y"}, Tipo: "entero"},
			},
		},
		Cuerpo: &ast.Cuerpo{
			Estatutos: []ast.Estatuto{
				&ast.Condicion{
					Expresion: &ast.Expresion{
						Izq: expID("x"),
						Op:  ">",
						Der: expValor(cteEnt(0)),
					},
					CuerpoSi: &ast.Cuerpo{
						Estatutos: []ast.Estatuto{
							&ast.Asigna{ID: "y", Expresion: exprSimple(expValor(cteEnt(1)))},
						},
					},
				},
			},
		},
	}
	gen := setupGenerator(prog)

	// estructura esperada: (>  x  0  t0) (GOTOF  t0  0  ??) (=  1  0  y) (END)
	ops := extractOps(gen.Quads)
	expected := []string{">", "GOTOF", "=", "END"}
	assertOps(t, ops, expected)

	// el GOTOF debe apuntar al END (índice 3)
	gotof := gen.Quads[1]
	if gotof.Result != 3 {
		t.Errorf("GOTOF debe saltar al índice 3 (END), got %d", gotof.Result)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 6. Condicional con sino
// ─────────────────────────────────────────────────────────────────────────────
func TestCondicionConSino(t *testing.T) {
	prog := &ast.Programa{
		ID: "cond_sino",
		Vars: &ast.Vars{
			Declaraciones: []*ast.DeclaracionVar{
				{IDs: []string{"x", "y"}, Tipo: "entero"},
			},
		},
		Cuerpo: &ast.Cuerpo{
			Estatutos: []ast.Estatuto{
				&ast.Condicion{
					Expresion: &ast.Expresion{
						Izq: expID("x"),
						Op:  ">",
						Der: expValor(cteEnt(0)),
					},
					CuerpoSi: &ast.Cuerpo{
						Estatutos: []ast.Estatuto{
							&ast.Asigna{ID: "y", Expresion: exprSimple(expValor(cteEnt(1)))},
						},
					},
					CuerpoSino: &ast.Cuerpo{
						Estatutos: []ast.Estatuto{
							&ast.Asigna{ID: "y", Expresion: exprSimple(expValor(cteEnt(0)))},
						},
					},
				},
			},
		},
	}
	gen := setupGenerator(prog)

	// (>)(GOTOF)(= 1)(GOTO)(= 0)(END)
	ops := extractOps(gen.Quads)
	expected := []string{">", "GOTOF", "=", "GOTO", "=", "END"}
	assertOps(t, ops, expected)

	// GOTOF salta al inicio del sino (índice 4, la segunda asignación)
	gotof := gen.Quads[1]
	if gotof.Result != 4 {
		t.Errorf("GOTOF debe saltar al índice 4 (sino), got %d", gotof.Result)
	}
	// GOTO salta al END (índice 5)
	goto_ := gen.Quads[3]
	if goto_.Result != 5 {
		t.Errorf("GOTO debe saltar al índice 5 (END), got %d", goto_.Result)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 7. Ciclo mientras
// ─────────────────────────────────────────────────────────────────────────────
func TestCiclo(t *testing.T) {
	prog := &ast.Programa{
		ID: "ciclo",
		Vars: &ast.Vars{
			Declaraciones: []*ast.DeclaracionVar{
				{IDs: []string{"x"}, Tipo: "entero"},
			},
		},
		Cuerpo: &ast.Cuerpo{
			Estatutos: []ast.Estatuto{
				&ast.Ciclo{
					Expresion: &ast.Expresion{
						Izq: expID("x"),
						Op:  ">",
						Der: expValor(cteEnt(0)),
					},
					Cuerpo: &ast.Cuerpo{
						Estatutos: []ast.Estatuto{
							&ast.Asigna{
								ID: "x",
								Expresion: exprSimple(&ast.Exp{
									Termino: &ast.Termino{Factor: &ast.Factor{Valor: &ast.Valor{ID: "x"}}},
									Op:      "-",
									Der:     expValor(cteEnt(1)),
								}),
							},
						},
					},
				},
			},
		},
	}
	gen := setupGenerator(prog)

	// (>)(GOTOF)(- x 1 t)(= t 0 x)(GOTO → 0)(END)
	ops := extractOps(gen.Quads)
	expected := []string{">", "GOTOF", "-", "=", "GOTO", "END"}
	assertOps(t, ops, expected)

	// GOTO debe volver al índice 0 (inicio de la condición)
	gotoQuad := gen.Quads[4]
	if gotoQuad.Result != 0 {
		t.Errorf("GOTO del ciclo debe volver a 0, got %d", gotoQuad.Result)
	}
	// GOTOF debe saltar al END (índice 5)
	gotof := gen.Quads[1]
	if gotof.Result != 5 {
		t.Errorf("GOTOF debe saltar al END (5), got %d", gotof.Result)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 8. Imprime con letrero y expresión
// ─────────────────────────────────────────────────────────────────────────────
func TestImprime(t *testing.T) {
	prog := &ast.Programa{
		ID: "imprime",
		Vars: &ast.Vars{
			Declaraciones: []*ast.DeclaracionVar{
				{IDs: []string{"x"}, Tipo: "entero"},
			},
		},
		Cuerpo: &ast.Cuerpo{
			Estatutos: []ast.Estatuto{
				&ast.Imprime{
					Items: []ast.ImprimeItem{
						{EsLetrero: true, Letrero: `"hola"`},
						{EsLetrero: false, Expr: exprSimple(expID("x"))},
					},
				},
			},
		},
	}
	gen := setupGenerator(prog)

	ops := extractOps(gen.Quads)
	expected := []string{"PRINT_STR", "PRINT", "END"}
	assertOps(t, ops, expected)
}

// ─────────────────────────────────────────────────────────────────────────────
// 9. Función con llamada
// ─────────────────────────────────────────────────────────────────────────────
func TestFuncionYLlamada(t *testing.T) {
	prog := &ast.Programa{
		ID: "confunc",
		Funcs: []*ast.Func{
			{
				TipoRetorno: "nula",
				ID:          "duplica",
				Params:      []*ast.Param{{ID: "n", Tipo: "entero"}},
				Cuerpo: &ast.Cuerpo{
					Estatutos: []ast.Estatuto{
						&ast.Imprime{
							Items: []ast.ImprimeItem{
								{EsLetrero: false, Expr: exprSimple(expID("n"))},
							},
						},
					},
				},
			},
		},
		Cuerpo: &ast.Cuerpo{
			Estatutos: []ast.Estatuto{
				&ast.Llamada{
					ID:   "duplica",
					Args: []*ast.Expresion{exprSimple(expValor(cteEnt(7)))},
				},
			},
		},
	}
	gen := setupGenerator(prog)

	ops := extractOps(gen.Quads)
	// GOTO(saltar al main) PRINT ENDFUNC PARAM GOSUB END
	expected := []string{"GOTO", "PRINT", "ENDFUNC", "PARAM", "GOSUB", "END"}
	assertOps(t, ops, expected)

	// GOTO debe apuntar al índice del PARAM (3)
	if gen.Quads[0].Result != 3 {
		t.Errorf("GOTO inicial debe apuntar al main (3), got %d", gen.Quads[0].Result)
	}
	// GOSUB debe apuntar al inicio de la función (1)
	gosub := gen.Quads[4]
	if gosub.Result != 1 {
		t.Errorf("GOSUB debe apuntar al inicio de la función (1), got %d", gosub.Result)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// helpers internos
// ─────────────────────────────────────────────────────────────────────────────

func extractOps(quads []Quadruple) []string {
	ops := make([]string, len(quads))
	for i, q := range quads {
		ops[i] = q.Op
	}
	return ops
}

func assertOps(t *testing.T, got, expected []string) {
	t.Helper()
	if len(got) != len(expected) {
		t.Fatalf("cantidad de cuádruplos: esperaba %v, got %v", expected, got)
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("quad[%d]: esperaba '%s', got '%s'", i, expected[i], got[i])
		}
	}
}
