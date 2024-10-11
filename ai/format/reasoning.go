package format

type Reasoning struct {
	Type       string `json:"type"`
	Properties struct {
		Reasoning struct {
			AnyOf []struct {
				Ref string `json:"$ref"`
			} `json:"anyOf"`
		} `json:"reasoning"`
	} `json:"properties"`
	Required []string `json:"required"`
	Defs     struct {
		Thought struct {
			Type        string `json:"type"`
			Description string `json:"description"`
		} `json:"thought"`
		Reflection struct {
			Type        string `json:"type"`
			Description string `json:"description"`
		} `json:"reflection"`
		Step struct {
			Type       string `json:"type"`
			Properties struct {
				Description struct {
					Type string `json:"type"`
				} `json:"description"`
				Action struct {
					Type string `json:"type"`
				} `json:"action"`
				Result struct {
					Type string `json:"type"`
				} `json:"result"`
			} `json:"properties"`
			Required []string `json:"required"`
		} `json:"step"`
		ChainOfThought struct {
			Type       string `json:"type"`
			Properties struct {
				Type struct {
					Const string `json:"const"`
				} `json:"type"`
				Steps struct {
					Type  string `json:"type"`
					Items struct {
						Ref string `json:"$ref"`
					} `json:"items"`
				} `json:"steps"`
				FinalAnswer struct {
					Type string `json:"type"`
				} `json:"final_answer"`
			} `json:"properties"`
			Required []string `json:"required"`
		} `json:"chain_of_thought"`
		TreeOfThought struct {
			Type       string `json:"type"`
			Properties struct {
				Type struct {
					Const string `json:"const"`
				} `json:"type"`
				Nodes struct {
					Type  string `json:"type"`
					Items struct {
						Type       string `json:"type"`
						Properties struct {
							Thought struct {
								Ref string `json:"$ref"`
							} `json:"thought"`
							Children struct {
								Type  string `json:"type"`
								Items struct {
									Ref string `json:"$ref"`
								} `json:"items"`
							} `json:"children"`
						} `json:"properties"`
						Required []string `json:"required"`
					} `json:"items"`
				} `json:"nodes"`
				FinalAnswer struct {
					Type string `json:"type"`
				} `json:"final_answer"`
			} `json:"properties"`
			Required []string `json:"required"`
		} `json:"tree_of_thought"`
		SelfReflection struct {
			Type       string `json:"type"`
			Properties struct {
				Type struct {
					Const string `json:"const"`
				} `json:"type"`
				Reflections struct {
					Type  string `json:"type"`
					Items struct {
						Ref string `json:"$ref"`
					} `json:"items"`
				} `json:"reflections"`
				FinalAnswer struct {
					Type string `json:"type"`
				} `json:"final_answer"`
			} `json:"properties"`
			Required []string `json:"required"`
		} `json:"self_reflection"`
	} `json:"$defs"`
}
