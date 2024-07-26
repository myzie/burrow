package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/myzie/burrow"
)

func main() {
	var target, proxy, method string
	flag.StringVar(&target, "url", "", "URL to send a request to")
	flag.StringVar(&proxy, "proxy", "", "URL of the proxy to use")
	flag.Parse()

	transport := burrow.NewTransport(proxy, "POST")
	client := &http.Client{Transport: transport}
	req, err := http.NewRequest(method, target, nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(body))
}
