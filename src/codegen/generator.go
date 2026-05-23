package codegen

import (
	"compilador/src/memory"
	"fmt"
)

type Generator struct {
	Operands  Stack[Operand]
	Operators Stack[string]
	Jumps     Stack[int]
	Quads     []Quadruple //decidí mejor usar una slice y no una cola
	Mem       *memory.MemoryManager
	tempCount int
}

func NewGenerator(mem *memory.MemoryManager) *Generator {
	return &Generator{
		Quads: make([]Quadruple, 0),
		Mem:   mem,
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
