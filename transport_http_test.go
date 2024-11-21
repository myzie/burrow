package burrow

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransport_RoundTrip(t *testing.T) {
	successEncoded := base64.StdEncoding.EncodeToString([]byte("success"))
	tests := []struct {
		name           string
		setupMockProxy func() *httptest.Server
		setupTransport func(*Transport)
		inputURL       string
		expectedErr    string
	}{
		{
			name: "successful request",
			setupMockProxy: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
					json.NewEncoder(w).Encode(Response{
						StatusCode: 200,
						Headers:    map[string]string{"Content-Type": "text/plain"},
						Body:       successEncoded,
					})
				}))
			},
			inputURL: "https://example.com",
		},
		{
			name: "proxy returns error",
			setupMockProxy: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(ProxyError{
						Message: "bad request",
						Type:    ProxyErrBadRequest,
					})
				}))
			},
			inputURL:    "https://example.com",
			expectedErr: "proxy error [1] bad request",
		},
		{
			name: "respects timeout",
			setupMockProxy: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Decode the incoming request to verify timeout was set
					var serReq Request
					json.NewDecoder(r.Body).Decode(&serReq)
					assert.Equal(t, float64(2), serReq.Timeout)
					resp := Response{
						StatusCode: 200,
						Body:       successEncoded,
					}
					json.NewEncoder(w).Encode(resp)
				}))
			},
			setupTransport: func(t *Transport) {
				t.WithTimeout(2 * time.Second)
			},
			inputURL: "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProxy := tt.setupMockProxy()
			defer mockProxy.Close()

			transport := NewTransport(mockProxy.URL, "POST")
			if tt.setupTransport != nil {
				tt.setupTransport(transport)
			}
			req, err := http.NewRequest("GET", tt.inputURL, nil)
			require.NoError(t, err)
			resp, err := transport.RoundTrip(req)
			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestTransportBuilders(t *testing.T) {
	transport := NewTransport("http://proxy", "POST")

	callback := func(ctx context.Context, req *Request, r *Response) {}
	transport.WithCallback(callback)
	assert.NotNil(t, transport.callback)

	transport.WithTimeout(5 * time.Second)
	assert.Equal(t, 5*time.Second, transport.timeout)

	transport.WithMaxResponseBytes(1000)
	assert.Equal(t, int64(1000), transport.maxResponseBytes)

	allowedTypes := []string{"application/json"}
	transport.WithAllowedContentTypes(allowedTypes)
	assert.Equal(t, allowedTypes, transport.allowedContentTypes)
}

func TestNewTransportWithClient(t *testing.T) {
	customClient := &http.Client{Timeout: 5 * time.Second}
	transport := NewTransportWithClient("http://proxy", "POST", customClient)

	assert.Equal(t, customClient, transport.client)
	assert.Equal(t, "http://proxy", transport.proxyURL)
	assert.Equal(t, "POST", transport.method)
}
