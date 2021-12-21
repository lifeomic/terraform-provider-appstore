# <provider> Provider

This provider is for managing LifeOmic app store resources. Typically these applets are then published on the marketplace (marketplace provider is still under development). If you're working to use this provider for external applets, contact LifeOmic for assistance. A self-serve experience in under development.

The provider uses your local AWS config in order to authenticate. Support for token-based authentication with the public graphql-proxy will probably come in the future.

## Example Usage

```hcl
provider "appstore" {}

resource "applet" "example" {
  provider       = appstore
  name           = "Example Applet for App Store"
  description    = "This applet is created and managed using terraform"
  author_display = "LifeOmic"
  url            = "applets.example.com/route/to/applet"
  image          = "applets.example.com/route/to/applet/icon.png"
}
```

## Argument Reference

* name: string
* description: string
* author_display: string
* url: string
* image: string

