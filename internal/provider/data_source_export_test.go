package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDataSourceExportAttrRendered(t *testing.T) {
	tests := []struct {
		Name     string
		Config   string
		Expected string
	}{{
		"export with working_dir",
		`
		data "cue_export" "test" {
			working_dir = "testdata"
		}`,
		`{"example":{"one":{"name":"one"},"two":{"name":"two"}}}`,
	}, {
		"export single file",
		`
		data "cue_export" "test" {
			working_dir = "testdata"
			files       = ["example_1.cue"]
		}`,
		`{"example":{"one":{"name":"one"}}}`,
	}, {
		"export with expression",
		`
		data "cue_export" "test" {
			working_dir = "testdata"
			expression = "example.one"
		}`,
		`{"name":"one"}`,
	}}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProviderFactories: ProviderFactories,
				Steps: []resource.TestStep{{
					Config: test.Config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.cue_export.test", "rendered", test.Expected),
					),
				}},
			})
		})
	}
}

func TestDataSourceExportCoroutineSafety(t *testing.T) {
	cfg := `
		data "cue_export" "test" {
			count = 3
			working_dir = "testdata"
		} 
        output "out_0" {
			value = data.cue_export.test[0].rendered
        }
        output "out_1" {
			value = data.cue_export.test[1].rendered
        }
        output "out_2" {
			value = data.cue_export.test[2].rendered
        }
    `
	expected := `{"example":{"one":{"name":"one"},"two":{"name":"two"}}}`

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{{
			Config: cfg,
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckOutput("out_0", expected),
				resource.TestCheckOutput("out_1", expected),
				resource.TestCheckOutput("out_2", expected),
			),
		}},
	})
}
