package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type BodyMarshaler interface {
	Marshal(body interface{}, writer io.Writer) error
	OnRequestReady(req *http.Request) error
}

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

func NewJSONBodyMarshaler() BodyMarshaler {
	return &JSONBodyMarshaler{}
}

type NoopBodyMarshaler struct{}

func (m *NoopBodyMarshaler) Marshal(body interface{}, writer io.Writer) error {
	if body == nil {
		return errors.New("body is nil")
	}

	var bodyBytes []byte
	switch v := body.(type) {
	case []byte:
		bodyBytes = v
	default:
		return errors.New(fmt.Sprintf("unsupported body type: %T", body))
	}

	if _, err := writer.Write(bodyBytes); err != nil {
		return err
	}

	return nil
}

func (m *NoopBodyMarshaler) OnRequestReady(req *http.Request) error {
	return nil
}

func NewNoopBodyMarshaler() BodyMarshaler {
	return &NoopBodyMarshaler{}
}
