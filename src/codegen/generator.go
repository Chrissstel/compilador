package codegen

import (
	"compilador/src/ast"
	"compilador/src/memory"
	"compilador/src/semantic"
	"fmt"
)

type Generator struct {
	Operands  Stack[Operand]
	Operators Stack[string]
	Jumps     Stack[int]
	Quads     []Quadruple //decidí mejor usar una slice y no una cola
	Mem       *memory.MemoryManager

	//usamos la tabla de símbolos para que consulte la dir de memoria
	Tabla *semantic.TablaSimbolos
}

func NewGenerator(mem *memory.MemoryManager, tabla *semantic.TablaSimbolos) *Generator {
	return &Generator{
		Quads: make([]Quadruple, 0),
		Mem:   mem,
		Tabla: tabla,
	}
}

// para agregar un cuádruplo a la slice
func (g *Generator) Emit(q Quadruple) {
	g.Quads = append(g.Quads, q)
}

// pide una dirección temporal entera
func (g *Generator) NewTempInt() Operand {
	return Operand{Address: g.Mem.NextTempInt(), Type: "entero"}
}

// pide una dirección temporal flotante
func (g *Generator) NewTempFloat() Operand {
	return Operand{Address: g.Mem.NextTempFloat(), Type: "flotante"}
}

// para decidir el tipo automáticamente
func (g *Generator) NewTemp(tipo string) Operand {
	if tipo == "flotante" {
		return g.NewTempFloat()
	}
	return g.NewTempInt()
}

// PushJump guarda el índice del último cuádruplo emitido
func (g *Generator) PushJump() {
	g.Jumps.Push(len(g.Quads) - 1)
}

// FillJump rellena el resultado de un cuádruplo de salto
func (g *Generator) FillJump(quadIdx int, target int) {
	g.Quads[quadIdx].Result = target
}

// PrintQuads imprime todos los cuádruplos para debug
func (g *Generator) PrintQuads() {
	fmt.Println("\n=== Cuádruplos ===")
	for i, q := range g.Quads {
		fmt.Printf("%3d: %s\n", i, q)
	}
}

func (g *Generator) LookupVar(name string) (Operand, error) {
	entrada, existe := g.Tabla.BuscarVar(name)
	if !existe {
		return Operand{}, fmt.Errorf("Variable '%s' no encontrada", name)
	}
	return Operand{
		Address: entrada.Direccion,
		Type:    string(entrada.Tipo),
	}, nil
}

// PARA LA ENTRADA PRINCIPAL

func (g *Generator) Visit(prog *ast.Programa) {
	//semantics ya le asignó dir a las vars globales

	//generamos las funciones
	for _, f := range prog.Funcs {
		g.visitFunc(f)
	}

	//visitamos el cuerpo principal
	g.visitCuerpo(prog.Cuerpo)

	g.Emit(Quadruple{Op: "END"})
}

//PARA LAS FUNCIONES

func (g *Generator) visitFunc(f *ast.Func) {
	//
	entrada, _ := g.Tabla.BuscarFunc(f.ID)
	snapshot := g.Tabla.EntrarScope(f.ID, entrada.Parametros)

	g.visitCuerpo(f.Cuerpo)

	g.Emit(Quadruple{Op: "ENDFUNC"})

	g.Tabla.SalirScope(snapshot)
}

//PARA EL CUERPO

func (g *Generator) visitCuerpo(c *ast.Cuerpo) {
	for _, e := range c.Estatutos {
		g.visitEstatuto(e)
	}
}

// PARA LOS ESTATUTOS
func (g *Generator) visitEstatuto(e ast.Estatuto) {
	switch n := e.(type) {
	case *ast.Asigna:
		g.visitAsigna(n)
	case *ast.Condicion:
		g.visitCondicion(n)
	case *ast.Ciclo:
		g.visitCiclo(n)
	case *ast.Llamada:
		g.visitLlamada(n)
	case *ast.Imprime:
		g.visitImprime(n)
	case *ast.BloqueEstatutos:
		for _, inner := range n.Estatutos {
			g.visitEstatuto(inner)
		}
	}
}

// PARA ASIGNACIÓN
func (g *Generator) visitAsigna(n *ast.Asigna) {
	g.visitExpresion(n.Expresion)
	result := g.Operands.Pop()

	dest, err := g.LookupVar(n.ID)
	if err != nil {
		fmt.Println("Codegen error:", err)
		return
	}

	g.Emit(Quadruple{
		Op:     "=",
		Left:   result.Address,
		Right:  0,
		Result: dest.Address,
	})
}

// PARA CONDICIONAL
func (g *Generator) visitCondicion(n *ast.Condicion) {
	// 1. evalúa la condición
	g.visitExpresion(n.Expresion)
	cond := g.Operands.Pop()

	// 2. GOTOF pendiente: si la condición es falsa salta a ???
	g.Emit(Quadruple{Op: "GOTOF", Left: cond.Address, Right: 0, Result: 0})
	g.PushJump()

	// 3. cuerpo del si
	g.visitCuerpo(n.CuerpoSi)

	if n.CuerpoSino != nil {
		// 4. al terminar el si, salta por encima del sino
		g.Emit(Quadruple{Op: "GOTO", Left: 0, Right: 0, Result: 0})

		// rellena el GOTOF: si la condición era falsa, llega aquí (inicio del sino)
		falseJump := g.Jumps.Pop()
		g.FillJump(falseJump, len(g.Quads))

		g.PushJump() // guarda el GOTO para rellenarlo al final del sino

		// 5. cuerpo del sino
		g.visitCuerpo(n.CuerpoSino)

		// 6. rellena el GOTO: el si salta aquí al terminar
		endJump := g.Jumps.Pop()
		g.FillJump(endJump, len(g.Quads))
	} else {
		// sin sino: solo rellena el GOTOF con la posición actual
		falseJump := g.Jumps.Pop()
		g.FillJump(falseJump, len(g.Quads))
	}
}

