package types

import (
	"strings"

	"robaertschi.xyz/robaertschi/tt/ast"
)

type Type interface {
	// Checks if the two types are the same
	IsSameType(Type) bool
	SupportsBinaryOperator(op ast.BinaryOperator) bool
	Name() string
}

type TypeId struct {
	id   int64
	name string
}

const (
	UnitId int64 = iota
	I64Id
	BoolId
)

var (
	Unit = New(UnitId, "()")
	I64  = New(I64Id, "i64")
	Bool = New(BoolId, "bool")
)

func (ti *TypeId) SupportsBinaryOperator(op ast.BinaryOperator) bool {
	if ti == Bool && !op.IsBooleanOperator() {
		return false
	}
	return true
}

func (ti *TypeId) IsSameType(t Type) bool {
	if ti2, ok := t.(*TypeId); ok {
		return ti.id == ti2.id
	}

	return false
}

func (ti *TypeId) Name() string {
	return ti.name
}

type FunctionType struct {
	ReturnType Type
	Parameters []Type
}

func (ft *FunctionType) SupportsBinaryOperator(op ast.BinaryOperator) bool {
	return false
}

func (ft *FunctionType) IsSameType(t Type) bool {
	if ft2, ok := t.(*FunctionType); ok {
		if !ft.ReturnType.IsSameType(ft2.ReturnType) {
			return false
		}

		if len(ft.Parameters) != len(ft2.Parameters) {
			return false
		}

		for i, t := range ft.Parameters {
			if !t.IsSameType(ft2.Parameters[i]) {
				return false
			}
		}
	}
	return false
}

func (ft *FunctionType) Name() string {
	b := strings.Builder{}

	b.WriteString("fn(")

	for i, param := range ft.Parameters {
		b.WriteString(param.Name())
		if i < (len(ft.Parameters) - 1) {
			b.WriteRune(',')
		}
	}

	b.WriteString("): " + ft.ReturnType.Name())

	return b.String()
}

var types map[string]Type = make(map[string]Type)

func New(id int64, name string) Type {
	typeId := &TypeId{id: id, name: name}
	types[name] = typeId
	return typeId
}

func From(name ast.Type) (Type, bool) {
	t, ok := types[string(name)]
	return t, ok
}
