package httpreqx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHttpClient(t *testing.T) {
	r := require.New(t)

	t.Run("NewHttpClient", func(t *testing.T) {
		client := NewHttpClient()
		r.NotNil(client)
		r.NotNil(client.client)
		r.Equal(20*time.Second, client.client.Timeout)
		r.NotNil(client.requestOptions)
		r.NotNil(client.requestOptions.BodyMarshaler)
		r.NotNil(client.requestOptions.BodyUnmarshaler)
	})

	t.Run("Clone", func(t *testing.T) {
		original := NewHttpClient().
			SetTimeout(30*time.Second).
			SetHeader("Authorization", "Bearer token").
			SetBodyMarshaler(NewJSONBodyMarshaler()).
			SetBodyUnmarshaler(NewJSONBodyUnmarshaler())

		clone := original.Clone()
		r.NotNil(clone)
		r.NotEqual(original, clone)
		r.Equal(original.client.Timeout, clone.client.Timeout)
		r.Equal(original.requestOptions.Headers, clone.requestOptions.Headers)
		r.Equal(original.requestOptions.BodyMarshaler, clone.requestOptions.BodyMarshaler)
		r.Equal(original.requestOptions.BodyUnmarshaler, clone.requestOptions.BodyUnmarshaler)

		// Verify independence
		clone.SetTimeout(10 * time.Second)
		r.Equal(10*time.Second, clone.client.Timeout)
		r.Equal(30*time.Second, original.client.Timeout)
	})

	t.Run("SetTimeout", func(t *testing.T) {
		client := NewHttpClient()
		timeout := 15 * time.Second
		client.SetTimeout(timeout)
		r.Equal(timeout, client.client.Timeout)
	})

	t.Run("SetHeader", func(t *testing.T) {
		client := NewHttpClient()
		client.SetHeader("Authorization", "Bearer token123")
		r.Equal("Bearer token123", client.requestOptions.Headers["Authorization"])
	})

	t.Run("SetHeaders", func(t *testing.T) {
		client := NewHttpClient()
		headers := map[string]string{
			"Authorization": "Bearer token123",
			"User-Agent":    "TestClient/1.0",
			"Accept":        "application/json",
		}
		client.SetHeaders(headers)
		r.Equal(headers, client.requestOptions.Headers)
	})

	t.Run("SetBodyMarshaler", func(t *testing.T) {
		client := NewHttpClient()
		jsonMarshaler := NewJSONBodyMarshaler()
		client.SetBodyMarshaler(jsonMarshaler)
		r.Equal(jsonMarshaler, client.requestOptions.BodyMarshaler)
	})

	t.Run("SetBodyUnmarshaler", func(t *testing.T) {
		client := NewHttpClient()
		jsonUnmarshaler := NewJSONBodyUnmarshaler()
		client.SetBodyUnmarshaler(jsonUnmarshaler)
		r.Equal(jsonUnmarshaler, client.requestOptions.BodyUnmarshaler)
	})

	t.Run("SetOnRequestReady", func(t *testing.T) {
		client := NewHttpClient()
		hook := func(req *http.Request) error {
			return nil
		}
		client.SetOnRequestReady(hook)
		r.NotNil(client.requestOptions.OnRequestReady)
	})

	t.Run("SetOnResponseReady", func(t *testing.T) {
		client := NewHttpClient()
		hook := func(resp *http.Response) error {
			return nil
		}
		client.SetOnResponseReady(hook)
		r.NotNil(client.requestOptions.OnResponseReady)
	})

	t.Run("SetDumpOnError", func(t *testing.T) {
		client := NewHttpClient()
		client.SetDumpOnError()
		r.True(client.requestOptions.StackTraceEnabled)
		r.Len(client.requestOptions.OnErrorHooks, 1)
	})

	t.Run("SetStackTraceEnabled", func(t *testing.T) {
		client := NewHttpClient()
		client.SetStackTraceEnabled(true)
		r.True(client.requestOptions.StackTraceEnabled)

		client.SetStackTraceEnabled(false)
		r.False(client.requestOptions.StackTraceEnabled)
	})
}

