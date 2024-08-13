package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/myzie/burrow"
)

func main() {
	var timeoutDur time.Duration
	var maxResponseBytes, maxRetries int64
	var allowedContentTypes string
	var target, proxy, method string
	flag.StringVar(&target, "url", "", "URL to send a request to")
	flag.StringVar(&proxy, "proxy", "", "URL of the proxy to use")
	flag.Int64Var(&maxResponseBytes, "max-response-bytes", 0, "Maximum response body size")
	flag.DurationVar(&timeoutDur, "timeout", 0, "Timeout")
	flag.Int64Var(&maxRetries, "retries", 0, "Maximum retries")
	flag.StringVar(&allowedContentTypes, "allowed-content-types", "", "Allowed content types")
	flag.Parse()

	allowedContentTypesList := strings.Split(allowedContentTypes, ",")

	opts := []burrow.ClientOption{
		burrow.WithProxyURL(proxy),
		burrow.WithRetries(int(maxRetries)),
		burrow.WithRetryableCodes([]int{404}),
		burrow.WithCallback(func(proxyResponse *burrow.Response) {
			fmt.Printf("proxy response: %+v\n", proxyResponse)
		}),
	}
	if maxResponseBytes > 0 {
		opts = append(opts, burrow.WithMaxResponseBytes(maxResponseBytes))
	}
	if timeoutDur > 0 {
		opts = append(opts, burrow.WithTimeout(timeoutDur))
	}
	if len(allowedContentTypesList) > 0 {
		opts = append(opts, burrow.WithAllowedContentTypes(allowedContentTypesList))
	}
	client := burrow.NewClient(opts...)

	req, err := http.NewRequest(method, target, nil)
	if err != nil {
		fmt.Println("failed to create request:", err)
		os.Exit(1)
	}

	resp, err := client.Do(req)
	if err != nil {
		var proxyErr *burrow.ProxyError
		if errors.As(err, &proxyErr) {
			fmt.Println("proxy error:", proxyErr.Message)
			fmt.Println("proxy error type:", proxyErr.Type)
			os.Exit(1)
		}
		fmt.Println("unknown error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("unexpected status code: %d\n", resp.StatusCode)
		os.Exit(1)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("failed to read response body:", err)
		os.Exit(1)
	}

	fmt.Println("================")
	fmt.Println(string(body))
}
