package object

import (
	"bytes"
	"github.com/flipez/rocket-lang/ast"
)

type Class struct {
	Name *ast.Identifier
	Env  *Environment
}

func (c *Class) Type() ObjectType { return CLASS_OBJ }
func (c *Class) Inspect() string {
	var out bytes.Buffer

	out.WriteString("class {\n")
	out.WriteString("}\n")

	return out.String()
}
func (c *Class) InvokeMethod(method string, env Environment, args ...Object) Object {
	return objectMethodLookup(c, method, args)
}
