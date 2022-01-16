package evaluator

import (
	"github.com/flipez/rocket-lang/ast"
	"github.com/flipez/rocket-lang/lexer"
	"github.com/flipez/rocket-lang/object"
	"github.com/flipez/rocket-lang/parser"
	"github.com/flipez/rocket-lang/stdlib"
	"github.com/flipez/rocket-lang/utilities"

	"fmt"
	"io/ioutil"
	"path/filepath"
)

var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	// Statements

	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.Block:
		return evalBlock(node, env)
	case *ast.Foreach:
		return evalForeach(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	// Expressions
	case *ast.Integer:
		return &object.Integer{Value: node.Value}
	case *ast.Float:
		return &object.Float{Value: node.Value}
	case *ast.Function:
		params := node.Parameters
		body := node.Body
		name := node.Name
		function := &object.Function{Parameters: params, Env: env, Body: body}

		if name != "" {
			env.Set(name, function)
		}

		return function
	case *ast.Import:
		return evalImport(node, env)
	case *ast.String:
		return &object.String{Value: node.Value}
	case *ast.Array:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}
	case *ast.Hash:
		return evalHash(node, env)

	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)

	case *ast.Infix:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)

	case *ast.If:
		return evalIf(node, env)

	case *ast.Call:
		function := Eval(node.Callable, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)

	case *ast.Index:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)

	case *ast.ObjectCall:
		res := evalObjectCall(node, env)
		return (res)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.Assign:
		return evalAssign(node, env)
	}

	return nil
}

func applyFunction(def object.Object, args []object.Object) object.Object {
	switch def := def.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(def, args)
		evaluated := Eval(def.Body, extendedEnv)
		return unwrapReturnValue(evaluated)

	case *object.Builtin:
		return def.Fn(args...)

	default:
		return newError("not a function: %s", def.Type())
	}
}

func extendFunctionEnv(def *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(def.Env)

	for paramIdx, param := range def.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}

