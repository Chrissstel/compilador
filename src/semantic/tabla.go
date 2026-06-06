package semantic

//más que tabla.go es semantic.go

import "fmt"

// puede ser una variable o parámetro
type VarEnDir struct {
	Nombre string //va a estar doble el nombre de la variable, pero es para facilitar la búsqueda
	Tipo   string
	//se añadió lo de la dirección
	Direccion int
}

// es una función ya declarada
type FuncEnDir struct {
	Nombre      string //aqui igual va a estar doble el nombre, pero por lo mismo
	TipoRetorno string
	//solo va a ser el nombre y van a estar en orden
	Params    []string            //esto es para validar las llamadas a funciones
	Variables map[string]VarEnDir //van a ser tanto parámetros como variables declaradas
	Recursos  int                 //cantidad de recursos que usa

	StartQuad int //índice del primer cuádruplo de la función
}

// TABLA
type TablaSimbolos struct {
	funciones map[string]*FuncEnDir //global también se guarda como función
	//scopeActual string                 //"global" o nombre de la función
}

func NuevaTabla() *TablaSimbolos {
	return &TablaSimbolos{
		funciones: make(map[string]*FuncEnDir),
		//scopeActual: "global",
	}
}

//PARA AGREGAR VARIABLES
func (t *TablaSimbolos) AgregarVar(scope string, id string, tipo string, dir int) error {
	// Implementación para agregar variable
	funcDir := t.funciones[scope] //es una FuncEnDir

	//checar si la variable ya existe
	if _, exists := funcDir.Variables[id]; exists {
		return fmt.Errorf("Error semántico: variable '%s' ya declarada en el scope '%s'", id, scope)
	}

	funcDir.Variables[id] = VarEnDir{
		Nombre: id,
		Tipo:   tipo,
		// asignar memoria
		Direccion: dir,
	}

	t.funciones[scope] = funcDir

	//aumentar la cantidad de recursos por funcion
	t.funciones[scope].Recursos++

	return nil
}

//Para agregar los parametros a cada funcion
//Solo es para distinguir entre params y variables
//se van a añadir en orden
func (t *TablaSimbolos) AgregarParam(scope string, id string) error {
	// Implementación para agregar parámetro
	funcDir := t.funciones[scope] //es una FuncEnDir
	//checar si el parámetro ya existe
	for _, param := range funcDir.Params {
		if param == id {
			return fmt.Errorf("Error semántico: parámetro '%s' ya declarado en la función '%s'", id, scope)
		}
	}
	funcDir.Params = append(funcDir.Params, id)
	t.funciones[scope] = funcDir
	return nil
}

//PARA BUSCAR VARIABLES
func (t *TablaSimbolos) BuscarVar(id string, scope string) (VarEnDir, bool) {
	//busca primero en el scope actual
	localFuncDir := t.funciones[scope]
	if localFuncDir == nil {
		return VarEnDir{}, false //si no existe pues lo va a regresar vacío y false
	}
	varEntry, exists := localFuncDir.Variables[id]
	if exists {
		return varEntry, true
	}

	//si no está en el scope actual, busca en global
	globalFuncDir := t.funciones["global"]
	if globalFuncDir != nil {
		varEntry, exists := globalFuncDir.Variables[id]
		if exists {
			return varEntry, true
		}
	}
	return VarEnDir{}, false
}

//PARA REGISTRAR FUNCIONES
func (t *TablaSimbolos) RegistrarFunc(id string, tipoRetorno string, params []VarEnDir) error {
	// Implementación para agregar función
	//también podríamos checar que no se añadan dos parámetros llamados igual

	//checar si la función ya existe
	if _, exists := t.funciones[id]; exists {
		return fmt.Errorf("Error semántico: función '%s' ya declarada", id)
	}

	t.funciones[id] = &FuncEnDir{ //aqui se añade la función a la tabla, aunque sin las variables, para que pueda haber recursividad o funciones que se llamen entre sí sin importar el orden
		Nombre:      id,
		TipoRetorno: tipoRetorno,
		Variables:   make(map[string]VarEnDir),
		Recursos:    0,
	}

	//agregar los parámetros como variables de la función
	//y también se guardan en params para validar las llamadas a funciones
	for _, param := range params {
		if err := t.AgregarVar(id, param.Nombre, param.Tipo, param.Direccion); err != nil {
			return err
		}
		if err := t.AgregarParam(id, param.Nombre); err != nil {
			return err
		}
	}

	return nil
}

//PARA BUSCAR FUNCIONES
func (t *TablaSimbolos) BuscarFunc(id string) (*FuncEnDir, bool) {
	funcDir, exists := t.funciones[id]
	return funcDir, exists
}

//DEBUG
