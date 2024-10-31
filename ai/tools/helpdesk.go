package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/invopop/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type Helpdesk struct {
	Operation string `json:"operation" jsonschema:"title=Operation,description=The operation to perform,enum=label,enum=create,enum=update,enum=close"`
	TicketID  int    `json:"ticketId" jsonschema:"title=Ticket ID,description=The ID of the ticket to label"`
	LabelIDs  []int  `json:"labelIds" jsonschema:"title=Label IDs,description=The IDs of the labels to assign to the ticket"`
}

func NewHelpdesk() *Helpdesk {
	return &Helpdesk{}
}

/*
Use the tool by passing the map of arguments that was returned from the AI.
*/
func (helpdesk *Helpdesk) Use(ctx context.Context, args map[string]any) string {
	if err := helpdesk.Unmarshal(args); err != nil {
		return errnie.Error(err).Error()
	}

	switch helpdesk.Operation {
	case "label":
		return helpdesk.LabelTicket()
	}
	return ""
}

/*
Unmarshal upgrades the map to the struct, for easier access to the data.
*/
func (helpdesk *Helpdesk) Unmarshal(args map[string]any) (err error) {
	var buf []byte
	if buf, err = json.Marshal(args); err != nil {
		return err
	}

	return json.Unmarshal(buf, helpdesk)
}

func (helpdesk *Helpdesk) GenerateSchema() string {
	schema := jsonschema.Reflect(&Helpdesk{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

func (helpdesk *Helpdesk) LabelTicket() string {
	for _, labelID := range helpdesk.LabelIDs {
		url := fmt.Sprintf(
			"https://app.trengo.com/api/v2/tickets/%d/labels", helpdesk.TicketID,
		)

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(
			[]byte(fmt.Sprintf(`{"label_id": %d}`, labelID)),
		))
		if err != nil {
			return errnie.Error(err).Error()
		}

		req.Header.Add("Authorization", fmt.Sprintf(
			"Bearer %s", os.Getenv("TRENGO_API_TOKEN"),
		))
		req.Header.Add("accept", "application/json")
		req.Header.Add("content-type", "application/json")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return errnie.Error(err).Error()
		}

		defer res.Body.Close()
		_, err = io.ReadAll(res.Body)
		if err != nil {
			return errnie.Error(err).Error()
		}
	}

	return "success"
}
