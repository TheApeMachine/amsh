package trengo

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/client"
	"github.com/theapemachine/errnie"
)

type Label struct {
	ID int `json:"label_id"`
}

type Presenter struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Response struct {
	Data []struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Slug      string `json:"slug"`
		Color     string `json:"color"`
		SortOrder int    `json:"sort_order"`
		Archived  any    `json:"archived"`
	} `json:"data"`
	Links struct {
		First string `json:"first"`
		Last  string `json:"last"`
		Prev  any    `json:"prev"`
		Next  any    `json:"next"`
	} `json:"links"`
	Meta struct {
		CurrentPage int `json:"current_page"`
		From        int `json:"from"`
		LastPage    int `json:"last_page"`
		Links       []struct {
			URL    any    `json:"url"`
			Label  string `json:"label"`
			Active bool   `json:"active"`
		} `json:"links"`
		Path    string `json:"path"`
		PerPage int    `json:"per_page"`
		To      int    `json:"to"`
		Total   int    `json:"total"`
	} `json:"meta"`
}

type ListLabels struct {
	conn    *client.Client
	baseURL string
	token   string
}

func NewListLabels() *ListLabels {
	return &ListLabels{
		conn:    client.New(),
		baseURL: "https://api.trengo.com/api/v2",
		token:   os.Getenv("TRENGO_API_TOKEN"),
	}
}

/*
List all the labels in a way that the language model can understand.
*/
func (l *ListLabels) List(ctx context.Context) ([]Presenter, error) {
	var (
		response     *client.Response
		labels       = make([]Presenter, 0)
		responseBody Response
		err          error
	)

	nextPage := 1
	nextPageURL := l.baseURL + "/labels?page=" + strconv.Itoa(nextPage)

	for {
		// First collect all the labels, making sure to take care of pagination.
		if response, err = l.conn.Get(nextPageURL, client.Config{
			Header: map[string]string{
				"Authorization": "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJhdWQiOiIxIiwianRpIjoiMWM0YjI1YTdhODA2MmQ3YWY5ZGZmYjkzZWMzODY5YzMyZTFkN2EzM2VmZDNkODQ5NjAyMWYxZGIzYmE3YWNkZWNlNmQ3YzYzZDg3ZGQ2ZTUiLCJpYXQiOjE3MzAxODc0OTEsIm5iZiI6MTczMDE4NzQ5MSwiZXhwIjo0ODU0MjM4NjkxLCJzdWIiOiI3MjI1OTYiLCJzY29wZXMiOltdLCJhZ2VuY3lfaWQiOjI3MzQ2OX0.iaDKmvL5IeDjFeIsSTwx2wFNAXv2-oo68XZXSTuwO52i-5aG0wrUuJuyLm2KNBSPeuwPEB44CrJrtBWgPsFuzw",
				"Accept":        "application/json",
			},
		}); err != nil {
			return nil, errnie.Error(err)
		}

		if response.StatusCode() != fiber.StatusOK {
			return nil, errnie.Error(errors.New("[STATUS " + strconv.Itoa(response.StatusCode()) + "]\n  " + response.String() + "\n[/STATUS]\n"))
		}

		json.Unmarshal(response.Body(), &responseBody)

		for _, label := range responseBody.Data {
			labels = append(labels, Presenter{
				ID:   label.ID,
				Name: label.Name,
			})
		}

		if responseBody.Links.Next != nil {
			nextPage = responseBody.Meta.CurrentPage + 1
			nextPageURL = responseBody.Links.Next.(string)
		}

		if responseBody.Links.Next == nil {
			break
		}
	}

	return labels, nil
}

type AssignLabels struct {
	conn    *client.Client
	baseURL string
	token   string
}

func NewAssignLabels() *AssignLabels {
	return &AssignLabels{
		conn:    client.New(),
		baseURL: "https://api.trengo.com/api/v2",
		token:   os.Getenv("TRENGO_API_TOKEN"),
	}
}

func (l *AssignLabels) Attach(ctx context.Context, labelID int, ticketID int) (err error) {
	var (
		response *client.Response
	)

	body := Label{
		ID: labelID,
	}

	if response, err = l.conn.Post(l.baseURL+"/tickets/"+strconv.Itoa(ticketID)+"/labels", client.Config{
		Header: map[string]string{
			"Authorization": "Bearer " + l.token,
			"Accept":        "application/json",
			"Content-Type":  "application/json",
		},
		Body: body,
	}); err != nil {
		return errnie.Error(err)
	}

	if response.StatusCode() != fiber.StatusOK {
		return errnie.Error(errors.New("[STATUS " + strconv.Itoa(response.StatusCode()) + "]\n  " + response.String() + "\n[/STATUS]\n"))
	}

	return
}
