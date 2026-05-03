package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"lingpad/ast"
	"lingpad/codegen"
	"lingpad/parser"
	"lingpad/preprocessor"
	"lingpad/semantic"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "[Main] Nenhum arquivo fornecido. Uso: main <arquivo>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Ler arquivo
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[Main] Erro ao ler arquivo: "+filename)
		os.Exit(1)
	}

	code := string(content) + "\n"

	// Pré-processamento: remover comentários
	prePro := &preprocessor.PrePro{}
	code = prePro.Filter(code)

	// Análise sintática
	astRoot := parser.Parse(code)

	// EXECUÇÃO: Criar tabela de símbolos e executar
	st := semantic.NewSymbolTable()
	astRoot.Evaluate(st)

	// Se existe função main(), chama automaticamente
	if mainFunc, exists := st.Functions["main"]; exists {
		funcDec := mainFunc.Node.(*ast.FunctionDec)
		localSt := semantic.NewSymbolTable()
		localSt.Functions = st.Functions
		funcDec.GetBody().Evaluate(localSt)
	}

	// GERAÇÃO DE CÓDIGO: Reinicializa para gerar assembly
	ast.CodeGenerator = ast.NewCodeGen()
	ast.NextNodeID = 0

	// Cria nova tabela de símbolos para geração
	stGen := semantic.NewSymbolTable()
	// Copia as funções definidas para a tabela de geração
	stGen.Functions = st.Functions

	// Gera código
	astGen := parser.Parse(code)
	astGen.Generate(stGen)

	// Salvar arquivo de assembly no mesmo diretório do arquivo de entrada
	var outputPath string

	if filename == "" {
		outputPath = "output.asm"
	} else {
		// Normaliza o caminho para usar separadores do sistema operacional
		absPath, err := filepath.Abs(filename)
		if err != nil {
			// Se não conseguir caminho absoluto, usa o caminho relativo
			absPath = filename
		}

		dir := filepath.Dir(absPath)
		baseName := filepath.Base(absPath)

		// Remove .ling e adiciona .asm
		if len(baseName) > 5 && baseName[len(baseName)-5:] == ".ling" {
			outputName := baseName[:len(baseName)-5] + ".asm"
			outputPath = filepath.Join(dir, outputName)
		} else {
			outputPath = filepath.Join(dir, baseName+".asm")
		}
	}

	codeGen := codegen.NewCode()
	for _, instr := range ast.CodeGenerator.GetInstructions() {
		codeGen.Append(instr)
	}
	codeGen.Dump(outputPath)

	fmt.Printf("[Main] Arquivo de assembly salvo: %s\n", outputPath)
}
