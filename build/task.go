package build

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"

	"robaertschi.xyz/robaertschi/tt/asm"
	"robaertschi.xyz/robaertschi/tt/asm/amd64"
	"robaertschi.xyz/robaertschi/tt/asm/qbe"
	"robaertschi.xyz/robaertschi/tt/lexer"
	"robaertschi.xyz/robaertschi/tt/parser"
	"robaertschi.xyz/robaertschi/tt/token"
	"robaertschi.xyz/robaertschi/tt/ttir"
	"robaertschi.xyz/robaertschi/tt/typechecker"
	"robaertschi.xyz/robaertschi/tt/utils"
)

type task interface {
	Run(id int, output io.Writer, doneChan chan taskResult)
	Name() string
	WithName(string)
}

type processTask struct {
	taskName string
	name     string
	args     []string
}

func NewProcessTask(name string, args ...string) task {
	return &processTask{
		name:     name,
		taskName: fmt.Sprintf("starting %q %v\n", name, args),
		args:     args,
	}
}

func (pt *processTask) WithName(name string) {
	pt.taskName = name
}

func (pt *processTask) Run(id int, output io.Writer, doneChan chan taskResult) {
	cmd := exec.Command(pt.name, pt.args...)
	cmd.Stdout = utils.NewPrefixWriterString(output, pt.name+" output: ")
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

func (pt *processTask) Name() string {
	return pt.taskName
}

type removeFileTask struct {
	file string
	name string
}

func NewRemoveFileTask(file string) task {
	return &removeFileTask{file: file, name: fmt.Sprintf("removing file %q", file)}
}

func (rft *removeFileTask) Run(id int, output io.Writer, doneChan chan taskResult) {
	err := os.Remove(rft.file)
	doneChan <- taskResult{
		Id:  id,
		Err: err,
	}
}

func (rft *removeFileTask) Name() string { return rft.name }

func (rft *removeFileTask) WithName(name string) { rft.name = name }

type createFileTask struct {
	file    string
	content string
	name    string
}

func NewCreateFileTask(file string, content string) task {
	return &createFileTask{
		file:    file,
		content: content,
		name:    fmt.Sprintf("writing file %q", file),
	}
}

func (cft *createFileTask) Run(id int, output io.Writer, doneChan chan taskResult) {
	file, err := os.Create(cft.file)
	if err != nil {
		doneChan <- taskResult{
			Id:  id,
			Err: err,
		}
		return
	}

	_, err = file.WriteString(cft.content)
	doneChan <- taskResult{
		Id:  id,
		Err: err,
	}
	return
}

func (cft *createFileTask) Name() string { return cft.name }

func (cft *createFileTask) WithName(name string) { cft.name = name }

type funcTask struct {
	f    func(io.Writer) error
	name string
}

func NewFuncTask(taskName string, f func(io.Writer) error) task {
	return &funcTask{f: f, name: taskName}
}

func (rft *funcTask) Run(id int, output io.Writer, doneChan chan taskResult) {
	doneChan <- taskResult{
		Id:  id,
		Err: rft.f(output),
	}
}

func (rft *funcTask) Name() string {
	return rft.name
}

func (rft *funcTask) WithName(name string) {
	rft.name = name
}

func build(outputWriter io.Writer, input string, output string, toPrint ToPrintFlags, backend asm.Backend) error {
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
	if (toPrint & PrintAst) != 0 {
		io.WriteString(outputWriter,
			fmt.Sprintf("AST:\n%s\n%+#v\n", program.String(), program))
	}

	tprogram, err := typechecker.New().CheckProgram(program)
	if err != nil {
		return err
	}
	if (toPrint & PrintTAst) != 0 {
		io.WriteString(outputWriter,
			fmt.Sprintf("TAST:\n%s\n%+#v\n", tprogram.String(), tprogram))
	}

	ir := ttir.EmitProgram(tprogram)
	if (toPrint & PrintIr) != 0 {
		io.WriteString(outputWriter,
			fmt.Sprintf("TTIR:\n%s\n%+#v\n", ir.String(), ir))
	}

	asmOutputFile, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create/truncate asm file %q because: %v\n", output, err)
	}
	defer asmOutputFile.Close()

	if backend == asm.Fasm {
		asm := amd64.CgProgram(ir)
		asmOutput := asm.Emit()
		_, err = asmOutputFile.WriteString(asmOutput)
		if err != nil {
			return fmt.Errorf("failed to write to file %q because: %v\n", output, err)
		}
	} else if backend == asm.Qbe {
		err := qbe.Emit(asmOutputFile, ir)
		if err != nil {
			return fmt.Errorf("failed to write to file %q because: %v\n", output, err)
		}

	}

	return nil
}

type taskResult struct {
	Id  int
	Err error
}

type node struct {
	next     []int
	previous []int
	task     task
}

type executionState int

const (
	notStarted executionState = iota
	executing
	finished
	failed
)

func runTasks(nodes map[int]*node, rootNodes []int, l *utils.Logger) error {

	done := make(map[int]executionState)
	output := make(map[int]*strings.Builder)
	running := []int{}
	doneChan := make(chan taskResult)
	errs := []error{}

	startTask := func(id int) {
		if done[id] != notStarted {
			panic(fmt.Sprintf("tried starting task %d twice", id))
		}
		l.Debugf("executing task %d", id)
		node := nodes[id]
		output[id] = &strings.Builder{}
		go node.task.Run(id, output[id], doneChan)
		running = append(running, id)
		done[id] = executing
	}

	for id, node := range nodes {
		// Initalize map
		done[id] = notStarted

		// because we are already going trough the whole map, we might as well
		// check the relations
		for _, next := range node.next {
			if node.next != nil {
				if _, ok := nodes[next]; !ok {
					panic(fmt.Sprintf("task with id %d has a invalid next node", id))
				}
			}
		}

		for _, prev := range node.previous {
			if _, ok := nodes[prev]; !ok {
				panic(fmt.Sprintf("task with id %d has a invalid prev node", id))
			}
		}
	}

	l.Debugf("starting rootNodes %v", rootNodes)
	for _, rootNode := range rootNodes {
		startTask(rootNode)
	}

	allFinished := false

	for !allFinished {
		select {
		case result := <-doneChan:
			l.Debugf("task %d is done with err: %v", result.Id, result.Err)
			for i, id := range running {
				if id == result.Id {
					running = slices.Delete(running, i, i+1)
					break
				}
			}

			if result.Err != nil {
				done[result.Id] = failed
				errs = append(errs, result.Err)
				break
			} else {
				done[result.Id] = finished
			}

			for _, next := range nodes[result.Id].next {
				nextNode := nodes[next]
				allDone := true
				for _, prev := range nextNode.previous {
					if done[prev] != finished {
						allDone = false
						break
					}
				}
				if allDone {
					startTask(next)
				}
			}
		default:
			if len(running) <= 0 {
				allFinished = true
			}
		}
	}

	for id, node := range nodes {
		if output[id] == nil {
			l.Warnf("output of task %q is nil", nodes[id].task.Name())
		} else if output[id].Len() > 0 {
			l.Infof("task %q output: %s", node.task.Name(), output[id])
		}
	}

	return errors.Join(errs...)

}
