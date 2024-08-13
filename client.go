package burrow

import (
	"net/http"
	"time"
)

// NewRoundRobinClient is a convenience function for creating an http.Client
// with a RoundRobinTransport that rotates through the provided proxy URLs.
func NewRoundRobinClient(proxyURLs []string) *http.Client {
	if len(proxyURLs) == 0 {
		return &http.Client{}
	}
	var transports []http.RoundTripper
	for _, proxyURL := range proxyURLs {
		transports = append(transports, NewTransport(proxyURL, "POST"))
	}
	return &http.Client{
		Transport: NewRoundRobinTransport(transports),
	}
}

// ClientOption defines a function that configures a client
type ClientOption func(*clientConfig)

type clientConfig struct {
	proxyURLs           []string
	retries             int
	retryableCodes      []int
	callback            ProxyCallback
	timeout             time.Duration
	maxResponseBytes    int64
	allowedContentTypes []string
}

// WithProxyURL sets a single proxy URL for the client
func WithProxyURL(url string) ClientOption {
	return func(c *clientConfig) {
		c.proxyURLs = []string{url}
	}
}

// WithProxyURLs sets the proxy URLs for the client
func WithProxyURLs(urls []string) ClientOption {
	return func(c *clientConfig) {
		c.proxyURLs = urls
	}
}

// WithRetries sets the number of retries for the client
func WithRetries(retries int) ClientOption {
	return func(c *clientConfig) {
		c.retries = retries
	}
}

// WithRetryableCodes sets the HTTP status codes that should trigger a retry
func WithRetryableCodes(codes []int) ClientOption {
	return func(c *clientConfig) {
		c.retryableCodes = codes
	}
}

// WithCallback sets a callback function that will be called each time a proxy
// request has completed successfully.
func WithCallback(callback ProxyCallback) ClientOption {
	return func(c *clientConfig) {
		c.callback = callback
	}
}

// WithTimeout sets the timeout that is passed to the proxy
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *clientConfig) {
		c.timeout = timeout
	}
}

// WithMaxResponseBytes sets the maximum response body size
func WithMaxResponseBytes(maxResponseBytes int64) ClientOption {
	return func(c *clientConfig) {
		c.maxResponseBytes = maxResponseBytes
	}
}

// WithAllowedContentTypes sets the allowed content types
func WithAllowedContentTypes(allowedContentTypes []string) ClientOption {
	return func(c *clientConfig) {
		c.allowedContentTypes = allowedContentTypes
	}
}

// NewClient creates an http.Client with the provided Burrow options.
// If no proxy URLs are provided, a vanilla http.Client is returned.
func NewClient(opts ...ClientOption) *http.Client {
	cfg := &clientConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	if len(cfg.proxyURLs) == 0 {
		return &http.Client{}
	}
	var transports []http.RoundTripper
	for _, proxyURL := range cfg.proxyURLs {
		transport := NewTransport(proxyURL, "POST")
		if cfg.callback != nil {
			transport.WithCallback(cfg.callback)
		}
		if cfg.timeout > 0 {
			transport.WithTimeout(cfg.timeout)
		}
		if cfg.maxResponseBytes > 0 {
			transport.WithMaxResponseBytes(cfg.maxResponseBytes)
		}
		if len(cfg.allowedContentTypes) > 0 {
			transport.WithAllowedContentTypes(cfg.allowedContentTypes)
		}
		transports = append(transports, transport)
	}
	rr := NewRoundRobinTransport(transports)
	if cfg.retries > 0 {
		rr.WithRetries(cfg.retries)
	}
	if cfg.retryableCodes != nil {
		rr.WithRetryableCodes(cfg.retryableCodes)
	}
	return &http.Client{Transport: rr}
}
