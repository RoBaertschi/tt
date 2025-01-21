package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"robaertschi.xyz/robaertschi/tt/lexer"
	"robaertschi.xyz/robaertschi/tt/parser"
	"robaertschi.xyz/robaertschi/tt/token"
	"robaertschi.xyz/robaertschi/tt/typechecker"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s [flags] input\nPossible flags:\n", os.Args[0])
		flag.PrintDefaults()
	}

	var output string
	flag.StringVar(&output, "o", "", "Output a executable named `file`")
	flag.StringVar(&output, "output", "", "Output a executable named `file`")
	flag.Parse()

	input := flag.Arg(0)
	if input == "" {
		flag.Usage()
		os.Exit(1)
	}

	if output == "" {
		output = strings.TrimRight(input, filepath.Ext(input))
	}

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

	tprogram, err := typechecker.New().CheckProgram(program)
	if err != nil {
		fmt.Printf("Typechecker failed with %e\n", err)
		os.Exit(1)
	}
}
