package main

import (
	"bytes"
	"io"
	"net/http"
)

func IsSuccessResponse(resp *http.Response) bool {
	return resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300
}

func CloneRequestBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return bodyBytes, nil
}
