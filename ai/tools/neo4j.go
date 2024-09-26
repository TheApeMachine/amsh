package tools

import (
	"context"
	"encoding/json"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
)

type KnowledgeBase struct {
	host   string
	port   string
	driver neo4j.DriverWithContext
	ctx    context.Context
	cypher string
	params map[string]any
}

func NewKnowledgeBase(ctx context.Context, cypher string, params map[string]any) *KnowledgeBase {
	var (
		driver neo4j.DriverWithContext
	)

	errnie.Op[*neo4j.DriverWithContext](neo4j.NewDriverWithContext(
		viper.GetString("stages.local.neo4j.uri"),
		neo4j.BasicAuth(
			viper.GetString("stages.local.neo4j.user"),
			viper.GetString("stages.local.neo4j.password"),
			"",
		),
	))

	return &KnowledgeBase{
		driver: driver,
		ctx:    ctx,
		cypher: cypher,
		params: params,
	}
}

func (knowledgeBase *KnowledgeBase) Read(p []byte) (n int, err error) {
	var (
		result *neo4j.EagerResult
		buf    []byte
	)

	if result, err = neo4j.ExecuteQuery(
		knowledgeBase.ctx, knowledgeBase.driver,
		knowledgeBase.cypher,
		knowledgeBase.params,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	); err != nil {
		return 0, err
	}

	if buf, err = json.Marshal(result.Records); err != nil {
		return 0, err
	}

	return data.New("knowledgebase", "read", "neo4j", buf).Read(p)
}

func (knowledgeBase *KnowledgeBase) Write(p []byte) (n int, err error) {
	return
}

func (knowledgeBase *KnowledgeBase) Close() error {
	return knowledgeBase.driver.Close(knowledgeBase.ctx)
}
