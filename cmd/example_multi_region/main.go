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

	var proxyURLs []string
	for _, proxyURL := range functions {
		proxyURLs = append(proxyURLs, proxyURL)
	}

	client := burrow.NewRoundRobinClient(proxyURLs)

	for {
		body, err := runRequest(client, target)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(body)
	}
}

func runRequest(c *http.Client, target string) (string, error) {
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
