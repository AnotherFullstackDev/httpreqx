package main

import "net/http"

type OnRequestReadyHook func(req *http.Request) error

type OnResponseReadyHook func(resp *http.Response) error
