package burrow

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var defaultMaxRedirects = 5

var defaultTransport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	TLSHandshakeTimeout:   5 * time.Second,
	ResponseHeaderTimeout: 10 * time.Second,
	MaxIdleConns:          100,
	MaxIdleConnsPerHost:   10,
	IdleConnTimeout:       90 * time.Second,
}

var DefaultClient = &http.Client{
	Transport: defaultTransport,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= defaultMaxRedirects {
			return fmt.Errorf("stopped after %d redirects", len(via))
		}
		return nil
	},
}

// Handler is a function used to process Burrow HTTP proxy requests.
type Handler func(ctx context.Context, req *Request) (*Response, error)

const (
	inlineResponseLimit = 1024 * 1024 // 1MB
	cacheKeyPrefix      = "cache"
	cacheTimestampKey   = "timestamp"
)

func checkCache(ctx context.Context, storage Storage, cacheKey string, cacheMaxAge float64) (*ObjectInfo, bool) {
	if cacheMaxAge <= 0 {
		return nil, false
	}
	info, err := storage.HeadObject(ctx, cacheKey)
	if err != nil || !info.Exists {
		return nil, false
	}
	cacheTime, ok := info.Metadata[cacheTimestampKey]
	if !ok {
		return nil, false
	}
	cacheTimeVal, err := time.Parse(time.RFC3339, cacheTime)
	if err != nil {
		return nil, false
	}
	age := time.Since(cacheTimeVal).Seconds()
	if age >= cacheMaxAge {
		return nil, false
	}
	return info, true
}

// GetHandler returns a Handler that proxies HTTP requests.
func GetHandler(client *http.Client, storage Storage) Handler {
	if client == nil {
		client = DefaultClient
	}
	return func(ctx context.Context, req *Request) (*Response, error) {
		if req.URL == "" {
			return nil, ProxyErrorf(ProxyErrBadRequest, "url is required")
		}
		var err error
		req.parsedURL, err = url.Parse(req.URL)
		if err != nil {
			return nil, ProxyErrorf(ProxyErrBadRequest, "invalid url: %v", err)
		}
		if req.parsedURL.Scheme != "http" && req.parsedURL.Scheme != "https" {
			return nil, ProxyErrorf(ProxyErrBadRequest, "unsupported url scheme")
		}

		// Generate cache key from URL
		cacheKey := GetCacheKey(req)

		// Check cache
		if info, hit := checkCache(ctx, storage, cacheKey, req.CacheMaxAge); hit {
			if req.Head {
				signedURL, err := storage.SignURL(ctx, cacheKey, time.Minute*15)
				if err != nil {
					return nil, ProxyErrorf(ProxyErrStorage, "failed to generate signed URL: %v", err)
				}
				headers := map[string]string{
					"Content-Type":   info.ContentType,
					"Content-Length": strconv.FormatInt(info.ContentLength, 10),
					"Cache-Time":     info.Metadata[cacheTimestampKey],
					"Cache-Key":      cacheKey,
				}
				return &Response{
					StatusCode: 200,
					Headers:    headers,
					SignedURL:  signedURL,
				}, nil
			}
			return getCachedResponse(ctx, storage, cacheKey, info.ContentLength)
		}

		// Proceed with actual HTTP request
		if req.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, time.Duration(req.Timeout*float64(time.Second)))
			defer cancel()
		}
		method := "GET"
		if req.Method != "" {
			method = req.Method
		}
		var decodedBody []byte
		if req.Body != "" {
			var err error
			decodedBody, err = base64.StdEncoding.DecodeString(req.Body)
			if err != nil {
				return nil, ProxyErrorf(ProxyErrBadRequest, "failed to decode request body: %v", err)
			}
		}
		httpReqBody := strings.NewReader(string(decodedBody))
		httpReq, err := http.NewRequestWithContext(ctx, method, req.URL, httpReqBody)
		if err != nil {
			return nil, ProxyErrorf(ProxyErrBadRequest, "failed to create http request: %v", err)
		}
		for k, v := range req.Headers {
			httpReq.Header.Set(k, v)
		}
		if req.Cookies != "" {
			httpReq.Header.Add("Cookie", req.Cookies)
		}
		resp, err := client.Do(httpReq)
		if err != nil {
			if isTimeoutError(err) {
				return nil, ProxyErrorf(ProxyErrTimeout, "http request timed out")
			}
			return nil, ProxyErrorf(ProxyErrUnknown, "failed to execute http request: %v", err)
		}
		defer resp.Body.Close()

		// Disallowed content types should cause a rejection
		if len(req.AllowedContentTypes) > 0 {
			contentType := resp.Header.Get("Content-Type")
			if !isContentTypeAllowed(contentType, req.AllowedContentTypes) {
				return nil, ProxyErrorf(ProxyErrDisallowedContentType, "response content type is disallowed: %s", contentType)
			}
		}

		contentLength := resp.ContentLength
		contentType := resp.Header.Get("Content-Type")
		shouldStream := contentLength == -1 || contentLength > inlineResponseLimit
		cacheTime := time.Now().UTC().Format(time.RFC3339)

		metadata := map[string]string{
			cacheTimestampKey: cacheTime,
			"url":             req.URL,
			"method":          req.Method,
			"region":          os.Getenv("AWS_REGION"),
		}

		var body []byte
		if shouldStream {
			// Stream directly to storage without reading into memory
			err = storage.PutObject(ctx, cacheKey, resp.Body, contentType, contentLength, metadata)
			if err != nil {
				return nil, ProxyErrorf(ProxyErrStorage, "failed to stream response to storage: %v", err)
			}
			// Get the actual content length from storage
			info, err := storage.HeadObject(ctx, cacheKey)
			if err != nil {
				return nil, ProxyErrorf(ProxyErrStorage, "failed to get stored object info: %v", err)
			}
			// Update contentLength with the actual size
			contentLength = info.ContentLength
		} else {
			// For small responses, read into memory
			limitReader := io.LimitReader(resp.Body, inlineResponseLimit+1)
			body, err = io.ReadAll(limitReader)
			if err != nil {
				if isTimeoutError(err) {
					return nil, ProxyErrorf(ProxyErrTimeout, "response body read timed out")
				}
				return nil, ProxyErrorf(ProxyErrUnknown, "failed to read response body: %v", err)
			}
			if int64(len(body)) > inlineResponseLimit {
				return nil, ProxyErrorf(ProxyErrExceededMaxBodySize, "response body exceeded maximum size: %d", inlineResponseLimit)
			}
			// Store small response in cache if caching is enabled
			if req.CacheMaxAge > 0 {
				err = storage.PutObject(ctx, cacheKey, bytes.NewReader(body), contentType, int64(len(body)), metadata)
				if err != nil {
					return nil, ProxyErrorf(ProxyErrStorage, "failed to store response in cache: %v", err)
				}
			}
		}

		// Prepare response headers. Only provide the headers that will also
		// be available when retrieving the response from the cache.
		headers := map[string]string{
			"Content-Type":   contentType,
			"Content-Length": strconv.FormatInt(contentLength, 10),
			"Cache-Time":     cacheTime,
			"Cache-Key":      cacheKey,
		}
		response := &Response{
			StatusCode: resp.StatusCode,
			Headers:    headers,
		}

		// For streamed or large responses, return a signed URL
		if shouldStream {
			signedURL, err := storage.SignURL(ctx, cacheKey, time.Minute*15)
			if err != nil {
				return nil, ProxyErrorf(ProxyErrStorage, "failed to generate signed URL: %v", err)
			}
			response.SignedURL = signedURL
		} else {
			response.Body = base64.StdEncoding.EncodeToString(body)
		}
		return response, nil
	}
}

