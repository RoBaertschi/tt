package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"robaertschi.xyz/robaertschi/tt/asm"
	"robaertschi.xyz/robaertschi/tt/build"
	"robaertschi.xyz/robaertschi/tt/term"
)

func main() {
	r, c, err := term.GetCursorPosition()
	fmt.Printf("%d, %d, %v\n", r, c, err)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s [flags] input\nPossible flags:\n", os.Args[0])
		flag.PrintDefaults()
	}

	var output string
	flag.StringVar(&output, "o", "", "Output a executable named `file`")
	flag.StringVar(&output, "output", "", "Output a executable named `file`")
	emitAsmOnly := flag.Bool("S", false, "Only emit the asembly file and exit")

	printAst := flag.Bool("ast", false, "Print the AST out to stdout")
	printTAst := flag.Bool("tast", false, "Print the typed AST out to stdout")
	printIr := flag.Bool("ttir", false, "Print the TTIR out to stdout")
	flag.Parse()

	input := flag.Arg(0)
	if input == "" {
		flag.Usage()
		os.Exit(1)
	}

	if output == "" {
		output = strings.TrimSuffix(input, filepath.Ext(input))
	}

	var toPrint build.ToPrintFlags
	if *printAst {
		toPrint |= build.PrintAst
	}
	if *printTAst {
		toPrint |= build.PrintTAst
	}
	if *printIr {
		toPrint |= build.PrintIr
	}

	logger := log.New(os.Stderr, "", log.Lshortfile)

	err = build.NewSourceProgram(input, output).Build(asm.Fasm, *emitAsmOnly, build.ToPrintFlags(toPrint))
	if err != nil {
		logger.Fatalln(err)
		os.Exit(1)
	}
}
