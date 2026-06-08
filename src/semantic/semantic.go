package semantic

//más que semantic.go es visitor.go

import (
	"compilador/src/ast"
	"compilador/src/codegen"
	"compilador/src/cube"
	"compilador/src/memory"
	"fmt"
)

//Analizador

type Analizador struct {
	Tabla   *TablaSimbolos
	errores []string
	Mem     *memory.MemoryManager
	//Se añadió el generador de cuádruplos
	Generador *codegen.QuadGenerator
}

func NuevoAnalizador(mem *memory.MemoryManager) *Analizador {
	return &Analizador{
		Tabla:     NuevaTabla(),
		errores:   make([]string, 0),
		Mem:       mem,
		Generador: codegen.NewGenerator(mem),
	}
}

//Entrada principal

func (a *Analizador) AnalizarPrograma(prog *ast.Programa) {
	//crear el scope global, añadimos la función global
	a.Tabla.funciones["global"] = &FuncEnDir{
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

	// generar el GoTo para saltar funciones
	if len(prog.Funcs) > 0 {
		a.Generador.EmitQuad("GOTO", 0, 0, 0)
		a.Generador.Jumps.Push(len(a.Generador.Quads) - 1)
	}

	//ahora sí se analiza el cuerpo de cada función
	for _, f := range prog.Funcs {
		a.analizarFunc(f)
	}

	//rellena el GoTo hacia main
	if len(prog.Funcs) > 0 {
		mainJump := a.Generador.Jumps.Pop()
		a.Generador.FillJump(mainJump, len(a.Generador.Quads))
	}

	//analizar el cuerpo principal
	a.analizarCuerpo(prog.Cuerpo, "global")

	//el último quad
	a.Generador.EmitQuad("END", 0, 0, 0)
}

//VARIABLES

func (a *Analizador) analizarVars(DeclsVars *ast.DeclsVars, scope string) {
	for _, decl := range DeclsVars.Decls {
		tipo := decl.Tipo
		for _, id := range decl.IDs {
			//pide dirección de memoria
			var dir int
			if scope == "global" {
				dir = a.Mem.GetDirGlobal(tipo)
			} else {
				dir = a.Mem.GetDirLocal(tipo)
			}

			//intenta agregar la variable a la tabla y si hay un error (como que ya existe) lo registra
			if err := a.Tabla.AgregarVar(scope, id, tipo, dir); err != nil {
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
			Direccion: a.Mem.GetDirLocal(p.Tipo), //asigna memoria para el parámetro
		}
	}
	if err := a.Tabla.RegistrarFunc(f.ID, f.TipoRetorno, params); err != nil {
		a.errores = append(a.errores, err.Error())
	}
}

func (a *Analizador) analizarFunc(f *ast.Func) {
	//registra el startQuad de la función
	a.Tabla.funciones[f.ID].StartQuad = len(a.Generador.Quads)

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
			varEntry, existe := a.Tabla.BuscarVar(f.Retorno.ID, f.ID)
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
			} else { //solo si todo esta bien
				//Gen de Quad - emitir el cuádruplo de retorno
				a.Generador.EmitQuad("RETURN", varEntry.Direccion, 0, 0)
			}
		}
	}
	a.Generador.EmitQuad("ENDFUNC", 0, 0, 0)

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
	//SE AÑADE EL DOWHILE
	case *ast.DoWhile:
		a.analizarDoWhile(tree, scope)
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
	varEntry, existe := a.Tabla.BuscarVar(n.ID, scope)
	if !existe {
		a.errores = append(a.errores, fmt.Sprintf("variable '%s' no fue declarada", n.ID))
		a.analizarExpresion(n.Expresion, scope) // igual hay que analizarla para no romper la pila
		return                                  // no emitir cuádruplo con dir inválida
	}

	//checar que el tipo de la expresión coincida con el tipo de la variable
	expType := a.analizarExpresion(n.Expresion, scope)
	if varEntry.Tipo != expType {
		a.errores = append(a.errores, fmt.Sprintf("tipo de la expresión no coincide con el de la variable '%s'", n.ID))
	}

	//Gen de Quad
	izq := a.Generador.Operands.Pop() //la expresión ya dejó su resultado en la pila de operandos
	a.Generador.EmitQuad("=", izq.Address, 0, varEntry.Direccion)
}

