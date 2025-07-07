package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type JSONBodyUnmarshaler struct{}

func (u *JSONBodyUnmarshaler) Unmarshal(result interface{}, reader io.Reader) error {
	if reader == nil {
		return errors.New("reader is nil")
	}

	if err := json.NewDecoder(reader).Decode(result); err != nil {
		return err
	}

	return nil
}

func (u *JSONBodyUnmarshaler) OnRequestReady(req *http.Request) error {
	req.Header.Set(HeaderAccept, "application/json")
	return nil
}

func NewJSONBodyUnmarshaler() BodyUnmarshaler {
	return &JSONBodyUnmarshaler{}
}
