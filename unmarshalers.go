package httpreqx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

// NewJSONBodyUnmarshaler creates a BodyUnmarshaler that unmarshalls the response body as JSON format.
// It automatically sets the Accept header to application/json.
// Unmarshalling is done via the json.NewDecoder function, that uses streaming decoding.
func NewJSONBodyUnmarshaler() BodyUnmarshaler {
	return &JSONBodyUnmarshaler{}
}

type NoopBodyUnmarshaler struct{}

func (u *NoopBodyUnmarshaler) Unmarshal(result interface{}, reader io.Reader) error {
	if reader == nil {
		return errors.New("reader is nil")
	}

	if result == nil {
		return errors.New("result destination is nil")
	}

	var writer io.Writer
	switch v := result.(type) {
	case io.Writer:
		writer = v

	// Extra handling for common types
	case *[]byte:
		buf := &bytes.Buffer{}
		if _, err := io.Copy(buf, reader); err != nil {
			return err
		}

		*v = buf.Bytes()

		return nil
	case *string:
		buf := &bytes.Buffer{}
		if _, err := io.Copy(buf, reader); err != nil {
			return err
		}

		*v = buf.String()

		return nil
	default:
		return fmt.Errorf("unsupported result destination for NoopBodyUnmarshaler: %T", result)
	}

	_, err := io.Copy(writer, reader)
	return err
}

func (u *NoopBodyUnmarshaler) OnRequestReady(_ *http.Request) error {
	return nil
}

// NewNoopBodyUnmarshaler creates a BodyUnmarshaler that does not modify the response body.
// It simply writes the response body to the destination without any additional processing.
// Allowed result destinations are:
// - io.Writer
// - *[]byte
// - *string
func NewNoopBodyUnmarshaler() BodyUnmarshaler {
	return &NoopBodyUnmarshaler{}
}
