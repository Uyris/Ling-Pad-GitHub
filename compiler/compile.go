package compiler

import (
	"fmt"
	"os"
	"path/filepath"
)

func CompileFile(filename string) (string, error) {
	if filename == "" {
		return "", fmt.Errorf("[Main] Nenhum arquivo fornecido. Uso: main <arquivo>")
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("[Main] Erro ao ler arquivo: %s", filename)
	}

	code := string(content) + "\n"

	// Pre-processamento: remover comentarios
	prePro := &PrePro{}
	code = prePro.Filter(code)

	// Analise sintatica
	ast := Run(code)

	// Execucao: criar tabela de simbolos e executar
	st := NewSymbolTable(nil)
	ast.Evaluate(st)

	// Geracao de codigo: reinicializa para gerar assembly
	codeInstructions = []string{}
	nextNodeID = 0

	// Cria nova tabela de simbolos para geracao
	stGen := NewSymbolTable(nil)

	// Gera o codigo assembly
	ast.Generate(stGen)

	// Gera o nome do arquivo de saida (.asm)
	outputFilename := filename[:len(filename)-len(filepath.Ext(filename))] + ".asm"

	// Escreve o arquivo assembly
	codeGenerator.Dump(outputFilename)
	return outputFilename, nil
}
