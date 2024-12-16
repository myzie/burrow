package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/myzie/burrow"
)

func main() {
	var timeoutDur time.Duration
	var maxRetries int64
	var cacheMaxAge time.Duration
	var target, proxy, method string
	flag.StringVar(&target, "url", "", "URL to send a request to")
	flag.StringVar(&proxy, "proxy", "", "URL of the proxy to use")
	flag.DurationVar(&timeoutDur, "timeout", 0, "Timeout")
	flag.Int64Var(&maxRetries, "retries", 0, "Maximum retries")
	flag.DurationVar(&cacheMaxAge, "cache-max-age", time.Hour, "Cache max age")
	flag.Parse()

	opts := []burrow.ClientOption{
		burrow.WithProxyURL(proxy),
		burrow.WithRetries(int(maxRetries)),
		burrow.WithRetryableCodes([]int{404}),
		burrow.WithCallback(func(ctx context.Context, req *burrow.Request, res *burrow.Response) {
			fmt.Printf("proxy response: %+v\n", res)
		}),
	}

	if timeoutDur > 0 {
		opts = append(opts, burrow.WithTimeout(timeoutDur))
	}
	if cacheMaxAge > 0 {
		opts = append(opts, burrow.WithCacheMaxAge(cacheMaxAge))
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

	isPrintable := resp.Header.Get("Content-Type") == "text/plain" ||
		resp.Header.Get("Content-Type") == "application/json"

	fmt.Println("================")
	if isPrintable {
		fmt.Println(string(body))
	} else {
		fmt.Printf("%d bytes\n", len(body))
		if err := os.WriteFile("response.bin", body, 0644); err != nil {
			fmt.Println("failed to write response to file:", err)
		}
		fmt.Println("response written to response.bin")
	}
}
