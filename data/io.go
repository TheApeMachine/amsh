package data

import (
	"io"

	"capnproto.org/go/capnp/v3"
)

/*
Read implements the io.Reader interface for the Artifact.
*/
func (artifact *Artifact) Read(p []byte) (n int, err error) {
	// Create a Cap'n Proto decoder that reads from the stream
	decoder := capnp.NewDecoder(model.reader)

	// Decode the Cap'n Proto message from the stream
	msg, err := decoder.Decode()
	if err != nil {
		return 0, err
	}

	// Extract the root Artifact struct from the message
	artifactCapnp, err := ReadRootArtifact(msg)
	if err != nil {
		return 0, err
	}

	// Extract the payload (or any other field)
	payload, err := artifactCapnp.Payload()
	if err != nil {
		return 0, err
	}

	// Copy the payload into the provided byte slice
	return copy(p, payload), nil
}

/*
Write implements the io.Writer interface for the Artifact.
*/
func (artifact *Artifact) Write(p []byte) (n int, err error) {
	// Create a Cap'n Proto message
	msg, _, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return 0, err
	}

	// Set fields of the artifact (convert p to the appropriate field, e.g., Payload)
	if err = artifact.SetPayload(p); err != nil {
		return 0, err
	}

	// Write the message to the writer (transport or connection)
	encoder := capnp.NewEncoder(model.writer)
	err = encoder.Encode(msg)
	if err != nil {
		return 0, err
	}

	// Return the number of bytes written
	return len(p), nil
}

/*
Close implements the io.Closer interface for the Artifact.
*/
func (artifact *Artifact) Close() error {
	if closer, ok := model.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