// la expresion tiene que ser de tipo booleana
// pero aquí el 0 es falso y cualquier otro int es verdadero
// tal vez tambien pueda ser flotante
func (a *Analizador) analizarCondicion(n *ast.Condicion, scope string) {
	tipo := a.analizarExpresion(n.Expresion, scope)
	if tipo != "entero" && tipo != "flotante" {
		a.errores = append(a.errores, "la expresión en la condición debe ser de tipo 'booleano'")
	}

	//Gen de Quad - salto condicional
	cond := a.Generador.Operands.Pop()
	a.Generador.EmitQuad("GOTOF", cond.Address, 0, 0)
	a.Generador.Jumps.Push(len(a.Generador.Quads) - 1)

	a.analizarCuerpo(n.CuerpoSi, scope)
	if n.CuerpoSino != nil {
		//Gen de Quad - salto incondicional para saltar el sino
		a.Generador.EmitQuad("GOTO", 0, 0, 0)
		falseJump := a.Generador.Jumps.Pop()
		a.Generador.FillJump(falseJump, len(a.Generador.Quads))
		a.Generador.Jumps.Push(len(a.Generador.Quads) - 1) // índice del GOTO

		a.analizarCuerpo(n.CuerpoSino, scope)

		endJump := a.Generador.Jumps.Pop()
		// para rellenar el GOTO que salta el sino
		a.Generador.FillJump(endJump, len(a.Generador.Quads))
	} else {
		falseJump := a.Generador.Jumps.Pop()
		// para rellenar el GOTOF que salta el sino (o el final del if si no hay sino)
		a.Generador.FillJump(falseJump, len(a.Generador.Quads))
	}
}

func (a *Analizador) analizarCiclo(n *ast.Ciclo, scope string) {
	loopStart := len(a.Generador.Quads)
	a.Generador.Jumps.Push(loopStart) //guarda el inicio del ciclo para poder saltar de vuelta al final

	//checar que la expresión sea de tipo "booleana"
	tipo := a.analizarExpresion(n.Expresion, scope)
	if tipo != "entero" && tipo != "flotante" {
		a.errores = append(a.errores, "la expresión en el ciclo debe ser de tipo 'booleano'")
	}

	cond := a.Generador.Operands.Pop()
	a.Generador.EmitQuad("GOTOF", cond.Address, 0, 0)
	a.Generador.Jumps.Push(len(a.Generador.Quads) - 1) // índice del GOTOF

	a.analizarCuerpo(n.Cuerpo, scope)

	exitJump := a.Generador.Jumps.Pop() // índice del GOTOF
	loopJump := a.Generador.Jumps.Pop() // índice de inicio
	a.Generador.EmitQuad("GOTO", 0, 0, loopJump)
	a.Generador.FillJump(exitJump, len(a.Generador.Quads))
}

func (a *Analizador) analizarDoWhile(n *ast.DoWhile, scope string) {
	loopStart := len(a.Generador.Quads)
	//guarda el inicio del ciclo para poder regresar
	a.Generador.Jumps.Push(loopStart)

	a.analizarCuerpo(n.Cuerpo, scope)

	//checar que la expresión sea de tipo "booleana"
	tipo := a.analizarExpresion(n.Expresion, scope)
	if tipo != "entero" && tipo != "flotante" {
		a.errores = append(a.errores, "la expresión en el ciclo debe ser de tipo 'booleano'")
	}

	//es donde se guardó el resultado de la condicion
	cond := a.Generador.Operands.Pop()
	//es el inicio del haz
	loopJump := a.Generador.Jumps.Pop()
	a.Generador.EmitQuad("GOTOT", cond.Address, 0, loopJump)

}

// tiene que regresar un tipo, aunque sea nula
func (a *Analizador) analizarLlamada(n *ast.Llamada, scope string) string {
	//la función debe estar declarada
	funcion, existe := a.Tabla.BuscarFunc(n.ID)
	if !existe {
		a.errores = append(a.errores, fmt.Sprintf("función '%s' no fue declarada", n.ID))
		return "nulo"
	}

	//el número de argumentos debe coincidir
	if len(n.Args) != len(funcion.Params) {
		a.errores = append(a.errores, fmt.Sprintf(
			"función '%s' espera %d argumento(s) pero recibió %d",
			n.ID, len(funcion.Params), len(n.Args),
		))
	}

	//ERA - reservamos espacio para la función
	rec := funcion.Recursos
	a.Generador.EmitQuad("ERA", rec, 0, 0) //puse directamente la cantidad de recursos

	//analizar cada argumento y validar que coincida con su parámetro
	for i, arg := range n.Args {
		argType := a.analizarExpresion(arg, scope)

		// Si hay más argumentos que parámetros, saltamos la validación del tipo
		if i < len(funcion.Params) {
			paramName := funcion.Params[i]
			//checamos que el tipo del argumento coincida con el tipo del parámetro
			if argType != funcion.Variables[paramName].Tipo {
				a.errores = append(a.errores, fmt.Sprintf(
					"en la llamada a '%s', el argumento %d es de tipo '%s' pero se esperaba '%s'",
					n.ID, i+1, argType, funcion.Variables[paramName].Tipo,
				))
			}
		}
		param := a.Generador.Operands.Pop()

		// Obtener la dirección del parámetro
		var paramDir int
		if i < len(funcion.Params) {
			paramName := funcion.Params[i]
			paramDir = funcion.Variables[paramName].Direccion
		}

		a.Generador.EmitQuad("PARAM", param.Address, 0, paramDir)
	}

	inicio := funcion.StartQuad
	a.Generador.EmitQuad("GOSUB", 0, 0, inicio)

	//aqui ya no me acordaba si usé nulo o nula
	if funcion.TipoRetorno != "nula" && funcion.TipoRetorno != "nulo" {
		// Usar dirección global en lugar de temporal, porque el valor de retorno
		// se necesita guardar después de ENDFUNC cuando el stack está vacío
		temp := a.Mem.GetDirTemp(funcion.TipoRetorno)
		a.Generador.EmitQuad("GETRETURN", 0, 0, temp)
		a.Generador.Operands.Push(codegen.Operand{Address: temp, Type: funcion.TipoRetorno})
	}

	return funcion.TipoRetorno
}

