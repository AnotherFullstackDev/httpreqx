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

func (r *Request) WriteBodyTo(result interface{}) *Request {
	r.unmarshalResultTo = result
	r.unmarshalResult = true
	return r
}

func (r *Request) SetBodyMarshaler(marshaler BodyMarshaler) *Request {
	r.options.SetBodyMarshaler(marshaler)
	return r
}

func (r *Request) SetBodyUnmarshaler(unmarshaler BodyUnmarshaler) *Request {
	r.options.SetBodyUnmarshaler(unmarshaler)
	return r
}

func (r *Request) SetHeaders(headers map[string]string) *Request {
	r.options.SetHeaders(headers)
	return r
}

func (r *Request) SetHeader(key, value string) *Request {
	r.options.SetHeader(key, value)
	return r
}

func (r *Request) SetOnRequestReady(onRequestReady OnRequestReadyHook) *Request {
	r.options.SetOnRequestReady(onRequestReady)
	return r
}

func (r *Request) SetOnResponseReady(onResponseReady OnResponseReadyHook) *Request {
	r.options.SetOnResponseReady(onResponseReady)
	return r
}

func (r *Request) SetDumpOnError() *Request {
	r.options.SetDumpOnError()
	return r
}

func (r *Request) SetStackTraceEnabled(enabled bool) *Request {
	r.options.SetStackTraceEnabled(enabled)
	return r
}

func (r *Request) Do() (*http.Response, error) {
	var beforeRequestHooks []OnRequestReadyHook

	bodyBuffer := bytes.NewBuffer(nil)
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

	resp, err := r.client.Do(req)
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
			defer func() {
				if err := resp.Body.Close(); err != nil {
					// Log the error, but do not return it, as we already have a response.
					fmt.Printf("Error closing response body: %v\n", err)
				}
			}()

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
