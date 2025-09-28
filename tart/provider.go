package tart

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Base URL of Tart API Controller",
			},
			"api_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Bearer token for auth",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"tart_vm": resourceVM(),
		},
		ConfigureFunc: configureProvider,
	}
}

type config struct {
	ApiURL   string
	ApiToken string
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	return &config{
		ApiURL:   d.Get("api_url").(string),
		ApiToken: d.Get("api_token").(string),
	}, nil
}
