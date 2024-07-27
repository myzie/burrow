package burrow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var _ http.RoundTripper = &Transport{}

// Transport implements the http.RoundTripper interface. Used to proxy HTTP
// requests via a Burrow HTTP endpoint.
type Transport struct {
	proxyURL string
	method   string
	client   *http.Client
}

// RoundTrip implements the http.RoundTripper interface
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	serReq, err := SerializeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}
	payload, err := json.Marshal(serReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	proxyReq, err := http.NewRequestWithContext(req.Context(), t.method, t.proxyURL,
		bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy request: %w", err)
	}
	proxyReq.Header.Set("Content-Type", "application/json")
	proxyResp, err := t.client.Do(proxyReq)
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
