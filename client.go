package burrow

import (
	"net/http"
)

// NewClient is a convenience function for creating an http.Client with a Burrow
// Transport. If proxyURL is empty, the client will not use a proxy.
func NewClient(proxyURL string) *http.Client {
	if proxyURL == "" {
		return &http.Client{}
	}
	return &http.Client{
		Transport: NewTransport(proxyURL, "POST"),
	}
}

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

// ClientOpts is used to configure an http.Client with proxying and retry
// behaviors.
type ClientOpts struct {
	ProxyURLs      []string
	Retries        int
	RetryableCodes []int
}

// NewClientWithOptions is a convenience function for creating an
// http.Client with a Burrow transport that rotates through the provided
// proxy URLs. It allows for configuration of retries via ClientOpts.
func NewClientWithOptions(opts ClientOpts) *http.Client {
	if len(opts.ProxyURLs) == 0 {
		return &http.Client{}
	}
	var transports []http.RoundTripper
	for _, proxyURL := range opts.ProxyURLs {
		transports = append(transports, NewTransport(proxyURL, "POST"))
	}
	rr := NewRoundRobinTransport(transports)
	if opts.Retries > 0 {
		rr.WithRetries(opts.Retries)
	}
	if opts.RetryableCodes != nil {
		rr.WithRetryableCodes(opts.RetryableCodes)
	}
	return &http.Client{
		Transport: rr,
	}
}
