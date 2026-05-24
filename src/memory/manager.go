package memory

type MemoryManager struct {
	globalInt   int
	globalFloat int

	localInt   int
	localFloat int

	tempInt   int
	tempFloat int

	constInt   int
	constFloat int

	//los mapas para las consts
	constIntMap    map[int]int     //valor a dir
	constFloatMap  map[float64]int //valor a dir
	constString    int
	constStringMap map[string]int //valor a dir
}

func NewMemoryManager() *MemoryManager {
	return &MemoryManager{
		globalInt:      1000,
		globalFloat:    2000,
		localInt:       5000,
		localFloat:     6000,
		tempInt:        9000,
		tempFloat:      10000,
		constInt:       13000,
		constFloat:     14000,
		constIntMap:    make(map[int]int),
		constFloatMap:  make(map[float64]int),
		constString:    15000,
		constStringMap: make(map[string]int),
	}
}

// regresa la dir en la que vamos y la aumenta
func (m *MemoryManager) NextGlobalInt() int {
	next := m.globalInt
	m.globalInt++
	return next
}

func (m *MemoryManager) NextGlobalFloat() int {
	next := m.globalFloat
	m.globalFloat++
	return next
}

func (m *MemoryManager) NextLocalInt() int {
	next := m.localInt
	m.localInt++
	return next
}

func (m *MemoryManager) NextLocalFloat() int {
	next := m.localFloat
	m.localFloat++
	return next
}

func (m *MemoryManager) NextTempInt() int {
	next := m.tempInt
	m.tempInt++
	return next
}

func (m *MemoryManager) NextTempFloat() int {
	next := m.tempFloat
	m.tempFloat++
	return next
}

func (m *MemoryManager) NextConstInt() int {
	next := m.constInt
	m.constInt++
	return next
}

func (m *MemoryManager) NextConstFloat() int {
	next := m.constFloat
	m.constFloat++
	return next
}

// para las constantes, primero revisamos si ya existe, si no, la agregamos

func (m *MemoryManager) GetConstInt(val int) int {
	if addr, exists := m.constIntMap[val]; exists {
		return addr
	}
	addr := m.NextConstInt()
	m.constIntMap[val] = addr
	return addr
}

func (m *MemoryManager) GetConstFloat(val float64) int {
	if addr, exists := m.constFloatMap[val]; exists {
		return addr
	}
	addr := m.NextConstFloat()
	m.constFloatMap[val] = addr
	return addr
}

func (m *MemoryManager) RegisterString(val string) int {
	if addr, exists := m.constStringMap[val]; exists {
		return addr
	}
	addr := m.constString
	m.constString++
	m.constStringMap[val] = addr
	return addr
}
