package main

import (
	"compilador/src/codegen"
	"compilador/src/lexer"
	"compilador/src/memory"
	"compilador/src/parser"
	"compilador/src/semantic"
	"fmt"
	"os"
)

func main() {

	//lee el archivo a un slice de bytes y un error
	src, err := os.ReadFile("./testdata/01.patito") //o completo.patito
	if err != nil {
		panic(err)
	}

	//Léxico
	tokens := lexer.Tokenize(string(src))

	//para ver todos los tokens
	for _, token := range tokens {
		token.Debug()
	}

	//Sintáctico
	p := parser.New(tokens)
	programa := p.ParsePrograma()
	fmt.Printf("\nPrograma '%s' parseado\n", programa.ID)
	programa.Print()

	//Semántico
	mem := memory.NewMemoryManager()
	analizador := semantic.NuevoAnalizador(mem)
	analizador.AnalizarPrograma(programa)

	if analizador.HayErrores() {
		fmt.Println("\n=== Errores semánticos ===")
		analizador.ImprimirErrores()
		os.Exit(1)
	}

	fmt.Println("\nAnálisis semántico correcto")
	analizador.ImprimirTabla()

	//Generación de cuadruplos
	gen := codegen.NewGenerator(mem, analizador.ObtenerTabla())
	gen.Visit(programa)
	gen.PrintQuads()

}
