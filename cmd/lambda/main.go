package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/myzie/burrow"
)

type RequestHandler struct {
	Burrow burrow.Handler
	Logger *slog.Logger
}

func (h RequestHandler) Handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var burrowReq burrow.Request
	if err := json.Unmarshal([]byte(request.Body), &burrowReq); err != nil {
		return NewGenericErrorResponse(400, fmt.Errorf("invalid request body (expected json)")), nil
	}
	if burrowReq.Method == "" {
		burrowReq.Method = "GET"
	}
	proxyName := fmt.Sprintf("aws.lambda.%s", getRegion())

	h.Logger.Info("request received",
		"proxy_name", proxyName,
		"url", burrowReq.URL,
		"method", burrowReq.Method,
		"timeout", burrowReq.Timeout,
		"max_response_bytes", burrowReq.MaxResponseBytes,
		"allowed_content_types", burrowReq.AllowedContentTypes,
		"client_ip", request.RequestContext.HTTP.SourceIP,
		"user_agent", request.RequestContext.HTTP.UserAgent)

	response, err := h.Burrow(ctx, &burrowReq)
	if err != nil {
		var proxyErr *burrow.ProxyError
		if errors.As(err, &proxyErr) {
			h.Logger.Error("proxy error", "error", err)
			return NewProxyErrorResponse(proxyErr), nil
		}
		h.Logger.Error("unknown error", "error", err)
		return NewGenericErrorResponse(500, err), nil
	}

	response.ClientDetails = &burrow.ClientDetails{
		SourceIP:  request.RequestContext.HTTP.SourceIP,
		UserAgent: request.RequestContext.HTTP.UserAgent,
	}
	response.ProxyName = proxyName
	responseBody, err := json.Marshal(response)
	if err != nil {
		h.Logger.Error("marshalling error", "error", err)
		return NewGenericErrorResponse(500, err), nil
	}

	h.Logger.Info("request completed",
		"proxy_name", proxyName,
		"url", burrowReq.URL,
		"method", burrowReq.Method,
		"duration", response.Duration,
		"status_code", response.StatusCode,
		"body_size", len(responseBody),
		"content_type", response.Headers["Content-Type"])

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       string(responseBody),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

func NewGenericErrorResponse(statusCode int, err error) events.APIGatewayV2HTTPResponse {
	body, err := json.Marshal(map[string]string{"message": err.Error()})
	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: statusCode,
			Body:       err.Error(),
			Headers:    map[string]string{"Content-Type": "text/plain"},
		}
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Body:       string(body),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

func NewProxyErrorResponse(proxyErr *burrow.ProxyError) events.APIGatewayV2HTTPResponse {
	statusCode := 500
	if proxyErr.Type == burrow.ProxyErrBadRequest {
		statusCode = 400
	}
	proxyErrBody, err := json.Marshal(proxyErr)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: statusCode,
			Body:       fmt.Sprintf(`{"message": "marshalling error: %s"}`, err.Error()),
			Headers:    map[string]string{"Content-Type": "application/json"},
		}
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Body:       string(proxyErrBody),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

func getRegion() string {
	if region := os.Getenv("AWS_REGION"); region != "" {
		return region
	}
	return os.Getenv("AWS_DEFAULT_REGION")
}

func main() {
	h := RequestHandler{
		Burrow: burrow.GetHandler(),
		Logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
	lambda.Start(h.Handle)
}
