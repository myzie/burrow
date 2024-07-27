# Burrow

Burrow is a serverless and globally-distributed HTTP proxy for Go.

It is designed to be completely compatible with the standard Go `*http.Client`
which means it can be transparently added to many existing applications. Burrow
provides an implementation of the `http.RoundTripper` interface that proxies
requests through one or more AWS Lambda functions exposed with
[Function URLs](https://docs.aws.amazon.com/lambda/latest/dg/urls-configuration.html).

A round-robin transport is also provided which makes it trivial to automatically
rotate through multiple Lambda functions in different regions.

## Features

- Easy-to-use proxy via `http.RoundTripper` implementation
- Optional round-robin transport for rotating through multiple lambda proxies
- Terraform for one command deployment to 17 AWS regions

## Usage

Add the Burrow package to your Go project:

```bash
go get github.com/myzie/burrow
```

Enable on an `http.Client`:

```go
proxy := "https://randomprefix.lambda-url.eu-west-2.on.aws/"
client := &http.Client{Transport: burrow.NewTransport(proxy, "POST")}
// Now use the *http.Client as you would normally
```

Create a round-robin transport:

```go
proxyURLs := []string{
    "https://randomprefix1.lambda-url.us-east-1.on.aws/",
    "https://randomprefix2.lambda-url.us-east-2.on.aws/",
    "https://randomprefix3.lambda-url.us-west-1.on.aws/",
}
var transports []http.RoundTripper
for _, u := range proxyURLs {
    transports = append(transports, NewTransport(u, "POST"))
}
client := &http.Client{
    Transport: NewRoundRobinTransport(transports),
}
// Client will now rotate through the provided proxies for each request
```

Or use the `burrow.NewRoundRobinClient` helper:

```go
client := burrow.NewRoundRobinClient([]string{
    "https://randomprefix1.lambda-url.us-east-1.on.aws/",
    "https://randomprefix2.lambda-url.us-east-2.on.aws/",
    "https://randomprefix3.lambda-url.us-west-1.on.aws/",
})
```

## AWS Multi-Region Deployment

Burrow includes Terraform configurations to deploy Burrow across the 17 default
enabled AWS regions in your account with a single command:

```bash
make deploy BUCKET_NAME=my-terraform-state-bucket
```

See the Makefile for more information. You'll need the following installed:

- Terraform
- AWS CLI
- Make
- Go
- JQ

## Future Enhancements

- Optional API key authentication in the Lambda proxy

## Examples

- [cmd/example_client](cmd/example_client)
- [cmd/example_multi_region](cmd/example_multi_region)

## Contributing

Contributions to Burrow are welcome! Please feel free to submit a Pull Request.

## License

Apache License 2.0
