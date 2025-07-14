package httpreqx

import (
	"net/http"
	"time"
)

type HttpClient struct {
	client         *http.Client
	requestOptions *RequestOptions
}

// NewHttpClient creates a new HttpClient with default settings.
// Default settings are:
// - Timeout: 20 seconds
// - BodyMarshaler: NoopBodyMarshaler - this marshaler does not modify the request body. Allows to create requests by passing the same type as the standard http.NewRequestWithContext accepts with some additions for convenience (see NewNoopBodyMarshaler).
// - BodyUnmarshaler: NoopBodyUnmarshaler - this unmarshaler does not modify the response body, just writes it to the destination with some added handling for convenience (see NewNoopBodyUnmarshaler).
func NewHttpClient() *HttpClient {
	return &HttpClient{
		client: &http.Client{
			Timeout: time.Second * 20,
		},
		requestOptions: &RequestOptions{
			BodyMarshaler:   NewNoopBodyMarshaler(),
			BodyUnmarshaler: NewNoopBodyUnmarshaler(),
		},
	}
}

// Clone creates a new HttpClient with the same settings as the original one.
// The cloned client can be modified independently without affecting the original client.
func (c *HttpClient) Clone() *HttpClient {
	clone := &HttpClient{
		client: &http.Client{
			Timeout: c.client.Timeout,
		},
		requestOptions: c.requestOptions.Clone(),
	}

	return clone
}

func (c *HttpClient) do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// SetBodyMarshaler sets the BodyMarshaler at the HttpClient level.
// This will affect all requests made with this client unless overridden at the request level.
func (c *HttpClient) SetBodyMarshaler(marshaler BodyMarshaler) *HttpClient {
	c.requestOptions.SetBodyMarshaler(marshaler)
	return c
}

// SetBodyUnmarshaler sets the BodyUnmarshaler at the HttpClient level.
// This will affect all requests made with this client unless overridden at the request level.
func (c *HttpClient) SetBodyUnmarshaler(unmarshaler BodyUnmarshaler) *HttpClient {
	c.requestOptions.SetBodyUnmarshaler(unmarshaler)
	return c
}

// SetHeaders sets the headers at the HttpClient level.
// Headers will affect all requests made with this client.
// When headers are set at the request level, they will be merged with the ones set at the client level.
// Headers set at the request level will override the ones set at the client level for that specific request.
func (c *HttpClient) SetHeaders(headers map[string]string) *HttpClient {
	c.requestOptions.SetHeaders(headers)
	return c
}

// SetHeader sets a single header at the HttpClient level.
// Headers merging and override precedence is the same as with SetHeaders.
func (c *HttpClient) SetHeader(key, value string) *HttpClient {
	c.requestOptions.SetHeader(key, value)
	return c
}

// SetTimeout sets the timeout for the underlying http.Client.
// This timeout will apply to all requests made with this client.
func (c *HttpClient) SetTimeout(timeout time.Duration) *HttpClient {
	c.client.Timeout = timeout
	return c
}

// SetOnRequestReady sets a hook that will be called right after an http.Request is created and all headers and body are set.
// This hook will be called for all requests made with this client unless overridden at the request level.
func (c *HttpClient) SetOnRequestReady(onRequestReady OnRequestReadyHook) *HttpClient {
	c.requestOptions.SetOnRequestReady(onRequestReady)
	return c
}

// SetOnResponseReady sets a hook that will be called right after the response is received and before it is processed.
// This hook will be called for all requests made with this client unless overridden at the request level.
func (c *HttpClient) SetOnResponseReady(onResponseReady OnResponseReadyHook) *HttpClient {
	c.requestOptions.SetOnResponseReady(onResponseReady)
	return c
}

// SetDumpOnError configures logging of the request, response and error when an error occurs.
// http.Request and http.Response bodies will be logged as well, if they are set.
// Original body passed by the caller code will be logged as well, if it is set.
// This method will also enable the StackTraceEnabled option, which will add a stack trace to the error if it occurs.
// This will affect all requests made with this client unless overridden at the request level.
func (c *HttpClient) SetDumpOnError() *HttpClient {
	c.requestOptions.SetDumpOnError()
	return c
}

// SetStackTraceEnabled enables or disables the stack trace in the error if it occurs.
// This will affect all requests made with this client unless overridden at the request level.
func (c *HttpClient) SetStackTraceEnabled(enabled bool) *HttpClient {
	c.requestOptions.SetStackTraceEnabled(enabled)
	return c
}
