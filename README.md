# terraform-provider-mailjet

[![Go Reference](https://pkg.go.dev/badge/github.com/enalean/terraform-provider-mailjet.svg)](https://pkg.go.dev/github.com/enalean/terraform-provider-mailjet)
[![Go Report Card](https://goreportcard.com/badge/github.com/enalean/terraform-provider-mailjet)](https://goreportcard.com/report/github.com/enalean/terraform-provider-mailjet)
![Github Actions](https://github.com/enalean/terraform-provider-mailjet/actions/workflows/CI.yml/badge.svg?branch=main)

This repository contains the source code for the [Mailjet Terraform provider](https://registry.terraform.io/providers/enalean/mailjet).
This Terraform provider lets you interact with the [Mailjet](https://www.mailjet.com/) API.

See the [documentation](https://registry.terraform.io/providers/enalean/mailjet/latest/docs) in the Terraform registry
for the most up-to-date information and latest release.

This provider is maintained by Enalean.

## Getting Started

To install this provider, copy and paste this code into your Terraform configuration. Then, run `terraform init`:

```terraform
terraform {
  required_providers {
    mailjet = {
      source = "enalean/mailjet"
      version = "~> 0.1" // Latest 0.1.x
    }
  }
}

provider "mailjet" {
  api_key_public = "..."
  api_key_private = "..."
}
```

In the `provider` block, set your API key in the `api_key_public` and `api_key_private` fields.
Alternatively, use the `MJ_APIKEY_PUBLIC` and `MJ_APIKEY_PRIVATE` environment variables.
