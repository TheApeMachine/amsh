package process

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/errnie"
)

type Programmer struct {
	Language  string     `json:"language" jsonschema:"required;title=Language;description=The programming language to use."`
	Fragments []Fragment `json:"fragments" jsonschema:"required;title=Fragments;description=The current code fragments to implement."`
	Commands  []Command  `json:"commands" jsonschema:"required;title=Commands;description=A list of bash terminal commands to run."`
}

type Fragment struct {
	Repository string `json:"repository" jsonschema:"required;title=Repository;description=The repository to implement the code in."`
	Branch     string `json:"branch" jsonschema:"required;title=Branch;description=The branch to implement the code in."`
	Filepath   string `json:"filepath" jsonschema:"required;title=Filepath;description=The filepath to implement the code in."`
	Action     string `json:"action" jsonschema:"required;title=Action;description=The action to perform on the code.;enum=add,modify,remove"`
	Code       string `json:"code" jsonschema:"required;title=Code;description=The code fragment to implement."`
}

type Command struct {
	Command string `json:"command" jsonschema:"required;title=Command;description=The bash terminal command to run."`
}

func NewProgrammer() *Programmer {
	return &Programmer{}
}

func (programmer *Programmer) SystemPrompt(key string) string {
	prompt := viper.GetViper().GetString(fmt.Sprintf("ai.setups.%s.processes.programmer.prompt", key))
	prompt = strings.ReplaceAll(prompt, "{{schemas}}", programmer.GenerateSchema())
	return prompt
}

func (programmer *Programmer) GenerateSchema() string {
	schema := jsonschema.Reflect(&Programmer{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

type DataScientist struct {
	Dataset string   `json:"dataset" jsonschema:"required;title=Dataset;description=The dataset to analyze."`
	Model   string   `json:"model" jsonschema:"required;title=Model;description=The machine learning model to use."`
	Metrics []Metric `json:"metrics" jsonschema:"required;title=Metrics;description=The metrics to evaluate the model on."`
}

type Metric struct {
	Name        string `json:"name" jsonschema:"required;title=Name;description=The name of the metric."`
	Description string `json:"description" jsonschema:"required;title=Description;description=The description of the metric."`
}

func NewDataScientist() *DataScientist {
	return &DataScientist{}
}

func (dataScientist *DataScientist) SystemPrompt(key string) string {
	prompt := viper.GetViper().GetString(fmt.Sprintf("ai.setups.%s.processes.data_scientist.prompt", key))
	prompt = strings.ReplaceAll(prompt, "{{schemas}}", dataScientist.GenerateSchema())
	return prompt
}

func (dataScientist *DataScientist) GenerateSchema() string {
	schema := jsonschema.Reflect(&DataScientist{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

type SecuritySpecialist struct {
	Code      string     `json:"code" jsonschema:"required;title=Code;description=The code to analyze."`
	Endpoints []Endpoint `json:"endpoints" jsonschema:"required;title=Endpoints;description=The endpoints to analyze the code against."`
	RedTeam   []RedTeam  `json:"red_team" jsonschema:"required;title=Red Team;description=The red team to analyze the code against."`
	BlueTeam  []BlueTeam `json:"blue_team" jsonschema:"required;title=Blue Team;description=The blue team to analyze the code against."`
}

type Endpoint struct {
	Method string `json:"method" jsonschema:"required;title=Method;description=The HTTP method of the endpoint."`
	Path   string `json:"path" jsonschema:"required;title=Path;description=The path of the endpoint."`
}

type RedTeam struct {
	Role        string `json:"role" jsonschema:"required;title=Role;description=The role of the red team."`
	Description string `json:"description" jsonschema:"required;title=Description;description=The description of the red team."`
}

type BlueTeam struct {
	Role        string `json:"role" jsonschema:"required;title=Role;description=The role of the blue team."`
	Description string `json:"description" jsonschema:"required;title=Description;description=The description of the blue team."`
}

func NewSecuritySpecialist() *SecuritySpecialist {
	return &SecuritySpecialist{}
}

func (securitySpecialist *SecuritySpecialist) SystemPrompt(key string) string {
	prompt := viper.GetViper().GetString(fmt.Sprintf("ai.setups.%s.processes.security_specialist.prompt", key))
	prompt = strings.ReplaceAll(prompt, "{{schemas}}", securitySpecialist.GenerateSchema())
	return prompt
}

func (securitySpecialist *SecuritySpecialist) GenerateSchema() string {
	schema := jsonschema.Reflect(&SecuritySpecialist{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

type QAEngineer struct {
	TestPlan   []TestCriteria `json:"test_plan" jsonschema:"required,title=Test Plan;description=The test plan to implement."`
	TestReport []TestReport   `json:"test_report" jsonschema:"required,title=Test Report;description=The test report to implement."`
}

type TestCriteria struct {
	Description string `json:"description" jsonschema:"required,title=Description;description=The description of the test criteria."`
	Criteria    string `json:"criteria" jsonschema:"required,title=Criteria;description=The criteria to test against."`
}

type TestReport struct {
	Description string `json:"description" jsonschema:"required,title=Description;description=The description of the test report."`
	Report      string `json:"report" jsonschema:"required,title=Report;description=The report of the test."`
}

func NewQAEngineer() *QAEngineer {
	return &QAEngineer{}
}

func (qaEngineer *QAEngineer) SystemPrompt(key string) string {
	prompt := viper.GetViper().GetString(fmt.Sprintf("ai.setups.%s.processes.qa_engineer.prompt", key))
	prompt = strings.ReplaceAll(prompt, "{{schemas}}", qaEngineer.GenerateSchema())
	return prompt
}

func (qaEngineer *QAEngineer) GenerateSchema() string {
	schema := jsonschema.Reflect(&QAEngineer{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}
