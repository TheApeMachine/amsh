package trengo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/theapemachine/errnie"
)

type MessageService struct {
	client  *http.Client
	baseURL string
	token   string
}

type Message struct {
	ID        int    `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	User      struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
}

type MessageResponse struct {
	Data []Message `json:"data"`
}

func NewMessageService(baseURL, token string) *MessageService {
	return &MessageService{
		client:  &http.Client{},
		baseURL: baseURL,
		token:   token,
	}
}

func (s *MessageService) ListMessages(ctx context.Context, ticketID int) ([]Message, error) {
	url := fmt.Sprintf("%s/tickets/%d/messages", s.baseURL, ticketID)

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

	var response MessageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errnie.Error(err)
	}

	return response.Data, nil
}

func (s *MessageService) FetchMessage(ctx context.Context, ticketID, messageID int) (*Message, error) {
	url := fmt.Sprintf("%s/tickets/%d/messages/%d", s.baseURL, ticketID, messageID)

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

	var message Message
	err = json.Unmarshal(body, &message)
	if err != nil {
		return nil, errnie.Error(err)
	}

	return &message, nil
}
