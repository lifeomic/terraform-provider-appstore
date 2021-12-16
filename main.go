package main

import (
	"github.com/lifeomic/terraform-provider-appstore/appstore"

	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: appstore.Provider})
}
