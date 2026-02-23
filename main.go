package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {

	// Verifica se um argumento que foi fornecido é válido
	if len(os.Args) < 2 {
		panic("Nenhum argumento fornecido. Por favor, forneça uma expressão matemática como argumento.")
	}

	input := os.Args[1]

	if len(input) == 0 {
		panic("Expressão vazia. Por favor, forneça uma expressão matemática válida.")
	}

	// Variáveis para armazenar o resultado, o número atual e a operação atual
	var result int = 0
	var currentNumber string = ""
	var operation rune = '+'
	var numberCompleted bool = false

	// Itera sobre cada caractere da expressão
	for _, caractere := range input {
		if caractere == ' ' {
			// Ignora espaços, mas marca que um número foi completado se houver um
			if currentNumber != "" {
				numberCompleted = true
			}
			continue
		}

		if caractere >= '0' && caractere <= '9' {
			// Se já completamos um número e não encontramos operador, é erro
			if numberCompleted {
				panic("Expressão inválida: dois números sem operador")
			}
			currentNumber += string(caractere)
		} else if caractere == '+' || caractere == '-' {
			if currentNumber == "" {
				panic("Expressão inválida: operador sem número")
			}
			num, err := strconv.Atoi(currentNumber)
			if err != nil {
				panic(err)
			}
			if operation == '+' {
				result += num
			} else {
				result -= num
			}
			currentNumber = ""
			operation = caractere
			numberCompleted = false
		} else {
			panic("Caractere inválido: " + string(caractere))
		}
	}

	if currentNumber == "" {
		panic("Expressão inválida: termina com operador")
	}
	num, err := strconv.Atoi(currentNumber)
	if err != nil {
		panic(err)
	}
	if operation == '+' {
		result += num
	} else {
		result -= num
	}

	fmt.Println(result)
}
