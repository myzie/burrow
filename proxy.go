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

// Request represents an HTTP request in a format that can be easily serialized
type Request struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// Response represents an HTTP response in a format that can be easily deserialized
type Response struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

func serializeRequest(req *http.Request) (*Request, error) {
	headers := make(map[string]string)
	for k, v := range req.Header {
		headers[k] = v[0]
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	encodedBody := base64.StdEncoding.EncodeToString(body)
	return &Request{
		Method:  req.Method,
		URL:     req.URL.String(),
		Headers: headers,
		Body:    encodedBody,
	}, nil
}

func deserializeResponse(serResp *Response) (*http.Response, error) {
	decodedBody, err := base64.StdEncoding.DecodeString(serResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
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
		decodedBody, err := base64.StdEncoding.DecodeString(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to decode request body: %w", err)
		}
		httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL,
			strings.NewReader(string(decodedBody)))
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP request: %w", err)
		}
		for k, v := range req.Headers {
			httpReq.Header.Set(k, v)
		}
		resp, err := client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		encodedBody := base64.StdEncoding.EncodeToString(body)
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
