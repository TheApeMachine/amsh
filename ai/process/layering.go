package process

import "github.com/theapemachine/amsh/utils"

type Layering struct {
	Layers []Layer `json:"layers" jsonschema:"title:Layers,description:The layers of the task,required"`
}

func (ta *Layering) SystemPrompt(key string) string {
	return utils.SystemPrompt(key, "layering", utils.GenerateSchema[Layering]())
}
