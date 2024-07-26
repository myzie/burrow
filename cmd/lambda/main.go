package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/myzie/burrow"
)

type RequestHandler struct {
	Burrow burrow.Handler
}

func (h RequestHandler) Handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var burrowReq burrow.Request
	if err := json.Unmarshal([]byte(request.Body), &burrowReq); err != nil {
		return NewErrorResponse(400, err), nil
	}
	log.Println("burrow url:", burrowReq.URL)
	log.Println("client source ip:", request.RequestContext.HTTP.SourceIP)
	log.Println("client user agent:", request.RequestContext.HTTP.UserAgent)

	burrowRes, err := h.Burrow(ctx, &burrowReq)
	if err != nil {
		if errors.Is(err, &burrow.BadInputError{}) {
			log.Println("bad input error:", err.Error())
			return NewErrorResponse(400, err), nil
		}
		log.Println("failed to handle request:", err)
		return NewErrorResponse(500, err), nil
	}

	burrowRes.ClientDetails = &burrow.ClientDetails{
		SourceIP:  request.RequestContext.HTTP.SourceIP,
		UserAgent: request.RequestContext.HTTP.UserAgent,
	}

	responseBody, err := json.Marshal(burrowRes)
	if err != nil {
		log.Println("failed to marshal response:", err)
		return NewErrorResponse(500, err), nil
	}
	log.Println("burrow success to:", burrowReq.URL)
	log.Println("burrow status code:", burrowRes.StatusCode)

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       string(responseBody),
	}, nil
}

func NewErrorResponse(statusCode int, err error) events.APIGatewayV2HTTPResponse {
	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Body:       err.Error(),
		Headers:    map[string]string{"Content-Type": "text/plain"},
	}
}

func main() {
	h := RequestHandler{Burrow: burrow.GetHandler()}
	lambda.Start(h.Handle)
}
