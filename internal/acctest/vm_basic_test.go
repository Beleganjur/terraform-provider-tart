//go:build acc
// +build acc

package acctest

import (
    "net/http/httptest"
    "testing"

    "github.com/beleganjur/terraform-provider-tart/tart"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_TartVM_Basic(t *testing.T) {
    // Start in-process API server for acceptance tests
    srv := httptest.NewServer(tart.SetupRouter())
    defer srv.Close()
    api := srv.URL + "/api"

    resource.Test(t, resource.TestCase{
        // SDKv2-based provider: use ProviderFactories
        ProviderFactories: map[string]func() (*schema.Provider, error){
            "tart": func() (*schema.Provider, error) { return tart.Provider(), nil },
        },
        Steps: []resource.TestStep{
            {
                Config: testAccConfigVM(api, "acc-vm-1", "ghcr.io/cirruslabs/tart-debian:13.20240922"),
            },
        },
    })
}

func testAccConfigVM(apiURL, name, image string) string {
    return `
terraform {
  required_providers {
    tart = {
      source  = "local/tart/tart"
      version = "0.1.0"
    }
  }
}

provider "tart" {
  api_url = "` + apiURL + `"
}

resource "tart_vm" "example" {
  name  = "` + name + `"
  image = "` + image + `"
}
`
}
