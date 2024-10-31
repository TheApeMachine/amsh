package process

/*
Process defines an interface that object can implement if the want to act
as a predefined process. Predefined processes are used to direct specific
behavior, useful is cases where we know what should be done based on an input.
*/
type Process interface {
	GenerateSchema() string
	SystemPrompt(string) string
}

/*
Analysis defines the interface that all analysis types must implement.
This allows for different types of analysis to be performed on the input
while maintaining a consistent interface.
*/
type Analysis interface {
	// Analyze performs the analysis on the given input and returns a result
	Analyze(input string) (interface{}, error)
}

var ProcessMap = map[string]Process{
	"task_analyzer":       NewTaskAnalyzer(),
	"default":             NewThinking(),
	"surface":             NewSurfaceAnalysis(),
	"pattern":             NewPatternAnalysis(),
	"quantum":             NewQuantumAnalysis(),
	"time":                NewTimeAnalysis(),
	"narrative":           NewNarrativeAnalysis(),
	"analogy":             NewAnalogyAnalysis(),
	"practical":           NewPracticalAnalysis(),
	"context":             NewContextAnalysis(),
	"moonshot":            NewMoonshot(),
	"sensible":            NewSensible(),
	"catalyst":            NewCatalyst(),
	"guardian":            NewGuardian(),
	"programmer":          NewProgrammer(),
	"data_scientist":      NewDataScientist(),
	"qa_engineer":         NewQAEngineer(),
	"security_specialist": NewSecuritySpecialist(),
	"trengo":              NewLabelling(),
	"slack":               NewPlanning(),
	"development":         NewDevelopment(),
}
