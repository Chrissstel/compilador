package memory

type MemoryManager struct {
	globalInt   int
	globalFloat int

	localInt   int
	localFloat int

	tempInt   int
	tempFloat int

	constInt    int
	constFloat  int
	constString int

	//los mapas para las consts
	constIntMap    map[int]int     //valor a dir
	constFloatMap  map[float64]int //valor a dir
	constStringMap map[string]int  //valor a dir
}

func NewMemoryManager() *MemoryManager {
	return &MemoryManager{
		globalInt:   1000,
		globalFloat: 2000,

		localInt:   5000,
		localFloat: 6000,

		tempInt:   9000,
		tempFloat: 10000,

		constInt:    13000,
		constFloat:  14000,
		constString: 15000,

		constIntMap:    make(map[int]int),     //valor a dir
		constFloatMap:  make(map[float64]int), //valor a dir
		constStringMap: make(map[string]int),  //valor a dir
	}
}

//FUNCIONES PARA OBTENER DIRS DE MEMORIA

func (m *MemoryManager) GetDirGlobal(tipo string) int {
	switch tipo {
	case "int", "entero":
		return m.GetDirGlobalInt()
	case "float", "flotante":
		return m.GetDirGlobalFloat()
	default:
		return 0
	}
}

func (m *MemoryManager) GetDirGlobalInt() int {
	dir := m.globalInt
	m.globalInt++
	return dir
}

func (m *MemoryManager) GetDirGlobalFloat() int {
	dir := m.globalFloat
	m.globalFloat++
	return dir
}

func (m *MemoryManager) GetDirLocalInt() int {
	dir := m.localInt
	m.localInt++
	return dir
}

func (m *MemoryManager) GetDirLocalFloat() int {
	dir := m.localFloat
	m.localFloat++
	return dir
}

func (m *MemoryManager) GetDirLocal(tipo string) int {
	switch tipo {
	case "int", "entero":
		return m.GetDirLocalInt()
	case "float", "flotante":
		return m.GetDirLocalFloat()
	default:
		return 0
	}
}

func (m *MemoryManager) GetDirTemp(tipo string) int {
	switch tipo {
	case "int", "entero":
		return m.GetDirTempInt()
	case "float", "flotante":
		return m.GetDirTempFloat()
	default:
		return 0
	}
}

func (m *MemoryManager) GetDirTempInt() int {
	dir := m.tempInt
	m.tempInt++
	return dir
}

func (m *MemoryManager) GetDirTempFloat() int {
	dir := m.tempFloat
	m.tempFloat++
	return dir
}

func (m *MemoryManager) GetDirConstInt(value int) int {
	if dir, exists := m.constIntMap[value]; exists {
		return dir
	}
	dir := m.constInt
	m.constInt++
	m.constIntMap[value] = dir
	return dir
}

func (m *MemoryManager) GetDirConstFloat(value float64) int {
	if dir, exists := m.constFloatMap[value]; exists {
		return dir
	}
	dir := m.constFloat
	m.constFloat++
	m.constFloatMap[value] = dir
	return dir
}

func (m *MemoryManager) GetDirConstString(value string) int {
	if dir, exists := m.constStringMap[value]; exists {
		return dir
	}
	dir := m.constString
	m.constString++
	m.constStringMap[value] = dir
	return dir
}
