package ai

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/errnie"
)

type Crew struct {
	ctx    context.Context
	conn   *Conn
	agents map[string]*Agent
}

func NewCrew(ctx context.Context, conn *Conn) *Crew {
	return &Crew{
		ctx:  ctx,
		conn: conn,
		agents: map[string]*Agent{
			"director":  NewAgent(ctx, conn, "director", colors[0]),
			"writer":    NewAgent(ctx, conn, "writer", colors[1]),
			"flow":      NewAgent(ctx, conn, "flow", colors[2]),
			"extractor": NewAgent(ctx, conn, "extractor", colors[3]),
		},
	}
}

/*
Direct the story if needed.
*/
func (crew *Crew) Direct(highlights string) (response *Direction, err error) {
	system := viper.GetViper().GetString("ai.crew.director.system")
	user := viper.GetViper().GetString("ai.crew.director.user")

	user = strings.ReplaceAll(user, "<{highlights}>", highlights)

	var resp string

	if resp, err = crew.agents["director"].ChatCompletion(system, user); err != nil {
		errnie.Error(err.Error())
		return
	}

	out := &Direction{}
	buf := crew.ExtractJSON(resp)

	if err = json.Unmarshal([]byte(buf), out); err != nil {
		errnie.Error(err.Error())
		return
	}

	spew.Dump(out)

	return out, nil
}

/*
Write a new scene, based on the input of the Director.
*/
func (crew *Crew) Write(scene string, directions *Direction) (response *Script, err error) {
	system := viper.GetViper().GetString("ai.crew.writer.system")
	user := viper.GetViper().GetString("ai.crew.writer.user")

	user = strings.ReplaceAll(user, "<{scene}>", scene)
	user = strings.ReplaceAll(user, "<{directions}>", directions.Description)

	errnie.Debug("crew.Write -> system %v", system)
	errnie.Debug("crew.Write -> user %v", user)

	var resp string

	if resp, err = crew.agents["writer"].ChatCompletion(system, user); err != nil {
		errnie.Error(err.Error())
		return
	}

	out := &Script{}
	buf := crew.ExtractJSON(resp)

	if err = json.Unmarshal([]byte(buf), out); err != nil {
		errnie.Error(err.Error())
		return
	}

	spew.Dump(out)

	return out, nil
}

/*
Flow allows us to control the flow of the story.
*/
func (crew *Crew) Flow(action string) (response *Flow, err error) {
	system := viper.GetViper().GetString("ai.crew.flow.system")
	user := viper.GetViper().GetString("ai.crew.flow.user")

	user = strings.ReplaceAll(user, "<{action}>", action)

	var resp string

	if resp, err = crew.agents["flow"].ChatCompletion(system, user); err != nil {
		errnie.Error(err.Error())
		return
	}

	out := &Flow{}
	buf := crew.ExtractJSON(resp)

	if err = json.Unmarshal([]byte(buf), out); err != nil {
		errnie.Error(err.Error())
		return
	}

	spew.Dump(out)

	return out, nil
}

/*
UpdateProfile uses an LLM to analyze the agent's history and update the profile.
*/
func (crew *Crew) UpdateProfile(agent *Agent) (err error) {
	system := viper.GetViper().GetString("ai.crew.extractor.system")
	user := viper.GetViper().GetString("ai.crew.extractor.user")

	user = strings.ReplaceAll(user, "<{profile}>", agent.profile.String())
	user = strings.ReplaceAll(user, "<{history}>", agent.history)

	added := ""
	if agent.profile.Name != "" {
		added += "\n\n" + agent.profile.Name + "'s profile:\n\n" + agent.profile.String()
	}

	var content string

	if content, err = crew.agents["extractor"].ChatCompletion(system+added, user); err != nil {
		errnie.Error(err.Error())
		return
	}

	agent.profile.Unmarshal(crew.ExtractJSON(content))
	spew.Dump(agent.profile)

	return
}

func (crew *Crew) ExtractJSON(content string) string {
	content = strings.ReplaceAll(content, "```json", "")
	content = strings.ReplaceAll(content, "```", "")
	content = strings.TrimSpace(content)

	return content
}

type Direction struct {
	Change      string `json:"change"`
	Description string `json:"description"`
}

type Script struct {
	Scene   string   `json:"scene"`
	Actions []string `json:"actions"`
}

type Flow struct {
	Flow  string `json:"flow"`
	Scope string `json:"scope"`
}
