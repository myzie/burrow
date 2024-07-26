package burrow

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

var _ http.RoundTripper = &LambdaTransport{}

// LambdaTransport implements the http.RoundTripper interface. Used to proxy
// HTTP requests via invoking an AWS Lambda function.
type LambdaTransport struct {
	client       *lambda.Client
	functionName string
}

// RoundTrip implements the http.RoundTripper interface
func (t *LambdaTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	serReq, err := serializeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}
	payload, err := json.Marshal(serReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	result, err := t.client.Invoke(req.Context(), &lambda.InvokeInput{
		FunctionName: aws.String(t.functionName),
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

// NewLambdaTransport creates a new LambdaTransport
func NewLambdaTransport(client *lambda.Client, functionName string) *LambdaTransport {
	return &LambdaTransport{
		client:       client,
		functionName: functionName,
	}
}
