package data

import (
	"errors"
	"time"

	"capnproto.org/go/capnp/v3"
	"github.com/theapemachine/amsh/errnie"
)

const version = "0.0.1"

/*
Empty is an empty artifact.
*/
var Empty = Artifact{}

/*
New creates a new artifact with the given origin, role, scope, and data.
*/
func New(origin, role, scope string, data []byte) Artifact {
	var (
		seg      *capnp.Segment
		err      error
		artifact Artifact
	)

	if _, seg, err = capnp.NewMessage(capnp.SingleSegment(nil)); err != nil {
		return Empty
	}

	if artifact, err = NewArtifact(seg); err != nil {
		return Empty
	}

	errnie.Error(artifact.SetOrigin(origin))
	errnie.Error(artifact.SetRole(role))
	errnie.Error(artifact.SetScope(scope))
	errnie.Error(artifact.SetPayload(data))

	artifact.SetTimestamp(uint64(time.Now().UnixNano()))
	artifact.SetVersion(version)

	return artifact
}

/*
Poke sets a value on the artifact, starting by looking for an existing field,
and falling back to using the attribute list.
*/
func (artifact Artifact) Poke(key string, value string) {
	var err error

	switch key {
	case "id":
		err = artifact.SetId(value)
	case "version":
		err = artifact.SetVersion(value)
	case "type":
		err = artifact.SetType(value)
	case "origin":
		err = artifact.SetOrigin(value)
	case "role":
		err = artifact.SetRole(value)
	case "scope":
		err = artifact.SetScope(value)
	case "payload":
		err = artifact.SetPayload([]byte(value))
	default:
		// If the key is not a top-level field, add it as an attribute.
		err = artifact.addAttribute(key, value)
	}

	errnie.Error(err)
}

/*
Peek retrieves a value from the artifact, starting by looking for an existing field,
and falling back to searching the attribute list.
*/
func (artifact Artifact) Peek(key string) string {
	var (
		value string
		data  []byte
		err   error
	)

	// Check if the key corresponds to a top-level field.
	switch key {
	case "id":
		value, err = artifact.Id()
	case "version":
		value, err = artifact.Version()
	case "type":
		value, err = artifact.Type()
	case "origin":
		value, err = artifact.Origin()
	case "role":
		value, err = artifact.Role()
	case "scope":
		value, err = artifact.Scope()
	case "payload":
		data, err = artifact.Payload()
		value = string(data)
	default:
		// If the key is not a top-level field, look in the attributes list.
		value, err = artifact.getAttributeValue(key)
	}

	errnie.Error(err)
	return value
}

// getAttributeValue searches the attribute list for the given key.
func (artifact Artifact) getAttributeValue(key string) (string, error) {
	attrs, err := artifact.Attributes()
	if errnie.Error(err) != nil {
		return "", err
	}

	// Iterate through the attributes list to find a matching key.
	for i := 0; i < attrs.Len(); i++ {
		attr := attrs.At(i) // Only one return value now.
		attrKey, err := attr.Key()
		if errnie.Error(err) != nil {
			return "", err
		}

		if attrKey == key {
			return attr.Value()
		}
	}

	return "", errors.New("key not found")
}

/*
addAttribute adds a new attribute to the artifact.
*/
func (artifact Artifact) addAttribute(key, value string) error {
	// Retrieve the existing attributes.
	attrs, err := artifact.Attributes()
	if err != nil {
		return errnie.Error(err)
	}

	// Create a new list of attributes, with length increased by 1 to accommodate the new attribute.
	newAttrs, err := NewAttribute_List(artifact.Segment(), int32(attrs.Len()+1))
	if err != nil {
		return errnie.Error(err)
	}

	// Copy existing attributes to the new list.
	for i := 0; i < attrs.Len(); i++ {
		if err := newAttrs.Set(i, attrs.At(i)); err != nil {
			return errnie.Error(err)
		}
	}

	// Add the new attribute at the last position.
	newAttr := newAttrs.At(attrs.Len())
	if err := newAttr.SetKey(key); err != nil {
		return errnie.Error(err)
	}
	if err := newAttr.SetValue(value); err != nil {
		return errnie.Error(err)
	}

	// Set the updated list of attributes back to the artifact.
	return errnie.Error(artifact.SetAttributes(newAttrs))
}
