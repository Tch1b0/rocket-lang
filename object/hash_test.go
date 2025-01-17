package object_test

import (
	"testing"

	"github.com/flipez/rocket-lang/object"
)

func TestHashObject(t *testing.T) {
	tests := []inputTestCase{
		{`{"a": 1} == {"a": 1}`, true},
		{`{"a": 1} == {"a": 1, "b": 2}`, false},
		{`{"a": 1} == {"b": 1}`, false},
		{`{"a": 1} == {"a": "c"}`, false},
		{`{{1: true}: "a"}.keys()`, `[{1: true}]`},
	}

	testInput(t, tests)
}

func TestHashObjectMethods(t *testing.T) {
	tests := []inputTestCase{
		{`{"a": 2}.keys()`, `["a"]`},
		{`{}.nope()`, "undefined method `.nope()` for HASH"},
		{`({}.wat().lines().size() == {}.methods().size() + 1).plz_s()`, "true"},
		{`{}.type()`, "HASH"},
		{`a = {"a": "b", "b":"a"};b = []; foreach key, value in a { b.yoink(key) }; b.size()`, 2},
		{`{"a": 1, "b": 2}["a"]`, 1},
		{`{"a": 1, "b": 2}.keys().size()`, 2},
		{`{"a": 1, "b": 2}.values().size()`, 2},
	}

	testInput(t, tests)
}

func TestHashInspect(t *testing.T) {
	tests := []inputTestCase{
		{"{}", "{}"},
		{`{"a": 1}`, `{"a": 1}`},
		{`{true: "a"}`, `{true: "a"}`},
	}

	for _, tt := range tests {
		hash := testEval(tt.input).(*object.Hash)
		hashInspect := hash.Inspect()

		if hash.Inspect() != tt.expected {
			t.Errorf("wrong string. expected=%#v, got=%#v", tt.expected, hashInspect)
		}
	}
}

func TestHashType(t *testing.T) {
	hash1 := object.NewHash(nil)

	if hash1.Type() != object.HASH_OBJ {
		t.Errorf("hash.Type() returns wrong type")
	}
}
