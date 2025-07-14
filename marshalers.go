package httpreqx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// BodyMarshaler is an interface for marshaling request bodies.
// Allows to configure the request body according to the marshaling type via the OnRequestReady method.
type BodyMarshaler interface {
	Marshal(body interface{}, writer io.Writer) error
	OnRequestReady(req *http.Request) error
}

// BodyUnmarshaler is an interface for unmarshalling response bodies.
// Allows to configure the request body according to the unmarshalling type via the OnRequestReady method.
type BodyUnmarshaler interface {
	Unmarshal(result interface{}, reader io.Reader) error
	OnRequestReady(req *http.Request) error
}

type JSONBodyMarshaler struct{}

func (m *JSONBodyMarshaler) Marshal(body interface{}, writer io.Writer) error {
	if body == nil {
		return errors.New("body is nil")
	}

	if err := json.NewEncoder(writer).Encode(body); err != nil {
		return err
	}

	return nil
}

func (m *JSONBodyMarshaler) OnRequestReady(req *http.Request) error {
	req.Header.Set(HeaderContentType, "application/json")
	return nil
}

// NewJSONBodyMarshaler creates a BodyMarshaler that marshals the body to JSON format.
// It automatically sets the Content-Type header to application/json.
// The body can be any type that is supported by the json.Marshal function.
// Marshaling is done using the json.NewEncoder function, that uses streaming encoding.
// A caveat is that a new line is added at the end of the body, which is a requirement for the JSON format (see https://go.dev/src/encoding/json/stream.go, (enc *Encoder) Encode method, line 221).
func NewJSONBodyMarshaler() BodyMarshaler {
	return &JSONBodyMarshaler{}
}

type NoopBodyMarshaler struct{}

func (m *NoopBodyMarshaler) Marshal(body interface{}, writer io.Writer) error {
	if body == nil {
		return errors.New("body is nil")
	}

	var reader io.Reader
	switch v := body.(type) {
	case []byte:
		reader = bytes.NewReader(v)
	case string:
		reader = strings.NewReader(v)
	case io.Reader:
		reader = v
	default:
		return fmt.Errorf("unsupported body type: %T", body)
	}

	_, err := io.Copy(writer, reader)
	return err
}

func (m *NoopBodyMarshaler) OnRequestReady(_ *http.Request) error {
	return nil
}

// NewNoopBodyMarshaler creates a BodyMarshaler that does not modify the request body.
// It allows to create requests by passing the same type as the standard http.NewRequestWithContext accepts, with some additions for convenience.
// The modifications are:
// - Automatically converts string to strings.Reader if the body is a string.
func NewNoopBodyMarshaler() BodyMarshaler {
	return &NoopBodyMarshaler{}
}
