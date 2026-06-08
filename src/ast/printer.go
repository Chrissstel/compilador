package ast

import "fmt"

func (p *Programa) Print() {
	fmt.Printf("Programa: %s\n", p.ID)
	if p.DeclsVars != nil {
		p.DeclsVars.print("├── ")
	}
	for i, f := range p.Funcs {
		prefix := "├── "
		if i == len(p.Funcs)-1 && p.Cuerpo == nil {
			prefix = "└── "
		}
		f.print(prefix, "│   ")
	}
	p.Cuerpo.print("└── ", "    ")
}

func (dv *DeclsVars) print(prefix string) {
	fmt.Printf("%sVars\n", prefix)
	for i, v := range dv.Decls {
		p := prefix + "├── "
		if i == len(dv.Decls)-1 {
			p = prefix + "└── "
		}
		fmt.Printf("%sDecl [%s] : %s\n", p, joinIDs(v.IDs), v.Tipo)
	}
}

func (f *Func) print(prefix, childPrefix string) {
	fmt.Printf("%sFuncion: %s -> %s\n", prefix, f.ID, f.TipoRetorno)
	for _, param := range f.Params {
		fmt.Printf("%s├── Param: %s : %s\n", childPrefix, param.ID, param.Tipo)
	}
	if f.Vars != nil {
		f.Vars.print(childPrefix + "├── ")
	}
	if f.Retorno != nil {
		f.Cuerpo.print(childPrefix+"├── ", childPrefix+"│   ")
		fmt.Printf("%s└── Retorno: %s\n", childPrefix, f.Retorno.ID)
	} else {
		f.Cuerpo.print(childPrefix+"└── ", childPrefix+"    ")
	}
}

func (c *Cuerpo) print(prefix, childPrefix string) {
	fmt.Printf("%sCuerpo\n", prefix)
	for i, e := range c.Estatutos {
		last := i == len(c.Estatutos)-1
		printEstatuto(e, childPrefix, last)
	}
}

func printEstatuto(e Estatuto, parentPrefix string, last bool) {
	pre := parentPrefix + "├── "
	childPre := parentPrefix + "│   "
	if last {
		pre = parentPrefix + "└── "
		childPre = parentPrefix + "    "
	}

	switch n := e.(type) {
	case *Asigna:
		fmt.Printf("%sAsigna: %s\n", pre, n.ID)
		printExpresion(n.Expresion, childPre, true)

	case *Condicion:
		fmt.Printf("%sCondicion\n", pre)
		fmt.Printf("%s├── Cond: ", childPre)
		printExpresionInline(n.Expresion)
		n.CuerpoSi.print(childPre+"├── ", childPre+"│   ")
		if n.CuerpoSino != nil {
			n.CuerpoSino.print(childPre+"└── Sino > ", childPre+"    ")
		}

	case *Ciclo:
		fmt.Printf("%sMientras\n", pre)
		fmt.Printf("%s├── Cond: ", childPre)
		printExpresionInline(n.Expresion)
		n.Cuerpo.print(childPre+"└── ", childPre+"    ")

	case *DoWhile:
		fmt.Printf("%sHaz\n", pre)
		n.Cuerpo.print(childPre+"├──", childPre+"│   ")
		fmt.Printf("%s└── Cond: ", childPre)
		printExpresionInline(n.Expresion)

	case *Llamada:
		fmt.Printf("%sLlamada: %s(%d args)\n", pre, n.ID, len(n.Args))
		for i, arg := range n.Args {
			printExpresion(arg, childPre, i == len(n.Args)-1)
		}

	case *Imprime:
		fmt.Printf("%sEscribe\n", pre)
		for i, item := range n.Items {
			last := i == len(n.Items)-1
			p := childPre + "├── "
			if last {
				p = childPre + "└── "
			}
			if item.EsLetrero {
				fmt.Printf("%sLetrero: %s\n", p, item.Letrero)
			} else {
				fmt.Printf("%sExpr: ", p)
				printExpresionInline(item.Expr)
			}
		}

	case *BloqueEstatutos:
		fmt.Printf("%sBloque\n", pre)
		for i, inner := range n.Estatutos {
			printEstatuto(inner, childPre, i == len(n.Estatutos)-1)
		}
	}
}

func printExpresion(e *Expresion, prefix string, last bool) {
	pre := prefix + "└── "
	fmt.Printf("%sExpr: ", pre)
	printExpresionInline(e)
}

func printExpresionInline(e *Expresion) {
	if e.Op != "" {
		fmt.Printf("%s %s %s\n", expStr(e.Izq), e.Op, expStr(e.Der))
	} else {
		fmt.Printf("%s\n", expStr(e.Izq))
	}
}

func expStr(e *Exp) string {
	if e == nil {
		return ""
	}
	s := termStr(e.Termino)
	if e.Op != "" {
		s += " " + e.Op + " " + expStr(e.Der)
	}
	return s
}

func termStr(t *Termino) string {
	if t == nil {
		return ""
	}
	s := factorStr(t.Factor)
	if t.Op != "" {
		s += " " + t.Op + " " + termStr(t.Der)
	}
	return s
}

func factorStr(f *Factor) string {
	if f.Expr != nil {
		return "(" + expStr(f.Expr.Izq) + ")"
	}
	if f.Llamada != nil {
		return f.Llamada.ID + "(...)"
	}
	return valorStr(f.Valor)
}

func valorStr(v *Valor) string {
	if v == nil {
		return "?"
	}
	if !v.EsCte {
		return v.Signo + v.ID
	}
	if v.CteEnt != nil {
		return v.Signo + fmt.Sprintf("%d", *v.CteEnt)
	}
	if v.CteFlot != nil {
		return v.Signo + fmt.Sprintf("%g", *v.CteFlot)
	}
	return "?"
}

func joinIDs(ids []string) string {
	result := ""
	for i, id := range ids {
		if i > 0 {
			result += ", "
		}
		result += id
	}
	return result
}
