## 0.2.0 (Upcoming)
    
CHANGES:

* Provider has been migrated from `terraform-plugin-sdkv2` to
  `terraform-plugin-framework`.

* Attributes `expression`, `files` and `working_dir` inside **cue_export** data
  source have been renamed to `path`, `args` and `dir` to better reflect the
  CUE terminology.

FEATURES:

* Add attribute `pkg` to **cue_export** for the package to be loaded.

## 0.1.0 (Initial Release)
