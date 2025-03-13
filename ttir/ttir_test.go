package ttir

import (
	"fmt"
	"testing"

	"robaertschi.xyz/robaertschi/tt/ast"
	"robaertschi.xyz/robaertschi/tt/lexer"
	"robaertschi.xyz/robaertschi/tt/parser"
	"robaertschi.xyz/robaertschi/tt/token"
	"robaertschi.xyz/robaertschi/tt/typechecker"
)

type ttirEmitterTest struct {
	input    string
	expected Program
}

func runTTIREmitterTest(t *testing.T, test ttirEmitterTest) {
	t.Helper()

	l, err := lexer.New(test.input, "test.tt")
	l.WithErrorCallback(func(l token.Loc, s string, a ...any) {
		format := fmt.Sprintf(s, a...)
		t.Errorf("Lexer error callback called: %s:%d:%d %s", l.File, l.Line, l.Col, format)
	})
	if err != nil {
		t.Fatalf("lexer error: %q", err)
	}

	p := parser.New(l)
	p.WithErrorCallback(func(tok token.Token, s string, a ...any) {
		format := fmt.Sprintf(s, a...)
		t.Errorf("Parser error callback called: %s:%d:%d %s", tok.Loc.File, tok.Loc.Line, tok.Loc.Col, format)
	})
	program := p.ParseProgram()
	tprogram, err := typechecker.New().CheckProgram(program)

	if err != nil {
		t.Fatalf("typechecker error: %q", err)
	}

	ttir := EmitProgram(tprogram)

	expectProgram(t, &test.expected, ttir)
}

func expectProgram(t *testing.T, expected *Program, actual *Program) {
	t.Helper()

	if len(expected.Functions) != len(actual.Functions) {
		t.Errorf("expected %d functions , got %d", len(expected.Functions), len(actual.Functions))
		return
	}

	for i, decl := range expected.Functions {
		expectFunction(t, decl, actual.Functions[i])
	}
}

func expectFunction(t *testing.T, expected *Function, actual *Function) {
	t.Helper()

	if expected.Name != actual.Name {
		t.Errorf("expected name %q, got %q", expected.Name, actual.Name)
	}

	if len(expected.Instructions) != len(actual.Instructions) {
		t.Errorf("expected %d instructions, got %d", len(expected.Instructions), len(actual.Instructions))
		return
	}

	for i, inst := range expected.Instructions {
		expectInstruction(t, inst, actual.Instructions[i])
	}
}

func expectInstruction(t *testing.T, inst Instruction, actual Instruction) {
	t.Helper()
	switch inst := inst.(type) {
	case *Ret:
		ret, ok := actual.(*Ret)

		if !ok {
			t.Errorf("expected inst to be *Ret, but got %T", actual)
			return
		}

		expectOperand(t, inst.Op, ret.Op)
	case *Binary:
		binary, ok := actual.(*Binary)

		if !ok {
			t.Errorf("expected inst to be %T, but got %T", inst, binary)
			return
		}

		expectOperand(t, inst.Lhs, binary.Lhs)
		expectOperand(t, inst.Rhs, binary.Rhs)
		expectOperand(t, inst.Dst, binary.Dst)
	}
}

func expectOperand(t *testing.T, expected Operand, actual Operand) {
	t.Helper()

	switch expected := expected.(type) {
	case *Constant:
		constant, ok := actual.(*Constant)

		if !ok {
			t.Errorf("expected operand to be *Constant, but got %T", actual)
			return
		}

		if expected.Value != constant.Value {
			t.Errorf("expected *Constant.Value to be %d, but got %d", expected.Value, constant.Value)
		}
	case *Var:
		v, ok := actual.(*Var)

		if !ok {
			t.Errorf("expected operand to be %T, but got %T", expected, actual)
			return
		}
		if expected.Value != v.Value {
			t.Errorf("expected var to be %q, but got %q", expected.Value, v.Value)
		}
	}
}

func TestBasicFunction(t *testing.T) {
	runTTIREmitterTest(t, ttirEmitterTest{
		input: "fn main(): i64 = 0;",
		expected: Program{
			Functions: []*Function{
				{
					Name: "main",
					Instructions: []Instruction{
						&Ret{
							Op: &Constant{Value: 0},
						},
					},
				},
			},
		},
	})
}

func TestBinaryExpression(t *testing.T) {
	runTTIREmitterTest(t, ttirEmitterTest{
		input: "fn main(): i64 = 3 + 3 + 3;",
		expected: Program{
			Functions: []*Function{
				{Name: "main", Instructions: []Instruction{
					&Binary{Operator: ast.Add, Lhs: &Constant{Value: 3}, Rhs: &Constant{Value: 3}, Dst: &Var{Value: "temp.1"}},
					&Binary{Operator: ast.Add, Lhs: &Var{Value: "temp.1"}, Rhs: &Constant{Value: 3}, Dst: &Var{Value: "temp.2"}},
					&Ret{Op: &Var{Value: "temp.2"}},
				}},
			},
		},
	})
}
