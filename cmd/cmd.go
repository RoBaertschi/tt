package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"robaertschi.xyz/robaertschi/tt/asm/amd64"
	"robaertschi.xyz/robaertschi/tt/lexer"
	"robaertschi.xyz/robaertschi/tt/parser"
	"robaertschi.xyz/robaertschi/tt/token"
	"robaertschi.xyz/robaertschi/tt/ttir"
	"robaertschi.xyz/robaertschi/tt/typechecker"
	"robaertschi.xyz/robaertschi/tt/utils"
)

type ToPrintFlags int

const (
	PrintAst ToPrintFlags = 1 << iota
	PrintTAst
	PrintIr
)

type Arguments struct {
	Output      string
	Input       string
	OnlyEmitAsm bool
	ToPrint     ToPrintFlags
}

func Compile(args Arguments) {
	output := args.Output
	input := args.Input
	onlyEmitAsm := args.OnlyEmitAsm

	asmOutputName := strings.TrimSuffix(input, filepath.Ext(input)) + ".asm"

	file, err := os.Open(input)
	if err != nil {
		fmt.Printf("Could not open file %q because: %e", input, err)
		os.Exit(1)
	}
	defer file.Close()

	inputText, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Could not read file %q because: %e", input, err)
		os.Exit(1)
	}

	l, err := lexer.New(string(inputText), input)
	if err != nil {
		fmt.Printf("Error while creating lexer: %e", err)
		os.Exit(1)
	}

	l.WithErrorCallback(func(l token.Loc, s string, a ...any) {
		fmt.Printf("%s:%d:%d: %s\n", l.File, l.Line, l.Col, fmt.Sprintf(s, a...))
	})

	p := parser.New(l)
	p.WithErrorCallback(func(t token.Token, s string, a ...any) {
		loc := t.Loc
		fmt.Printf("%s:%d:%d: %s\n", loc.File, loc.Line, loc.Col, fmt.Sprintf(s, a...))
	})

	program := p.ParseProgram()
	if p.Errors() > 0 {
		fmt.Printf("Parser encountered 1 or more errors, quiting...\n")
		os.Exit(1)
	}
	if (args.ToPrint & PrintAst) != 0 {
		fmt.Printf("AST:\n%s\n%+#v\n", program.String(), program)
	}

	tprogram, err := typechecker.New().CheckProgram(program)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	if (args.ToPrint & PrintTAst) != 0 {
		fmt.Printf("TAST:\n%s\n%+#v\n", tprogram.String(), tprogram)
	}

	ir := ttir.EmitProgram(tprogram)
	if (args.ToPrint & PrintIr) != 0 {
		fmt.Printf("TTIR:\n%s\n%+#v\n", ir.String(), ir)
	}
	asm := amd64.CgProgram(ir)

	asmOutput := asm.Emit()

	asmOutputFile, err := os.Create(asmOutputName)
	if err != nil {
		fmt.Printf("Failed to create/truncate asm file %q because: %v\n", asmOutputName, err)
		os.Exit(1)
	}

	_, err = asmOutputFile.WriteString(asmOutput)
	asmOutputFile.Close()
	if err != nil {
		fmt.Printf("Failed to write to file %q because: %v\n", asmOutputName, err)
		os.Exit(1)
	}

	fasmPath, err := exec.LookPath("fasm")
	if err != nil {
		fasmPath, err = exec.LookPath("fasm2")
		if err != nil {
			fmt.Printf("Could not find fasm or fasm2, please install any those two using your systems package manager or from https://flatassembler.net\n")
			os.Exit(1)
		}
	}

	if !onlyEmitAsm {
		args := []string{asmOutputName, output}
		cmd := exec.Command(fasmPath, args...)
		cmd.Stdout = utils.NewPrefixWriterString(os.Stdout, "fasm output: ")
		cmd.Stderr = cmd.Stdout

		err = cmd.Run()
		if err != nil {
			fmt.Printf("Failed to run fasm because: %v\nCheck the asm file %q for errors and report these to the author!\n", err, asmOutputName)
			os.Exit(1)
		}

		removeErr := os.Remove(asmOutputName)
		if removeErr != nil {
			fmt.Printf("Failed to remove %q file, please remove it yourself. Err: %v\n", asmOutputName, err)
		}
	}
}
