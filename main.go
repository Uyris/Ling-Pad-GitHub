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
		fmt.Fprintln(os.Stderr, "[Main] ERROR: Nenhum arquivo fornecido. Uso: main <arquivo>")
		os.Exit(1)
	}

	filename := os.Args[1]
	fmt.Printf("[Main] DEBUG: Tentando abrir arquivo: %s\n", filename)

	// Ler arquivo
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Main] ERROR ao ler arquivo '%s': %v\n", filename, err)
		os.Exit(1)
	}
	fmt.Printf("[Main] DEBUG: Arquivo lido com sucesso (%d bytes)\n", len(content))

	code := string(content) + "\n"

	// Pré-processamento: remover comentários
	prePro := &preprocessor.PrePro{}
	code = prePro.Filter(code)
	fmt.Printf("[Main] DEBUG: Pré-processamento concluído\n")

	// Análise sintática
	fmt.Printf("[Main] DEBUG: Iniciando análise sintática...\n")
	astRoot := parser.Parse(code)
	fmt.Printf("[Main] DEBUG: Análise sintática concluída\n")

	// EXECUÇÃO: Criar tabela de símbolos e executar
	fmt.Printf("[Main] DEBUG: Iniciando avaliação semântica...\n")
	st := semantic.NewSymbolTable()
	astRoot.Evaluate(st)
	fmt.Printf("[Main] DEBUG: Avaliação semântica concluída\n")

	// Se existe função main(), chama automaticamente
	if mainFunc, exists := st.Functions["main"]; exists {
		fmt.Printf("[Main] DEBUG: Função main() encontrada, executando...\n")
		funcDec := mainFunc.Node.(*ast.FunctionDec)
		localSt := semantic.NewSymbolTable()
		localSt.Functions = st.Functions
		funcDec.GetBody().Evaluate(localSt)
		fmt.Printf("[Main] DEBUG: Função main() executada\n")
	} else {
		fmt.Printf("[Main] DEBUG: Função main() não encontrada\n")
	}

	// GERAÇÃO DE CÓDIGO: Reinicializa para gerar assembly
	fmt.Printf("[Main] DEBUG: Reinicializando gerador de código...\n")
	ast.CodeGenerator = ast.NewCodeGen()
	ast.NextNodeID = 0

	// Cria nova tabela de símbolos para geração
	stGen := semantic.NewSymbolTable()
	// Copia as funções definidas para a tabela de geração
	stGen.Functions = st.Functions

	// Gera código
	fmt.Printf("[Main] DEBUG: Iniciando geração de código assembly...\n")
	astGen := parser.Parse(code)
	astGen.Generate(stGen)
	fmt.Printf("[Main] DEBUG: Geração de código concluída\n")

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

	fmt.Printf("[Main] DEBUG: Caminho de saída calculado: %s\n", outputPath)

	codeGen := codegen.NewCode()
	for _, instr := range ast.CodeGenerator.GetInstructions() {
		codeGen.Append(instr)
	}

	fmt.Printf("[Main] DEBUG: Salvando %d instruções de assembly...\n", len(ast.CodeGenerator.GetInstructions()))
	codeGen.Dump(outputPath)

	// Verificar se o arquivo foi criado
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Printf("[Main] SUCCESS: Arquivo de assembly salvo com sucesso: %s\n", outputPath)
	} else {
		fmt.Fprintf(os.Stderr, "[Main] ERROR: Falha ao salvar arquivo: %v\n", err)
		os.Exit(1)
	}
}
