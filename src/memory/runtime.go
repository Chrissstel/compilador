package memory

import (
	stack "compilador/src"
)

type RuntimeMemory struct {
	globals map[int]interface{}
	consts  map[int]interface{}

	//es para el pueblo, para que vm pueda acceder a el
	Stack stack.Stack[*ActivationRecord] //esto es la "memoria local" para cada función
}

type ActivationRecord struct {
	Values map[int]interface{}
}

func NewRuntimeMemory() *RuntimeMemory {
	return &RuntimeMemory{
		globals: make(map[int]interface{}),
		consts:  make(map[int]interface{}),
	}
}

func NewActivationRecord() *ActivationRecord {
	return &ActivationRecord{
		Values: make(map[int]interface{}),
	}
}

func (m *RuntimeMemory) GetValue(addr int) interface{} {

	// Globales
	if addr >= 1000 && addr < 3000 {
		return m.globals[addr]
	}

	// Constantes
	if addr >= 13000 && addr < 16000 {
		return m.consts[addr]
	}

	// Locales y temporales
	top := m.Stack.Top()

	return top.Values[addr]
}

func (m *RuntimeMemory) SetValue(addr int, value interface{}) {

	// Globales
	if addr >= 1000 && addr < 3000 {
		m.globals[addr] = value
		return
	}

	// Constantes no se pueden modificar
	if addr >= 13000 && addr < 16000 {
		panic("no se puede modificar una constante")
	}

	// Locales y temporales
	top := m.Stack.Top()
	top.Values[addr] = value
}

// Helpers para saber el tipo
func IsInt(addr int) bool {
	return (addr >= 1000 && addr < 2000) ||
		(addr >= 5000 && addr < 6000) ||
		(addr >= 9000 && addr < 10000) ||
		(addr >= 13000 && addr < 14000)
}

func IsFloat(addr int) bool {
	return (addr >= 2000 && addr < 3000) ||
		(addr >= 6000 && addr < 7000) ||
		(addr >= 10000 && addr < 11000) ||
		(addr >= 14000 && addr < 15000)
}
