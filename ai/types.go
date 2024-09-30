package ai

import (
	"encoding/json"

	"github.com/theapemachine/amsh/errnie"
)

type Parser interface {
	Parse(response string) error
}

/*
Direction is a type that structures the directors instructions to the other agents.
*/
type Direction struct {
	Direction string `json:"direction"`
}

func (direction *Direction) Parse(response string) error {
	errnie.Trace()

	if err := json.Unmarshal(
		ExtractJSON(response),
		&direction,
	); err != nil {
		return errnie.Error(err)
	}

	return nil
}

/*
Side is a type that structures the writer's instructions to the actors.
*/
type Side struct {
	Scene      string      `json:"scene"`
	Characters []Character `json:"characters"`
	Actions    []string    `json:"actions"`
}

func (side *Side) Parse(response string) error {
	errnie.Trace()

	if err := json.Unmarshal(
		ExtractJSON(response),
		&side,
	); err != nil {
		return errnie.Error(err)
	}

	return nil
}

/*
Character is a type that structures the writer's instructions to the actors.
*/
type Character struct {
	Name        string `json:"name"`
	Role        string `json:"role"`
	Description string `json:"description"`
}

func (character *Character) Parse(response string) error {
	errnie.Trace()

	if err := json.Unmarshal(
		ExtractJSON(response),
		&character,
	); err != nil {
		return errnie.Error(err)
	}

	return nil
}

/*
Location is a type that structures the writer's instructions to the actors.
*/
type Location struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (location *Location) Parse(response string) error {
	errnie.Trace()

	if err := json.Unmarshal(
		ExtractJSON(response),
		&location,
	); err != nil {
		return errnie.Error(err)
	}

	return nil
}

/*
Edit is a type that structures the editor's instructions to the flow.
*/
type Edit struct {
	Flow  string `json:"flow"`
	Scope string `json:"scope"`
}

func (edit *Edit) Parse(response string) error {
	errnie.Trace()

	if err := json.Unmarshal(
		ExtractJSON(response),
		&edit,
	); err != nil {
		return errnie.Error(err)
	}

	return nil
}

/*
Extract is a type that structures the producer's instructions to the flow.
*/
type Extract struct {
	Target string `json:"target"`
	Data   string `json:"data"`
}

func (extract *Extract) Parse(response string) error {
	errnie.Trace()

	if err := json.Unmarshal(
		ExtractJSON(response),
		&extract,
	); err != nil {
		return errnie.Error(err)
	}

	return nil
}
