provider "aws" {
  region = var.aws_region
}



resource "aws_iam_role" "lambda" {
  name = "osrs_progress_lambda"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Sid    = ""
      Principal = {
        Service = "lambda.amazonaws.com"
      }
      }
    ]
  })
}

resource "aws_cloudwatch_log_group" "lambda_logs" {
  name              = "/aws/lambda/${var.lambda_function_name}"
  retention_in_days = 7
}

resource "aws_iam_role_policy_attachment" "lambda_basic_execution" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_cloudwatch_event_rule" "daily" {
  name        = "trigger-${var.lambda_function_name}"
  description = "Triggers the Osrs progress lambda on a fixed cron schedule"

  schedule_expression = "cron(50 12 * * *)"
}

resource "aws_cloudwatch_event_target" "lambda" {
  target_id = var.lambda_function_name
  arn       = aws_lambda_function.osrs_progress.arn
  rule      = aws_cloudwatch_event_rule.daily.name
}

resource "aws_lambda_permission" "eventbridge" {
  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.osrs_progress.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.daily.arn
}

resource "aws_lambda_function" "osrs_progress" {
  function_name = var.lambda_function_name
  description   = "Fetches data from Wise Old Man API for a list of usernames and then posts a sorted list to Discord"
  role          = aws_iam_role.lambda.arn
  handler       = "bootstrap"

  runtime          = "provided.al2023"
  architectures    = ["arm64"]
  source_code_hash = filebase64sha256("${path.module}/../build/osrs-progress-lambda.zip")
  memory_size      = 128
  timeout          = 300
  filename         = "${path.module}/../build/osrs-progress-lambda.zip"

  logging_config {
    log_group  = aws_cloudwatch_log_group.lambda_logs.name
    log_format = "JSON"
  }

  environment {
    variables = {
      USERNAMES        = var.usernames
      SORT_BY          = var.sort_by
      WEBHOOK_URL      = var.webhook_url
      GAINS_QUERY_MODE = var.gains_query_mode
      GAINS_PERIOD     = var.gains_period
      TIMEZONE         = var.timezone
      EMBED_COLOUR     = var.embed_colour
      THUMBNAIL_URL    = var.thumbnail_url
      IMAGE_URL        = var.image_url
    }
  }

  depends_on = [
    aws_cloudwatch_log_group.lambda_logs,
    aws_iam_role_policy_attachment.lambda_basic_execution
  ]
}