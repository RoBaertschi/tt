package main

import (
	"flag"
	"fmt"
	"os"
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
}
