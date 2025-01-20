package ttir

type Program struct {
	Functions []Function
}

type Function struct {
	Name         string
	Instructions []Instruction
}

type Instruction interface {
	String() string
	instruction()
}

type Ret struct {
	op Operand
}

func (r *Ret) String()      {}
func (r *Ret) instruction() {}

type Operand interface {
	String() string
	operand()
}

type Constant struct {
	Value int64
}

func (c *Constant) operand() {}
