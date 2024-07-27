module github.com/myzie/burrow/lambda

go 1.22.2

require (
	github.com/aws/aws-lambda-go v1.47.0
	github.com/myzie/burrow v0.0.1
)

replace github.com/myzie/burrow => ../..
