package types

type Type interface {
	// Checks if the two types are the same
	IsSameType(Type) bool
	Name() string
}

type TypeId struct {
	id   int64
	name string
}

const (
	I64Id int64 = iota
)

var (
	I64 = New(I64Id, "i64")
)

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
