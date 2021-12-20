# Example of how to use the resource

terraform {
  required_providers {
    appstore = {
      version = "~> 1.0.0"
      source  = "lifeomic.com/tf/appstore" # Doesn't mean anything
    }
  }
}

variable "name" {
  type = string
}

variable "icons" {
  type = list(object({
    src   = string
    sizes = string
  }))
}

variable "start_url" {
  type = string
}

variable "description" {
  type = string
}

# Not the right name, just for testing
variable "url_base" {
  type    = string
  default = "https://lifeapplets.dev.lifeomic.com" # The anxiety part should probably be separate
}

locals {
  app_url = "${var.url_base}/anxiety"
}

provider "appstore" {}

resource "applet" "anxiety" {
  provider       = appstore
  name           = var.name
  description    = var.description
  author_display = "LifeOmic"
  url            = "${local.app_url}/${var.start_url}"
  image          = "${local.app_url}/${var.icons[0].src}"
}
