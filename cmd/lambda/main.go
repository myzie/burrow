package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/myzie/burrow"
)

func main() {
	lambda.Start(burrow.GetHandler())
}
