package main

import (
	"fmt"
	"os"

	"lingpad/compiler"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "[Main] Nenhum arquivo fornecido. Uso: main <arquivo>")
		os.Exit(1)
	}

	if _, err := compiler.CompileFile(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
