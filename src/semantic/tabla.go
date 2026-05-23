package semantic

import "fmt"

type TipoDato string //tipos de entrada en la tabla

const (
	TipoEntero   TipoDato = "entero"
	TipoFlotante TipoDato = "flotante"
	TipoNula     TipoDato = "nula"
)

// puede ser una variable o parámetro
type EntradaVar struct {
	Nombre string
	Tipo   TipoDato
	//se añadió lo de la dirección
	Direccion int
}

// es una función ya declarada
type EntradaFunc struct {
	Nombre      string
	TipoRetorno TipoDato
	Parametros  []EntradaVar
}

//snapshot con nombre para guardarlo en el historial
type scopeGuardado struct {
	nombre    string
	variables map[string]EntradaVar
}

// TABLA
type TablaSimbolos struct {
	variables   map[string]EntradaVar  //del scope actual
	funciones   map[string]EntradaFunc //siempre global
	scopeActual string                 //"global" o nombre de la función
	historial   []scopeGuardado        //para guardar los scopes que se cierran
}

func NuevaTabla() *TablaSimbolos {
	return &TablaSimbolos{
		variables:   make(map[string]EntradaVar),
		funciones:   make(map[string]EntradaFunc),
		scopeActual: "global",
		historial:   make([]scopeGuardado, 0),
	}
}

//VARIABLES

func (t *TablaSimbolos) AgregarVar(nombre string, tipo TipoDato) error {
	if _, existe := t.variables[nombre]; existe {
		return fmt.Errorf("variable '%s' ya fue declarada en scope '%s'", nombre, t.scopeActual)
	}
	t.variables[nombre] = EntradaVar{Nombre: nombre, Tipo: tipo}
	return nil
}

func (t *TablaSimbolos) BuscarVar(nombre string) (EntradaVar, bool) {
	entrada, existe := t.variables[nombre]
	return entrada, existe
}

//FUNCIONES

func (t *TablaSimbolos) AgregarFunc(nombre string, tipoRetorno TipoDato, params []EntradaVar) error {
	if _, existe := t.funciones[nombre]; existe {
		return fmt.Errorf("función '%s' ya fue declarada", nombre)
	}
	t.funciones[nombre] = EntradaFunc{
		Nombre:      nombre,
		TipoRetorno: tipoRetorno,
		Parametros:  params,
	}
	return nil
}

func (t *TablaSimbolos) BuscarFunc(nombre string) (EntradaFunc, bool) {
	entrada, existe := t.funciones[nombre]
	return entrada, existe
}

//SCOPES

//guarda las variables globales y abre un scope local limpio
func (t *TablaSimbolos) EntrarScope(nombreFunc string, params []EntradaVar) map[string]EntradaVar {
	//guarda el scope actual para obtenerlo después
	snapshot := t.variables

	//nuevo scope local y le ponemos los parámetros
	t.variables = make(map[string]EntradaVar)
	for _, p := range params {
		t.variables[p.Nombre] = p
	}
	t.scopeActual = nombreFunc
	return snapshot
}

//para restaurar el scope anterior
func (t *TablaSimbolos) SalirScope(snapshot map[string]EntradaVar) {
	//guarda el scope local en el historial antes de cerrarlo
	copia := make(map[string]EntradaVar)
	for k, v := range t.variables {
		copia[k] = v
	}
	t.historial = append(t.historial, scopeGuardado{
		nombre:    t.scopeActual,
		variables: copia,
	})

	t.variables = snapshot
	t.scopeActual = "global"
}

//DEBUG

func (t *TablaSimbolos) Imprimir() {
	fmt.Printf("\n=== Tabla de Símbolos ===\n")

	fmt.Println("────── Funciones ──────")
	if len(t.funciones) == 0 {
		fmt.Println(" (ninguna)")
	}
	for _, f := range t.funciones {
		params := ""
		for i, p := range f.Parametros {
			if i > 0 {
				params += ", "
			}
			params += fmt.Sprintf("%s:%s", p.Nombre, p.Tipo)
		}
		fmt.Printf("  %s(%s) -> %s\n", f.Nombre, params, f.TipoRetorno)
	}

	//scope global
	fmt.Println("\n── Variables globales ───────────────────")
	if len(t.variables) == 0 {
		fmt.Println("  (ninguna)")
	}
	for _, v := range t.variables {
		fmt.Printf("  %s : %s\n", v.Nombre, v.Tipo)
	}

	// scopes locales del historial
	for _, scope := range t.historial {
		fmt.Printf("\n── Variables locales [%s] ────────────────\n", scope.nombre)
		if len(scope.variables) == 0 {
			fmt.Println("  (ninguna)")
		}
		for _, v := range scope.variables {
			fmt.Printf("  %s : %s\n", v.Nombre, v.Tipo)
		}
	}

	fmt.Println("=====================================")
}

func (a *Analizador) ImprimirTabla() {
	a.tabla.Imprimir()
}
