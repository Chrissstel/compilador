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
		Variables:   make(map[string]VarEnDir),
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
	expType := a.analizarExpresion(n.Expresion, scope)
	if varEntry.Tipo != expType {
		a.errores = append(a.errores, fmt.Sprintf("tipo de la expresión no coincide con el de la variable '%s'", n.ID))
	}
}

// la expresion tiene que ser de tipo booleana
// pero aquí el 0 es falso y cualquier otro int es verdadero
// tal vez tambien pueda ser flotante
func (a *Analizador) analizarCondicion(n *ast.Condicion, scope string) {
	tipo := a.analizarExpresion(n.Expresion, scope)
	if tipo != "entero" && tipo != "flotante" {
		a.errores = append(a.errores, fmt.Sprintf("la expresión en la condición debe ser de tipo 'booleano'"))
	}
	a.analizarCuerpo(n.CuerpoSi, scope)
	if n.CuerpoSino != nil {
		a.analizarCuerpo(n.CuerpoSino, scope)
	}
}

func (a *Analizador) analizarCiclo(n *ast.Ciclo, scope string) {
	//checar que la expresión sea de tipo "booleana"
	tipo := a.analizarExpresion(n.Expresion, scope)
	if tipo != "entero" && tipo != "flotante" {
		a.errores = append(a.errores, fmt.Sprintf("la expresión en el ciclo debe ser de tipo 'booleano'"))
	}
	a.analizarCuerpo(n.Cuerpo, scope)
}

// tiene que regresar un tipo, aunque sea nula
func (a *Analizador) analizarLlamada(n *ast.Llamada, scope string) string {
	//la función debe estar declarada
	entrada, existe := a.tabla.BuscarFunc(n.ID)
	if !existe {
		a.errores = append(a.errores, fmt.Sprintf("función '%s' no fue declarada", n.ID))
		return "nulo"
	}

	//el número de argumentos debe coincidir
	if len(n.Args) != len(entrada.Params) {
		a.errores = append(a.errores, fmt.Sprintf(
			"función '%s' espera %d argumento(s) pero recibió %d",
			n.ID, len(entrada.Params), len(n.Args),
		))
	}

	//analizar cada argumento y validar que coincida con su parámetro
	for i, arg := range n.Args {
		argType := a.analizarExpresion(arg, scope)
		paramName := entrada.Params[i]

		//checamos que el tipo del argumento coincida con el tipo del parámetro
		if argType != entrada.Variables[paramName].Tipo {
			a.errores = append(a.errores, fmt.Sprintf(
				"en la llamada a '%s', el argumento %d es de tipo '%s' pero se esperaba '%s'",
				n.ID, i+1, argType, entrada.Variables[paramName].Tipo,
			))
		}
	}
	return entrada.TipoRetorno
}

func (a *Analizador) analizarImprime(n *ast.Imprime, scope string) {
	for _, item := range n.Items {
		if !item.EsLetrero {
			//checar que no sea de tipo nula la expresión
			expType := a.analizarExpresion(item.Expr, scope)
			if expType == "nulo" {
				a.errores = append(a.errores, fmt.Sprintf("no se puede imprimir una expresión de tipo 'nulo'"))
			}
		}
		//si sí es letrero no hay nada que verificar
	}
}

//EXPRESIONES

// va a regresar el tipo de la expresión
func (a *Analizador) analizarExpresion(e *ast.Expresion, scope string) string {
	tipoIzq := a.analizarExp(e.Izq, scope)
	if e.Der != nil {
		tipoDer := a.analizarExp(e.Der, scope)
		//checar que los tipos sean compatibles con el operador
		//CHECAR MAS DETALLADAMENTE CON EL CUBO SEMANTICO
		if tipoIzq != tipoDer {
			a.errores = append(a.errores, fmt.Sprintf("tipos incompatibles en la expresión: '%s' vs '%s'", tipoIzq, tipoDer))
		}

	}
	return tipoIzq
}

func (a *Analizador) analizarExp(e *ast.Exp, scope string) string {
	tipoIzq := a.analizarTermino(e.Termino, scope)
	if e.Der != nil {
		tipoDer := a.analizarExp(e.Der, scope)
		if tipoIzq != tipoDer {
			a.errores = append(a.errores, fmt.Sprintf("tipos incompatibles en la expresión: '%s' vs '%s'", tipoIzq, tipoDer))
		}
	}
	return tipoIzq
}

func (a *Analizador) analizarTermino(t *ast.Termino, scope string) string {
	tipo := a.analizarFactor(t.Factor, scope)
	if t.Der != nil {
		tipoDer := a.analizarTermino(t.Der, scope)
		//CHECAR MAS DETALLADAMENTE CON EL CUBO SEMANTICO
		if tipo != tipoDer {
			a.errores = append(a.errores, fmt.Sprintf("tipos incompatibles en la expresión: '%s' vs '%s'", tipo, tipoDer))
		}
	}
	return tipo
}

func (a *Analizador) analizarFactor(f *ast.Factor, scope string) string {
	var tipo string
	switch {
	case f.Expr != nil:
		tipo = a.analizarExpresion(f.Expr, scope)
	case f.Llamada != nil:
		tipo = a.analizarLlamada(f.Llamada, scope)

	case f.Valor != nil:
		tipo = a.analizarValor(f.Valor, scope)
	}
	return tipo
}

func (a *Analizador) analizarValor(v *ast.Valor, scope string) string {
	var tipo string
	//si es un ID debe estar declarado
	if !v.EsCte {
		if varEntry, existe := a.tabla.BuscarVar(v.ID, scope); !existe {
			a.errores = append(a.errores, fmt.Sprintf("variable '%s' no fue declarada", v.ID))
		} else {
			tipo = varEntry.Tipo
		}
	} else {
		//si es constante, el tipo depende de si es int o float
		if v.CteEnt != nil {
			tipo = "entero"
		} else {
			tipo = "flotante"
		}
	}

	return tipo
}

//PARA IMPRIMIR ERRORES Y TABLA

// checa si hay errores registrados
func (a *Analizador) HayErrores() bool {
	return len(a.errores) > 0
}

func (a *Analizador) ImprimirErrores() {
	fmt.Println("\n=== Errores semánticos ===")
	for _, err := range a.errores {
		fmt.Println("- " + err)
	}
}

func (a *Analizador) ImprimirTabla() {
	fmt.Println("\n=== Tabla de símbolos ===")
	for funcName, funcEntry := range a.tabla.funciones {
		fmt.Printf("Función '%s' (retorna '%s'):\n", funcName, funcEntry.TipoRetorno)
		if len(funcEntry.Params) > 0 {
			fmt.Println("  Parámetros:")
			for _, param := range funcEntry.Params {
				varEntry := funcEntry.Variables[param]
				fmt.Printf("    - %s: %s\n", varEntry.Nombre, varEntry.Tipo)
			}
		} else {
			fmt.Println("  Sin parámetros")
		}
		if len(funcEntry.Variables) > 0 {
			fmt.Println("  Variables locales:")
			for varName, varEntry := range funcEntry.Variables {
				fmt.Printf("    - %s: %s\n", varName, varEntry.Tipo)
			}
		} else {
			fmt.Println("  Sin variables locales")
		}
		fmt.Printf("  Recursos usados: %d\n", funcEntry.Recursos)
		fmt.Println()
	}
}
