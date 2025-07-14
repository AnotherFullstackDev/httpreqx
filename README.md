# Go httpreqx

A thin wrapper around the standard `net/http` package that simplifies sending HTTP requests in Go.

[![Go Version](https://img.shields.io/badge/go-1.20%2B-brightgreen)](https://golang.org)

## Features

- **Fluent API**: Chain-based method calls for easy request building
- **Marshalers/Unmarshalers**: Built-in JSON, bytes, string support with extensible interface
- **Request/Response Hooks**: Middleware-like functionality for request/response processing
- **Error Handling**: Optional request/response dumping and stack traces for debugging
- **Native Go Integration**: Built on top of standard `net/http` package with zero external dependencies

## Installation

```bash
go get github.com/AnotherFullstackDev/httpreqx
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/AnotherFullstackDev/httpreqx"
)

func main() {
    ctx := context.Background()
    
    // Create a new HTTP client (uses NoopBodyMarshaler and NoopBodyUnmarshaler by default)
    client := httpreqx.NewHttpClient()
    
    // Simple GET request with response body capture
    var responseBody string
    resp, err := client.NewGetRequest(ctx, "https://httpbin.org/get").
        WriteBodyTo(&responseBody).
        Do()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Status: %s\n", resp.Status)
    fmt.Printf("Response: %s\n", responseBody)
}
```

## Usage Examples

### Basic HTTP Methods

```go
ctx := context.Background()
client := httpreqx.NewHttpClient() // Uses NoopBodyMarshaler by default

// GET request
resp, err := client.NewGetRequest(ctx, "https://api.example.com/users").Do()

// POST request with different body types (NoopBodyMarshaler supports all of these)
resp, err = client.NewPostRequest(ctx, "https://api.example.com/users", []byte(`{"name": "John"}`)).Do()
resp, err = client.NewPostRequest(ctx, "https://api.example.com/users", `{"name": "John"}`).Do() // string
resp, err = client.NewPostRequest(ctx, "https://api.example.com/users", strings.NewReader(`{"name": "John"}`)).Do() // io.Reader

// PUT request
resp, err = client.NewPutRequest(ctx, "https://api.example.com/users/1", []byte(`{"name": "Jane"}`)).Do()

// DELETE request
resp, err = client.NewDeleteRequest(ctx, "https://api.example.com/users/1").Do()

// Other HTTP methods
resp, err = client.NewPatchRequest(ctx, "https://api.example.com/users/1", data).Do()
resp, err = client.NewOptionsRequest(ctx, "https://api.example.com/users").Do()
resp, err = client.NewHeadRequest(ctx, "https://api.example.com/users").Do()
```

### JSON Marshaling/Unmarshaling

```go
ctx := context.Background()

// Create client with JSON support
client := httpreqx.NewHttpClient().
    SetBodyMarshaler(httpreqx.NewJSONBodyMarshaler()).
    SetBodyUnmarshaler(httpreqx.NewJSONBodyUnmarshaler())

// POST with JSON body and JSON response
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

user := User{Name: "John", Email: "john@example.com"}
var result map[string]interface{}

// WriteBodyTo automatically handles response body consumption and closing
// This is the recommended way to handle response bodies to prevent resource leaks
resp, err := client.NewPostRequest(ctx, "https://httpbin.org/post", user).
    WriteBodyTo(&result).
    Do()

if err != nil {
    log.Fatal(err)
}

fmt.Printf("Result: %+v\n", result)
```

**Note:** The JSON marshaler adds a newline at the end of the JSON body, which is a requirement for the JSON format specification.

### Manual Response Body Handling

```go
// If you don't use WriteBodyTo, you must manually close the response body
resp, err := client.NewGetRequest(ctx, "https://api.example.com/data").Do()
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close() // Important: prevent resource leaks!

// Read body manually
body, err := io.ReadAll(resp.Body)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Raw response: %s\n", string(body))
```

### Raw Response Body Handling

The library comes with a `NoopBodyUnmarshaler` that allows you to capture raw response bodies in different formats:

```go
ctx := context.Background()
client := httpreqx.NewHttpClient() // Uses NoopBodyUnmarshaler by default

// Capture as []byte
var bodyBytes []byte
resp, err := client.NewGetRequest(ctx, "https://api.example.com/data").
    WriteBodyTo(&bodyBytes).
    Do()

// Capture as string
var bodyString string
resp, err = client.NewGetRequest(ctx, "https://api.example.com/data").
    WriteBodyTo(&bodyString).
    Do()

// Stream directly to a writer (e.g., file)
file, err := os.Create("response.txt")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

resp, err = client.NewGetRequest(ctx, "https://api.example.com/data").
    WriteBodyTo(file). // Streams directly to file
    Do()
```

### Custom Headers

```go
client := httpreqx.NewHttpClient().
    SetHeader("Authorization", "Bearer token123").
    SetHeader("User-Agent", "MyApp/1.0")

// Or set multiple headers at once
headers := map[string]string{
    "Authorization": "Bearer token123",
    "User-Agent":    "MyApp/1.0",
    "Accept":        "application/json",
}
client = httpreqx.NewHttpClient().SetHeaders(headers)

// Per-request headers
resp, err := client.NewGetRequest(ctx, "https://api.example.com/data").
    SetHeader("X-Request-ID", "req-123").
    Do()
```

### Request/Response Hooks

```go
client := httpreqx.NewHttpClient().
    SetOnRequestReady(func(req *http.Request) error {
        // Log request details
        fmt.Printf("Making request to: %s\n", req.URL)
        // Add custom headers or modify request
        req.Header.Set("X-Timestamp", time.Now().Format(time.RFC3339))
        return nil
    }).
    SetOnResponseReady(func(resp *http.Response) error {
        // Log response details
        fmt.Printf("Response status: %s\n", resp.Status)
        return nil
    })

// Per-request hooks
resp, err := client.NewGetRequest(ctx, "https://api.example.com/data").
    SetOnRequestReady(func(req *http.Request) error {
        // Request-specific logic
        return nil
    }).
    SetOnResponseReady(func(resp *http.Response) error {
        // Response-specific logic
        return nil
    }).
    Do()
```

### Error Handling and Debugging

```go
// Enable request/response dumping on errors
// This automatically enables stack traces as well
client := httpreqx.NewHttpClient().SetDumpOnError()

// Enable only stack traces (without dumping)
client = httpreqx.NewHttpClient().SetStackTraceEnabled(true)

// SetDumpOnError logs:
// - HTTP request details (method, URL, headers, body)
// - HTTP response details (status, headers, body)
// - Original request body passed by caller
// - Stack trace of the error
resp, err := client.NewGetRequest(ctx, "https://httpbin.org/status/404").Do()
if err != nil {
    log.Printf("Error with detailed info: %v\n", err)
}

// Per-request debugging (overrides client-level settings)
resp, err = client.NewGetRequest(ctx, "https://httpbin.org/status/500").
    SetDumpOnError().  // Only for this request
    SetStackTraceEnabled(true).
    Do()
```

### Client Cloning

```go
// Create a base client
baseClient := httpreqx.NewHttpClient().
    SetBodyMarshaler(httpreqx.NewJSONBodyMarshaler()).
    SetBodyUnmarshaler(httpreqx.NewJSONBodyUnmarshaler()).
    SetHeader("User-Agent", "MyApp/1.0")

// Clone for specific use case
apiClient := baseClient.Clone().
    SetHeader("Authorization", "Bearer api-token")

// Clone for another use case
adminClient := baseClient.Clone().
    SetHeader("Authorization", "Bearer admin-token").
    SetTimeout(30 * time.Second)
```

### Timeout Configuration

```go
client := httpreqx.NewHttpClient().
    SetTimeout(10 * time.Second) // 10 second timeout

resp, err := client.NewGetRequest(ctx, "https://slow-api.example.com/data").Do()
```

## Important Notes

### Default Behavior

By default, `NewHttpClient()` creates a client with:
- **20-second timeout**
- **NoopBodyMarshaler** - handles raw bytes and common types for request bodies
- **NoopBodyUnmarshaler** - handles raw response bodies with support for `io.Writer`, `*[]byte`, and `*string`

This means you can immediately start using `WriteBodyTo()` with basic types without additional configuration:

```go
client := httpreqx.NewHttpClient()

// Works out of the box
var response string
resp, err := client.NewGetRequest(ctx, url).WriteBodyTo(&response).Do()
```

### Resource Management

**Always use `WriteBodyTo()` when possible** - it automatically handles response body consumption and closing to prevent resource leaks.

```go
// ✅ Recommended: Automatic resource management
var result string // or []byte, or io.Writer
resp, err := client.NewGetRequest(ctx, url).WriteBodyTo(&result).Do()

// ❌ Manual management: You must close the body yourself
resp, err := client.NewGetRequest(ctx, url).Do()
if err != nil {
    return err
}
defer resp.Body.Close() // Required to prevent resource leaks
```

### Request vs Client Configuration

Most configuration methods can be set at both the client and request level:

- **Client-level**: Affects all requests made by that client
- **Request-level**: Overrides client settings for that specific request only

**Header Merging Behavior:**
- Client-level headers are applied to all requests
- Request-level headers are merged with client-level headers
- Request-level headers override client-level headers with the same name
- Headers with different names are combined

```go
// Client-level configuration
client := httpreqx.NewHttpClient().
    SetHeader("User-Agent", "MyApp/1.0").
    SetHeader("Authorization", "Bearer token123").
    SetBodyMarshaler(httpreqx.NewJSONBodyMarshaler())

// Request-level override (doesn't affect client)
resp, err := client.NewGetRequest(ctx, url).
    SetHeader("User-Agent", "SpecialRequest/1.0").  // Overrides client's User-Agent
    SetHeader("X-Request-ID", "req-456").           // Adds new header
    SetBodyMarshaler(httpreqx.NewNoopBodyMarshaler()). // Overrides client setting
    Do()
// Final headers: User-Agent: SpecialRequest/1.0, Authorization: Bearer token123, X-Request-ID: req-456
```

## API Reference

### HttpClient Methods

- `NewHttpClient() *HttpClient` - Creates a new HTTP client with default settings:
  - 20-second timeout
  - NoopBodyMarshaler (handles raw bytes, string, and io.Reader for request bodies)
  - NoopBodyUnmarshaler (handles raw response bodies to io.Writer, *[]byte, and *string)
- `(*HttpClient) Clone() *HttpClient` - Creates a copy of the client with the same configuration. The cloned client can be modified independently without affecting the original client.
- `(*HttpClient) SetTimeout(timeout time.Duration) *HttpClient` - Sets the timeout for the underlying http.Client. This timeout will apply to all requests made with this client.
- `(*HttpClient) SetHeader(key, value string) *HttpClient` - Sets a single header at the HttpClient level. Headers merging and override precedence is the same as with SetHeaders.
- `(*HttpClient) SetHeaders(headers map[string]string) *HttpClient` - Sets headers at the HttpClient level. Headers will affect all requests made with this client. When headers are set at the request level, they will be merged with client-level headers, with request-level headers taking precedence.
- `(*HttpClient) SetBodyMarshaler(marshaler BodyMarshaler) *HttpClient` - Sets the BodyMarshaler at the HttpClient level. This will affect all requests made with this client unless overridden at the request level.
- `(*HttpClient) SetBodyUnmarshaler(unmarshaler BodyUnmarshaler) *HttpClient` - Sets the BodyUnmarshaler at the HttpClient level. This will affect all requests made with this client unless overridden at the request level.
- `(*HttpClient) SetOnRequestReady(hook OnRequestReadyHook) *HttpClient` - Sets a hook that will be called right after an http.Request is created and all headers and body are set. This hook will be called for all requests made with this client unless overridden at the request level.
- `(*HttpClient) SetOnResponseReady(hook OnResponseReadyHook) *HttpClient` - Sets a hook that will be called right after the response is received and before it is processed. This hook will be called for all requests made with this client unless overridden at the request level.
- `(*HttpClient) SetDumpOnError() *HttpClient` - Configures logging of the request, response and error when an error occurs. http.Request and http.Response bodies will be logged as well, if they are set. Original body passed by the caller code will be logged as well. This method will also enable the StackTraceEnabled option. This will affect all requests made with this client unless overridden at the request level.
- `(*HttpClient) SetStackTraceEnabled(enabled bool) *HttpClient` - Enables or disables the stack trace in the error if it occurs. This will affect all requests made with this client unless overridden at the request level.

### Request Creation Methods

- `(*HttpClient) NewRequest(ctx context.Context, method, path string, body interface{}) *Request` - Creates a new Request with the specified method, path, and body. Mostly used internally. For convenience, use NewGetRequest, NewPostRequest, etc.
- `(*HttpClient) NewGetRequest(ctx context.Context, path string) *Request` - Creates a GET request
- `(*HttpClient) NewPostRequest(ctx context.Context, path string, body interface{}) *Request` - Creates a POST request
- `(*HttpClient) NewPutRequest(ctx context.Context, path string, body interface{}) *Request` - Creates a PUT request
- `(*HttpClient) NewPatchRequest(ctx context.Context, path string, body interface{}) *Request` - Creates a PATCH request
- `(*HttpClient) NewDeleteRequest(ctx context.Context, path string) *Request` - Creates a DELETE request
- `(*HttpClient) NewOptionsRequest(ctx context.Context, path string) *Request` - Creates an OPTIONS request
- `(*HttpClient) NewHeadRequest(ctx context.Context, path string) *Request` - Creates a HEAD request
- `(*HttpClient) NewConnectRequest(ctx context.Context, path string) *Request` - Creates a CONNECT request
- `(*HttpClient) NewTraceRequest(ctx context.Context, path string) *Request` - Creates a TRACE request

### Request Configuration Methods

- `(*Request) WriteBodyTo(result interface{}) *Request` - Sets the destination for unmarshalling the response body. This method will consume the response body and close it after reading. This is the recommended way to consume the response body as it prevents resource leaks, provides type safety and a unified way to work with body. In case this method is not used, the caller must close the response body manually after reading it to prevent resource leaks!
- `(*Request) SetHeader(key, value string) *Request` - Sets a single header for the request. This will override header with the same name set at the client level but only for this request.
- `(*Request) SetHeaders(headers map[string]string) *Request` - Sets the headers for the request. This will override headers with the same name set at the client level but only for this request.
- `(*Request) SetBodyMarshaler(marshaler BodyMarshaler) *Request` - Sets the BodyMarshaler at the request level. Does not affect the client.
- `(*Request) SetBodyUnmarshaler(unmarshaler BodyUnmarshaler) *Request` - Sets the BodyUnmarshaler at the request level. Does not affect the client.
- `(*Request) SetOnRequestReady(hook OnRequestReadyHook) *Request` - Sets a hook that will be called right after an http.Request is created and all headers and body are set. This method will override any hooks set at the client level, without affecting the client, but only for this request.
- `(*Request) SetOnResponseReady(hook OnResponseReadyHook) *Request` - Sets a hook that will be called right after the response is received and before it is processed. This method will override any hooks set at the client level, without affecting the client, but only for this request.
- `(*Request) SetDumpOnError() *Request` - Configures logging of the request, response and error when an error occurs. http.Request and http.Response bodies will be logged as well, if they are set. Original body passed by the caller code will be logged as well. This method will also enable the StackTraceEnabled option, which will add a stack trace to the error if it occurs.
- `(*Request) SetStackTraceEnabled(enabled bool) *Request` - Enables or disables the stack trace in the error if it occurs.
- `(*Request) Do() (*http.Response, error)` - Executes the configured HTTP request and returns the http.Response.

### Built-in Marshalers/Unmarshalers

- `NewJSONBodyMarshaler() BodyMarshaler` - Creates a BodyMarshaler that marshals the body to JSON format. It automatically sets the Content-Type header to application/json. The body can be any type that is supported by the json.Marshal function. Marshaling is done using the json.NewEncoder function, that uses streaming encoding. A caveat is that a new line is added at the end of the body, which is a requirement for the JSON format.
- `NewJSONBodyUnmarshaler() BodyUnmarshaler` - Creates a BodyUnmarshaler that unmarshals the response body as JSON format. It automatically sets the Accept header to application/json. Unmarshaling is done via the json.NewDecoder function, that uses streaming decoding.
- `NewNoopBodyMarshaler() BodyMarshaler` - Creates a BodyMarshaler that does not modify the request body. It allows to create requests by passing the same type as the standard http.NewRequestWithContext accepts, with some additions for convenience. The modifications are: automatically converts string to strings.Reader if the body is a string. Supports `[]byte`, `string`, and `io.Reader` body types.
- `NewNoopBodyUnmarshaler() BodyUnmarshaler` - Creates a BodyUnmarshaler that does not modify the response body. It simply writes the response body to the destination without any additional processing. Allowed result destinations are: `io.Writer`, `*[]byte`, and `*string`.

### Utility Functions

- `IsSuccessResponse(resp *http.Response) bool` - Checks if response status is 2xx
- `CloneRequestBody(req *http.Request) ([]byte, error)` - Clones request body for inspection

### HTTP Header Constants

The library provides constants for common HTTP headers:

```go
// Authorization and Authentication
httpreqx.HeaderAuthorization    // "Authorization"
httpreqx.HeaderCookie          // "Cookie"
httpreqx.HeaderSetCookie       // "Set-Cookie"

// Content and Encoding
httpreqx.HeaderContentType     // "Content-Type"
httpreqx.HeaderContentLength   // "Content-Length"
httpreqx.HeaderAccept          // "Accept"
httpreqx.HeaderAcceptEncoding  // "Accept-Encoding"
httpreqx.HeaderAcceptLanguage  // "Accept-Language"

// Request Information
httpreqx.HeaderUserAgent       // "User-Agent"
httpreqx.HeaderHost            // "Host"
httpreqx.HeaderOrigin          // "Origin"
httpreqx.HeaderReferer         // "Referer"

// Caching and Conditional Requests
httpreqx.HeaderCacheControl    // "Cache-Control"
httpreqx.HeaderETag            // "ETag"
httpreqx.HeaderIfModifiedSince // "If-Modified-Since"
httpreqx.HeaderIfNoneMatch     // "If-None-Match"

// Security and Proxy
httpreqx.HeaderXRequestedWith     // "X-Requested-With"
httpreqx.HeaderXForwardedFor      // "X-Forwarded-For"
httpreqx.HeaderXFrameOptions      // "X-Frame-Options"
httpreqx.HeaderStrictTransportSec // "Strict-Transport-Security"

// Location and Redirection
httpreqx.HeaderLocation        // "Location"
```

Usage example:

```go
client := httpreqx.NewHttpClient().
    SetHeader(httpreqx.HeaderAuthorization, "Bearer token123").
    SetHeader(httpreqx.HeaderUserAgent, "MyApp/1.0")
```

## Compatibility

- Minimum supported Go version: **1.20**
- Tested with: Go 1.20–1.24

## License

This project is licensed under the MIT License.
