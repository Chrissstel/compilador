package main

import (
	"compilador/src/lexer"
	"compilador/src/memory"
	"compilador/src/parser"
	"compilador/src/semantic"
	"compilador/src/vm"

	//"compilador/src/parser"
	"fmt"
	"os"
)

func main() {

	//lee el archivo a un slice de bytes y un error
	src, err := os.ReadFile("./testdata/completo.patito") //o completo.patito
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

	//Semántico y CodeGen
	mem := memory.NewMemoryManager()
	//el analizador también va a generar cuádruplos
	analizador := semantic.NuevoAnalizador(mem)
	analizador.AnalizarPrograma(programa)

	if analizador.HayErrores() {
		analizador.ImprimirErrores()
		os.Exit(1)
	}

	fmt.Println("\nAnálisis semántico correcto")
	analizador.ImprimirTabla()

	//Mostrar cuádruplos
	analizador.Generador.PrintQuads()

	//Máquina virtual
	vm := vm.NewVM(analizador.Generador.Quads, mem)
	vm.Run()

}
