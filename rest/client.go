package rest

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Method contains the supported HTTP verbs.
type Method string

// Supported HTTP verbs.
const (
	Get Method = "GET"
)

// Request holds the request to an API Call.
type Request struct {
	Method      Method
	BaseURL     string // https://swapi.dev
	Headers     map[string]string
	QueryParams map[string]string
	Body        []byte
}

// RestError is a struct for an error handling.
type RestError struct {
	Response *Response
}

// Error is the implementation of the error interface.
func (e *RestError) Error() string {
	return e.Response.Body
}

// DefaultClient is used if no custom HTTP client is defined
var DefaultClient = &Client{HTTPClient: &http.Client{}}

type Client struct {
	HTTPClient *http.Client
}

// Response holds the response from an API call.
type Response struct {
	StatusCode int
	Body       string
	Headers    map[string][]string
}

// AddQueryParameters adds query parameters to the URL.
func AddQueryParameters(baseURL string, queryParams map[string]string) string {
	baseURL += "?"
	params := url.Values{}
	for key, value := range queryParams {
		params.Add(key, value)
	}
	return baseURL + params.Encode()
}

// BuildRequestObject creates the HTTP request object.
func BuildRequestObject(request Request) (*http.Request, error) {
	// Add any query parameters to the URL.
	if len(request.QueryParams) != 0 {
		request.BaseURL = AddQueryParameters(request.BaseURL, request.QueryParams)
	}
	req, err := http.NewRequest(string(request.Method), request.BaseURL, bytes.NewBuffer(request.Body))
	if err != nil {
		return req, err
	}
	for key, value := range request.Headers {
		req.Header.Set(key, value)
	}
	_, exists := req.Header["Content-Type"]
	if len(request.Body) > 0 && !exists {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, err
}

// MakeRequest makes the API call.
func MakeRequest(req *http.Request) (*http.Response, error) {
	return DefaultClient.HTTPClient.Do(req)
}

// BuildResponse builds the response struct.
func BuildResponse(res *http.Response) (*Response, error) {
	body, err := ioutil.ReadAll(res.Body)
	response := Response{
		StatusCode: res.StatusCode,
		Body:       string(body),
		Headers:    res.Header,
	}
	res.Body.Close() // nolint
	return &response, err
}

// MakeRequest makes the API call.
func (c *Client) MakeRequest(req *http.Request) (*http.Response, error) {
	return c.HTTPClient.Do(req)
}

func (c *Client) API(request Request) (*Response, error) {
	return c.Send(request)
}

func (c *Client) Send(request Request) (*Response, error) {
	return c.SendWithContext(context.Background(), request)
}

func (c *Client) SendWithContext(ctx context.Context, request Request) (*Response, error) {
	// Build the HTTP request object.
	req, err := BuildRequestObject(request)
	if err != nil {
		return nil, err
	}
	// Pass in the user provided context
	req = req.WithContext(ctx)

	// Build the HTTP client and make the request.
	res, err := c.MakeRequest(req)
	if err != nil {
		return nil, err
	}

	// Build Response object.
	return BuildResponse(res)
}
