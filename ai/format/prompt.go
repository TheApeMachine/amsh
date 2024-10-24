package format

type Prompt struct {
	UserIntent []string `json:"user_intent" jsonschema:"description=The users intent, objective, and goals extracted from the prompt"`
	Confusers  []string `json:"confusers" jsonschema:"description=Any part of the user's prompt that may cause confusion for the agents"`
	Optimize   bool     `json:"done" jsonschema:"description=You will have infinite iterations to reason, until you set this to true"`
	NextSteps  []Step   `json:"next_steps" jsonschema:"description=The next steps to be taken"`
}
