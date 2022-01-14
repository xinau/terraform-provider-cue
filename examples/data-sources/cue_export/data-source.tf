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
