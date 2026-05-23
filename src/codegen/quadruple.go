package codegen

import (
	"fmt"
)

type Operand struct {
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

func (q Quadruple) String() string {
	return fmt.Sprintf("(%s  %d  %d  %d)", q.Op, q.Left, q.Right, q.Result)
}
