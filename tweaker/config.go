package tweaker

import (
	"fmt"

	"github.com/spf13/viper"
)

var cfg *Config

func init() {
	fmt.Println("init")
	cfg = NewConfig()
}

type Config struct {
	v *viper.Viper
}

func NewConfig() *Config {
	fmt.Println("NewConfig")
	return &Config{
		v: viper.GetViper(),
	}
}

func SetViper(v *viper.Viper) { cfg.v = v }

/* LogLevel returns the log level from the config. */
func LogLevel() string { return cfg.v.GetString("log.level") }

/* GetInt returns an integer from the config. */
func GetInt(key string) int { return cfg.v.GetInt(key) }

type Setup struct {
	Name   string `yaml:"name"`
	Prompt struct {
		Prefix string `yaml:"prefix"`
		Suffix string `yaml:"suffix"`
		Script []struct {
			Name    string `yaml:"name"`
			System  string `yaml:"system"`
			Actions []struct {
				User string `yaml:"user"`
			} `yaml:"actions"`
		} `yaml:"script"`
	} `yaml:"prompt"`
	Agents []struct {
		Type             string `yaml:"type"`
		Scope            string `yaml:"scope"`
		Responsibilities string `yaml:"responsibilities"`
		Replicas         int    `yaml:"replicas"`
	} `yaml:"agents"`
	Flow []struct {
		Agent        string `yaml:"agent"`
		Instructions string `yaml:"instructions"`
	} `yaml:"flow"`
}

/*
GetSetup return a setup from the config.
*/
func GetSetup(key string) Setup {
	fmt.Println("GetSetup with hardcoded values")

	return Setup{
		Name: "The Ape Machine",
		Prompt: struct {
			Prefix string `yaml:"prefix"`
			Suffix string `yaml:"suffix"`
			Script []struct {
				Name    string `yaml:"name"`
				System  string `yaml:"system"`
				Actions []struct {
					User string `yaml:"user"`
				} `yaml:"actions"`
			} `yaml:"script"`
		}{
			Prefix: "You are part of an advanced AI simulation, modeled on real-world dynamics.\nThe simulation goes through an initial setup phase, where the actors are given an opportunity to define their characters.\nAfter the setup phase, the simulation enters its main loop, where the actors are left to interact with each other.\nA special crew of agents is responsible for ensuring that the simulation behaves as expected.",
			Suffix: "All responses should be formatted as a valid Markdown fragment.",
			Script: []struct {
				Name    string `yaml:"name"`
				System  string `yaml:"system"`
				Actions []struct {
					User string `yaml:"user"`
				} `yaml:"actions"`
			}{
				{
					Name:   "setup",
					System: "## Setup\n\nThe simulation is now starting its setup phase.\nEach actor will be given an opportunity to define their character.\nAfter the setup phase, the simulation will enter its main loop.",
					Actions: []struct {
						User string `yaml:"user"`
					}{
						{User: "## Name\n\nRespond to the task below.\n\n> Select a name for your character. You should select a full name, with a first name and last name."},
						{User: "## Backstory\n\nRespond to the task below.\n\n> Provide a backstory for your character. Make your backstory interesting and engaging. Your character should be unique and memorable, with realistic motivations, traits and goals."},
						{User: "## Resume\n\nRespond to the task below.\n\n> Provide a resume for your character. Your resume should be formatted as a valid Markdown document."},
					},
				},
			},
		},
		Agents: []struct {
			Type             string `yaml:"type"`
			Scope            string `yaml:"scope"`
			Responsibilities string `yaml:"responsibilities"`
			Replicas         int    `yaml:"replicas"`
		}{
			{Type: "director", Scope: "process", Responsibilities: "providing high level directions to the crew", Replicas: 1},
			{Type: "writer", Scope: "process", Responsibilities: "writing prompts for the actors", Replicas: 1},
			{Type: "editor", Scope: "process", Responsibilities: "editing the story by manipulating the flow of actions", Replicas: 1},
			{Type: "producer", Scope: "process", Responsibilities: "extracting structured data from the current context", Replicas: 1},
			{Type: "actor", Scope: "worker", Responsibilities: "acting out the story", Replicas: 1},
		},
		Flow: []struct {
			Agent        string `yaml:"agent"`
			Instructions string `yaml:"instructions"`
		}{
			{Agent: "director", Instructions: "- Analyze the current context to determine if the story is on track.\n- Your response should be formatted as JSON, inside a JSON code block.\n- Respond only with the JSON, no other text."},
			{Agent: "writer", Instructions: "- Write a prompt for the actor.\n- Your response should be formatted as JSON, inside a JSON code block.\n- Respond only with the JSON, no other text."},
			{Agent: "actor", Instructions: "- Act out the story.\n- Your response should include your dialog that corresponds to the current prompt."},
			{Agent: "editor", Instructions: "- Edit the story by manipulating the flow of actions.\n- Your response should be formatted as JSON, inside a JSON code block.\n- Respond only with the JSON, no other text."},
			{Agent: "producer", Instructions: "- Extract structured data from the current context.\n- Your response should be formatted as JSON, inside a JSON code block.\n- Respond only with the JSON, no other text."},
		},
	}
}

type Template struct {
	System string `yaml:"system"`
	User   string `yaml:"user"`
}

/*
GetTemplate returns a template from the config.
*/
func GetTemplate() Template {
	fmt.Println("GetTemplate with hardcoded values")

	return Template{
		System: "# <{name}>\n\n<{prefix}>\n\n## Role\n\nYou are a <{role}>, responsible for <{responsibilities}>.\n\n## Instructions\n\n<{instructions}>",
		User:   "## Context\n\n<details>\n  <summary>Current Context</summary>\n  \n  <{context}>\n</details>\n\n## Response\n\n",
	}
}
