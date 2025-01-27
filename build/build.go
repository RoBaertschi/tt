// Build allows the user to build a tt file. The goal is to make it easy to support multiple backends with different requirements
package build

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"robaertschi.xyz/robaertschi/tt/asm"
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
			}
		}
	}

}

func (sp *SourceProgram) Build(backend asm.Backend) {

}
