package codegen

import (
	"compilador/src/cube"
	"compilador/src/memory"
	"fmt"
)

type Operand struct {
	Name    string
	Address int
	Type    string
}

// son int porque se van a guardar directamente las dirs de memoria
// si no aplica, se le pone 0
type Quadruple struct {
	Op     string
	Left   int
	Right  int
	Result int
}

type QuadGenerator struct {
	Operands  Stack[Operand] //es operand para ahorrarnos la pila de tipos y dirs por separado
	Operators Stack[string]
	Jumps     Stack[int]
	Quads     []Quadruple
	Mem       *memory.MemoryManager

	// se quita FuncStart map[string]int // nombre → índice del primer quad de la función
}

func NewGenerator(mem *memory.MemoryManager) *QuadGenerator {
	return &QuadGenerator{
		Quads: make([]Quadruple, 0),
		Mem:   mem,
	}
}

func (q Quadruple) PrintQuad() string {
	return fmt.Sprintf("(%s  %d  %d  %d)", q.Op, q.Left, q.Right, q.Result)
}

func (g *QuadGenerator) EmitQuad(Op string, Left int, Right int, Result int) {
	g.Quads = append(g.Quads, Quadruple{Op: Op, Left: Left, Right: Right, Result: Result})
}

func (g *QuadGenerator) PrintQuads() {
	fmt.Println("\n=== Cuádruplos ===")
	fmt.Printf("%-5s %-12s %-8s %-8s %-8s\n", "idx", "op", "left", "right", "result")
	fmt.Println("─────────────────────────────────────────")
	for i := 0; i < len(g.Quads); i++ {
		q := g.Quads[i]
		fmt.Printf("%-5d %-12s %-8d %-8d %-8d\n",
			i, q.Op, q.Left, q.Right, q.Result)
	}
	fmt.Println()
}

// PARA RESOLVER LA OPERACIÓN
func (g *QuadGenerator) ResolverOp() {
	right := g.Operands.Pop()
	left := g.Operands.Pop()
	op := g.Operators.Pop()

	resultType, err := cube.SemanticCube(left.Type, right.Type, op)
	if err != nil {
		fmt.Println("Error de tipos:", err)
		return
	}

	temp := g.Mem.GetDirTemp(resultType) //es una dirección de mem
	g.EmitQuad(op, left.Address, right.Address, temp)
	g.Operands.Push(Operand{Address: temp, Type: resultType})
}

// PARA LLENAR UN JUMP
func (g *QuadGenerator) FillJump(donde int, parche int) {
	//reemplaza el resultado del cuádruplo en el índice 'donde' con 'parche'
	g.Quads[donde] = Quadruple{
		Op:     g.Quads[donde].Op,
		Left:   g.Quads[donde].Left,
		Right:  g.Quads[donde].Right,
		Result: parche,
	}
}
