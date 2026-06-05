package ast

//convertí cada regla en un nodo
//Nodo raíz
type Programa struct {
	ID        string
	DeclsVars *DeclsVars //puede ser nil
	Funcs     []*Func
	Cuerpo    *Cuerpo
}

//Variables

type DeclsVars struct {
	Decls []*Vars //cada vars es una lista de variables del mismo tipo
}

//funciona porque solo se puede declarar variables en <VARS>
//Vars es lo mismo que una decl
type Vars struct { //se usa en programa y en funciones
	IDs  []string
	Tipo string
}

//Funciones
type Func struct {
	TipoRetorno string //puede ser nulo, entero o flotante
	ID          string
	Params      []*Param
	Vars        *DeclsVars //puede ser nil
	Cuerpo      *Cuerpo
	Retorno     *Retorno //puede ser nil
}

//AGREGUÉ LO DE RETORNAR
type Retorno struct {
	ID string //solo puede retornar ids
}

type Param struct {
	ID   string
	Tipo string
}

//Cuerpo y estatutos
type Cuerpo struct {
	Estatutos []Estatuto
}

//interfaz para los distintos tipos de estatutos
//esto define que cualquier tipo que implemente el método estatutoNode() es un Estatuto
type Estatuto interface{ estatutoNode() }

type Asigna struct {
	ID        string
	Expresion *Expresion
}

func (*Asigna) estatutoNode() {}

type Condicion struct {
	Expresion  *Expresion
	CuerpoSi   *Cuerpo
	CuerpoSino *Cuerpo //puede ser nil
}

func (*Condicion) estatutoNode() {}

type Ciclo struct {
	Expresion *Expresion
	Cuerpo    *Cuerpo
}

func (*Ciclo) estatutoNode() {}

type Llamada struct {
	ID   string
	Args []*Expresion
}

func (*Llamada) estatutoNode() {}

type Imprime struct {
	Items []ImprimeItem //expresion o letrero
}

func (*Imprime) estatutoNode() {}

type ImprimeItem struct {
	EsLetrero bool
	Letrero   string     //si EsLetrero
	Expr      *Expresion //si !EsLetrero
}

type BloqueEstatutos struct {
	Estatutos []Estatuto
}

func (*BloqueEstatutos) estatutoNode() {}

//Expresiones
type Expresion struct {
	Izq *Exp
	Op  string //vacío si no hay
	Der *Exp   //nil si no hay op
}

type Exp struct {
	Termino *Termino
	Op      string
	Der     *Exp
}

type Termino struct {
	Factor *Factor
	Op     string
	Der    *Termino
}

type Factor struct {
	Expr    *Expresion
	Llamada *Llamada
	Signo   string
	Valor   *Valor
}

type Valor struct {
	EsCte   bool
	ID      string //si !EsCte
	CteEnt  *int
	CteFlot *float64
}
