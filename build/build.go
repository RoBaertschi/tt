// Build allows the user to build a tt file. The goal is to make it easy to support multiple backends with different requirements
package build

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"robaertschi.xyz/robaertschi/tt/asm"
	"robaertschi.xyz/robaertschi/tt/asm/qbe"
	"robaertschi.xyz/robaertschi/tt/utils"
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
	l := utils.NewLogger(os.Stderr, "[build] ", utils.Info)

	nodes := make(map[int]*node)
	rootNodes := []int{}
	id := 0

	addRootNode := func(task task) int {
		l.Debugf("registering root task %d", id)
		node := &node{task: task}
		nodes[id] = node
		rootNodes = append(rootNodes, id)
		id += 1
		return id - 1
	}

	addNode := func(task task, deps ...int) int {
		l.Debugf("registering task %d", id)
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

	if backend == asm.Fasm {
		err := sp.buildFasm(addRootNode, addNode, emitAsmOnly, toPrint)
		if err != nil {
			return err
		}
	} else if backend == asm.Qbe {
		err := sp.buildQbe(addRootNode, addNode, emitAsmOnly, toPrint)
		if err != nil {
			return err
		}
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

	asmFile := addRootNode(NewFuncTask("generating assembly for "+sp.InputFile, func(output io.Writer) error {
		return build(output, sp.InputFile, mainAsmOutput, toPrint, asm.Fasm)
	}))

	if !emitAsmOnly {
		task := NewProcessTask(fasmPath, mainAsmOutput, sp.OutputFile)
		task.WithName("assembling " + mainAsmOutput)
		fasmTask := addNode(task, asmFile)

		// Cleanup

		addNode(NewRemoveFileTask(mainAsmOutput), fasmTask)
	}

	return nil
}

func (sp *SourceProgram) buildQbe(addRootNode func(task) int, addNode func(task, ...int) int, emitAsmOnly bool, toPrint ToPrintFlags) error {
	fasmPath, err := exec.LookPath("fasm")
	if err != nil {
		fasmPath, err = exec.LookPath("fasm2")
		if err != nil {
			return fmt.Errorf("could not find fasm or fasm2, please install any those two using your systems package manager or from https://flatassembler.net")
		}
	}
	qbePath, err := exec.LookPath("qbe")
	if err != nil {
		return fmt.Errorf("could not find qbe, please install using your systems package manager or from https://https://c9x.me/compile")
	}

	asPath, err := exec.LookPath("as")
	if err != nil {
		return fmt.Errorf("could not find the system `as` assembler, please install it using your systems package manager")
	}

	ldPath, err := exec.LookPath("ld")
	if err != nil {
		return fmt.Errorf("could not find the system `ld` linker, please install it using your systems package manager")
	}

	mainAsmOutput := strings.TrimSuffix(sp.InputFile, filepath.Ext(sp.InputFile)) + ".qbe"

	asmFile := addRootNode(NewFuncTask("generating assembly for "+sp.InputFile, func(output io.Writer) error {
		return build(output, sp.InputFile, mainAsmOutput, toPrint, asm.Qbe)
	}))

	if !emitAsmOnly {

		objectFileTasks := []int{}
		qbeStubAsm := "qbe_stub.asm"
		qbeStubO := "qbe_stub.o"
		generatedAsmFile := addRootNode(NewCreateFileTask(qbeStubAsm, qbe.Stub))
		id := addNode(NewProcessTask(fasmPath, qbeStubAsm, qbeStubO), generatedAsmFile)
		objectFileTasks = append(objectFileTasks, id)
		sp.ObjectFiles = append(sp.ObjectFiles, qbeStubO)

		qbeOutput := strings.TrimSuffix(mainAsmOutput, filepath.Ext(mainAsmOutput)) + ".S"
		task := NewProcessTask(qbePath, mainAsmOutput, "-o", qbeOutput)
		task.WithName("running qbe on " + mainAsmOutput)
		qbeTask := addNode(task, asmFile)
		sp.InputAssemblies = append(sp.InputAssemblies, qbeOutput)

		for _, asmFile := range sp.InputAssemblies {
			outputFile := strings.TrimSuffix(asmFile, filepath.Ext(asmFile)) + ".o"

			if filepath.Ext(asmFile) == ".asm" {
				id := addRootNode(NewProcessTask(fasmPath, asmFile, outputFile))
				objectFileTasks = append(objectFileTasks, id)
			} else if filepath.Ext(asmFile) == ".S" {
				id := addRootNode(NewProcessTask(asPath, asmFile, "-o", outputFile))
				objectFileTasks = append(objectFileTasks, id)
			} else {
				panic(fmt.Sprintf("unkown asm file extension %q", filepath.Ext(asmFile)))
			}

			sp.ObjectFiles = append(sp.ObjectFiles, outputFile)
		}
		ldTask := NewProcessTask(ldPath, append([]string{"-o", sp.OutputFile}, sp.ObjectFiles...)...)
		ldTaskId := addNode(ldTask, append([]int{generatedAsmFile}, objectFileTasks...)...)

		for _, object := range sp.ObjectFiles {
			// Cleanup object files
			addNode(NewRemoveFileTask(object), ldTaskId)
		}

		// Cleanup
		addNode(NewRemoveFileTask(mainAsmOutput), ldTaskId, qbeTask)
		addNode(NewRemoveFileTask(qbeStubAsm), ldTaskId, generatedAsmFile)
		addNode(NewRemoveFileTask(qbeOutput), ldTaskId)
	}

	return nil
}
