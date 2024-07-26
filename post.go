package burrow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// PostTransport implements the http.RoundTripper interface
type PostTransport struct {
	ProxyURL   string
	HTTPClient *http.Client
}

// RoundTrip implements the http.RoundTripper interface
func (t *PostTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	serReq, err := serializeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}
	payload, err := json.Marshal(serReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	proxyReq, err := http.NewRequestWithContext(req.Context(), "POST", t.ProxyURL,
		bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy request: %w", err)
	}
	proxyReq.Header.Set("Content-Type", "application/json")
	proxyResp, err := t.HTTPClient.Do(proxyReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to proxy: %w", err)
	}
	defer proxyResp.Body.Close()
	if proxyResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("proxy returned non-200 status code: %d", proxyResp.StatusCode)
	}
	body, err := io.ReadAll(proxyResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read proxy response body: %w", err)
	}
	var serResp Response
	if err := json.Unmarshal(body, &serResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return deserializeResponse(&serResp)
}

// NewPostTransport creates a new PostTransport
func NewPostTransport(proxyURL string, c ...*http.Client) *PostTransport {
	var client *http.Client
	if len(c) > 0 {
		client = c[0]
	} else {
		client = &http.Client{}
	}
	return &PostTransport{
		ProxyURL:   proxyURL,
		HTTPClient: client,
	}
}
