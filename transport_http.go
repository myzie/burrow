package burrow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var _ http.RoundTripper = &Transport{}

type ErrorCode int

const (
	ProxyErrUnknown               ErrorCode = 0
	ProxyErrBadRequest            ErrorCode = 1
	ProxyErrExceededMaxBodySize   ErrorCode = 2
	ProxyErrDisallowedContentType ErrorCode = 3
	ProxyErrTimeout               ErrorCode = 4
)

type ProxyError struct {
	Message string    `json:"message"`
	Type    ErrorCode `json:"type"`
}

func (e *ProxyError) Error() string {
	return fmt.Sprintf("proxy error [%d] %s", e.Type, e.Message)
}

func ProxyErrorf(code ErrorCode, format string, args ...any) *ProxyError {
	return &ProxyError{
		Type:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// ProxyCallback is a callback function that will be called each time a proxy
// request has completed successfully.
type ProxyCallback func(ctx context.Context, req *Request, res *Response)

// Transport implements the http.RoundTripper interface. Used to proxy HTTP
// requests via a Burrow HTTP endpoint.
type Transport struct {
	proxyURL            string
	method              string
	client              *http.Client
	callback            ProxyCallback
	timeout             time.Duration
	maxResponseBytes    int64
	allowedContentTypes []string
}

// RoundTrip implements the http.RoundTripper interface
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	serReq, err := SerializeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}
	serReq.Timeout = t.timeout.Seconds()
	serReq.MaxResponseBytes = t.maxResponseBytes
	serReq.AllowedContentTypes = t.allowedContentTypes
	payload, err := json.Marshal(serReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	proxyReq, err := http.NewRequestWithContext(req.Context(), t.method, t.proxyURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy request: %w", err)
	}
	proxyReq.Header.Set("Content-Type", "application/json")
	proxyResp, err := t.client.Do(proxyReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to proxy: %w", err)
	}
	defer proxyResp.Body.Close()
	body, err := io.ReadAll(proxyResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read proxy response body: %w", err)
	}
	if proxyResp.StatusCode != http.StatusOK {
		var errResp ProxyError
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, &ProxyError{
				Message: fmt.Sprintf("proxy returned non-200 status code: %d", proxyResp.StatusCode),
				Type:    ProxyErrUnknown,
			}
		}
		return nil, &errResp
	}
	var serResp Response
	if err := json.Unmarshal(body, &serResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if t.callback != nil {
		t.callback(req.Context(), serReq, &serResp)
	}
	return DeserializeResponse(&serResp)
}

// NewTransport creates a new Transport
func NewTransport(proxyURL string, method string, c ...*http.Client) *Transport {
	return &Transport{
		proxyURL: proxyURL,
		method:   "POST",
		client:   &http.Client{},
	}
}

// NewTransportWithClient creates a new Transport that uses the provided
// HTTP client internally. If you're not sure, use NewTransport instead.
func NewTransportWithClient(proxyURL string, method string, c *http.Client) *Transport {
	return &Transport{
		proxyURL: proxyURL,
		method:   method,
		client:   c,
	}
}

// WithCallback sets a callback function that will be called each time a proxy
// request has completed successfully.
func (t *Transport) WithCallback(callback ProxyCallback) *Transport {
	t.callback = callback
	return t
}

// WithTimeout sets the timeout that is passed to the proxy
func (t *Transport) WithTimeout(timeout time.Duration) *Transport {
	t.timeout = timeout
	return t
}

// WithMaxResponseBytes sets the maximum response body size
func (t *Transport) WithMaxResponseBytes(maxResponseBytes int64) *Transport {
	t.maxResponseBytes = maxResponseBytes
	return t
}

// WithAllowedContentTypes sets the allowed content types
func (t *Transport) WithAllowedContentTypes(allowedContentTypes []string) *Transport {
	t.allowedContentTypes = allowedContentTypes
	return t
}
