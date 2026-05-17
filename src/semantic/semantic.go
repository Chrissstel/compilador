package semantic

import (
	"compilador/src/ast"
	"fmt"
)

//Analizador

type Analizador struct {
	tabla   *TablaSimbolos
	errores []string
}

func NuevoAnalizador() *Analizador {
	return &Analizador{
		tabla:   NuevaTabla(),
		errores: make([]string, 0),
	}
}

// agregar un error sin tener que detener el análisis
func (a *Analizador) error(msg string) {
	a.errores = append(a.errores, "SemanticError: "+msg)
}

func (a *Analizador) HayErrores() bool {
	return len(a.errores) > 0
}

func (a *Analizador) ImprimirErrores() {
	for _, e := range a.errores {
		fmt.Println(e)
	}
}

//Entrada principal

func (a *Analizador) AnalizarPrograma(prog *ast.Programa) {
	//registrar variables globales
	if prog.Vars != nil {
		a.analizarVars(prog.Vars)
	}

	//registrar funciones, pero sin el cuerpo para que una función pueda llamar a otra que se declara después
	for _, f := range prog.Funcs {
		a.registrarFirmaFunc(f)
	}

	//ahora sí se analiza el cuerpo de cada función
	for _, f := range prog.Funcs {
		a.analizarFunc(f)
	}

	//analizar el cuerpo principal
	a.analizarCuerpo(prog.Cuerpo)
}

//VARIABLES

func (a *Analizador) analizarVars(vars *ast.Vars) {
	for _, decl := range vars.Declaraciones {
		tipo := TipoDato(decl.Tipo)
		for _, id := range decl.IDs {
			if err := a.tabla.AgregarVar(id, tipo); err != nil {
				a.error(err.Error())
			}
		}
	}
}

//FUNCIONES

// solo mete la función a la tabla sin analizar su cuerpo
func (a *Analizador) registrarFirmaFunc(f *ast.Func) {
	params := make([]EntradaVar, len(f.Params))
	for i, p := range f.Params {
		params[i] = EntradaVar{Nombre: p.ID, Tipo: TipoDato(p.Tipo)}
	}
	if err := a.tabla.AgregarFunc(f.ID, TipoDato(f.TipoRetorno), params); err != nil {
		a.error(err.Error())
	}
}

func (a *Analizador) analizarFunc(f *ast.Func) {
	//obtener los params registrados en la tabla
	entrada, _ := a.tabla.BuscarFunc(f.ID)

	//entrar al scope local con los parámetros ya dentro
	snapshot := a.tabla.EntrarScope(f.ID, entrada.Parametros)

	//agregar vars locales
	if f.Vars != nil {
		a.analizarVars(f.Vars)
	}

	//analizar el cuerpo
	a.analizarCuerpo(f.Cuerpo)

	//salir del scope
	a.tabla.SalirScope(snapshot)
}

// CUERPO Y ESTATUTOS
func (a *Analizador) analizarCuerpo(c *ast.Cuerpo) {
	for _, e := range c.Estatutos {
		a.analizarEstatuto(e)
	}
}

func (a *Analizador) analizarEstatuto(e ast.Estatuto) { //estatuto es una interfaz
	switch n := e.(type) {
	case *ast.Asigna:
		a.analizarAsigna(n)
	case *ast.Condicion:
		a.analizarCondicion(n)
	case *ast.Ciclo:
		a.analizarCiclo(n)
	case *ast.Llamada:
		a.analizarLlamada(n)
	case *ast.Imprime:
		a.analizarImprime(n)
	case *ast.BloqueEstatutos:
		for _, inner := range n.Estatutos {
			a.analizarEstatuto(inner)
		}
	}
}

// CADA TIPO DE ESTATUTO
func (a *Analizador) analizarAsigna(n *ast.Asigna) {
	//la variable debe estar declarada
	if _, existe := a.tabla.BuscarVar(n.ID); !existe {
		a.error(fmt.Sprintf("variable '%s' no fue declarada", n.ID))
	}
	a.analizarExpresion(n.Expresion)
}

func (a *Analizador) analizarCondicion(n *ast.Condicion) {
	a.analizarExpresion(n.Expresion)
	a.analizarCuerpo(n.CuerpoSi)
	if n.CuerpoSino != nil {
		a.analizarCuerpo(n.CuerpoSino)
	}
}

func (a *Analizador) analizarCiclo(n *ast.Ciclo) {
	a.analizarExpresion(n.Expresion)
	a.analizarCuerpo(n.Cuerpo)
}

func (a *Analizador) analizarLlamada(n *ast.Llamada) {
	//la función debe estar declarada
	entrada, existe := a.tabla.BuscarFunc(n.ID)
	if !existe {
		a.error(fmt.Sprintf("función '%s' no fue declarada", n.ID))
		return
	}

	//el número de argumentos debe coincidir
	if len(n.Args) != len(entrada.Parametros) {
		a.error(fmt.Sprintf(
			"función '%s' espera %d argumento(s) pero recibió %d",
			n.ID, len(entrada.Parametros), len(n.Args),
		))
	}

	//analizar cada argumento
	for _, arg := range n.Args {
		a.analizarExpresion(arg)
	}
}

func (a *Analizador) analizarImprime(n *ast.Imprime) {
	for _, item := range n.Items {
		if !item.EsLetrero {
			a.analizarExpresion(item.Expr)
		}
		//si sí es letrero no hay nada que verificar
	}
}

//EXPRESIONES

func (a *Analizador) analizarExpresion(e *ast.Expresion) {
	a.analizarExp(e.Izq)
	if e.Der != nil {
		a.analizarExp(e.Der)
	}
}

func (a *Analizador) analizarExp(e *ast.Exp) {
	a.analizarTermino(e.Termino)
	if e.Der != nil {
		a.analizarExp(e.Der)
	}
}

func (a *Analizador) analizarTermino(t *ast.Termino) {
	a.analizarFactor(t.Factor)
	if t.Der != nil {
		a.analizarTermino(t.Der)
	}
}

func (a *Analizador) analizarFactor(f *ast.Factor) {
	switch {
	case f.Expr != nil:
		a.analizarExpresion(f.Expr)

	case f.Llamada != nil:
		a.analizarLlamada(f.Llamada)

	case f.Valor != nil:
		a.analizarValor(f.Valor)
	}
}

func (a *Analizador) analizarValor(v *ast.Valor) {
	//si es un ID debe estar declarado
	if !v.EsCte {
		if _, existe := a.tabla.BuscarVar(v.ID); !existe {
			a.error(fmt.Sprintf("variable '%s' no fue declarada", v.ID))
		}
	}
	//si es constante no hay nada que verificar
}
