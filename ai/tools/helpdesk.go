package tools

type Helpdesk struct{}

func NewHelpdesk() *Helpdesk {
	return &Helpdesk{}
}

func (helpdesk *Helpdesk) Use(args map[string]any) string {
	return ""
}
