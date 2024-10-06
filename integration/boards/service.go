package boards

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
)

type Service struct {
	orgURL      string
	projectID   string
	projectName string
	pat         string
}

func NewService() *Service {
	return &Service{
		orgURL:      os.Getenv("AZDO_ORG_URL"),
		pat:         os.Getenv("AZDO_PAT"),
		projectID:   "9b0e875e-c3b5-481d-a8c5-48e96f624015",
		projectName: "fanapp",
	}
}

func (srv *Service) GetClient(ctx context.Context) (client core.Client) {
	var err error
	if client, err = core.NewClient(ctx, azuredevops.NewPatConnection(srv.orgURL, srv.pat)); err != nil {
		return nil
	}

	return
}

func (srv *Service) GetProjects(ctx context.Context) (projects []core.TeamProjectReference, err error) {
	var (
		responseValue *core.GetProjectsResponseValue
		index         = 0
	)

	if responseValue, err = srv.GetClient(ctx).GetProjects(ctx, core.GetProjectsArgs{}); err != nil {
		return nil, err
	}

	spew.Dump(responseValue)

	for responseValue != nil {
		// Log the page of team project names
		for _, teamProjectReference := range (*responseValue).Value {
			log.Printf("Name[%v] = %v", index, *teamProjectReference.Name)
			index++
		}

		// if continuationToken has a value, then there is at least one more page of projects to get
		if responseValue.ContinuationToken != "" {

			continuationToken, err := strconv.Atoi(responseValue.ContinuationToken)
			if err != nil {
				log.Fatal(err)
			}

			// Get next page of team projects
			projectArgs := core.GetProjectsArgs{
				ContinuationToken: &continuationToken,
			}
			responseValue, err = srv.GetClient(ctx).GetProjects(ctx, projectArgs)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			responseValue = nil
		}
	}

	return
}