// PARA EL CICLO
func (g *Generator) visitCiclo(n *ast.Ciclo) {
	// 1. guarda dónde empieza la condición para volver aquí
	loopStart := len(g.Quads)
	g.Jumps.Push(loopStart)

	// 2. evalúa la condición
	g.visitExpresion(n.Expresion)
	cond := g.Operands.Pop()

	// 3. GOTOF pendiente: si es falsa sale del loop
	g.Emit(Quadruple{Op: "GOTOF", Left: cond.Address, Right: 0, Result: 0})
	g.PushJump()

	// 4. cuerpo del loop
	g.visitCuerpo(n.Cuerpo)

	// 5. vuelve al inicio de la condición
	exitJump := g.Jumps.Pop() // índice del GOTOF
	loopJump := g.Jumps.Pop() // índice de inicio del loop
	g.Emit(Quadruple{Op: "GOTO", Result: loopJump})

	// 6. rellena el GOTOF: si era falsa llega aquí (después del loop)
	g.FillJump(exitJump, len(g.Quads))
}

// PARA IMPRIME
func (g *Generator) visitImprime(n *ast.Imprime) {
	for _, item := range n.Items {
		if item.EsLetrero {
			// guarda el string en el mapa de constantes y emite PRINT_STR
			addr := g.Mem.RegisterString(item.Letrero)
			g.Emit(Quadruple{Op: "PRINT_STR", Left: addr, Right: 0, Result: 0})
		} else {
			g.visitExpresion(item.Expr)
			val := g.Operands.Pop()
			g.Emit(Quadruple{Op: "PRINT", Left: val.Address, Right: 0, Result: 0})
		}
	}
}

// PARA UNA LLAMADA
func (g *Generator) visitLlamada(n *ast.Llamada) {
	// evalúa cada argumento y lo pasa con PARAM
	for _, arg := range n.Args {
		g.visitExpresion(arg)
		param := g.Operands.Pop()
		g.Emit(Quadruple{Op: "PARAM", Left: param.Address, Right: 0, Result: 0})
	}
	g.Emit(Quadruple{Op: "GOSUB", Left: 0, Right: 0, Result: 0})
	// el Result del GOSUB lo rellenas cuando sepas la dirección de la función
	// lo dejamos pendiente
}

// PARA EXPRESIONES
func (g *Generator) visitExpresion(e *ast.Expresion) {
	g.visitExp(e.Izq)

	if e.Op != "" {
		// hay operador relacional: push, visita el lado derecho, resuelve
		g.Operators.Push(e.Op)
		g.visitExp(e.Der)
		g.resolveOp()
	}
}

func (g *Generator) visitExp(e *ast.Exp) {
	g.visitTermino(e.Termino)

	if e.Op != "" {
		g.Operators.Push(e.Op)
		g.visitExp(e.Der)
		g.resolveOp()
	}
}

func (g *Generator) visitTermino(t *ast.Termino) {
	g.visitFactor(t.Factor)

	if t.Op != "" {
		g.Operators.Push(t.Op)
		g.visitTermino(t.Der)
		g.resolveOp()
	}
}

func (g *Generator) visitFactor(f *ast.Factor) {
	switch {
	case f.Expr != nil:
		// ( <EXPRESION> ) — solo visitas adentro, los paréntesis no generan cuádruplo
		g.visitExpresion(f.Expr)

	case f.Llamada != nil:
		g.visitLlamada(f.Llamada)

	case f.Valor != nil:
		g.visitValor(f.Signo, f.Valor)
	}
}

func (g *Generator) visitValor(signo string, v *ast.Valor) {
	var op Operand

	if v.EsCte {
		if v.CteEnt != nil {
			addr := g.Mem.GetConstInt(*v.CteEnt)
			op = Operand{Address: addr, Type: "entero"}
		} else {
			addr := g.Mem.GetConstFloat(*v.CteFlot)
			op = Operand{Address: addr, Type: "flotante"}
		}
	} else {
		var err error
		op, err = g.LookupVar(v.ID)
		if err != nil {
			fmt.Println("Codegen error:", err)
			return
		}
	}

	// si tiene signo negativo, emite un cuádruplo de negación
	if signo == "-" {
		temp := g.NewTemp(op.Type)
		g.Emit(Quadruple{Op: "NEG", Left: op.Address, Right: 0, Result: temp.Address})
		g.Operands.Push(temp)
		return
	}

	g.Operands.Push(op)
}

// PARA RESOLVER LA OPERACIÓN
func (g *Generator) resolveOp() {
	right := g.Operands.Pop()
	left := g.Operands.Pop()
	op := g.Operators.Pop()

	resultType, err := SemanticCube(left.Type, right.Type, op)
	if err != nil {
		fmt.Println("Error de tipos:", err)
		return
	}

	temp := g.NewTemp(resultType)
	g.Emit(Quadruple{
		Op:     op,
		Left:   left.Address,
		Right:  right.Address,
		Result: temp.Address,
	})
	g.Operands.Push(temp)
}
