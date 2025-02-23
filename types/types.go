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

var types map[string]Type = make(map[string]Type)

func New(id int64, name string) Type {
	typeId := &TypeId{id: id, name: name}
	types[name] = typeId
	return typeId
}

func From(name string) (Type, bool) {
	t, ok := types[name]
	return t, ok
}
