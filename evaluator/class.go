package evaluator

import (
	"github.com/flipez/rocket-lang/ast"
	"github.com/flipez/rocket-lang/object"
)

func evalClass(c *ast.Class, env *object.Environment) object.Object {
	class := &object.Class{
		Name: c.Name,
		Env:  object.NewEnvironment(),
	}

	nestedEnv := object.NewEnclosedEnvironment(env)
	Eval(c.Body, nestedEnv)

	env.Set(c.Name.Value, class)

	return class
}
