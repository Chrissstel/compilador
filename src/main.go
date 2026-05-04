package main

import (
	"compilador/src/lexer"
	"compilador/src/parser"
	"fmt"
	"os"
)

func main() {
	//lee el archivo a un slice de bytes y un error
	src, err := os.ReadFile("./examples/01.patito")
	if err != nil {
		panic(err)
	}

	tokens := lexer.Tokenize(string(src))

	//para ver todos los tokens
	for _, token := range tokens {
		token.Debug()
	}

	p := parser.New(tokens)
	programa := p.ParsePrograma()

	fmt.Printf("\nPrograma '%s' parseado\n", programa.ID)

}
