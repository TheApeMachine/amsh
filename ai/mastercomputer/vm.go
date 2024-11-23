package mastercomputer

import (
	"context"

	"github.com/theapemachine/amsh/ai/boogie"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/errnie"
)

type VM struct {
	ctx          context.Context
	lexer        *boogie.Lexer
	parser       *boogie.Parser
	compiler     *boogie.Compiler
	instructions []boogie.Instruction
	processors   []*Processor
	buffer       *Buffer
	LoadStream   chan provider.Event
}

func NewVM(ctx context.Context) *VM {
	errnie.Log("vm.NewVM()")

	return &VM{
		ctx:        ctx,
		lexer:      boogie.NewLexer(),
		parser:     boogie.NewParser(),
		compiler:   boogie.NewCompiler(),
		processors: make([]*Processor, 0),
		buffer:     NewBuffer(),
	}
}

func (vm *VM) Load(program string) {
	errnie.Log("vm.Load(%s)", program)

	compiler := boogie.NewCompiler()
	compiler.Generate(
		boogie.NewParser().Generate(
			boogie.NewLexer().Generate(program),
		),
	)

	vm.instructions = compiler.Load()
	errnie.Log("vm.instructions(%v)", vm.instructions)
}

func (vm *VM) Generate(instruction boogie.Instruction) {
	errnie.Log("vm.Generate(%v)", instruction)

	switch instruction.Type {
	case boogie.INSTRUCTION_SPAWN:
		vm.processors = append(vm.processors, NewProcessor(vm.ctx, instruction))
	case boogie.INSTRUCTION_JOIN:
		vm.parallel(instruction.Next)
	}
}

func (vm *VM) parallel(indices []int) {
	errnie.Log("vm.parallel(%v)", indices)

	for _, idx := range indices {
		vm.Generate(vm.instructions[idx])
	}
}
