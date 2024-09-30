package ai

import (
	"encoding/json"
	"fmt"

	"github.com/theapemachine/amsh/errnie"
)

type Parser interface {
	Parse(response string) error
	Markdown() string
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

func (direction *Direction) Markdown() string {
	return fmt.Sprintf("| Director | %s |", direction.Direction)
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

func (side *Side) Markdown() string {
	out := fmt.Sprintf("**Writer**\n\n%s\n\n", side.Scene)

	for _, character := range side.Characters {
		out += character.Markdown() + "\n\n"
	}

	for _, action := range side.Actions {
		out += fmt.Sprintf("* %s\n", action)
	}

	return out
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

func (character *Character) Markdown() string {
	return fmt.Sprintf("**%s**\n\n%s\n\n%s", character.Name, character.Role, character.Description)
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

func (location *Location) Markdown() string {
	return fmt.Sprintf("**%s**\n\n%s", location.Name, location.Description)
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

func (edit *Edit) Markdown() string {
	return fmt.Sprintf("| Editor | %s | %s |", edit.Flow, edit.Scope)
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

func (extract *Extract) Markdown() string {
	return fmt.Sprintf("| Producer | %s | %s |", extract.Target, extract.Data)
}
