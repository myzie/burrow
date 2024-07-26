package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/myzie/burrow"
)

func readFunctionURLs(path string) (map[string]string, error) {
	functions := make(map[string]string)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&functions); err != nil {
		return nil, err
	}
	return functions, nil
}

func main() {
	var target, functionSpec string
	flag.StringVar(&target, "url", "https://api.ipify.org?format=json", "URL to send a request to")
	flag.StringVar(&functionSpec, "functions", "./function_urls.json", "Function URLs JSON file")
	flag.Parse()

	functions, err := readFunctionURLs(functionSpec)
	if err != nil {
		log.Fatal(err)
	}

	clients := map[string]*http.Client{}
	for region, proxyURL := range functions {
		clients[region] = &http.Client{
			Transport: burrow.NewHTTPTransport(proxyURL, "POST"),
		}
	}

	for region, client := range clients {
		fmt.Println("----")
		fmt.Println(" region:", region)
		fmt.Println(" proxy:", functions[region])
		req, err := http.NewRequest("GET", target, nil)
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
		fmt.Println(" body:", string(body))
	}
}
