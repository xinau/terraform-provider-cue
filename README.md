# terraform-provider-cue

Terraform provider for interacting with [CUE](https://cuelang.org/).

## Installation

For terraform >=0.13 add the provider to the `required_providers` inside the `terraform` configuration block.

```hcl
terraform {
  required_providers {
    cue = {
      source = "xinau/cue"
    }
  }
}
```

## Usage

See [provider documentation on the Terraform Registry](https://registry.terraform.io/providers/xinau/cue/latest/docs)

## Development

To compile the provider and test locally, run `go install` and setup `dev_overrides` inside the `.terraformrc` file.
This will build the provider and put the provider binary in the `$GOPATH/bin` for terraform to use, I.e.

```hcl
provider_installation {
  dev_overrides {
    "xinau/cue" = "/home/xinau/go/bin"
  }
  direct {}
}
```

To generate or update the provider documentation, run `go generate`.
