
output "arn" {
  value = aws_lambda_function.lambda.arn
}

output "name" {
  value = aws_lambda_function.lambda.function_name
}

output "url" {
  value = aws_lambda_function_url.lambda.function_url
}
