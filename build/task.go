package build

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"slices"

	"robaertschi.xyz/robaertschi/tt/asm/amd64"
	"robaertschi.xyz/robaertschi/tt/lexer"
	"robaertschi.xyz/robaertschi/tt/parser"
	"robaertschi.xyz/robaertschi/tt/token"
	"robaertschi.xyz/robaertschi/tt/ttir"
	"robaertschi.xyz/robaertschi/tt/typechecker"
	"robaertschi.xyz/robaertschi/tt/utils"
)

type task interface {
	Run(id int, doneChan chan taskResult)
}

type processTask struct {
	name string
	args []string
}

func NewProcessTask(name string, args ...string) task {
	return &processTask{
		name: name,
		args: args,
	}
}

func (pt *processTask) Run(id int, doneChan chan taskResult) {
	cmd := exec.Command(pt.name, pt.args...)
	cmd.Stdout = utils.NewPrefixWriterString(os.Stdout, pt.name+" output: ")
	cmd.Stderr = cmd.Stdout

	fmt.Printf("starting %q %v\n", pt.name, pt.args)
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
	err := os.Remove(rft.file)
	doneChan <- taskResult{
		Id:  id,
		Err: err,
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

func build(input string, output string, toPrint ToPrintFlags) error {
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
		fmt.Printf("AST:\n%s\n%+#v\n", program.String(), program)
	}

	tprogram, err := typechecker.New().CheckProgram(program)
	if err != nil {
		return err
	}
	if (toPrint & PrintTAst) != 0 {
		fmt.Printf("TAST:\n%s\n%+#v\n", tprogram.String(), tprogram)
	}

	ir := ttir.EmitProgram(tprogram)
	if (toPrint & PrintIr) != 0 {
		fmt.Printf("TTIR:\n%s\n%+#v\n", ir.String(), ir)
	}
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

func runTasks(nodes map[int]*node, rootNodes []int, l *log.Logger) error {

	done := make(map[int]executionState)
	running := []int{}
	doneChan := make(chan taskResult)
	errs := []error{}

	startTask := func(id int) {
		if done[id] != notStarted {
			panic(fmt.Sprintf("tried starting task %d twice", id))
		}
		// fmt.Printf("executing task %d\n", id)
		node := nodes[id]
		go node.task.Run(id, doneChan)
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

	// fmt.Printf("starting rootNodes %v\n", rootNodes)
	for _, rootNode := range rootNodes {
		startTask(rootNode)
	}

	allFinished := false

	for !allFinished {
		select {
		case result := <-doneChan:
			// fmt.Printf("task %d is done with err: %v\n", result.Id, result.Err)
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

	return errors.Join(errs...)

}
