package burrow

import (
	"bytes"
	"io"
	"math"
	"net/http"
	"sync"
	"time"
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
	return &RoundRobinTransport{
		transports: transports,
		retryable:  defaultRetryableCodes,
	}
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
	// Clone the request body if it exists
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body.Close()
	}
	var lastResp *http.Response
	for i := 0; i <= r.retries; i++ {
		// Recreate the body for each attempt
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
		transport := r.nextTransport()
		response, err := transport.RoundTrip(req)
		if err != nil {
			// This means the proxying itself failed, which we will not retry
			return nil, err
		}
		// Return immediately if the status code is in the 2xx range
		if response.StatusCode >= 200 && response.StatusCode < 300 {
			return response, nil
		}
		// Return immediately if the status code is not retryable
		if !r.isRetryable(response.StatusCode) {
			return response, nil
		}
		// Close the response body if we're not returning it
		if lastResp != nil {
			lastResp.Body.Close()
		}
		lastResp = response

		// Skip sleep on the last iteration
		if i < r.retries {
			// Calculate backoff duration starting from 100ms
			backoff := time.Duration(math.Pow(2, float64(i))*100) * time.Millisecond
			// Use context-aware sleep
			timer := time.NewTimer(backoff)
			select {
			case <-req.Context().Done():
				timer.Stop()
				return lastResp, req.Context().Err()
			case <-timer.C:
				// Continue with next retry
			}
		}
	}
	return lastResp, nil
}

func (r *RoundRobinTransport) nextTransport() http.RoundTripper {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	transport := r.transports[r.index]
	r.index = (r.index + 1) % len(r.transports)
	return transport
}

func (r *RoundRobinTransport) isRetryable(code int) bool {
	return r.retryable[code]
}

var defaultRetryableCodes = map[int]bool{
	http.StatusTooManyRequests:    true,
	http.StatusRequestTimeout:     true,
	http.StatusServiceUnavailable: true,
	http.StatusGatewayTimeout:     true,
	http.StatusBadGateway:         true,
	999:                           true,
}
