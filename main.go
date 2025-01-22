package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"robaertschi.xyz/robaertschi/tt/cmd"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s [flags] input\nPossible flags:\n", os.Args[0])
		flag.PrintDefaults()
	}

	var output string
	flag.StringVar(&output, "o", "", "Output a executable named `file`")
	flag.StringVar(&output, "output", "", "Output a executable named `file`")
	onlyEmitAsm := flag.Bool("S", false, "Only emit the asembly file and exit")

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

	var toPrint cmd.ToPrintFlags
	if *printAst {
		toPrint |= cmd.PrintAst
	}
	if *printTAst {
		toPrint |= cmd.PrintTAst
	}
	if *printIr {
		toPrint |= cmd.PrintIr
	}

	cmd.Compile(cmd.Arguments{Output: output, Input: input, OnlyEmitAsm: *onlyEmitAsm, ToPrint: toPrint})
}
