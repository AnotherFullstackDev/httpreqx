package httpreqx

import (
	"net/http"
)

type RequestOptions struct {
	BodyMarshaler     BodyMarshaler
	BodyUnmarshaler   BodyUnmarshaler
	Headers           map[string]string
	OnRequestReady    OnRequestReadyHook
	OnResponseReady   OnResponseReadyHook
	OnErrorHooks      []onErrorHook
	StackTraceEnabled bool
}

func (o *RequestOptions) Clone() *RequestOptions {
	clone := &RequestOptions{
		BodyMarshaler:     o.BodyMarshaler,
		BodyUnmarshaler:   o.BodyUnmarshaler,
		Headers:           make(map[string]string),
		OnRequestReady:    o.OnRequestReady,
		OnResponseReady:   o.OnResponseReady,
		OnErrorHooks:      append([]onErrorHook{}, o.OnErrorHooks...),
		StackTraceEnabled: o.StackTraceEnabled,
	}

	for k, v := range o.Headers {
		clone.Headers[k] = v
	}

	return clone
}

func (o *RequestOptions) SetBodyMarshaler(marshaler BodyMarshaler) {
	o.BodyMarshaler = marshaler
}

func (o *RequestOptions) SetBodyUnmarshaler(unmarshaler BodyUnmarshaler) {
	o.BodyUnmarshaler = unmarshaler
}

func (o *RequestOptions) SetHeaders(headers map[string]string) {
	if o.Headers == nil {
		o.Headers = make(map[string]string)
	}

	for k, v := range headers {
		o.Headers[k] = v
	}
}

func (o *RequestOptions) SetHeader(key, value string) {
	o.SetHeaders(map[string]string{key: value})
}

func (o *RequestOptions) SetOnRequestReady(onRequestReady OnRequestReadyHook) {
	o.OnRequestReady = onRequestReady
}

func (o *RequestOptions) SetOnResponseReady(onResponseReady OnResponseReadyHook) {
	o.OnResponseReady = onResponseReady
}

func (o *RequestOptions) SetDumpOnError() {
	o.SetStackTraceEnabled(true)
	o.OnErrorHooks = make([]onErrorHook, 0)
	o.OnErrorHooks = append(o.OnErrorHooks, func(req *http.Request, resp *http.Response, _ error, body interface{}) {
		dumpRequest(req)
		dumpResponse(resp)
		dumpBody(body)
	})
}

func (o *RequestOptions) SetStackTraceEnabled(enabled bool) {
	o.StackTraceEnabled = enabled
}
