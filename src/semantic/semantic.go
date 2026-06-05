package semantic

//más que semantic.go es visitor.go

import (
	"compilador/src/ast"
	"compilador/src/memory"
	"fmt"
)

//Analizador

type Analizador struct {
	tabla *TablaSimbolos

	errores []string
	Mem     *memory.MemoryManager
}

func NuevoAnalizador(mem *memory.MemoryManager) *Analizador {
	return &Analizador{
		tabla:   NuevaTabla(),
		errores: make([]string, 0),
		Mem:     mem,
	}
}

//Entrada principal

func (a *Analizador) AnalizarPrograma(prog *ast.Programa) {
	//crear el scope global, añadimos la función global
	a.tabla.funciones["global"] = &FuncEnDir{
		Nombre:      "global",
		TipoRetorno: "nula",
		Variables:   nil,
		Recursos:    0,
	}

	//registrar variables globales
	if prog.DeclsVars != nil {
		a.analizarVars(prog.DeclsVars, "global")
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
	a.analizarCuerpo(prog.Cuerpo, "global")
}

//VARIABLES

func (a *Analizador) analizarVars(DeclsVars *ast.DeclsVars, scope string) {
	for _, decl := range DeclsVars.Decls {
		tipo := decl.Tipo
		for _, id := range decl.IDs {
			//pide dirección de memoria

			//intenta agregar la variable a la tabla y si hay un error (como que ya existe) lo registra
			if err := a.tabla.AgregarVar(scope, id, tipo); err != nil {
				a.errores = append(a.errores, err.Error())
			}
		}
	}
}

//FUNCIONES

// solo mete la función a la tabla sin analizar su cuerpo
func (a *Analizador) registrarFirmaFunc(f *ast.Func) {
	params := make([]VarEnDir, len(f.Params))
	for i, p := range f.Params {
		//falta asignar dir de memoria

		params[i] = VarEnDir{
			Nombre: p.ID,
			Tipo:   p.Tipo,
			//falta la dirección de memoria
		}
	}
	if err := a.tabla.RegistrarFunc(f.ID, f.TipoRetorno, params); err != nil {
		a.errores = append(a.errores)
	}
}

func (a *Analizador) analizarFunc(f *ast.Func) {
	//analizar parámetros ya se hizo en registrarFirmaFunc
	//analizar vars locales
	if f.Vars != nil {
		a.analizarVars(f.Vars, f.ID) //el segundo es el scope
	}

	//analizar el cuerpo
	a.analizarCuerpo(f.Cuerpo, f.ID)

	//validar lo del retorno
	if f.TipoRetorno == "nulo" {
		// función nula NO debe tener retornar
		if f.Retorno != nil {
			a.errores = append(a.errores, fmt.Sprintf(
				"función '%s' es de tipo 'nulo' y no debe retornar un valor",
				f.ID,
			))
		}
	} else {
		// función no nula SÍ debe tener retornar
		if f.Retorno == nil {
			a.errores = append(a.errores, fmt.Sprintf(
				"función '%s' debe retornar un valor de tipo '%s'",
				f.ID, f.TipoRetorno,
			))
		} else {
			// verifica que el id exista en el scope actual, o en el global
			varEntry, existe := a.tabla.BuscarVar(f.Retorno.ID, f.ID)
			if !existe {
				a.errores = append(a.errores, fmt.Sprintf(
					"función '%s': variable '%s' en retornar no fue declarada",
					f.ID, f.Retorno.ID,
				))

			} else if string(varEntry.Tipo) != f.TipoRetorno {
				// verifica que el tipo coincida
				a.errores = append(a.errores, fmt.Sprintf(
					"función '%s': retorna '%s' de tipo '%s' pero se esperaba '%s'",
					f.ID, f.Retorno.ID, varEntry.Tipo, f.TipoRetorno,
				))
			}
		}
	}

}

// CUERPO Y ESTATUTOS

// el cuerpo puede estar dentro de una función o en el programa principal, por eso se le pasa el scope
func (a *Analizador) analizarCuerpo(c *ast.Cuerpo, scope string) {
	for _, e := range c.Estatutos {
		a.analizarEstatuto(e, scope)
	}
}

func (a *Analizador) analizarEstatuto(e ast.Estatuto, scope string) { //estatuto es una interfaz
	switch tree := e.(type) {
	case *ast.Asigna:
		a.analizarAsigna(tree, scope)
	case *ast.Condicion:
		a.analizarCondicion(tree, scope)
	case *ast.Ciclo:
		a.analizarCiclo(tree, scope)
	case *ast.Llamada:
		a.analizarLlamada(tree, scope)
	case *ast.Imprime:
		a.analizarImprime(tree, scope)
	case *ast.BloqueEstatutos:
		for _, inner := range tree.Estatutos {
			a.analizarEstatuto(inner, scope)
		}
	}
}

// CADA TIPO DE ESTATUTO
func (a *Analizador) analizarAsigna(n *ast.Asigna, scope string) {
	//la variable debe estar declarada
	if _, existe := a.tabla.BuscarVar(n.ID, scope); !existe {
		a.errores = append(a.errores, fmt.Sprintf("variable '%s' no fue declarada", n.ID))
	}

	//checar que el tipo de la expresión coincida con el tipo de la variable
	varEntry, _ := a.tabla.BuscarVar(n.ID, scope)
	expType := a.analizarExpresion(n.Expresion)
	if varEntry.Tipo != expType {
		a.errores = append(a.errores, fmt.Sprintf("tipo de la expresión no coincide con el de la variable '%s'", n.ID))
	}
}

func (a *Analizador) analizarCondicion(n *ast.Condicion, scope string) {
	a.analizarExpresion(n.Expresion)
	a.analizarCuerpo(n.CuerpoSi, scope)
	if n.CuerpoSino != nil {
		a.analizarCuerpo(n.CuerpoSino, scope)
	}
}

func (a *Analizador) analizarCiclo(n *ast.Ciclo, scope string) {
	a.analizarExpresion(n.Expresion)
	a.analizarCuerpo(n.Cuerpo)
}

func (a *Analizador) analizarLlamada(n *ast.Llamada, scope string) {
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

func (a *Analizador) analizarImprime(n *ast.Imprime, scope string) {
	for _, item := range n.Items {
		if !item.EsLetrero {
			a.analizarExpresion(item.Expr)
		}
		//si sí es letrero no hay nada que verificar
	}
}

//EXPRESIONES

// va a regresar el tipo de la expresión
func (a *Analizador) analizarExpresion(e *ast.Expresion) string {
	a.analizarExp(e.Izq)
	if e.Der != nil {
		a.analizarExp(e.Der)
	}
	return ""
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

func (a *Analizador) ObtenerTabla() *TablaSimbolos {
	return a.tabla
}
