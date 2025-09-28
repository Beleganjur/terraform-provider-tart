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

# Debian (public image)
resource "tart_vm" "debian_13" {
  name  = "debian"
  image = "ghcr.io/cirruslabs/debian:latest"
}

# Ubuntu (public image)
resource "tart_vm" "ubuntu_2204" {
  name  = "ubuntu"
  image = "ghcr.io/cirruslabs/ubuntu:latest"
}

# Fedora (public image)
resource "tart_vm" "fedora" {
  name  = "fedora"
  image = "ghcr.io/cirruslabs/fedora:latest"
}

# DietPi ARM64
resource "tart_vm" "dietpi_arm64" {
  name  = "dietpi-arm64"
  image = "ghcr.io/cirruslabs/tart-dietpi:20240901-arm64"
}

output "debian_13_vm_name" { value = tart_vm.debian_13.name }
output "debian_13_vm_status" { value = tart_vm.debian_13.status }

output "ubuntu_2204_vm_name" { value = tart_vm.ubuntu_2204.name }
output "ubuntu_2204_vm_status" { value = tart_vm.ubuntu_2204.status }

output "fedora_vm_name" { value = tart_vm.fedora.name }
output "fedora_vm_status" { value = tart_vm.fedora.status }

output "dietpi_vm_name" { value = tart_vm.dietpi_arm64.name }
output "dietpi_vm_status" { value = tart_vm.dietpi_arm64.status }
