## 0.2.0 (Upcoming)
    
CHANGES:

* Provider has been migrated from `terraform-plugin-sdkv2` to
  `terraform-plugin-framework`.

* Attributes `expression`, `files` and `working_dir` inside **cue_export** data
  source have been renamed to `expr`, `paths` and `dir`.

FEATURES:

* Add attribute `pkg` to **cue_export** for the package to be loaded.

* Add attribute `tags` to **cue_export** for injecting values as fields.

## 0.1.0 (Initial Release)
