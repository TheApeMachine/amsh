package process

type Requirement struct {
	Title       string `json:"title" jsonschema:"title:Title,description:The title of the requirement,required"`
	Description string `json:"description" jsonschema:"title:Description,description:The description of the requirement,required"`
}
