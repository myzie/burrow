package burrow

import (
	"fmt"
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
	retries    int
	retryable  map[int]bool
}

// NewRoundRobinTransport creates a new RoundRobinTransport that rotates through
// the provided http.Transports with each request.
func NewRoundRobinTransport(transports []http.RoundTripper) *RoundRobinTransport {
	return &RoundRobinTransport{transports: transports, retryable: defaultRetryableCodes}
}

// WithRetries sets the allowed number of retry attempts for each request.
func (r *RoundRobinTransport) WithRetries(retries int) *RoundRobinTransport {
	if retries < 0 {
		retries = 0
	}
	r.retries = retries
	return r
}

// WithRetryableCodes sets the list of HTTP status codes that should be retried.
func (r *RoundRobinTransport) WithRetryableCodes(codes []int) *RoundRobinTransport {
	retryable := map[int]bool{}
	for _, code := range codes {
		retryable[code] = true
	}
	r.retryable = retryable
	return r
}

// RoundTrip implements the http.RoundTripper interface.
func (r *RoundRobinTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var lastErr error
	for i := 0; i <= r.retries; i++ {
		transport := r.nextTransport()
		response, err := transport.RoundTrip(req)
		if err == nil {
			return response, nil
		}
		lastErr = err
		if !r.isRetryable(err) {
			return nil, err
		}
		fmt.Println("retrying request", req.URL, "attempt", i+1, "err", err)
	}
	return nil, lastErr
}

func (r *RoundRobinTransport) nextTransport() http.RoundTripper {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	transport := r.transports[r.index]
	r.index = (r.index + 1) % len(r.transports)
	return transport
}

func (r *RoundRobinTransport) isRetryable(err error) bool {
	if err == nil {
		return false
	}
	switch err := err.(type) {
	case *TransportError:
		return r.retryable[err.StatusCode]
	default:
		return false
	}
}

var defaultRetryableCodes = map[int]bool{
	http.StatusUnauthorized:       true,
	http.StatusForbidden:          true,
	http.StatusTooManyRequests:    true,
	http.StatusRequestTimeout:     true,
	http.StatusServiceUnavailable: true,
	http.StatusGatewayTimeout:     true,
	http.StatusBadGateway:         true,
	999:                           true,
}
