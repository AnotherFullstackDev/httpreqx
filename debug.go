package httpreqx

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
)

//type EnrichedError error

type EnrichedError struct {
	Err   error
	Stack string
}

func (e *EnrichedError) Error() string {
	if e == nil {
		return "<nil>"
	}

	if e.Err == nil {
		return fmt.Sprintf("<nil>\nStack trace:\n%s", e.Stack)
	}

	if e.Stack == "" {
		return e.Err.Error()
	}

	return fmt.Sprintf("%s\nStack trace:\n%s", e.Err.Error(), e.Stack)
}

func enrichErrorWithStackTrace(err error) error {
	stacktrace := string(debug.Stack())
	return &EnrichedError{err, stacktrace}
}

func dumpError(err error) {
	if err == nil {
		return
	}

	var enrichedError *EnrichedError
	if !errors.As(err, &enrichedError) {
		err = enrichErrorWithStackTrace(err)
	}

	fmt.Printf("ERROR: %s\n", err)
}

func dumpRequest(req *http.Request) {
	if req == nil {
		return
	}

	var headers []string
	for name, values := range req.Header {
		for _, value := range values {
			headers = append(headers, fmt.Sprintf("%s: %s", name, value))
		}
	}

	var body string
	// Handles scenarios when the request body is already consumed
	bodyReader, _ := req.GetBody()
	if bodyReader != nil {
		bodyBytes, _ := io.ReadAll(bodyReader)
		body = string(bodyBytes)
	}

	if body != "" {
		body = strings.TrimSpace(body)
	} else {
		body = "<empty>"
	}

	fmt.Printf("Request: %s %s\n", req.Method, req.URL.String())
	fmt.Printf("Request headers:\n%s\n", strings.Join(headers, "\n"))
	fmt.Printf("Request body: %v\n", body)
}

func dumpResponse(resp *http.Response) {
	if resp == nil {
		return
	}

	var headers []string
	for name, values := range resp.Header {
		for _, value := range values {
			headers = append(headers, fmt.Sprintf("%s: %s", name, value))
		}
	}

	var body string
	if resp.Body != nil {
		bodyBytes, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(strings.NewReader(string(bodyBytes))) // Reset body for further use
		body = string(bodyBytes)
	}

	if body == "" {
		body = "<empty>"
	}

	fmt.Printf("Response status: %s\n", resp.Status)
	fmt.Printf("Response status code: %d\n", resp.StatusCode)
	fmt.Printf("Response headers:\n%s\n", strings.Join(headers, "\n"))
	fmt.Printf("Response body: %s\n", body)
}

func dumpBody(body interface{}) {
	if body == nil {
		fmt.Println("Original body: <nil>")
		return
	}

	var normalizedBody string
	switch v := body.(type) {
	case string:
		normalizedBody = v
	case []byte:
		normalizedBody = string(v)
	default:
		normalizedBody = fmt.Sprintf("%v", v)
	}

	fmt.Printf("Original body: %s\n", normalizedBody)
}
