package httpreqx

import (
	"net/http"
	"time"
)

// Goals
//
// - API must follow and use the native golang API as much as possible.
// - No useless operations, like unnecessary unmarshalling or so.
// - Chain based style of usage. Avoid unnecessary imports and extra error handling.
// - No global state to make it goroutine safe.

type HttpClient struct {
	client         *http.Client
	requestOptions *RequestOptions
}

func NewHttpClient() *HttpClient {
	return &HttpClient{
		client: &http.Client{
			Timeout: time.Second * 20,
		},
		requestOptions: &RequestOptions{
			BodyMarshaler: NewNoopBodyMarshaler(),
		},
	}
}

func (c *HttpClient) Clone() *HttpClient {
	clone := &HttpClient{
		client:         c.client,
		requestOptions: c.requestOptions.Clone(),
	}
	return clone
}

func (c *HttpClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

func (c *HttpClient) SetBodyMarshaler(marshaler BodyMarshaler) *HttpClient {
	c.requestOptions.SetBodyMarshaler(marshaler)
	return c
}

func (c *HttpClient) SetBodyUnmarshaler(unmarshaler BodyUnmarshaler) *HttpClient {
	c.requestOptions.SetBodyUnmarshaler(unmarshaler)
	return c
}

func (c *HttpClient) SetHeaders(headers map[string]string) *HttpClient {
	c.requestOptions.SetHeaders(headers)
	return c
}

func (c *HttpClient) SetHeader(key, value string) *HttpClient {
	c.requestOptions.SetHeader(key, value)
	return c
}

func (c *HttpClient) SetTimeout(timeout time.Duration) *HttpClient {
	c.client.Timeout = timeout
	return c
}

func (c *HttpClient) SetOnRequestReady(onRequestReady OnRequestReadyHook) *HttpClient {
	c.requestOptions.SetOnRequestReady(onRequestReady)
	return c
}

func (c *HttpClient) SetOnResponseReady(onResponseReady OnResponseReadyHook) *HttpClient {
	c.requestOptions.SetOnResponseReady(onResponseReady)
	return c
}

func (c *HttpClient) SetDumpOnError() *HttpClient {
	c.requestOptions.SetDumpOnError()
	return c
}

func (c *HttpClient) SetStackTraceEnabled(enabled bool) *HttpClient {
	c.requestOptions.SetStackTraceEnabled(enabled)
	return c
}
