package amd64

import (
	_ "embed"
	"strings"
	"testing"

	"robaertschi.xyz/robaertschi/tt/ast"
	"robaertschi.xyz/robaertschi/tt/ttir"
)

func expectProgram(t *testing.T, expected Program, actual Program) {
	t.Helper()
	if len(expected.Functions) != len(actual.Functions) {
		t.Errorf("Expected %d functions but got %d", len(expected.Functions), len(actual.Functions))
		return
	}

	for i := range expected.Functions {
		expectFunction(t, expected.Functions[i], actual.Functions[i])
	}
}

func expectFunction(t *testing.T, expected Function, actual Function) {
	t.Helper()
	if expected.Name != actual.Name {
		t.Errorf("Expected function name %q but got %q", expected.Name, actual.Name)
	}

	if len(expected.Instructions) != len(actual.Instructions) {
		t.Errorf("Expected %d instructions but got %d, expected: %v, actual: %v", len(expected.Instructions), len(actual.Instructions), expected.Instructions, actual.Instructions)
		return
	}

	for i := range expected.Instructions {
		expectInstruction(t, expected.Instructions[i], actual.Instructions[i])
	}
}

func expectInstruction(t *testing.T, expected Instruction, actual Instruction) {
	t.Helper()

	switch expected := expected.(type) {
	case *SimpleInstruction:
		actual, ok := actual.(*SimpleInstruction)
		if !ok {
			t.Errorf("Expected SimpleInstruction but got %T", actual)
			return
		}

		if expected.Opcode != actual.Opcode {
			t.Errorf("Expected opcode %q but got %q", expected.Opcode, actual.Opcode)
		}

		switch expected.Opcode {
		case Mov:
			expectOperand(t, expected.Lhs, actual.Lhs)
			expectOperand(t, expected.Rhs, actual.Rhs)
		case Ret:
			// nothing to do
		}
	case *SetCCInstruction:
		actual, ok := actual.(*SetCCInstruction)
		if !ok {
			t.Errorf("Expected SetCCInstruction but got %T", actual)
			return
		}

		if expected.Cond != actual.Cond {
			t.Errorf("Expected condition %q but got %q", expected.Cond, actual.Cond)
		}

		expectOperand(t, expected.Dst, actual.Dst)
	}

}

func expectOperand(t *testing.T, expected Operand, actual Operand) {
	t.Helper()

	switch expected := expected.(type) {
	case Register:
		actual, ok := actual.(Register)
		if !ok {
			t.Errorf("Expected Register but got %T", actual)
			return
		}

		if expected != actual {
			t.Errorf("Expected Register %q but got %q", expected, actual)
		}
	case Imm:
		actual, ok := actual.(Imm)
		if !ok {
			t.Errorf("Expected Immediate but got %T", actual)
			return
		}

		if expected != actual {
			t.Errorf("Expected Immediate %q but got %q", expected, actual)
		}
	case Stack:
		actual, ok := actual.(Stack)

		if !ok {
			t.Errorf("Expected Stack but got %T", actual)
		}

		if expected != actual {
			t.Errorf("Expected Stack value %q but got %q", expected, actual)
		}
	case Pseudo:
		actual, ok := actual.(Pseudo)

		if !ok {
			t.Errorf("Expected Stack but got %T", actual)
		}

		if expected != actual {
			t.Errorf("Expected Stack value %q but got %q", expected, actual)
		}
	default:
		t.Errorf("Unknown operand type %T", expected)
	}
}

func trim(s string) string {
	return strings.Trim(s, " \n\t")
}

func TestOperands(t *testing.T) {
	var op Operand

	op = AX
	if str := op.OperandString(One); str != "al" {
		t.Errorf("The register AX should be \"al\" but got %q", str)
	}

	op = Imm(3)
	if str := op.OperandString(One); str != "3" {
		t.Errorf("The immediate value 3 should be \"3\" but got %q", str)
	}
}

//go:embed basic_test.txt
var basicTest string

func TestCodegen(t *testing.T) {
	program := &ttir.Program{
		Functions: []ttir.Function{
			{
				Name:           "main",
				HasReturnValue: true,
				Instructions: []ttir.Instruction{
					&ttir.Ret{Op: &ttir.Constant{Value: 0}},
				},
			},
		},
	}

	expectedProgram := Program{

		Functions: []Function{
			{
				Name: "main",
				Instructions: []Instruction{
					&SimpleInstruction{
						Opcode: Mov,
						Lhs:    AX,
						Rhs:    Imm(0),
					},
					&SimpleInstruction{
						Opcode: Ret,
					},
				},
				HasReturnValue: true,
			},
		},
	}

	actualProgram := CgProgram(program)
	expectProgram(t, expectedProgram, *actualProgram)

	actual := actualProgram.Emit()
	expected := basicTest
	if trim(actual) != trim(expected) {
		t.Errorf("Expected program to be:\n>>%s<<\nbut got:\n>>%s<<\n", trim(expected), trim(actual))
	}
}

//go:embed binary_test.txt
var binaryTest string

func TestBinary(t *testing.T) {
	program := &ttir.Program{
		Functions: []ttir.Function{
			{
				Name: "main",
				Instructions: []ttir.Instruction{
					&ttir.Binary{
						Lhs:      &ttir.Constant{Value: 3},
						Rhs:      &ttir.Constant{Value: 3},
						Operator: ast.Add,
						Dst:      &ttir.Var{Value: "temp.1"},
					},
					&ttir.Ret{Op: &ttir.Var{Value: "temp.1"}},
				},
				HasReturnValue: true,
			},
		},
	}

	actual := CgProgram(program).Emit()
	if trim(actual) != trim(binaryTest) {
		t.Errorf("Expected program to be:\n>>%s<<\nbut got:\n>>%s<<\n", trim(binaryTest), trim(actual))
	}
}

//go:embed equality_test.txt
var equalityTest string

// There was once a bug with how the cmp instructions were generated, this check should fail if it happens again
func TestEqualityOperators(t *testing.T) {
	program := ttir.Program{
		Functions: []ttir.Function{
			{
				Name:           "main",
				HasReturnValue: false,
				Instructions: []ttir.Instruction{
					&ttir.Binary{
						Lhs:      &ttir.Constant{Value: 5},
						Rhs:      &ttir.Constant{Value: 4},
						Dst:      &ttir.Var{Value: "temp.1"},
						Operator: ast.LessThanEqual,
					},
					&ttir.Binary{
						Lhs:      &ttir.Constant{Value: 5},
						Rhs:      &ttir.Constant{Value: 4},
						Dst:      &ttir.Var{Value: "temp.2"},
						Operator: ast.LessThan,
					},
					&ttir.Binary{
						Lhs:      &ttir.Constant{Value: 5},
						Rhs:      &ttir.Constant{Value: 4},
						Dst:      &ttir.Var{Value: "temp.3"},
						Operator: ast.GreaterThanEqual,
					},
					&ttir.Binary{
						Lhs:      &ttir.Constant{Value: 5},
						Rhs:      &ttir.Constant{Value: 4},
						Dst:      &ttir.Var{Value: "temp.4"},
						Operator: ast.GreaterThan,
					},
					&ttir.Ret{},
				},
			},
		},
	}

	actual := CgProgram(&program).Emit()
	actualTrimmed := trim(actual)
	expectedTrimmed := trim(actualTrimmed)

	if expectedTrimmed != actualTrimmed {
		t.Errorf("Expected program to be:\n>>%s<<\nbut got:\n>>%s<<\n", expectedTrimmed, actualTrimmed)
	}
}
