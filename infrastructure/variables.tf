variable "aws_region" {
  type        = string
  description = "Region of which the lambda operates in"
  default     = "eu-west-1"
}

variable "lambda_function_name" {
  type        = string
  description = "Lambda function name"
  default     = "osrs-progress-lambda"
}

variable "sort_by" {
  type        = string
  description = "Sets the descending order of the discord embed"
  default     = "Exp"
}

variable "usernames" {
  type        = string
  default     = ""
  description = "List of usernames"
  sensitive   = true
}

variable "webhook_url" {
  type        = string
  default     = ""
  sensitive   = true
  description = "Webhook Url for the server which the embeds will be sent to."
}

variable "gains_query_mode" {
  type        = string
  default     = "period"
  description = "Query parameters for the wise old man api. Period or range."
}

variable "timezone" {
  type        = string
  default     = ""
  description = "The timezone to use when gains query mode is set to range."
}

variable "image_url" {
  type        = string
  default     = "https://oldschool.runescape.wiki/images/Cheer_%28Penguin%29_emote_icon.png?e60bd"
  description = "Url to an Image which will appear in the Discord embed"
}

variable "thumbnail_url" {
  type        = string
  default     = "https://oldschool.runescape.wiki/images/Cheer_%28Penguin%29_emote_icon.png?e60bd"
  description = "Url to an Image which will appear in the Discord embed"
}

variable "gains_period" {
  type        = string
  default     = "day"
  description = "The period of time to query the API for gains."
}

variable "embed_colour" {
  type        = string
  default     = ""
  description = "The colour of the embed"
}