package main

import (
	"context"
	"fmt"
	"github.com/AnotherFullstackDev/httpreqx"
	"log"
	"net/http"
)

func main() {
	ctx := context.Background()

	// simple client
	c := httpreqx.NewHttpClient().SetBodyUnmarshaler(httpreqx.NewJSONBodyUnmarshaler())
	var result map[string]any
	_, err := c.NewPostRequest(ctx, "https://httpbin.org/post", []byte(`{"request_data": 123}`)).
		WriteBodyTo(&result).
		Do()
	log.Println(result, err)

	// JSON client
	c = httpreqx.NewHttpClient().
		SetBodyMarshaler(httpreqx.NewJSONBodyMarshaler()).
		SetBodyUnmarshaler(httpreqx.NewJSONBodyUnmarshaler())

	var result1 map[string]any
	_, err = c.NewGetRequest(ctx, "https://httpbin.org/uuid").
		WriteBodyTo(&result1).
		Do()
	log.Println(result1, err)

	body := map[string]any{
		"key": "value",
	}
	var result2 map[string]any
	_, err = c.NewPostRequest(ctx, "https://httpbin.org/post", body).
		WriteBodyTo(&result2).
		Do()
	log.Println(result2, err)

	// Cloned JSON client
	clone := c.Clone().
		SetOnRequestReady(func(req *http.Request) error {
			log.Println("OnRequestReady")
			clonedBody, err := httpreqx.CloneRequestBody(req)
			if err != nil {
				return err
			}
			req.Header.Set("x-on-request-ready", "true")
			req.Header.Set("x-cloned-body-length", fmt.Sprintf("%d", len(clonedBody)))
			return nil
		}).
		SetOnResponseReady(func(resp *http.Response) error {
			log.Println("OnResponseReady. Status:", resp.Status)
			return nil
		})
	var result3 map[string]any
	_, err = clone.NewGetRequest(ctx, "https://httpbin.org/get").WriteBodyTo(&result3).Do()
	log.Println(result3, err)

	var result4 map[string]any
	body = map[string]any{
		"key": "value",
	}
	_, err = clone.NewPostRequest(ctx, "https://httpbin.org/post", body).
		WriteBodyTo(&result4).
		Do()
	log.Println(result4, err)

	// With errors dump
	c = c.Clone().SetDumpOnError()
	result5 := map[string]any{}
	_, err = c.NewPostRequest(ctx, "https://httpbin.org/status/404", map[string]any{
		"key": "value",
	}).WriteBodyTo(&result5).Do()
	log.Println(err)

	// With dump and stack traces
	c = c.Clone().SetStackTraceEnabled(true)
	_, err = c.NewGetRequest(ctx, "https://httpbin.org/status/500").Do()
	log.Println(err)

	// With stack traces only
	c = httpreqx.NewHttpClient().SetStackTraceEnabled(true)
	_, err = c.NewGetRequest(ctx, "https://httpbin.org/status/401").Do()
	log.Println(err)
}
