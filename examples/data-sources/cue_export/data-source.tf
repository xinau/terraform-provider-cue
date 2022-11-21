data "cue_export" "example" {}

output "example" {
  value = jsondecode(data.cue_export.example.rendered)
}
