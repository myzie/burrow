package burrow

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Request represents an http request in a format that can be easily serialized
type Request struct {
	URL     string            `json:"url"`
	Method  string            `json:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
	Cookies []string          `json:"cookies,omitempty"`
}

// Response represents an http response in a format that can be easily deserialized
type Response struct {
	StatusCode    int               `json:"status_code"`
	Headers       map[string]string `json:"headers,omitempty"`
	Body          string            `json:"body,omitempty"`
	Cookies       []string          `json:"cookies,omitempty"`
	ClientDetails *ClientDetails    `json:"client_details,omitempty"`
}

// ClientDetails represents the details of the client that made the request
type ClientDetails struct {
	SourceIP  string `json:"source_ip"`
	UserAgent string `json:"user_agent"`
}

func serializeRequest(req *http.Request) (*Request, error) {
	headers := make(map[string]string)
	for k, v := range req.Header {
		headers[k] = v[0]
	}
	var encodedBody string
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body = io.NopCloser(bytes.NewBuffer(body))
		encodedBody = base64.StdEncoding.EncodeToString(body)
	}
	return &Request{
		Method:  req.Method,
		URL:     req.URL.String(),
		Headers: headers,
		Body:    encodedBody,
	}, nil
}

func deserializeResponse(serResp *Response) (*http.Response, error) {
	var decodedBody []byte
	if serResp.Body != "" {
		var err error
		decodedBody, err = base64.StdEncoding.DecodeString(serResp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to decode response body: %w", err)
		}
	}
	resp := &http.Response{
		StatusCode: serResp.StatusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBuffer(decodedBody)),
	}
	for k, v := range serResp.Headers {
		resp.Header.Set(k, v)
	}
	return resp, nil
}

type Handler func(ctx context.Context, req *Request) (*Response, error)

func GetHandler(c ...*http.Client) Handler {
	var client *http.Client
	if len(c) > 0 {
		client = c[0]
	} else {
		client = &http.Client{}
	}
	return func(ctx context.Context, req *Request) (*Response, error) {
		if req.URL == "" {
			return nil, &BadInputError{Err: fmt.Errorf("url is required")}
		}
		method := "GET"
		if req.Method != "" {
			method = req.Method
		}
		var decodedBody []byte
		if req.Body != "" {
			var err error
			decodedBody, err = base64.StdEncoding.DecodeString(req.Body)
			if err != nil {
				return nil, &BadInputError{
					Err: fmt.Errorf("failed to decode request body: %w", err),
				}
			}
		}
		httpReq, err := http.NewRequestWithContext(ctx, method, req.URL,
			strings.NewReader(string(decodedBody)))
		if err != nil {
			return nil, &BadInputError{
				Err: fmt.Errorf("failed to create http request: %w", err),
			}
		}
		for k, v := range req.Headers {
			httpReq.Header.Set(k, v)
		}
		resp, err := client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("failed to perform http request: %w", err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		var encodedBody string
		if len(body) > 0 {
			encodedBody = base64.StdEncoding.EncodeToString(body)
		}
		headers := map[string]string{}
		for k, v := range resp.Header {
			headers[k] = v[0]
		}
		return &Response{
			StatusCode: resp.StatusCode,
			Headers:    headers,
			Body:       encodedBody,
		}, nil
	}
}

type BadInputError struct {
	Err error
}

func (b *BadInputError) Error() string {
	return fmt.Sprintf("bad input: %v", b.Err)
}

func (b *BadInputError) Unwrap() error {
	return b.Err
}
