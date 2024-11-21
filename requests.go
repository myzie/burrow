package burrow

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

// Request represents an http request in a format that can be easily serialized
type Request struct {
	URL                 string            `json:"url"`
	Method              string            `json:"method,omitempty"`
	Headers             map[string]string `json:"headers,omitempty"`
	Body                string            `json:"body,omitempty"`
	Cookies             string            `json:"cookies,omitempty"`
	Timeout             float64           `json:"timeout,omitempty"`
	MaxResponseBytes    int64             `json:"max_response_bytes,omitempty"`
	AllowedContentTypes []string          `json:"allowed_content_types,omitempty"`
}

// Response represents an http response in a format that can be easily deserialized
type Response struct {
	StatusCode    int               `json:"status_code"`
	Headers       map[string]string `json:"headers,omitempty"`
	Body          string            `json:"body,omitempty"`
	ClientDetails *ClientDetails    `json:"client_details,omitempty"`
	Duration      float64           `json:"duration,omitempty"`
	ProxyName     string            `json:"proxy_name,omitempty"`
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

func DeserializeResponse(serResp *Response) (*http.Response, error) {
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
