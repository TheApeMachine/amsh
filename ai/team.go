package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

/*
Team is a management construct to coordinate multiple AI agents and align
them to work towards a unified goal.
*/
type Team struct {
	agents map[string]*Agent
	leader *Agent
}

/*
NewTeam constructs the Team of Agents.
*/
func NewTeam(agents map[string]*Agent, leaderName string) *Team {
	team := &Team{
		agents: agents,
	}
	if leader, ok := agents[leaderName]; ok {
		team.leader = leader
	} else {
		// Create a new leader if not found in the agents map
		team.leader = NewAgent(NewConn(), REVIEWER, nil, "TeamLeader")
	}
	return team
}

/*
ExecuteTeamTask is a method that orchestrates the execution of a task by a team of agents.
It starts by checking if the team leader is set. If not, it returns an error.
The method then asks the leader to break down the task and assign it to team members.
The breakdown is expected to be in JSON format and is used to create individual tasks for each agent.
After assigning tasks, it executes these tasks using the respective agents.
If the coder agent completes a task, its result is passed to the reviewer agent for review.
Finally, it compiles and summarizes the results from all agents into a single response.
*/
func (t *Team) ExecuteTeamTask(ctx context.Context, task string) (map[string]interface{}, error) {
	if t.leader == nil {
		return nil, fmt.Errorf("team leader not set")
	}

	// Ask the leader to break down the task and assign to team members
	breakdownPrompt := viper.GetString("prompt.template.task.breakdown")
	taskBreakdown, err := t.leader.ExecuteTask(ctx, strings.Replace(breakdownPrompt, "{{.task}}", task, -1))
	if err != nil {
		return nil, fmt.Errorf("failed to break down task: %w", err)
	}

	// Print taskBreakdown for debugging
	fmt.Printf("Task Breakdown: %+v\n", taskBreakdown)

	// Parse the task breakdown
	var taskAssignments map[string]interface{}
	if result, ok := taskBreakdown["result"].(map[string]interface{}); ok {
		taskAssignments = result
	} else {
		return nil, fmt.Errorf("invalid task breakdown format")
	}

	// Execute individual tasks
	results := make(map[string]interface{})
	var coderResult map[string]interface{}

	for taskKey, taskInfo := range taskAssignments {
		taskDetails, ok := taskInfo.(map[string]interface{})
		if !ok {
			continue
		}

		agentName, ok := taskDetails["assigned_to"].(string)
		if !ok {
			continue
		}

		agentTask, ok := taskDetails["description"].(string)
		if !ok {
			continue
		}

		agent, ok := t.agents[strings.ToLower(agentName)]
		if !ok {
			continue
		}

		result, err := agent.ExecuteTask(ctx, agentTask)
		if err != nil {
			return nil, fmt.Errorf("agent %s failed to execute task: %w", agentName, err)
		}

		results[taskKey] = result

		if strings.ToLower(agentName) == "coder" {
			coderResult = result
		}
	}

	// If we have a coder result, pass it to the reviewer
	if coderResult != nil {
		reviewPrompt := viper.GetString("prompt.template.task.review")
		reviewerTask := strings.Replace(reviewPrompt, "{{.code}}", coderResult["result"].(string), -1)
		reviewResult, err := t.agents["reviewer"].ExecuteTask(ctx, reviewerTask)
		if err != nil {
			return nil, fmt.Errorf("reviewer failed to execute task: %w", err)
		}
		results["review"] = reviewResult
	}

	// Print results for debugging
	resultsJSON, _ := json.Marshal(results)
	fmt.Printf("Individual Results:\n%s\n", string(resultsJSON))

	// Ask the leader to compile and summarize the results
	compilePrompt := viper.GetString("prompt.template.task.compile")
	summary, err := t.leader.ExecuteTask(ctx, strings.Replace(compilePrompt, "{{.results}}", string(resultsJSON), -1))
	if err != nil {
		return nil, fmt.Errorf("failed to compile results: %w", err)
	}

	return summary, nil
}
