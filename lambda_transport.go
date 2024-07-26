package burrow

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

// LambdaProxyTransport implements the http.RoundTripper interface
type LambdaProxyTransport struct {
	LambdaClient *lambda.Client
	FunctionName string
}

// RoundTrip implements the http.RoundTripper interface
func (t *LambdaProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	serReq, err := serializeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}
	payload, err := json.Marshal(serReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	result, err := t.LambdaClient.Invoke(req.Context(), &lambda.InvokeInput{
		FunctionName: aws.String(t.FunctionName),
		Payload:      payload,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to invoke Lambda function: %w", err)
	}
	var serResp Response
	if err := json.Unmarshal(result.Payload, &serResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return deserializeResponse(&serResp)
}

// NewLambdaProxyTransport creates a new LambdaProxyTransport
func NewLambdaProxyTransport(lambdaClient *lambda.Client, functionName string) *LambdaProxyTransport {
	return &LambdaProxyTransport{
		LambdaClient: lambdaClient,
		FunctionName: functionName,
	}
}
