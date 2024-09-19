package ai

import (
	"context"
	"fmt"
	"net"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"github.com/theapemachine/amsh/data"
)

type Model struct {
	conn     net.Conn
	err      error
	host     string
	port     string
	artifact data.Artifact
}

func NewModel() *Model {
	return &Model{
		host: "localhost",
		port: "8123",
	}
}

func (model *Model) Read(p []byte) (n int, err error) {
	return model.conn.Read(p)
}

func (model *Model) Write(p []byte) (n int, err error) {
	if model.conn, model.err = net.Dial(
		"tcp", fmt.Sprintf("%s:%s", model.host, model.port),
	); model.err != nil {
		return 0, model.err
	}
	defer model.conn.Close()

	transport := rpc.NewStreamTransport(model.conn)
	clientConn := rpc.NewConn(transport, nil)
	defer clientConn.Close()

	client := data.ModelService(clientConn.Bootstrap(context.Background()))

	future, release := client.Query(context.Background(), func(r data.ModelService_query_Params) error {
		msg, err := capnp.Unmarshal(p)
		if err != nil {
			return err
		}

		if model.artifact, model.err = data.ReadRootArtifact(msg); model.err != nil {
			return model.err
		}

		return nil
	})

	defer release()

	return
}

func (model *Model) Close() error {
	return nil
}
