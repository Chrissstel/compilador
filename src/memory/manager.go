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

		constIntMap:    make(map[int]int),
		constFloatMap:  make(map[float64]int),
		constStringMap: make(map[string]int),
	}
}
