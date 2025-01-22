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
	flag.Parse()

	input := flag.Arg(0)
	if input == "" {
		flag.Usage()
		os.Exit(1)
	}

	if output == "" {
		output = strings.TrimSuffix(input, filepath.Ext(input))
	}

	cmd.Compile(cmd.Arguments{Output: output, Input: input, OnlyEmitAsm: *onlyEmitAsm})
}
