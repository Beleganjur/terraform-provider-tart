terraform {
  required_providers {
    tart = {
      source  = "local/tart/tart"
      version = "0.1.3"
    }
  }
}

provider "tart" {
  api_url   = "http://localhost:8085/api"
  # api_token = "optional-bearer-token"
}

# Create a Debian 13 VM
resource "tart_vm" "debian_13" {
  name  = "debian-13"
  image = "ghcr.io/cirruslabs/tart-debian:13.20240922"
}

# Create a Sequoia base VM
resource "tart_vm" "sequoia_base" {
  name  = "sequoia-base"
  image = "sequoia-base"
}

output "debian_13_vm_name" {
  value = tart_vm.debian_13.name
}

output "debian_13_vm_status" {
  value = tart_vm.debian_13.status
}

output "sequoia_base_vm_name" {
  value = tart_vm.sequoia_base.name
}

output "sequoia_base_vm_status" {
  value = tart_vm.sequoia_base.status
}
