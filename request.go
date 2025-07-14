package httpreqx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
)

type Request struct {
	client            *HttpClient
	method            string
	path              string
	ctx               context.Context
	body              interface{}
	unmarshalResultTo interface{}
	unmarshalResult   bool
	options           *RequestOptions
}

// NewRequest creates a new Request with the specified method, path, and body.
// It is mostly used internally.
// For convenience, you can use the NewGetRequest, NewPostRequest, etc. methods to create requests with common HTTP methods.
func (c *HttpClient) NewRequest(ctx context.Context, method, path string, body interface{}) *Request {
	return &Request{
		client:  c,
		method:  method,
		path:    path,
		ctx:     ctx,
		body:    body,
		options: c.requestOptions,
	}
}

func (c *HttpClient) NewGetRequest(ctx context.Context, path string) *Request {
	return c.NewRequest(ctx, http.MethodGet, path, nil)
}

func (c *HttpClient) NewPostRequest(ctx context.Context, path string, body interface{}) *Request {
	return c.NewRequest(ctx, http.MethodPost, path, body)
}

func (c *HttpClient) NewPutRequest(ctx context.Context, path string, body interface{}) *Request {
	return c.NewRequest(ctx, http.MethodPut, path, body)
}

func (c *HttpClient) NewPatchRequest(ctx context.Context, path string, body interface{}) *Request {
	return c.NewRequest(ctx, http.MethodPatch, path, body)
}

func (c *HttpClient) NewDeleteRequest(ctx context.Context, path string) *Request {
	return c.NewRequest(ctx, http.MethodDelete, path, nil)
}

func (c *HttpClient) NewOptionsRequest(ctx context.Context, path string) *Request {
	return c.NewRequest(ctx, http.MethodOptions, path, nil)
}

func (c *HttpClient) NewConnectRequest(ctx context.Context, path string) *Request {
	return c.NewRequest(ctx, http.MethodConnect, path, nil)
}

func (c *HttpClient) NewHeadRequest(ctx context.Context, path string) *Request {
	return c.NewRequest(ctx, http.MethodHead, path, nil)
}

func (c *HttpClient) NewTraceRequest(ctx context.Context, path string) *Request {
	return c.NewRequest(ctx, http.MethodTrace, path, nil)
}

// WriteBodyTo sets the destination for unmarshalling the response body.
// This method will consume the response body and close it after reading.
// This is the recommended way to consume the response body as it prevents resource leaks, provides type safety and a unified way to work with body.
// In case this method in not used, the caller must close the response body manually after reading it to prevent resource leaks!
func (r *Request) WriteBodyTo(result interface{}) *Request {
	r.unmarshalResultTo = result
	r.unmarshalResult = true
	return r
}

// SetBodyMarshaler sets the BodyMarshaler at the request level. Does not affect the client.
func (r *Request) SetBodyMarshaler(marshaler BodyMarshaler) *Request {
	r.options.SetBodyMarshaler(marshaler)
	return r
}

// SetBodyUnmarshaler sets the BodyUnmarshaler at the request level. Does not affect the client.
func (r *Request) SetBodyUnmarshaler(unmarshaler BodyUnmarshaler) *Request {
	r.options.SetBodyUnmarshaler(unmarshaler)
	return r
}

// SetHeaders sets the headers for the request. This will override headers with the same name set at the client level but only for this request.
func (r *Request) SetHeaders(headers map[string]string) *Request {
	r.options.SetHeaders(headers)
	return r
}

// SetHeader sets a single header for the request. This will override header with the same name set at the client level but only for this request.
func (r *Request) SetHeader(key, value string) *Request {
	r.options.SetHeader(key, value)
	return r
}

// SetOnRequestReady sets a hook that will be called right after an http.Request is created and all headers and body are set.
// This method will override any hooks set at the client level, without affecting the client, but only for this request.
func (r *Request) SetOnRequestReady(onRequestReady OnRequestReadyHook) *Request {
	r.options.SetOnRequestReady(onRequestReady)
	return r
}

// SetOnResponseReady sets a hook that will be called right after the response is received and before it is processed.
// This method will override any hooks set at the client level, without affecting the client, but only for this request.
func (r *Request) SetOnResponseReady(onResponseReady OnResponseReadyHook) *Request {
	r.options.SetOnResponseReady(onResponseReady)
	return r
}

