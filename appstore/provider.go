package appstore

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	return BuildAppStoreClient()
}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		ConfigureFunc: providerConfigure,
		Schema:        map[string]*schema.Schema{},
		ResourcesMap: map[string]*schema.Resource{
			"applet": appletResource(),
		},
	}
}
