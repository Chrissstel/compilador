package ast

import "compilador/src/lexer"

//---------------------
//LITERAL EXPRESSIONS
//---------------------
type FlotExpr struct {
	Value float64
}

func (n FlotExpr) expr() {}

type EntExpr struct {
	Value int64
}

func (n EntExpr) expr() {}

type LetreroExpr struct {
	Value string
}

func (n LetreroExpr) expr() {}

type IdentExpr struct {
	Value string
}

func (n IdentExpr) expr() {}

//---------------------
//COMPLEX EXPRESSIONS
//---------------------

type BinaryExpr struct {
	Left     Expr
	Operator lexer.Token
	Right    Expr
}

func (n BinaryExpr) expr() {}
