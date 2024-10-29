package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/theapemachine/amsh/ai/process"
	"github.com/theapemachine/amsh/errnie"
)

type Helpdesk struct{}

func NewHelpdesk() *Helpdesk {
	return &Helpdesk{}
}

func (helpdesk *Helpdesk) Use(args map[string]any) string {
	return ""
}

func (helpdesk *Helpdesk) LabelTicket(args string) {
	// Extract the JSON from the Markdown JSON block.
	re := regexp.MustCompile("(?s)json\\s*(\\{.*?\\})\\s*")
	matches := re.FindAllStringSubmatch(args, -1)
	jsonContent := []string{}

	for _, match := range matches {
		jsonContent = append(jsonContent, strings.TrimSpace(match[1]))
	}

	out := process.Labelling{}
	if err := json.Unmarshal([]byte(jsonContent[0]), &out); err != nil {
		errnie.Error(err)
		return
	}

	for _, labelID := range out.LabelIDs {
		url := fmt.Sprintf(
			"https://app.trengo.com/api/v2/tickets/%d/labels", out.TicketID,
		)

		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(
			[]byte(fmt.Sprintf(`{"label_id": %d}`, labelID)),
		))

		req.Header.Add("Authorization", fmt.Sprintf(
			"Bearer %s", os.Getenv("TRENGO_API_TOKEN"),
		))
		req.Header.Add("accept", "application/json")
		req.Header.Add("content-type", "application/json")

		res, _ := http.DefaultClient.Do(req)

		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)

		fmt.Println(string(body))
	}
}
