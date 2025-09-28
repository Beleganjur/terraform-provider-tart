# Terraform Provider for Tart Virtualization

A minimal custom Terraform provider for managing Tart virtual machines via a secure Go API controller.

## Usage
See `USAGE.md` for end-to-end instructions, including:

- How to authenticate to GHCR for official Tart images (`tart login ghcr.io`)
- Proper `tart login` syntax and CI-friendly examples
- Quick start with Homebrew, cloning a macOS base image, and running it
- Running the sample Terraform configurations (`test.tf`)

Quick links:

- Usage guide: [USAGE.md](./USAGE.md)
- Tart CLI login docs: https://docs.cirruslabs.org/tart/cli/login/

## Notes on images used in examples

- The images referenced in `main.tf` use public Linux image repositories (e.g., `ghcr.io/cirruslabs/ubuntu:latest`, `ghcr.io/cirruslabs/debian:latest`, `ghcr.io/cirruslabs/fedora:latest`) and typically do not require authentication to pull.
- Some alternative images (for example, DietPi or the `tart-*` repos) may require GHCR authentication. See `USAGE.md` for how to obtain a token and run `tart login ghcr.io` if you encounter 403 (DENIED) errors.
