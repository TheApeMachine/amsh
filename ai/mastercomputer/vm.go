// vm.go
package mastercomputer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/qpool"
)

// Opcode represents VM instructions
type Opcode int

const (
	OpNop Opcode = iota
	OpSend
	OpReceive
	OpStore
	OpLoad
	OpCall
	OpReturn
	OpJump
	OpBranch
)

// Instruction represents a single VM instruction
type Instruction struct {
	Op       Opcode
	Operands []interface{}
}

/*
VMState represents the current state of the virtual machine
*/
type VMState struct {
	PC          int                    // Program counter
	Stack       []interface{}          // Data stack
	Memory      map[string]interface{} // Shared memory space
	Accumulator interface{}            // Current working value
}

/*
VM is a construct that executes boogie code, which is an Agent driven language.
This means that each instruction is executed by an Agent.
*/
type VM struct {
	id           string
	state        *VMState
	instructions []Instruction
	pool         *qpool.Q
	comm         *AgentCommunication
	mu           sync.RWMutex
}

/*
NewVM creates a new Virtual Machine instance.
*/
func NewVM(pool *qpool.Q, comm *AgentCommunication) *VM {
	return &VM{
		id: fmt.Sprintf("vm-%s", uuid.New().String()),
		state: &VMState{
			Memory: make(map[string]interface{}),
			Stack:  make([]interface{}, 0),
		},
		pool: pool,
		comm: comm,
	}
}

/*
Execute the VM with a given set of instructions. These instructions are ultimately
created by an Agent, and represent a series of intentions to be carried out.
A short-lived Agent is created for the VM to "perform" the instructions.
*/
func (vm *VM) Execute(ctx context.Context, instructions []Instruction) error {
	vm.mu.Lock()
	vm.instructions = instructions
	vm.state.PC = 0
	vm.mu.Unlock()

	// Schedule execution in the pool
	result := vm.pool.Schedule(
		fmt.Sprintf("vm-exec-%s", uuid.New().String()),
		func() (any, error) {
			return vm.runLoop(ctx)
		},
		qpool.WithCircuitBreaker(vm.id, 3, time.Second*30),
	)

	// Wait for result
	if value := <-result; value.Error != nil {
		return value.Error
	}

	return nil
}

// runLoop executes instructions until completion or error
func (vm *VM) runLoop(ctx context.Context) (interface{}, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if done, err := vm.executeNext(); err != nil {
				return nil, err
			} else if done {
				return vm.state.Accumulator, nil
			}
		}
	}
}

// executeNext executes the next instruction
func (vm *VM) executeNext() (bool, error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if vm.state.PC >= len(vm.instructions) {
		return true, nil
	}

	inst := vm.instructions[vm.state.PC]
	vm.state.PC++

	switch inst.Op {
	case OpSend:
		return false, vm.executeSend(inst)
	case OpReceive:
		return false, vm.executeReceive(inst)
	case OpStore:
		return false, vm.executeStore(inst)
	case OpLoad:
		return false, vm.executeLoad(inst)
	case OpCall:
		return false, vm.executeCall(inst)
	case OpReturn:
		return true, nil
	case OpJump:
		return false, vm.executeJump(inst)
	case OpBranch:
		return false, vm.executeBranch(inst)
	case OpNop:
		return false, nil
	default:
		return false, fmt.Errorf("unknown opcode: %v", inst.Op)
	}
}

// Instruction execution methods
func (vm *VM) executeSend(inst Instruction) error {
	if len(inst.Operands) < 2 {
		return errnie.Error(fmt.Errorf("send requires target and value operands"))
	}

	target := inst.Operands[0].(string)
	value := inst.Operands[1]

	result := vm.pool.Schedule(
		fmt.Sprintf("vm-send-%s", uuid.New().String()),
		func() (any, error) {
			_, err := vm.comm.SendInstruction(vm.id, target, value)
			return nil, err
		},
	)

	// Wait for send confirmation
	if value := <-result; value.Error != nil {
		return value.Error
	}

	return nil
}

func (vm *VM) executeReceive(inst Instruction) error {
	if len(inst.Operands) < 1 {
		return errnie.Error(fmt.Errorf("receive requires source operand"))
	}

	source := inst.Operands[0].(string)

	// Create quantum channel for receiving
	result := vm.pool.Schedule(
		fmt.Sprintf("vm-receive-%s", uuid.New().String()),
		func() (any, error) {
			ch, err := vm.comm.JoinDiscussion(source)
			if err != nil {
				return nil, err
			}

			// Wait for first message
			msg := <-ch
			vm.state.Accumulator = msg.Value
			return nil, nil
		},
	)

	// Wait for receive completion
	if value := <-result; value.Error != nil {
		return value.Error
	}

	return nil
}

func (vm *VM) executeStore(inst Instruction) error {
	if len(inst.Operands) < 2 {
		return errnie.Error(fmt.Errorf("store requires key and value operands"))
	}

	key := inst.Operands[0].(string)
	value := inst.Operands[1]
	vm.state.Memory[key] = value
	return nil
}

func (vm *VM) executeLoad(inst Instruction) error {
	if len(inst.Operands) < 1 {
		return errnie.Error(fmt.Errorf("load requires key operand"))
	}

	key := inst.Operands[0].(string)
	if value, exists := vm.state.Memory[key]; exists {
		vm.state.Accumulator = value
		return nil
	}
	return errnie.Error(fmt.Errorf("key not found: %s", key))
}

func (vm *VM) executeCall(inst Instruction) error {
	if len(inst.Operands) < 1 {
		return errnie.Error(fmt.Errorf("call requires function name operand"))
	}

	funcName := inst.Operands[0].(string)
	args := inst.Operands[1:]

	// Schedule function execution in pool
	result := vm.pool.Schedule(
		fmt.Sprintf("vm-call-%s", uuid.New().String()),
		func() (any, error) {
			// Pass args to show intent of future use
			return nil, fmt.Errorf("function not implemented: %s (with %d args)", funcName, len(args))
		},
	)

	// Wait for function completion
	if value := <-result; value.Error != nil {
		return value.Error
	}

	return nil
}

func (vm *VM) executeJump(inst Instruction) error {
	if len(inst.Operands) < 1 {
		return errnie.Error(fmt.Errorf("jump requires target operand"))
	}

	target := inst.Operands[0].(int)
	if target < 0 || target >= len(vm.instructions) {
		return errnie.Error(fmt.Errorf("jump target out of bounds: %d", target))
	}

	vm.state.PC = target
	return nil
}

func (vm *VM) executeBranch(inst Instruction) error {
	if len(inst.Operands) < 2 {
		return errnie.Error(fmt.Errorf("branch requires condition and target operands"))
	}

	condition := inst.Operands[0].(bool)
	target := inst.Operands[1].(int)

	if condition {
		if target < 0 || target >= len(vm.instructions) {
			return errnie.Error(fmt.Errorf("branch target out of bounds: %d", target))
		}
		vm.state.PC = target
	}
	return nil
}
