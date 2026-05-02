package main

import (
	"fmt"
	"os"
)

func main() {
	//lee el archivo a un slice de bytes y un error
	bytes, _ := os.ReadFile("./examples/00.patito")
	source := string(bytes)

	fmt.Printf("Code: %s\n", source)
}
