package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
	start := time.Now()
	proxyName := fmt.Sprintf("aws.lambda.%s", getRegion())

	h.Logger.Info("request received",
		"proxy_name", proxyName,
		"url", burrowReq.URL,
		"method", burrowReq.Method,
		"timeout", burrowReq.Timeout,
		"cache_max_age", burrowReq.CacheMaxAge,
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
	response.Duration = time.Since(start).Seconds()
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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	bucketName := os.Getenv("BUCKET_NAME")
	bucketRegion := os.Getenv("BUCKET_REGION")

	if bucketName == "" || bucketRegion == "" {
		logger.Error("BUCKET_REGION and BUCKET_NAME must be set")
		os.Exit(1)
	}

	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(bucketRegion),
	)
	if err != nil {
		logger.Error("failed to load aws config", "error", err)
		os.Exit(1)
	}
	storage := burrow.NewS3Storage(s3.NewFromConfig(cfg), bucketName)

	h := RequestHandler{
		Burrow: burrow.GetHandler(burrow.DefaultClient, storage),
		Logger: logger,
	}
	lambda.Start(h.Handle)
}