func (a *Analizador) analizarImprime(n *ast.Imprime, scope string) {
	for _, item := range n.Items {
		if item.EsLetrero {
			addr := a.Mem.GetDirConstString(item.Letrero)
			a.Generador.EmitQuad("PRINT_STR", addr, 0, 0)
		} else {
			//checar que no sea de tipo nula la expresión
			expType := a.analizarExpresion(item.Expr, scope)
			if expType == "nulo" {
				a.errores = append(a.errores, "no se puede imprimir una expresión de tipo 'nulo'")
			}
			//analizarExpresion ya dejó el resultado en la pila
			val := a.Generador.Operands.Pop()
			a.Generador.EmitQuad("PRINT", val.Address, 0, 0)
		}

	}
}

//EXPRESIONES

// va a regresar el tipo de la expresión
func (a *Analizador) analizarExpresion(e *ast.Expresion, scope string) string {
	tipoIzq := a.analizarExp(e.Izq, scope)
	if e.Der != nil {
		a.Generador.Operators.Push(e.Op)

		tipoDer := a.analizarExp(e.Der, scope)
		//checar que los tipos sean compatibles con el operador
		tipo, err := cube.SemanticCube(tipoIzq, tipoDer, e.Op)
		if err != nil {
			a.errores = append(a.errores, err.Error())
		}
		//Gen de Quad - resolver la operación
		a.Generador.ResolverOp()

		return tipo
	}
	return tipoIzq
}

func (a *Analizador) analizarExp(e *ast.Exp, scope string) string {
	tipoIzq := a.analizarTermino(e.Termino, scope)
	if e.Der != nil {
		a.Generador.Operators.Push(e.Op)

		tipoDer := a.analizarExp(e.Der, scope)
		//CHECAR DETALLADAMENTE CON EL CUBO SEMANTICO
		tipo, err := cube.SemanticCube(tipoIzq, tipoDer, e.Op)
		if err != nil {
			a.errores = append(a.errores, err.Error())
		}
		//Gen de Quad - resolver la operación
		a.Generador.ResolverOp()

		return tipo
	}
	return tipoIzq
}

func (a *Analizador) analizarTermino(t *ast.Termino, scope string) string {
	tipoIzq := a.analizarFactor(t.Factor, scope)

	if t.Op != "" { //si hay operador, hay lado derecho
		//Gen de Quad - push el operator
		a.Generador.Operators.Push(t.Op)

		tipoDer := a.analizarTermino(t.Der, scope)
		//CHECAR MAS DETALLADAMENTE CON EL CUBO SEMANTICO
		tipo, err := cube.SemanticCube(tipoIzq, tipoDer, t.Op)
		if err != nil {
			a.errores = append(a.errores, err.Error())
		}

		//Gen de Quad - resolver la operación
		a.Generador.ResolverOp()

		return tipo
	}
	return tipoIzq
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
	var op codegen.Operand
	//si es un ID debe estar declarado
	if !v.EsCte {
		if varEntry, existe := a.Tabla.BuscarVar(v.ID, scope); !existe {
			a.errores = append(a.errores, fmt.Sprintf("variable '%s' no fue declarada", v.ID))
		} else { //si sí existe
			tipo = varEntry.Tipo

			//Gen de Quad
			op = codegen.Operand{Name: v.ID, Address: varEntry.Direccion, Type: tipo}

		}
	} else {
		//si es constante, el tipo depende de si es int o float
		if v.CteEnt != nil {
			tipo = "entero"
			//gen de quad
			op = codegen.Operand{Address: a.Mem.GetDirConstInt(*v.CteEnt), Type: tipo}

		} else {
			tipo = "flotante"
			//gen de quad
			op = codegen.Operand{Address: a.Mem.GetDirConstFloat(*v.CteFlot), Type: tipo}

		}
	}

	// si tiene signo negativo, emite un cuádruplo de negación
	if v.Signo == "-" {
		temp := a.Mem.GetDirTemp(op.Type) //obtiene una dirección
		a.Generador.EmitQuad("NEG", op.Address, 0, temp)
		a.Generador.Operands.Push(codegen.Operand{Address: temp, Type: op.Type})
		return tipo
	}

	//Gen de Quad
	a.Generador.Operands.Push(op)

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
	for funcName, funcEntry := range a.Tabla.funciones {
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

func (a *Analizador) Errores() []string {
	return a.errores
}
