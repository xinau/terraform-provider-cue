---
layout: ""
page_title: "CUE Provider"
description: |-
The CUE provider provides resources to interact with [CUE](https://cuelang.org/).
---

# CUE Provider

The CUE provider provides resources to interact with [CUE](https://cuelang.org/).

## Example Usage

```terraform
data "cue_export" "config" {}

locals {
  config = jsondecode(data.cue_export.config.rendered)
}

output "name" {
  value = config.name
}

output "port" {
  value = config.port
}
```
