// Build allows the user to build a tt file. The goal is to make it easy to support multiple backends with different requirements
package build

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"robaertschi.xyz/robaertschi/tt/asm"
	"robaertschi.xyz/robaertschi/tt/asm/amd64"
	"robaertschi.xyz/robaertschi/tt/lexer"
	"robaertschi.xyz/robaertschi/tt/parser"
	"robaertschi.xyz/robaertschi/tt/token"
	"robaertschi.xyz/robaertschi/tt/ttir"
	"robaertschi.xyz/robaertschi/tt/typechecker"
	"robaertschi.xyz/robaertschi/tt/utils"
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
	// The linkded executable
	OutputFile string
}

type task interface {
	Run(id int, doneChan chan taskResult)
}

type processTask struct {
	name string
	args []string
}

func NewProcessTask(id int, name string, args ...string) task {
	return &processTask{
		name: name,
		args: args,
	}
}

func (pt *processTask) Run(id int, doneChan chan taskResult) {
	cmd := exec.Command(pt.name, pt.args...)
	cmd.Stdout = utils.NewPrefixWriterString(os.Stdout, pt.name+" output: ")
	cmd.Stderr = cmd.Stdout

	err := cmd.Run()
	var exitError error
	if cmd.ProcessState.ExitCode() != 0 {
		exitError = fmt.Errorf("command %q failed with exit code %d", pt.name, cmd.ProcessState.ExitCode())
	}
	doneChan <- taskResult{
		Id:  id,
		Err: errors.Join(err, exitError),
	}
}

type removeFileTask struct {
	file string
}

func NewRemoveFileTask(file string) task {
	return &removeFileTask{file: file}
}

func (rft *removeFileTask) Run(id int, doneChan chan taskResult) {
	doneChan <- taskResult{
		Id:  id,
		Err: os.Remove(rft.file),
	}
}

type funcTask struct {
	f func() error
}

func NewFuncTask(f func() error) task {
	return &funcTask{f: f}
}

func (rft *funcTask) Run(id int, doneChan chan taskResult) {
	doneChan <- taskResult{
		Id:  id,
		Err: rft.f(),
	}
}

type taskResult struct {
	Id  int
	Err error
}

type node struct {
	next     *int
	previous []int
	task     task
}

func runTasks(nodes map[int]*node, rootNodes []int) {
	done := make(map[int]bool)
	running := []int{}
	doneChan := make(chan taskResult)

	for id, node := range nodes {
		// Initalize map
		done[id] = false

		// because we are already going trough the whole map, we might as well
		// check the relations
		if node.next != nil {
			if _, ok := nodes[*node.next]; !ok {
				panic(fmt.Sprintf("task with id %d has a invalid next node", id))
			}
		}

		for _, prev := range node.previous {
			if _, ok := nodes[prev]; !ok {
				panic(fmt.Sprintf("task with id %d has a invalid prev node", id))
			}
		}
	}

	for _, rootNode := range rootNodes {
		node := nodes[rootNode]
		go node.task.Run(rootNode, doneChan)
		running = append(running, rootNode)
	}

	for {
		select {
		case result := <-doneChan:
			done[result.Id] = true
			for i, id := range running {
				if id == result.Id {
					running = make([]int, len(running)-1)
					running = append(running[:i], running[i+1:]...)
				}
			}

			node := nodes[result.Id]
			if node.next != nil {
				allDone := true
				for _, prev := range node.previous {
					if !done[prev] {
						allDone = false
						break
					}
				}
				if allDone {
					node := nodes[*node.next]
					go node.task.Run(*node.next, doneChan)
					running = append(running, *node.next)
				}
			}
		}
	}

}

func build(input string, output string) error {
	file, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("could not open file %q because: %v", input, err)
	}
	defer file.Close()

	inputText, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("Could not read file %q because: %v", input, err)
	}

	l, err := lexer.New(string(inputText), input)
	if err != nil {
		return fmt.Errorf("error while creating lexer: %v", err)
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
		return fmt.Errorf("parser encountered 1 or more errors")
	}

	tprogram, err := typechecker.New().CheckProgram(program)
	if err != nil {
		return err
	}

	ir := ttir.EmitProgram(tprogram)
	asm := amd64.CgProgram(ir)

	asmOutput := asm.Emit()

	asmOutputFile, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create/truncate asm file %q because: %v\n", output, err)
	}

	_, err = asmOutputFile.WriteString(asmOutput)
	asmOutputFile.Close()
	if err != nil {
		return fmt.Errorf("failed to write to file %q because: %v\n", output, err)
	}

	return nil
}

func (sp *SourceProgram) Build(backend asm.Backend) error {
	var gasPath string
	fasmPath, err := exec.LookPath("fasm")
	if err != nil {
		fasmPath, err = exec.LookPath("fasm2")
		if err != nil {
			return fmt.Errorf("could not find fasm or fasm2, please install any those two using your systems package manager or from https://flatassembler.net")
		}
	}
	if backend == asm.Qbe {
		gasPath, err = exec.LookPath("as")
		if err != nil {
			return fmt.Errorf("could not find as assembler, pleas install it")
		}
	}

	nodes := make(map[int]*node)
	rootNodes := []int{}
	id := 0

	addRootNode := func(task task) int {
		node := &node{task: task}
		nodes[id] = node
		rootNodes = append(rootNodes, id)
		id += 1
		return id - 1
	}

	addNode := func(task task, deps ...int) int {
		node := &node{task: task}
		nodes[id] = node

		for _, dep := range deps {
			nodeDep := nodes[dep]
			if nodeDep.next != nil {
				panic(fmt.Sprintf("dep %d already has an next", dep))
			}
			newId := id
			nodeDep.next = &newId
			node.previous = append(node.previous, dep)
		}

		rootNodes = append(rootNodes, id)
		id += 1
		return id - 1
	}

	var mainFile int
	switch backend {
	case asm.Fasm:
		mainAsmOutput := strings.TrimSuffix(sp.InputFile, filepath.Ext(sp.InputFile)) + ".asm"
		sp.InputAssemblies = append(sp.InputAssemblies, mainAsmOutput)
		mainFile = addRootNode(NewFuncTask(func() error {
			return build(sp.InputFile, mainAsmOutput)
		}))
	case asm.Qbe:
		panic("qbe support not finished")
	}

	asmFiles := []int{}
	for _, file := range sp.InputAssemblies {
		output := strings.TrimSuffix(file, filepath.Ext(file)) + ".o"
		if filepath.Ext(file) == ".asm" {
			asmFiles = append(asmFiles, addNode(
				NewProcessTask(id, fasmPath, file, output), mainFile,
			))
		} else {
			asmFiles = append(asmFiles, addNode(NewProcessTask(id, gasPath, file, "-o", output)))
		}
		sp.ObjectFiles = append(sp.ObjectFiles, output)
	}

	return runTasks(nodes, rootNodes)
}
