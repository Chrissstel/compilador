package ast

type CuerpoStmt struct {
	Body []Stmt
}

func (n CuerpoStmt) stmt() {}

// para asegurar que termine con ;
type ExpressionStmt struct {
	Expression Expr
}

func (n ExpressionStmt) stmt() {}
