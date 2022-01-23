package parser

import (
	"fmt"

	"github.com/flipez/rocket-lang/ast"
)

func (p *Parser) parseClass() ast.Expression {
	class := &ast.Class{Token: p.curToken}

	p.nextToken()

	class.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	class.Body = p.parseBlock()

	print(fmt.Sprintf("%v\n", class.Body))

	return class
}