func TestRequest(t *testing.T) {
	r := require.New(t)

	t.Run("NewRequest", func(t *testing.T) {
		client := NewHttpClient()
		ctx := context.Background()
		body := map[string]string{"key": "value"}

		req := client.NewRequest(ctx, http.MethodPost, "/test", body)
		r.NotNil(req)
		r.Equal(client, req.client)
		r.Equal(http.MethodPost, req.method)
		r.Equal("/test", req.path)
		r.Equal(ctx, req.ctx)
		r.Equal(body, req.body)
		r.Equal(client.requestOptions, req.options)
	})

	t.Run("HTTP Method Requests", func(t *testing.T) {
		client := NewHttpClient()
		ctx := context.Background()

		testCases := []struct {
			name    string
			method  string
			path    string
			body    interface{}
			creator func() *Request
		}{
			{
				name:   "GET",
				method: http.MethodGet,
				path:   "/get",
				body:   nil,
				creator: func() *Request {
					return client.NewGetRequest(ctx, "/get")
				},
			},
			{
				name:   "POST",
				method: http.MethodPost,
				path:   "/post",
				body:   "test body",
				creator: func() *Request {
					return client.NewPostRequest(ctx, "/post", "test body")
				},
			},
			{
				name:   "PUT",
				method: http.MethodPut,
				path:   "/put",
				body:   "test body",
				creator: func() *Request {
					return client.NewPutRequest(ctx, "/put", "test body")
				},
			},
			{
				name:   "PATCH",
				method: http.MethodPatch,
				path:   "/patch",
				body:   "test body",
				creator: func() *Request {
					return client.NewPatchRequest(ctx, "/patch", "test body")
				},
			},
			{
				name:   "DELETE",
				method: http.MethodDelete,
				path:   "/delete",
				body:   nil,
				creator: func() *Request {
					return client.NewDeleteRequest(ctx, "/delete")
				},
			},
			{
				name:   "OPTIONS",
				method: http.MethodOptions,
				path:   "/options",
				body:   nil,
				creator: func() *Request {
					return client.NewOptionsRequest(ctx, "/options")
				},
			},
			{
				name:   "CONNECT",
				method: http.MethodConnect,
				path:   "/connect",
				body:   nil,
				creator: func() *Request {
					return client.NewConnectRequest(ctx, "/connect")
				},
			},
			{
				name:   "HEAD",
				method: http.MethodHead,
				path:   "/head",
				body:   nil,
				creator: func() *Request {
					return client.NewHeadRequest(ctx, "/head")
				},
			},
			{
				name:   "TRACE",
				method: http.MethodTrace,
				path:   "/trace",
				body:   nil,
				creator: func() *Request {
					return client.NewTraceRequest(ctx, "/trace")
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := tc.creator()
				r.Equal(tc.method, req.method)
				r.Equal(tc.path, req.path)
				r.Equal(tc.body, req.body)
			})
		}
	})

	t.Run("WriteBodyTo", func(t *testing.T) {
		client := NewHttpClient()
		ctx := context.Background()
		req := client.NewGetRequest(ctx, "/test")

		var result string
		req.WriteBodyTo(&result)
		r.True(req.unmarshalResult)
		r.Equal(&result, req.unmarshalResultTo)
	})

	t.Run("SetBodyMarshaler", func(t *testing.T) {
		client := NewHttpClient()
		ctx := context.Background()
		req := client.NewPostRequest(ctx, "/test", "body")

		jsonMarshaler := NewJSONBodyMarshaler()
		req.SetBodyMarshaler(jsonMarshaler)
		r.Equal(jsonMarshaler, req.options.BodyMarshaler)
	})

	t.Run("SetBodyUnmarshaler", func(t *testing.T) {
		client := NewHttpClient()
		ctx := context.Background()
		req := client.NewGetRequest(ctx, "/test")

		jsonUnmarshaler := NewJSONBodyUnmarshaler()
		req.SetBodyUnmarshaler(jsonUnmarshaler)
		r.Equal(jsonUnmarshaler, req.options.BodyUnmarshaler)
	})

	t.Run("SetHeader", func(t *testing.T) {
		client := NewHttpClient()
		ctx := context.Background()
		req := client.NewGetRequest(ctx, "/test")

		req.SetHeader("X-Test", "test-value")
		r.Equal("test-value", req.options.Headers["X-Test"])
	})

	t.Run("SetHeaders", func(t *testing.T) {
		client := NewHttpClient()
		ctx := context.Background()
		req := client.NewGetRequest(ctx, "/test")

		headers := map[string]string{
			"X-Test1": "value1",
			"X-Test2": "value2",
		}
		req.SetHeaders(headers)
		r.Equal(headers, req.options.Headers)
	})

	t.Run("SetOnRequestReady", func(t *testing.T) {
		client := NewHttpClient()
		ctx := context.Background()
		req := client.NewGetRequest(ctx, "/test")

		hook := func(req *http.Request) error {
			return nil
		}
		req.SetOnRequestReady(hook)
		r.NotNil(req.options.OnRequestReady)
	})

	t.Run("SetOnResponseReady", func(t *testing.T) {
		client := NewHttpClient()
		ctx := context.Background()
		req := client.NewGetRequest(ctx, "/test")

		hook := func(resp *http.Response) error {
			return nil
		}
		req.SetOnResponseReady(hook)
		r.NotNil(req.options.OnResponseReady)
	})

	t.Run("SetDumpOnError", func(t *testing.T) {
		client := NewHttpClient()
		ctx := context.Background()
		req := client.NewGetRequest(ctx, "/test")

		req.SetDumpOnError()
		r.True(req.options.StackTraceEnabled)
		r.Len(req.options.OnErrorHooks, 1)
	})

	t.Run("SetStackTraceEnabled", func(t *testing.T) {
		client := NewHttpClient()
		ctx := context.Background()
		req := client.NewGetRequest(ctx, "/test")

		req.SetStackTraceEnabled(true)
		r.True(req.options.StackTraceEnabled)

		req.SetStackTraceEnabled(false)
		r.False(req.options.StackTraceEnabled)
	})
}

