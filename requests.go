package burrow

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// Request represents an http request in a format that can be easily serialized
type Request struct {
	URL                 string            `json:"url"`
	Method              string            `json:"method,omitempty"`
	Headers             map[string]string `json:"headers,omitempty"`
	Body                string            `json:"body,omitempty"`
	Cookies             string            `json:"cookies,omitempty"`
	Timeout             float64           `json:"timeout,omitempty"`
	AllowedContentTypes []string          `json:"allowed_content_types,omitempty"`
	Head                bool              `json:"head,omitempty"`
	CacheMaxAge         float64           `json:"cache_max_age,omitempty"`
	parsedURL           *url.URL          `json:"-"`
}

// Response represents an http response in a format that can be easily deserialized
type Response struct {
	StatusCode    int               `json:"status_code"`
	Headers       map[string]string `json:"headers,omitempty"`
	Body          string            `json:"body,omitempty"`
	ClientDetails *ClientDetails    `json:"client_details,omitempty"`
	Duration      float64           `json:"duration,omitempty"`
	ProxyName     string            `json:"proxy_name,omitempty"`
	SignedURL     string            `json:"signed_url,omitempty"`
}

// ClientDetails represents the details of the client that made the request
type ClientDetails struct {
	SourceIP  string `json:"source_ip"`
	UserAgent string `json:"user_agent"`
}

func SerializeRequest(req *http.Request) (*Request, error) {
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

func DeserializeResponse(ctx context.Context, serResp *Response) (*http.Response, error) {
	resp := &http.Response{
		StatusCode: serResp.StatusCode,
		Header:     make(http.Header),
	}
	for k, v := range serResp.Headers {
		resp.Header.Set(k, v)
	}

	// Set resp.ContentLength according to the header
	if contentLength := serResp.Headers["Content-Length"]; contentLength != "" {
		if length, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
			resp.ContentLength = length
		}
	}

	if serResp.Body != "" {
		decodedBody, err := base64.StdEncoding.DecodeString(serResp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to decode response body: %w", err)
		}
		resp.Body = io.NopCloser(bytes.NewBuffer(decodedBody))
		resp.ContentLength = int64(len(decodedBody))
	} else if serResp.SignedURL != "" {
		body, _, err := getSignedUrl(ctx, serResp.SignedURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get signed url: %w", err)
		}
		resp.Body = body
	} else {
		resp.Body = io.NopCloser(bytes.NewBuffer([]byte{}))
		resp.ContentLength = 0
	}

	return resp, nil
}

func getSignedUrl(ctx context.Context, signedUrl string) (io.ReadCloser, http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, signedUrl, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return resp.Body, resp.Header, nil
}
