package boogie

type InstructionType int

const (
	INSTRUCTION_SPAWN InstructionType = iota
	INSTRUCTION_FLOW
	INSTRUCTION_JOIN
	INSTRUCTION_MATCH
)

type Instruction struct {
	Type      InstructionType
	Operation string
	Behavior  string
	Next      []int
	Fallbacks []int
}

type Compiler struct {
	instructions []Instruction
}

func NewCompiler() *Compiler {
	return &Compiler{
		instructions: make([]Instruction, 0),
	}
}

func (compiler *Compiler) Load() []Instruction {
	return compiler.instructions
}

func (compiler *Compiler) Generate(node *Node) int {
	switch node.Type {
	case NODE_PROGRAM:
		compiler.Generate(node.Next[0])
	case NODE_CLOSURE:
		lastIdx := -1

		for _, op := range node.Next {
			lastIdx = compiler.Generate(op)
		}

		return lastIdx
	case NODE_OPERATION:
		idx := len(compiler.instructions)

		instruction := Instruction{
			Type:      INSTRUCTION_SPAWN,
			Operation: node.Value,
			Next:      make([]int, 0),
			Fallbacks: make([]int, 0),
		}

		if node.Behavior != nil {
			instruction.Behavior = node.Behavior.Value
		}

		for _, next := range node.Next {
			nextIdx := compiler.Generate(next)
			instruction.Next = append(instruction.Next, nextIdx)
		}

		compiler.instructions = append(compiler.instructions, instruction)
		return idx
	case NODE_JOIN:
		joinIdx := len(compiler.instructions)
		parallelIdices := make([]int, 0)

		for _, closure := range node.Next {
			parallelIdices = append(parallelIdices, compiler.Generate(closure))
		}

		instruction := Instruction{
			Type: INSTRUCTION_JOIN,
			Next: parallelIdices,
		}

		compiler.instructions = append(compiler.instructions, instruction)
		return joinIdx
	}

	return -1
}