func TestRequestExecution(t *testing.T) {
	r := require.New(t)

	t.Run("Successful GET Request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/success":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status": "success", "message": "Hello World"}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		client := NewHttpClient()
		ctx := context.Background()

		var result string
		resp, err := client.NewGetRequest(ctx, server.URL+"/success").
			WriteBodyTo(&result).
			Do()

		r.NoError(err)
		r.NotNil(resp)
		r.Equal(http.StatusOK, resp.StatusCode)
		r.Contains(result, "success")
		r.Contains(result, "Hello World")
	})

	t.Run("Successful POST Request with JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/post":
				if r.Method != http.MethodPost {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				if r.Header.Get("Content-Type") != "application/json" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				var requestBody map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				response := map[string]interface{}{
					"received": requestBody,
					"status":   "success",
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		client := NewHttpClient().
			SetBodyMarshaler(NewJSONBodyMarshaler()).
			SetBodyUnmarshaler(NewJSONBodyUnmarshaler())

		ctx := context.Background()
		requestBody := map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
		}

		var result map[string]interface{}
		resp, err := client.NewPostRequest(ctx, server.URL+"/post", requestBody).
			WriteBodyTo(&result).
			Do()

		r.NoError(err)
		r.NotNil(resp)
		r.Equal(http.StatusOK, resp.StatusCode)
		r.Equal("success", result["status"])
		r.Equal(requestBody, result["received"])
	})

	t.Run("Request with Custom Headers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/headers":
				response := map[string]string{
					"authorization": r.Header.Get("Authorization"),
					"user-agent":    r.Header.Get("User-Agent"),
					"x-custom":      r.Header.Get("X-Custom"),
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		client := NewHttpClient().
			SetHeader("Authorization", "Bearer client-token").
			SetHeader("User-Agent", "TestClient/1.0")

		ctx := context.Background()

		var result string
		resp, err := client.NewGetRequest(ctx, server.URL+"/headers").
			SetHeader("X-Custom", "request-header").
			WriteBodyTo(&result).
			Do()

		r.NoError(err)
		r.NotNil(resp)
		r.Equal(http.StatusOK, resp.StatusCode)
		r.Contains(result, "Bearer client-token")
		r.Contains(result, "TestClient/1.0")
		r.Contains(result, "request-header")
	})

	t.Run("Request with Hooks", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/hooks":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status": "success"}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		client := NewHttpClient().
			SetOnRequestReady(func(req *http.Request) error {
				req.Header.Set("X-Request-Hook", "called")
				return nil
			}).
			SetOnResponseReady(func(resp *http.Response) error {
				return nil
			})

		ctx := context.Background()

		resp, err := client.NewGetRequest(ctx, server.URL+"/hooks").Do()

		r.NoError(err)
		r.NotNil(resp)
		r.Equal(http.StatusOK, resp.StatusCode)
	})

	t.Run("Request with NoopBodyMarshaler", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/noop":
				body, _ := json.Marshal(map[string]string{
					"received_body": r.Header.Get("X-Body-Length"),
				})
				w.Header().Set("Content-Type", "application/json")
				w.Write(body)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		client := NewHttpClient().
			SetBodyUnmarshaler(NewJSONBodyUnmarshaler())

		ctx := context.Background()

		testCases := []struct {
			name     string
			body     interface{}
			expected string
		}{
			{
				name:     "String Body",
				body:     "test string body",
				expected: "16",
			},
			{
				name:     "Byte Slice Body",
				body:     []byte("test byte body"),
				expected: "14",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var result map[string]string
				var bodyLength int
				switch v := tc.body.(type) {
				case string:
					bodyLength = len(v)
				case []byte:
					bodyLength = len(v)
				default:
					bodyLength = len(fmt.Sprintf("%v", tc.body))
				}
				resp, err := client.NewPostRequest(ctx, server.URL+"/noop", tc.body).
					SetHeader("X-Body-Length", fmt.Sprintf("%d", bodyLength)).
					WriteBodyTo(&result).
					Do()

				r.NoError(err)
				r.NotNil(resp)
				r.Equal(http.StatusOK, resp.StatusCode)
				r.Equal(tc.expected, result["received_body"])
			})
		}
	})

	t.Run("Request with NoopBodyUnmarshaler", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/noop-unmarshal":
				w.Write([]byte("Hello, World!"))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		client := NewHttpClient()
		ctx := context.Background()

		t.Run("String Result", func(t *testing.T) {
			var result string
			resp, err := client.NewGetRequest(ctx, server.URL+"/noop-unmarshal").
				WriteBodyTo(&result).
				Do()

			r.NoError(err)
			r.NotNil(resp)
			r.Equal(http.StatusOK, resp.StatusCode)
			r.Equal("Hello, World!", result)
		})

		t.Run("Byte Slice Result", func(t *testing.T) {
			var result []byte
			resp, err := client.NewGetRequest(ctx, server.URL+"/noop-unmarshal").
				WriteBodyTo(&result).
				Do()

			r.NoError(err)
			r.NotNil(resp)
			r.Equal(http.StatusOK, resp.StatusCode)
			r.Equal([]byte("Hello, World!"), result)
		})
	})

	t.Run("Error Handling - 4xx Status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/not-found":
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error": "Not Found"}`))
			case "/bad-request":
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "Bad Request"}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		client := NewHttpClient()
		ctx := context.Background()

		testCases := []struct {
			name       string
			path       string
			statusCode int
		}{
			{
				name:       "404 Not Found",
				path:       "/not-found",
				statusCode: http.StatusNotFound,
			},
			{
				name:       "400 Bad Request",
				path:       "/bad-request",
				statusCode: http.StatusBadRequest,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := client.NewGetRequest(ctx, server.URL+tc.path).Do()

				r.Error(err)
				r.NotNil(resp)
				r.Equal(tc.statusCode, resp.StatusCode)
				r.Contains(err.Error(), fmt.Sprintf(":%d", tc.statusCode))
			})
		}
	})

	t.Run("Error Handling - 5xx Status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/internal-error":
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Internal Server Error"}`))
			case "/service-unavailable":
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{"error": "Service Unavailable"}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		client := NewHttpClient()
		ctx := context.Background()

		testCases := []struct {
			name       string
			path       string
			statusCode int
		}{
			{
				name:       "500 Internal Server Error",
				path:       "/internal-error",
				statusCode: http.StatusInternalServerError,
			},
			{
				name:       "503 Service Unavailable",
				path:       "/service-unavailable",
				statusCode: http.StatusServiceUnavailable,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := client.NewGetRequest(ctx, server.URL+tc.path).Do()

				r.Error(err)
				r.NotNil(resp)
				r.Equal(tc.statusCode, resp.StatusCode)
				r.Contains(err.Error(), fmt.Sprintf(":%d", tc.statusCode))
			})
		}
	})

	t.Run("Request with DumpOnError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/error":
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Something went wrong"}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		client := NewHttpClient().SetDumpOnError()
		ctx := context.Background()

		resp, err := client.NewGetRequest(ctx, server.URL+"/error").Do()

		r.Error(err)
		r.NotNil(resp)
		r.Equal(http.StatusInternalServerError, resp.StatusCode)
		r.Contains(err.Error(), ":500")
	})

	t.Run("Request with Stack Trace", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/error":
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "Bad Request"}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		client := NewHttpClient().SetStackTraceEnabled(true)
		ctx := context.Background()

		resp, err := client.NewGetRequest(ctx, server.URL+"/error").Do()

		r.Error(err)
		r.NotNil(resp)
		r.Equal(http.StatusBadRequest, resp.StatusCode)
		r.Contains(err.Error(), ":400")
	})

	t.Run("Manual Response Body Handling", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/manual":
				w.Write([]byte("Manual body handling"))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		client := NewHttpClient()
		ctx := context.Background()

		resp, err := client.NewGetRequest(ctx, server.URL+"/manual").Do()

		r.NoError(err)
		r.NotNil(resp)
		r.Equal(http.StatusOK, resp.StatusCode)

		// Manually read and close the body
		body, err := json.Marshal(resp.Body)
		r.NoError(err)
		resp.Body.Close()

		// Verify body was read correctly
		r.NotEmpty(body)
	})

	t.Run("Request with Different HTTP Methods", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/methods":
				response := map[string]string{
					"method": r.Method,
					"status": "success",
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		client := NewHttpClient().SetBodyUnmarshaler(NewJSONBodyUnmarshaler())
		ctx := context.Background()

		testCases := []struct {
			name   string
			method string
			body   interface{}
		}{
			{
				name:   "GET",
				method: http.MethodGet,
				body:   nil,
			},
			{
				name:   "POST",
				method: http.MethodPost,
				body:   "test body",
			},
			{
				name:   "PUT",
				method: http.MethodPut,
				body:   "test body",
			},
			{
				name:   "PATCH",
				method: http.MethodPatch,
				body:   "test body",
			},
			{
				name:   "DELETE",
				method: http.MethodDelete,
				body:   nil,
			},
			{
				name:   "OPTIONS",
				method: http.MethodOptions,
				body:   nil,
			},
			{
				name:   "HEAD",
				method: http.MethodHead,
				body:   nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var result map[string]string
				var resp *http.Response
				var err error

				switch tc.method {
				case http.MethodGet:
					resp, err = client.NewGetRequest(ctx, server.URL+"/methods").
						WriteBodyTo(&result).Do()
				case http.MethodPost:
					resp, err = client.NewPostRequest(ctx, server.URL+"/methods", tc.body).
						WriteBodyTo(&result).Do()
				case http.MethodPut:
					resp, err = client.NewPutRequest(ctx, server.URL+"/methods", tc.body).
						WriteBodyTo(&result).Do()
				case http.MethodPatch:
					resp, err = client.NewPatchRequest(ctx, server.URL+"/methods", tc.body).
						WriteBodyTo(&result).Do()
				case http.MethodDelete:
					resp, err = client.NewDeleteRequest(ctx, server.URL+"/methods").
						WriteBodyTo(&result).Do()
				case http.MethodOptions:
					resp, err = client.NewOptionsRequest(ctx, server.URL+"/methods").
						WriteBodyTo(&result).Do()
				case http.MethodHead:
					resp, err = client.NewHeadRequest(ctx, server.URL+"/methods").Do()
				}

				if tc.method == http.MethodHead {
					// HEAD requests typically don't have a body
					r.NoError(err)
					r.NotNil(resp)
					r.Equal(http.StatusOK, resp.StatusCode)
				} else {
					r.NoError(err)
					r.NotNil(resp)
					r.Equal(http.StatusOK, resp.StatusCode)
					r.Equal(tc.method, result["method"])
					r.Equal("success", result["status"])
				}
			})
		}
	})

	t.Run("Request with JSON Marshaler/Unmarshaler", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/json":
				if r.Header.Get("Content-Type") != "application/json" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				var requestBody map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				response := map[string]interface{}{
					"received": requestBody,
					"status":   "success",
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		client := NewHttpClient().
			SetBodyMarshaler(NewJSONBodyMarshaler()).
			SetBodyUnmarshaler(NewJSONBodyUnmarshaler())

		ctx := context.Background()

		type TestStruct struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Age   int    `json:"age"`
		}

		requestBody := TestStruct{
			Name:  "Jane Doe",
			Email: "jane@example.com",
			Age:   30,
		}

		var result map[string]interface{}
		resp, err := client.NewPostRequest(ctx, server.URL+"/json", requestBody).
			WriteBodyTo(&result).
			Do()

		r.NoError(err)
		r.NotNil(resp)
		r.Equal(http.StatusOK, resp.StatusCode)
		r.Equal("success", result["status"])

		received := result["received"].(map[string]interface{})
		r.Equal("Jane Doe", received["name"])
		r.Equal("jane@example.com", received["email"])
		r.Equal(float64(30), received["age"]) // JSON numbers are unmarshaled as float64
	})

	t.Run("Request with Request-Level Overrides", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/override":
				response := map[string]string{
					"client_header":  r.Header.Get("X-Client-Header"),
					"request_header": r.Header.Get("X-Request-Header"),
					"content_type":   r.Header.Get("Content-Type"),
					"accept":         r.Header.Get("Accept"),
					"status":         "success",
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		client := NewHttpClient().
			SetHeader("X-Client-Header", "client-value").
			SetBodyMarshaler(NewJSONBodyMarshaler()).
			SetBodyUnmarshaler(NewJSONBodyUnmarshaler())

		ctx := context.Background()

		var result string
		resp, err := client.NewGetRequest(ctx, server.URL+"/override").
			SetHeader("X-Request-Header", "request-value").
			SetBodyMarshaler(NewNoopBodyMarshaler()).
			SetBodyUnmarshaler(NewNoopBodyUnmarshaler()).
			WriteBodyTo(&result).
			Do()

		r.NoError(err)
		r.NotNil(resp)
		r.Equal(http.StatusOK, resp.StatusCode)
		r.Contains(result, "client-value")
		r.Contains(result, "request-value")
		r.Contains(result, "success")
	})

	t.Run("Request with Error in Hook", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/hook-error":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status": "success"}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		ctx := context.Background()

		// Test request hook error
		t.Run("Request Hook Error", func(t *testing.T) {
			resp, err := NewHttpClient().NewGetRequest(ctx, server.URL+"/hook-error").
				SetOnRequestReady(func(req *http.Request) error {
					return fmt.Errorf("request hook error")
				}).
				Do()

			r.Error(err)
			r.Nil(resp)
			r.Contains(err.Error(), "request hook error")
		})

		// Test response hook error
		t.Run("Response Hook Error", func(t *testing.T) {
			resp, err := NewHttpClient().NewGetRequest(ctx, server.URL+"/hook-error").
				SetOnResponseReady(func(resp *http.Response) error {
					return fmt.Errorf("response hook error")
				}).
				Do()

			r.Error(err)
			r.NotNil(resp)
			r.Contains(err.Error(), "response hook error")
		})
	})

	t.Run("Request with Missing Body Marshaler", func(t *testing.T) {
		client := NewHttpClient()
		// Clear the body marshaler to test the error case
		client.requestOptions.BodyMarshaler = nil
		ctx := context.Background()

		resp, err := client.NewPostRequest(ctx, "http://example.com", "body").Do()

		r.Error(err)
		r.Nil(resp)
		r.Contains(err.Error(), "body marshaler is not set")
	})

	t.Run("Request with Missing Body Unmarshaler", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/missing-unmarshaler":
				w.Write([]byte("test response"))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		client := NewHttpClient()
		// Clear the body unmarshaler to test the error case
		client.requestOptions.BodyUnmarshaler = nil
		ctx := context.Background()

		var result string
		resp, err := client.NewGetRequest(ctx, server.URL+"/missing-unmarshaler").
			WriteBodyTo(&result).
			Do()

		r.Error(err)
		r.NotNil(resp)
		r.Contains(err.Error(), "body unmarshaler is not set")
	})
}
