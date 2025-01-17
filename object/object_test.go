package object_test

import (
	"errors"

	"github.com/flipez/rocket-lang/evaluator"
	"github.com/flipez/rocket-lang/lexer"
	"github.com/flipez/rocket-lang/object"
	"github.com/flipez/rocket-lang/parser"

	"testing"
)

type inputTestCase struct {
	input    string
	expected interface{}
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	imports := make(map[string]struct{})
	p := parser.New(l, imports)
	program, _ := p.ParseProgram()
	env := object.NewEnvironment()

	return evaluator.Eval(program, env)
}

func testInput(t *testing.T, tests []inputTestCase) {
	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case float64:
			testFloatObject(t, evaluated, float64(expected))
		case string:
			arrObj, ok := evaluated.(*object.Array)
			if ok {
				testStringObject(t, object.NewString(arrObj.Inspect()), expected)
				continue
			}
			strObj, ok := evaluated.(*object.String)
			if ok {
				testStringObject(t, strObj, expected)
				continue
			}
			_, ok = evaluated.(*object.Null)
			if ok {
				continue
			}

			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)", evaluated, evaluated)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q", expected, errObj.Message)
			}
		case bool:
			testBooleanObject(t, evaluated, expected)
		}
	}
}

func TestIsError(t *testing.T) {
	trueErrors := []object.Object{
		object.NewError(errors.New("test error")),
		object.NewError("test error"),
		object.NewErrorFormat("test %s", "error"),
	}

	for _, err := range trueErrors {
		if !object.IsError(err) {
			t.Errorf("'%s' should be an error", err.Inspect())
		}
	}

	falseErrors := []object.Object{
		nil,
		object.NewString("a"),
		object.NULL,
	}
	for _, err := range falseErrors {
		if object.IsError(nil) {
			t.Errorf("'%#v' is not an error", err)
		}
	}
}

func TestIsNumber(t *testing.T) {
	if !object.IsNumber(object.NewInteger(1)) {
		t.Error("INTEGER_OBJ should be a number")
	}
	if !object.IsNumber(object.NewFloat(1.1)) {
		t.Error("FLOAT_OBJ should be a number")
	}
	if object.IsNumber(object.NULL) {
		t.Error("NULL_OBJ is not a number")
	}
}

func TestIsTruthy(t *testing.T) {
	if !object.IsTruthy(object.TRUE) {
		t.Error("BOOLEAN_OBJ=true should be truthy")
	}
	if !object.IsTruthy(object.NewString("")) {
		t.Error("STRING_OBJ should be truthy")
	}
	if object.IsTruthy(object.NULL) {
		t.Error("NULL_OBJ should not be truthy")
	}
	if object.IsTruthy(object.FALSE) {
		t.Errorf("BOOLEAN_OBJ=false, should not be truthy")
	}
}

func TestIsFalsy(t *testing.T) {
	if object.IsFalsy(object.TRUE) {
		t.Error("BOOLEAN_OBJ=true should not be falsy")
	}
	if object.IsFalsy(object.NewString("")) {
		t.Error("STRING_OBJ should not be falsy")
	}
	if !object.IsFalsy(object.NULL) {
		t.Error("NULL_OBJ should be falsy")
	}
	if !object.IsFalsy(object.FALSE) {
		t.Errorf("BOOLEAN_OBJ=false, should be falsy")
	}
}
