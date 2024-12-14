package trengo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/theapemachine/errnie"
)

type TicketService struct {
	client  *http.Client
	baseURL string
	token   string
}

type Ticket struct {
	ID          int    `json:"id"`
	Subject     string `json:"subject"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	AssignedTo  string `json:"assigned_to"`
	LastMessage string `json:"last_message"`
}

type TicketResponse struct {
	Data []Ticket `json:"data"`
}

func NewTicketService() *TicketService {
	return &TicketService{
		client:  &http.Client{},
		baseURL: "https://app.trengo.com/api/v2",
		token:   os.Getenv("TRENGO_API_TOKEN"),
	}
}

func (s *TicketService) ListTickets(ctx context.Context, page int) ([]Ticket, error) {
	url := fmt.Sprintf("%s/tickets?page=%d&sort=-date", s.baseURL, page)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errnie.Error(err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+s.token)

	res, err := s.client.Do(req)
	if err != nil {
		return nil, errnie.Error(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errnie.Error(err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, errnie.Error(fmt.Errorf("API request failed with status code: %d, body: %s", res.StatusCode, string(body)))
	}

	var response TicketResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errnie.Error(err)
	}

	return response.Data, nil
}
