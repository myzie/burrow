output "arn" {
  value = aws_lambda_function.this.arn
}

output "name" {
  value = aws_lambda_function.this.function_name
}
