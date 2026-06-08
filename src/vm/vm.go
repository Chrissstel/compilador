package vm

import (
	stack "compilador/src"
	"compilador/src/codegen"
	"compilador/src/memory"
	"fmt"
)

type VM struct {
	quads        []codegen.Quadruple
	pos          int //la posición
	mem          *memory.RuntimeMemory
	pendingFrame *memory.ActivationRecord //para guardar el frame de la función que llama mientras se ejecuta la llamada
	returnStack  stack.Stack[int]
	//cuando llamas una funcion tienes que guardar a donde regresar
	lastReturn interface{} //para guardar el valor de retorno de la función que se acaba de ejecutar
}

func NewVM(quads []codegen.Quadruple, manager *memory.MemoryManager) *VM {
	rt := memory.NewRuntimeMemory()
	//para que ya tenga las constantes cargadas
	manager.CargarConstantes(rt)

	// Crear un frame global inicial para que siempre haya un frame en el stack
	// Esto permite que los temporales (9000-11000) tengan un lugar válido para almacenarse
	globalFrame := memory.NewActivationRecord()
	rt.Stack.Push(globalFrame)

	return &VM{
		quads: quads,
		pos:   0,
		mem:   rt,
	}
}

// Función para ejecutar los cuádruplos
func (vm *VM) Run() {
	for vm.pos < len(vm.quads) {

		quad := vm.quads[vm.pos]

		switch quad.Op {
		case "+":
			vm.OperacionBinaria(
				quad,
				func(a, b int) int { return a + b },
				func(a, b float64) float64 { return a + b },
			)
		case "-":
			vm.OperacionBinaria(
				quad,
				func(a, b int) int { return a - b },
				func(a, b float64) float64 { return a - b },
			)
		case "*":
			vm.OperacionBinaria(
				quad,
				func(a, b int) int { return a * b },
				func(a, b float64) float64 { return a * b },
			)
		case "/":
			vm.OperacionBinaria(
				quad,
				func(a, b int) int { return a / b },
				func(a, b float64) float64 { return a / b },
			)
		case "=":
			vm.asigna(quad)

		case ">":
			vm.comparar(quad, func(l, r float64) bool { return l > r })
		case "<":
			vm.comparar(quad, func(l, r float64) bool { return l < r })
		case "==":
			vm.comparar(quad, func(l, r float64) bool { return l == r })
		case "!=":
			vm.comparar(quad, func(l, r float64) bool { return l != r })
		case ">=":
			vm.comparar(quad, func(l, r float64) bool { return l >= r })
		case "<=":
			vm.comparar(quad, func(l, r float64) bool { return l <= r })

		case "GOTO":
			vm.goTo(quad)
		case "GOTOF":
			vm.gotoF(quad)
		case "GOTOT":
			vm.goToT(quad)

		case "ERA":
			vm.era()
		case "PARAM":
			vm.param(quad)
		case "GOSUB":
			vm.goSub(quad)
		case "RETURN":
			vm.ret(quad)
		case "ENDFUNC":
			vm.endFunc()
		case "GETRETURN":
			vm.getReturn(quad)

		case "PRINT":
			vm.print(quad)
		case "PRINT_STR":
			vm.print(quad)
		case "END":
			return

		}
		vm.pos++
	}
}

// convierte un valor a float, total, si el resultado es float no pasa nada
func toFloat(v interface{}) float64 {
	switch x := v.(type) {
	case int:
		return float64(x)
	case float64:
		return x
	default:
		panic("tipo invalido")
	}
}

// Esta función es para + - / *, para que no quede tan largo el código
func (vm *VM) OperacionBinaria(
	quad codegen.Quadruple,
	intOp func(int, int) int,
	floatOp func(float64, float64) float64,
) {

	left := vm.mem.GetValue(quad.Left)
	right := vm.mem.GetValue(quad.Right)

	//si el resultado es entero usamos la operación de enteros
	if memory.IsInt(quad.Result) {
		vm.mem.SetValue(
			quad.Result,
			intOp(left.(int), right.(int)),
		)
		return
	}

	vm.mem.SetValue(
		quad.Result,
		floatOp(toFloat(left), toFloat(right)),
	)
}

func (vm *VM) asigna(quad codegen.Quadruple) {
	// = 13001 0 1001
	//aqui está bien porque getvalue regresa una interface{} y setvalue recibe interface{}
	value := vm.mem.GetValue(quad.Left)
	vm.mem.SetValue(quad.Result, value)
}

// Para reutilizarla en todas las comparaciones
func (vm *VM) comparar(quad codegen.Quadruple, compare func(float64, float64) bool) {
	left := toFloat(vm.mem.GetValue(quad.Left))
	right := toFloat(vm.mem.GetValue(quad.Right))

	if compare(left, right) {
		vm.mem.SetValue(quad.Result, 1)
	} else {
		vm.mem.SetValue(quad.Result, 0)
	}
}

func (vm *VM) goTo(quad codegen.Quadruple) {
	// GOTO 0 0 25
	vm.pos = quad.Result - 1 // -1 porque después del switch se hace vm.pos++
}

func (vm *VM) gotoF(quad codegen.Quadruple) {
	// GOTOF 9003 0 25
	cond := vm.mem.GetValue(quad.Left)
	if cond.(int) == 0 {
		vm.pos = quad.Result - 1 // -1 porque después del switch se hace vm.pos++
	}
}

func (vm *VM) goToT(quad codegen.Quadruple) {
	// GOTOT 9003 0 25
	cond := vm.mem.GetValue(quad.Left)
	if cond.(int) == 1 {
		vm.pos = quad.Result - 1 // -1 porque después del switch se hace vm.pos++
	}
}

// AHORA LAS FUNCIONES PADRES
func (vm *VM) era() {
	// ERA 0 0 func1
	// Aquí es donde se crea el activation record de la función que se va a llamar, pero no se mete a la pila todavía porque falta pasarle los parámetros
	vm.pendingFrame = memory.NewActivationRecord()
	//como en realidad no usamos los recursos, no ocupamos el quad
}

func (vm *VM) param(quad codegen.Quadruple) {
	// PARAM 1000 0 5000
	value := vm.mem.GetValue(quad.Left)
	vm.pendingFrame.Values[quad.Result] = value
}

func (vm *VM) goSub(quad codegen.Quadruple) {
	// GOSUB 0 0 9	//el 9 es el quad donde empieza la función
	//dejas tu migajita de pan para saber a donde regresar
	vm.returnStack.Push(vm.pos + 1)
	//ahora sí activamos el frame
	vm.mem.Stack.Push(vm.pendingFrame)
	//y luego nos movemos al startquad de la función
	vm.pos = quad.Result - 1 // -1 porque después del switch se hace vm.pos++
}

func (vm *VM) getReturn(quad codegen.Quadruple) {
	// GETRETURN 0 0 5000
	vm.mem.SetValue(quad.Result, vm.lastReturn)
}

func (vm *VM) ret(quad codegen.Quadruple) {
	// RETURN 9003 0 0
	vm.lastReturn = vm.mem.GetValue(quad.Left) //guardamos el valor de retorno para que pueda ser usado por el que hizo la llamada
}

func (vm *VM) endFunc() {
	//sacamos el frame de la función que se acaba de ejecutar
	vm.mem.Stack.Pop()

	//recuperamos la migajita de pan
	returnPos := vm.returnStack.Pop()
	vm.pos = returnPos - 1 // -1 porque después del switch se hace vm.pos++
}

func (vm *VM) print(quad codegen.Quadruple) {
	value := vm.mem.GetValue(quad.Left)
	fmt.Println(value)
}
