package ast

//convertí cada regla en un nodo
//Nodo raíz
type Programa struct {
	ID     string
	Vars   *Vars //puede ser nil
	Funcs  []*Func
	Cuerpo *Cuerpo
}

//Variables
type Vars struct {
	Declaraciones []*DeclaracionVar
}

type DeclaracionVar struct {
	IDs  []string
	Tipo string
}

//Funciones
type Func struct {
	TipoRetorno string //puede ser nula, entero o flotante
	ID          string
	Params      []*Param
	Vars        *Vars
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
