package codegen

import (
	"fmt"
	"os"
)

// Code armazena e gerencia o código assembly gerado
type Code struct {
	Instructions []string
}

// NewCode cria um novo gerador de código
func NewCode() *Code {
	return &Code{Instructions: []string{}}
}

// Append adiciona uma instrução assembly
func (c *Code) Append(instruction string) {
	c.Instructions = append(c.Instructions, instruction)
}

// Clear limpa as instruções
func (c *Code) Clear() {
	c.Instructions = []string{}
}

// GetInstructions retorna as instruções
func (c *Code) GetInstructions() []string {
	return c.Instructions
}

// Dump escreve o código assembly em um arquivo
func (c *Code) Dump(filename string) error {
	header := `section .data
  format_out: db "%d", 10, 0 ; format do printf
  format_in: db "%d", 0 ; format do scanf
  scan_int: dd 0; 32-bits integer

section .text
  extern printf ; usar printf
  extern scanf ; usar scanf
  extern exit ; exit function
  global _start ; início do programa

_start:
  push ebp ; guarda o EBP
  mov ebp, esp ; zera a pilha

  ; aqui começa o codigo gerado:`

	footer := `
  ; aqui termina o código gerado

  mov esp, ebp ; reestabelece a pilha
  pop ebp

  ; chamada de exit(0)
  push 0
  call exit`

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo '%s': %w", filename, err)
	}
	defer file.Close()

	if _, err := file.WriteString(header + "\n"); err != nil {
		return fmt.Errorf("erro ao escrever header em '%s': %w", filename, err)
	}

	for _, instr := range c.Instructions {
		if _, err := file.WriteString(instr + "\n"); err != nil {
			return fmt.Errorf("erro ao escrever instrução em '%s': %w", filename, err)
		}
	}

	if _, err := file.WriteString(footer + "\n"); err != nil {
		return fmt.Errorf("erro ao escrever footer em '%s': %w", filename, err)
	}

	return nil
}
