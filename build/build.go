// Build allows the user to build a tt file. The goal is to make it easy to support multiple backends with different requirements
package build

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"robaertschi.xyz/robaertschi/tt/asm"
)

type ToPrintFlags int

const (
	PrintAst ToPrintFlags = 1 << iota
	PrintTAst
	PrintIr
)

type SourceProgram struct {
	// The tt source file
	InputFile string
	// A list of additional assembly files to compile
	// This file could be extended by different backends
	// .asm is for fasm, .S for gas
	InputAssemblies []string
	// Additional object files, will also include the generated object files
	ObjectFiles []string
	// The linked executable
	OutputFile string
}

func NewSourceProgram(inputFile string, outputFile string) *SourceProgram {
	return &SourceProgram{InputFile: inputFile, OutputFile: outputFile}
}

func (sp *SourceProgram) Build(backend asm.Backend, emitAsmOnly bool, toPrint ToPrintFlags) error {
	l := log.New(os.Stderr, "[build] ", log.Lshortfile)

	nodes := make(map[int]*node)
	rootNodes := []int{}
	id := 0

	addRootNode := func(task task) int {
		l.Printf("registering root task %d", id)
		node := &node{task: task}
		nodes[id] = node
		rootNodes = append(rootNodes, id)
		id += 1
		return id - 1
	}

	addNode := func(task task, deps ...int) int {
		l.Printf("registering task %d", id)
		if len(deps) <= 0 {
			panic("node without dep is useless")
		}

		node := &node{task: task}
		nodes[id] = node

		for _, dep := range deps {
			nodeDep := nodes[dep]
			nodeDep.next = append(nodeDep.next, id)
			node.previous = append(node.previous, dep)
		}

		id += 1
		return id - 1
	}

	err := sp.buildFasm(addRootNode, addNode, emitAsmOnly, toPrint)
	if err != nil {
		return err
	}

	return runTasks(nodes, rootNodes, l)
}

func (sp *SourceProgram) buildFasm(addRootNode func(task) int, addNode func(task, ...int) int, emitAsmOnly bool, toPrint ToPrintFlags) error {
	fasmPath, err := exec.LookPath("fasm")
	if err != nil {
		fasmPath, err = exec.LookPath("fasm2")
		if err != nil {
			return fmt.Errorf("could not find fasm or fasm2, please install any those two using your systems package manager or from https://flatassembler.net")
		}
	}

	mainAsmOutput := strings.TrimSuffix(sp.InputFile, filepath.Ext(sp.InputFile)) + ".asm"

	asmFile := addRootNode(NewFuncTask(func() error {
		return build(sp.InputFile, mainAsmOutput, toPrint)
	}))

	if !emitAsmOnly {
		fasmTask := addNode(NewProcessTask(fasmPath, mainAsmOutput, sp.OutputFile), asmFile)

		// Cleanup

		addNode(NewRemoveFileTask(mainAsmOutput), fasmTask)
	}

	return nil
}
