output "email" {
  description = "The verified SES email identity."
  value       = aws_ses_email_identity.sender.email
}

output "arn" {
  description = "ARN of the SES email identity."
  value       = aws_ses_email_identity.sender.arn
}
