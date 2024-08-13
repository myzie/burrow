package burrow

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

var defaultMaxRedirects = 5

var defaultMaxResponseBytes = int64(5 * 1024 * 1024) // 5MB default

var defaultTransport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	TLSHandshakeTimeout:   5 * time.Second,
	ResponseHeaderTimeout: 10 * time.Second,
	MaxIdleConns:          100,
	MaxIdleConnsPerHost:   10,
	IdleConnTimeout:       90 * time.Second,
}

// Handler is a function used to process Burrow HTTP proxy requests.
type Handler func(ctx context.Context, req *Request) (*Response, error)

// GetHandler returns a Handler that proxies HTTP requests.
func GetHandler(c ...*http.Client) Handler {
	var client *http.Client
	if len(c) > 0 {
		client = c[0]
	} else {
		client = &http.Client{
			Transport: defaultTransport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= defaultMaxRedirects {
					return fmt.Errorf("stopped after %d redirects", len(via))
				}
				return nil
			},
		}
	}
	return func(ctx context.Context, req *Request) (*Response, error) {
		if req.URL == "" {
			return nil, ProxyErrorf(ProxyErrBadRequest, "url is required")
		}
		start := time.Now()
		if req.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, time.Duration(req.Timeout*float64(time.Second)))
			defer cancel()
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
				return nil, ProxyErrorf(ProxyErrBadRequest, "failed to decode request body: %v", err)
			}
		}
		httpReqBody := strings.NewReader(string(decodedBody))
		httpReq, err := http.NewRequestWithContext(ctx, method, req.URL, httpReqBody)
		if err != nil {
			return nil, ProxyErrorf(ProxyErrBadRequest, "failed to create http request: %v", err)
		}
		for k, v := range req.Headers {
			httpReq.Header.Set(k, v)
		}
		if req.Cookies != "" {
			httpReq.Header.Add("Cookie", req.Cookies)
		}
		resp, err := client.Do(httpReq)
		if err != nil {
			if isTimeoutError(err) {
				return nil, ProxyErrorf(ProxyErrTimeout, "http request timed out")
			}
			return nil, ProxyErrorf(ProxyErrUnknown, "failed to execute http request: %v", err)
		}
		defer resp.Body.Close()

		if len(req.AllowedContentTypes) > 0 {
			contentType := resp.Header.Get("Content-Type")
			if !isContentTypeAllowed(contentType, req.AllowedContentTypes) {
				return nil, ProxyErrorf(ProxyErrDisallowedContentType, "response content type is disallowed: %s", contentType)
			}
		}
		maxSize := defaultMaxResponseBytes
		if req.MaxResponseBytes > 0 {
			maxSize = req.MaxResponseBytes
		}
		// Add 1 so that we can detect if the body was truncated
		limitReader := io.LimitReader(resp.Body, maxSize+1)
		body, err := io.ReadAll(limitReader)
		if err != nil {
			if isTimeoutError(err) {
				return nil, ProxyErrorf(ProxyErrTimeout, "response body read timed out")
			}
			return nil, ProxyErrorf(ProxyErrUnknown, "failed to read response body: %v", err)
		}
		if int64(len(body)) > maxSize {
			return nil, ProxyErrorf(ProxyErrExceededMaxBodySize, "response body exceeded maximum size: %d", maxSize)
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
			Duration:   time.Since(start).Seconds(),
		}, nil
	}
}

func isContentTypeAllowed(contentType string, allowedContentTypes []string) bool {
	for _, allowedContentType := range allowedContentTypes {
		if strings.HasPrefix(contentType, allowedContentType) {
			return true
		}
	}
	return false
}

func isTimeoutError(err error) bool {
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout()
	}
	if err == context.DeadlineExceeded {
		return true
	}
	return false
}
