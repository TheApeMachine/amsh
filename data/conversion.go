package data

import (
	"fmt"
	"strings"

	"capnproto.org/go/capnp/v3"
	"github.com/theapemachine/amsh/errnie"
)

func (artifact *Artifact) Card() string {
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

func (artifact *Artifact) Marshal() []byte {
	var (
		buf []byte
		err error
	)

	if buf, err = artifact.Message().Marshal(); err != nil {
		errnie.Error(fmt.Errorf("marshal error: %w", err))
		return nil
	}

	return buf
}

func Unmarshal(buf []byte) *Artifact {
	var (
		msg *capnp.Message
		af  Artifact
		err error
	)

	if msg, err = capnp.Unmarshal(buf); errnie.Error(err) != nil {
		errnie.Error(fmt.Errorf("unmarshal error: %w", err))
		return nil
	}

	if errnie.Error(err) != nil {
		return nil
	}

	if af, err = ReadRootArtifact(msg); errnie.Error(err) != nil {
		return nil
	}

	return &af
}
