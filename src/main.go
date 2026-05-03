package main

import (
	"compilador/src/lexer"
	"os"
)

func main() {
	//lee el archivo a un slice de bytes y un error
	bytes, _ := os.ReadFile("./examples/01.patito")
	source := string(bytes)

	tokens := lexer.Tokenize(source)

	for _, token := range tokens {
		token.Debug()
	}

}
