package tools

type Project struct{}

func NewProject() *Project {
	return &Project{}
}

func (project *Project) Use(args map[string]any) string {
	return ""
}
