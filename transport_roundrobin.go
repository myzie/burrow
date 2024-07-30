package burrow

import (
	"net/http"
	"sync"
)

var _ http.RoundTripper = &RoundRobinTransport{}

// RoundRobinTransport is an http.RoundTripper that sends requests using a
// rotating set of http.Transports.
type RoundRobinTransport struct {
	transports []http.RoundTripper
	mutex      sync.Mutex
	index      int
}

// NewRoundRobinTransport creates a new RoundRobinTransport that rotates through
// the provided http.Transports with each request.
func NewRoundRobinTransport(transports []http.RoundTripper) *RoundRobinTransport {
	return &RoundRobinTransport{transports: transports}
}

// RoundTrip implements the http.RoundTripper interface
func (r *RoundRobinTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r.mutex.Lock()
	index := r.index
	r.index = (r.index + 1) % len(r.transports)
	transport := r.transports[index]
	r.mutex.Unlock()

	return transport.RoundTrip(req)
}
