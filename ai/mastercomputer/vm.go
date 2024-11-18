package mastercomputer

import (
	"context"

	"github.com/theapemachine/amsh/ai/boogie"
)

type VM struct {
	ctx          context.Context
	lexer        *boogie.Lexer
	parser       *boogie.Parser
	compiler     *boogie.Compiler
	instructions []boogie.Instruction
	processors   []*Processor
	buffer       *Buffer
}

func NewVM(ctx context.Context) *VM {
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
	compiler := boogie.NewCompiler()
	compiler.Generate(
		boogie.NewParser().Generate(
			boogie.NewLexer().Generate(program),
		),
	)

	vm.instructions = compiler.Load()
}

func (vm *VM) Generate(instruction boogie.Instruction) {
	switch instruction.Type {
	case boogie.INSTRUCTION_SPAWN:
		vm.processors = append(vm.processors, NewProcessor(vm.ctx, instruction))
	case boogie.INSTRUCTION_JOIN:
		vm.parallel(instruction.Next)
	}
}

func (vm *VM) parallel(indices []int) {
	for _, idx := range indices {
		vm.Generate(vm.instructions[idx])
	}
}
