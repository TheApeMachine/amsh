package ai

/*
RoleType is an enum representing the different roles an AI agent can play.
*/
type RoleType uint

const (
	// TEAMLEAD is an AI agent that is responsible for orchestrating agents that are joined together in a Team.
	TEAMLEAD RoleType = iota
	// ARCHITECT is an AI agent that specializes in technical breakdown and vision, and is able to diagram out the details of a system.
	ARCHITECT
	// CODER is an AI agent that specializes in writing code. It is language-agnostic and can write code in any language.
	CODER
	// REVIEWER is an AI agent that performs an in-depth review of code for bugs, and guards against diversions from the overal vision and guidelines.
	REVIEWER
	// TESTER is an AI agent that specializes in testing code for bugs and ensuring that it is working as expected.
	TESTER
	// AUDITOR monitors the failure and success rates of prompts and their outputs, and can test and review alternative prompts in a sandboxed environment.
	AUDITOR
)
