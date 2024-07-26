package burrow

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Handler is a function used to process Burrow HTTP proxy requests.
type Handler func(ctx context.Context, req *Request) (*Response, error)

// GetHandler returns a Handler that proxies HTTP requests.
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
