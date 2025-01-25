package types

import "robaertschi.xyz/robaertschi/tt/ast"

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
	I64Id int64 = iota
	BoolId
)

var (
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

func New(id int64, name string) Type {
	return &TypeId{id: id, name: name}
}