// SetDumpOnError configures logging of the request, response and error when an error occurs.
// http.Request and http.Response bodies will be logged as well, if they are set.
// Original body passed by the caller code will be logged as well, if it is set.
// This method will also enable the StackTraceEnabled option, which will add a stack trace to the error if it occurs.
func (r *Request) SetDumpOnError() *Request {
	r.options.SetDumpOnError()
	return r
}

// SetStackTraceEnabled enables or disables the stack trace in the error if it occurs.
func (r *Request) SetStackTraceEnabled(enabled bool) *Request {
	r.options.SetStackTraceEnabled(enabled)
	return r
}

// Do method executes the configured HTTP request and returns the http.Response.
func (r *Request) Do() (*http.Response, error) {
	var beforeRequestHooks []OnRequestReadyHook

	// TODO: consider using sync.Pool to reuse buffers for the request body. Might be beneficial for performance in high-load scenarios.
	bodyBuffer := &bytes.Buffer{}
	if r.body != nil {
		bodyMarshaler := r.options.BodyMarshaler

		if bodyMarshaler == nil {
			return nil, r.processError(nil, nil, errors.New("body marshaler is not set"), r.body)
		}

		beforeRequestHooks = append(beforeRequestHooks, bodyMarshaler.OnRequestReady)

		if err := bodyMarshaler.Marshal(r.body, bodyBuffer); err != nil {
			return nil, r.processError(nil, nil, fmt.Errorf("body marshaling: %w", err), r.body)
		}
	}

	req, err := http.NewRequestWithContext(r.ctx, r.method, r.path, bodyBuffer)
	if err != nil {
		return nil, r.processError(req, nil, err, r.body)
	}

	if r.options.Headers != nil {
		for key, value := range r.options.Headers {
			req.Header.Set(key, value)
		}
	}

	if r.options.BodyUnmarshaler != nil {
		beforeRequestHooks = append(beforeRequestHooks, r.options.BodyUnmarshaler.OnRequestReady)
	}
	if r.options.OnRequestReady != nil {
		beforeRequestHooks = append(beforeRequestHooks, r.options.OnRequestReady)
	}
	for _, beforeHook := range beforeRequestHooks {
		if err := beforeHook(req); err != nil {
			return nil, r.processError(req, nil, fmt.Errorf("on request ready hook: %w", err), r.body)
		}
	}

	resp, err := r.client.do(req)

	// Ensure the response body is closed to prevent resource leaks.
	defer func() {
		// If unmarshalling is false the body consumption will not happen inside the Do method,
		// therefore the body must be passed to the caller to handle it.
		if !r.unmarshalResult {
			return
		}

		if err := resp.Body.Close(); err != nil {
			// Log the error, but do not return it, as we already have a response.
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if err != nil {
		return nil, r.processError(req, nil, err, r.body)
	}

	var afterRequestHooks []OnResponseReadyHook
	if r.options.OnResponseReady != nil {
		afterRequestHooks = append(afterRequestHooks, r.options.OnResponseReady)
	}
	for _, afterHook := range afterRequestHooks {
		if err := afterHook(resp); err != nil {
			return resp, r.processError(req, resp, fmt.Errorf("on response ready hook: %w", err), r.body)
		}
	}

	if !IsSuccessResponse(resp) {
		err = fmt.Errorf("%s:%d", resp.Status, resp.StatusCode)
		return resp, r.processError(req, resp, err, r.body)
	}

	if r.unmarshalResult {
		if r.options.BodyUnmarshaler != nil {
			if err := r.options.BodyUnmarshaler.Unmarshal(r.unmarshalResultTo, resp.Body); err != nil {
				return resp, r.processError(req, resp, fmt.Errorf("body unmarshaling: %w", err), r.body)
			}
		} else {
			return resp, r.processError(req, resp, errors.New("result destination is provided but body unmarshaler is not set"), r.body)
		}
	}

	return resp, nil
}

func (r *Request) processError(req *http.Request, resp *http.Response, err error, body interface{}) error {
	if r.options.StackTraceEnabled {
		err = enrichErrorWithStackTrace(err)
	}

	for _, hook := range r.options.OnErrorHooks {
		hook(req, resp, err, body)
	}

	return err
}
