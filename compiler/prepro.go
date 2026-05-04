package compiler

import "regexp"

// PrePro realiza o pre-processamento do codigo-fonte
type PrePro struct{}

// Filter remove comentarios inline do codigo
func (p *PrePro) Filter(code string) string {
	// Remove tudo entre "//" e "\n", mantendo o "\n"
	re := regexp.MustCompile(`//[^\n]*`)
	return re.ReplaceAllString(code, "")
}
