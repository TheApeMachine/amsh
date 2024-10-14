package data

import (
	"fmt"
	"strings"

	"capnproto.org/go/capnp/v3"
	"github.com/theapemachine/amsh/errnie"
)

func (artifact Artifact) Card() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("[ARTIFACT (%s)]\n", artifact.Peek("id")))
	builder.WriteString(fmt.Sprintf("\tVERSION  : %s\n", artifact.Peek("version")))
	builder.WriteString(fmt.Sprintf("\tTYPE     : %s\n", artifact.Peek("type")))
	builder.WriteString(fmt.Sprintf("\tTIMESTAMP: %s\n", artifact.Peek("timestamp")))
	builder.WriteString(fmt.Sprintf("\tORIGIN   : %s\n", artifact.Peek("origin")))
	builder.WriteString(fmt.Sprintf("\tROLE     : %s\n", artifact.Peek("role")))
	builder.WriteString(fmt.Sprintf("\tSCOPE    : %s\n", artifact.Peek("scope")))

	builder.WriteString("\t[ATTRIBUTES]\n")

	attributes, err := artifact.Attributes()

	if err != nil {
		return err.Error()
	}

	for i := 0; i < attributes.Len(); i++ {
		attribute := attributes.At(i)
		key, err := attribute.Key()

		if err != nil {
			return err.Error()
		}

		value, err := attribute.Value()

		if err != nil {
			return err.Error()
		}

		builder.WriteString(fmt.Sprintf("\t\t%s = %s\n", key, value))
	}

	builder.WriteString("\t[/ATTRIBUTES]\n\n")
	builder.WriteString("\t[PAYLOAD]\n")

	payload, err := artifact.Payload()

	if err != nil {
		return err.Error()
	}

	builder.WriteString("\t\t" + string(payload))
	builder.WriteString("\t[/PAYLOAD]\n")
	builder.WriteString("[/ARTIFACT]\n")

	return builder.String()
}

func (artifact Artifact) Marshal() []byte {
	errnie.Trace()
	var (
		buf []byte
		err error
	)

	if buf, err = artifact.Message().Marshal(); err != nil {
		return nil
	}

	return buf
}

func (artifact Artifact) Unmarshal(buf []byte) Artifact {
	errnie.Trace()
	var (
		msg *capnp.Message
		err error
	)

	if msg, err = capnp.Unmarshal(buf); errnie.Error(err) != nil {
		return Empty
	}

	if artifact, err = ReadRootArtifact(msg); errnie.Error(err) != nil {
		return Empty
	}

	return artifact
}
