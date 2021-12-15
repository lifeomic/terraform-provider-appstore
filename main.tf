terraform {
  required_providers {
    appstore = {
      version = "~> 1.0.0"
      source  = "lifeomic.com/tf/appstore" # Doesn't mean anything
    }
  }
}

provider "appstore" {}

resource "applet" "anxiety" {
  provider = appstore
  name = "Anxiety Applet"
  description = "Some fancy description, with edits"
  author_display = "LifeOmic"
  url = "https://lifeapplets.dev.lifeomic.com/anxiety/"
  image = "https://lifeapplets.dev.lifeomic.com/anxiety/icon-240.png"
}

# output "app_name" {
#   value = data.application.anxiety.name
# }
