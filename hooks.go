package httpreqx

import "net/http"

type OnRequestReadyHook func(req *http.Request) error

type OnResponseReadyHook func(resp *http.Response) error

type onErrorHook func(req *http.Request, resp *http.Response, err error, body interface{})
