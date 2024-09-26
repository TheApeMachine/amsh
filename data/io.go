package data

import (
	"capnproto.org/go/capnp/v3"
	"github.com/theapemachine/amsh/errnie"
)

/*
Read implements the io.Reader interface for the Artifact.
It marshals the entire artifact into the provided byte slice.
*/
func (artifact *Artifact) Read(p []byte) (n int, err error) {
	var buf []byte

	// Marshal the artifact into bytes.
	if buf, err = artifact.Message().Marshal(); err != nil {
		return 0, err
	}

	// If the provided byte slice is too small, grow it.
	if len(p) < len(buf) {
		// Grow the slice to fit the marshaled data.
		p = make([]byte, len(buf))
	}

	// Copy the marshaled bytes into the provided byte slice.
	return copy(p, buf), nil
}

/*
Write implements the io.Writer interface for the Artifact.
It writes the entire artifact to the provided stream.
*/
func (artifact *Artifact) Write(p []byte) (n int, err error) {
	var (
		msg        *capnp.Message
		buf        Artifact
		ID         string
		checksum   []byte
		pubkey     []byte
		version    string
		Type       string
		timestamp  uint64
		origin     string
		role       string
		scope      string
		attributes capnp.StructList[Attribute]
		payload    []byte
	)

	if msg, err = capnp.Unmarshal(p); err != nil {
		return 0, err
	}

	if buf, err = ReadRootArtifact(msg); err != nil {
		return 0, err
	}

	if ID, err = buf.Id(); err != nil {
		return 0, err
	}

	if checksum, err = buf.Checksum(); err != nil {
		return 0, err
	}

	if pubkey, err = buf.Pubkey(); err != nil {
		return 0, err
	}

	if version, err = buf.Version(); err != nil {
		return 0, err
	}

	if Type, err = buf.Type(); err != nil {
		return 0, err
	}

	if timestamp = buf.Timestamp(); err != nil {
		return 0, err
	}

	if origin, err = buf.Origin(); err != nil {
		return 0, err
	}

	if role, err = buf.Role(); err != nil {
		return 0, err
	}

	if scope, err = buf.Scope(); err != nil {
		return 0, err
	}

	if payload, err = buf.Payload(); err != nil {
		return 0, err
	}

	if attributes, err = buf.Attributes(); err != nil {
		return 0, err
	}

	errnie.Op[*Artifact](artifact.SetId(ID), "error setting id")
	errnie.Op[*Artifact](artifact.SetChecksum(checksum), "error setting checksum")
	errnie.Op[*Artifact](artifact.SetPubkey(pubkey), "error setting pubkey")
	errnie.Op[*Artifact](artifact.SetVersion(version), "error setting version")
	errnie.Op[*Artifact](artifact.SetType(Type), "error setting type")
	artifact.SetTimestamp(timestamp)
	errnie.Op[*Artifact](artifact.SetOrigin(origin), "error setting origin")
	errnie.Op[*Artifact](artifact.SetRole(role), "error setting role")
	errnie.Op[*Artifact](artifact.SetScope(scope), "error setting scope")
	errnie.Op[*Artifact](artifact.SetAttributes(attributes), "error setting attributes")
	errnie.Op[*Artifact](artifact.SetPayload(payload), "error setting payload")

	return len(p), nil
}

/*
Close implements the io.Closer interface for the Artifact.
*/
func (artifact *Artifact) Close() error {
	// No-op for this example, but could be extended to manage resources.
	return nil
}