func isContentTypeAllowed(contentType string, allowedContentTypes []string) bool {
	// Empty allowedContentTypes is a wildcard
	if len(allowedContentTypes) == 0 {
		return true
	}
	// Default content type to HTML if not set
	if contentType == "" {
		contentType = "text/html"
	}
	// Split multiple content types if present
	contentTypes := strings.Split(contentType, ",")

	for _, ct := range contentTypes {
		mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(ct))
		if err != nil {
			continue
		}
		mediaType = strings.ToLower(mediaType)
		for _, allowedType := range allowedContentTypes {
			allowedType = strings.ToLower(strings.TrimSpace(allowedType))
			if strings.HasSuffix(allowedType, "/*") {
				prefix := strings.TrimSuffix(allowedType, "/*")
				if strings.HasPrefix(mediaType, prefix) {
					return true
				}
			} else if mediaType == allowedType {
				return true
			}
		}
	}
	return false
}

func isTimeoutError(err error) bool {
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout()
	}
	if err == context.DeadlineExceeded {
		return true
	}
	return false
}

func getCachedResponse(ctx context.Context, storage Storage, cacheKey string, contentLength int64) (*Response, error) {
	reader, info, err := storage.GetObject(ctx, cacheKey)
	if err != nil {
		return nil, ProxyErrorf(ProxyErrStorage, "failed to retrieve cached response: %v", err)
	}
	defer reader.Close()

	response := &Response{
		StatusCode: 200,
		Headers:    info.Metadata,
	}
	// For large responses, return a signed URL
	if contentLength > inlineResponseLimit {
		signedURL, err := storage.SignURL(ctx, cacheKey, 1*time.Hour)
		if err != nil {
			return nil, ProxyErrorf(ProxyErrStorage, "failed to generate signed URL: %v", err)
		}
		response.SignedURL = signedURL
	} else {
		// Read and encode small responses
		body, err := io.ReadAll(reader)
		if err != nil {
			return nil, ProxyErrorf(ProxyErrStorage, "failed to read cached response: %v", err)
		}
		response.Body = base64.StdEncoding.EncodeToString(body)
	}
	return response, nil
}

func GetCacheKey(req *Request) string {
	h := sha256.New()
	io.WriteString(h, req.Method)
	io.WriteString(h, req.URL)
	if req.Body != "" {
		io.WriteString(h, req.Body)
	}
	// Combine cache prefix with host and hash
	return fmt.Sprintf("%s/%s/%x",
		cacheKeyPrefix,
		req.parsedURL.Host,
		h.Sum(nil),
	)
}