func evalHash(node *ast.Hash, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	case left.Type() == object.STRING_OBJ:
		return evalStringIndexExpression(left, index)
	case left.Type() == object.MODULE_OBJ:
		return evalModuleIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalModuleIndexExpression(module, index object.Object) object.Object {
	moduleObject := module.(*object.Module)

	return evalHashIndexExpression(moduleObject.Attributes, index)
}

func evalImport(ie *ast.Import, env *object.Environment) object.Object {
	name := Eval(ie.Name, env)

	if isError(name) {
		return name
	}

	if s, ok := name.(*object.String); ok {
		attributes := EvalModule(s.Value)

		if isError(attributes) {
			return attributes
		}

		env.Set(filepath.Base(s.Value), &object.Module{Name: s.Value, Attributes: attributes})
		return &object.Null{}
	}

	return newError("Import Error: invalid import path '%s'", name)
}

func evalStringIndexExpression(left, index object.Object) object.Object {
	stringObject := left.(*object.String)
	idx := index.(*object.Integer).Value
	max := int64(len(stringObject.Value) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	return &object.String{Value: string(stringObject.Value[idx])}
}

func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	return arrayObject.Elements[idx]
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := stdlib.Builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: " + node.Value)
}

func evalBlock(block *ast.Block, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalIf(ie *ast.If, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)

	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newError("devision by zero not allowed")
		}
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Float).Value
	rightVal := right.(*object.Float).Value

	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newError("devision by zero not allowed")
		}
		return &object.Float{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case operator == "==":
		return nativeBoolToBooleanObject(object.CompareObjects(left, right))
	case operator == "!=":
		return nativeBoolToBooleanObject(!object.CompareObjects(left, right))
	case object.IsNumber(left) && object.IsNumber(right):
		if left.Type() == right.Type() && operator != "/" {
			if left.Type() == object.INTEGER_OBJ {
				return evalIntegerInfixExpression(operator, left, right)
			} else if left.Type() == object.FLOAT_OBJ {
				return evalFloatInfixExpression(operator, left, right)
			}
		}

		leftOrig, rightOrig := left, right
		if left.Type() == object.INTEGER_OBJ {
			left = left.(*object.Integer).ToFloat()
		}
		if right.Type() == object.INTEGER_OBJ {
			right = right.(*object.Integer).ToFloat()
		}

		result := evalFloatInfixExpression(operator, left, right)

		if object.IsNumber(result) && leftOrig.Type() == object.INTEGER_OBJ && rightOrig.Type() == object.INTEGER_OBJ {
			return result.(*object.Float).TryInteger()
		}
		return result
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case left.Type() == object.ARRAY_OBJ && right.Type() == object.ARRAY_OBJ:
		return evalArrayInfixExpression(operator, left, right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch operator {
	case "+":
		return &object.String{Value: leftVal + rightVal}
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalArrayInfixExpression(operator string, left, right object.Object) object.Object {
	leftArray := left.(*object.Array)
	rightArray := right.(*object.Array)

	switch operator {
	case "+":
		length := len(leftArray.Elements) + len(rightArray.Elements)
		elements := make([]object.Object, length, length)
		copy(elements, leftArray.Elements)
		copy(elements[len(leftArray.Elements):], rightArray.Elements)
		return &object.Array{Elements: elements}
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStatements(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range stmts {
		result = Eval(statement, env)

		if returnValue, ok := result.(*object.ReturnValue); ok {
			return returnValue.Value
		}
	}

	return result
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func evalObjectCall(call *ast.ObjectCall, env *object.Environment) object.Object {
	obj := Eval(call.Object, env)
	if method, ok := call.Call.(*ast.Call); ok {
		args := evalExpressions(call.Call.(*ast.Call).Arguments, env)
		ret := obj.InvokeMethod(method.Callable.String(), *env, args...)
		if ret != nil {
			return ret
		}
	}

	return newError("Failed to invoke method: %s", call.Call.(*ast.Call).Callable.String())
}

func evalForeach(fle *ast.Foreach, env *object.Environment) object.Object {
	val := Eval(fle.Value, env)

	helper, ok := val.(object.Iterable)
	if !ok {
		return newError("%s object doesn't implement the Iterable interface", val.Type())
	}

	var permit []string
	permit = append(permit, fle.Ident)
	if fle.Index != "" {
		permit = append(permit, fle.Index)
	}

	//
	// This will allow writing EVERYTHING to the parent scope,
	// except the two variables named in the permit-array
	child := object.NewTemporaryScope(env, permit)

	helper.Reset()

	ret, idx, ok := helper.Next()

	for ok {

		child.Set(fle.Ident, ret)

		idxName := fle.Index
		if idxName != "" {
			child.Set(fle.Index, idx)
		}

		rt := Eval(fle.Body, child)

		//
		// If we got an error/return then we handle it.
		//
		if rt != nil && !isError(rt) && (rt.Type() == object.RETURN_VALUE_OBJ || rt.Type() == object.ERROR_OBJ) {
			return rt
		}

		ret, idx, ok = helper.Next()
	}

	return val
}

func evalAssign(a *ast.Assign, env *object.Environment) (val object.Object) {
	evaluated := Eval(a.Value, env)
	if isError(evaluated) {
		return evaluated
	}

	env.Set(a.Name.String(), evaluated)
	return evaluated
}

func EvalModule(name string) object.Object {
	filename := utilities.FindModule(name)

	if filename == "" {
		return newError("Import Error: no module named '%s' found", name)
	}

	b, err := ioutil.ReadFile(filename)

	if err != nil {
		return newError("IO Error: error reading module '%s': %s", name, err)
	}

	l := lexer.New(string(b))
	imports := make(map[string]struct{})
	p := parser.New(l, imports)

	module, _ := p.ParseProgram()

	if len(p.Errors()) != 0 {
		return newError("Parse Error: %s", p.Errors())
	}

	env := object.NewEnvironment()
	Eval(module, env)

	return env.Exported()
}
