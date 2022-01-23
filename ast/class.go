package ast

import (
	"bytes"

	"github.com/flipez/rocket-lang/token"
)

type Class struct {
	Expression
	Token token.Token
	Name  *Identifier
	Body  *Block
}

func (cs *Class) TokenLiteral() string { return cs.Token.Literal }
func (cs *Class) String() string {
	var out bytes.Buffer

	out.WriteString("class ")
	out.WriteString(cs.Name.Value)
	out.WriteString(" {\n")
	out.WriteString("\n}")

	return out.String()
}
