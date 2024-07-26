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
