package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestExportDataSourceDir(t *testing.T) {
	UnitTestRendered(t,
		`{ dir = "testdata/multiple" }`,
		`{"Alice":"Bob","Foo":{"Bar":"Baz"},"Hello":", World!"}`,
	)
}

func TestExportDataSourceSingleArgs(t *testing.T) {
	UnitTestRendered(t,
		`{ args = ["testdata/multiple/example_1.cue"] }`,
		`{"Alice":"Bob"}`,
	)
}

func TestExportDataSourceMultipleArgs(t *testing.T) {
	UnitTestRendered(t,
		`{ args = ["testdata/multiple/example_1.cue", "testdata/multiple/example_3.cue"] }`,
		`{"Alice":"Bob","Hello":", World!"}`,
	)
}

func TestExportDataSourcePath(t *testing.T) {
	UnitTestRendered(t,
		`{
			dir  = "testdata/multiple"
			path = "Foo.Bar"
		}`,
		`"Baz"`,
	)
}

func TestExportDataSourceFull(t *testing.T) {
	UnitTestSteps(t, resource.TestStep{
		Config: `data "cue_export" "test" { dir = "testdata/single" }`,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.cue_export.test", "id",
				"715eda0e975747591d5ed7b5d40c9d95183397598e42023fcc2eeb2ff8e69a24"),
			resource.TestCheckResourceAttr("data.cue_export.test", "rendered",
				"{\"Hello\":\", World!\"}"),
		),
	})
}

func TestExportDataSourceLoadingError(t *testing.T) {
	UnitTestSteps(t, resource.TestStep{
		Config:      `data "cue_export" "test" { dir = "testdata/missing" }`,
		ExpectError: regexp.MustCompile(`Unexpected CUE Loading Error`),
	})
}

func TestExportDataSourceValidationError(t *testing.T) {
	UnitTestSteps(t, resource.TestStep{
		Config:      `data "cue_export" "test" { dir = "testdata/incomplete" }`,
		ExpectError: regexp.MustCompile(`Unexpected CUE Validation Error`),
	})
}

func TestExportDataSourceParsePathError(t *testing.T) {
	UnitTestSteps(t, resource.TestStep{
		Config: `data "cue_export" "test" { 
			dir  = "testdata/multiple"
			path = "path,not,found"
		}`,
		ExpectError: regexp.MustCompile(`Unexpected CUE Parse Path Error`),
	})
}

func TestExportDataSourceLookupPathError(t *testing.T) {
	UnitTestSteps(t, resource.TestStep{
		Config: `data "cue_export" "test" { 
			dir  = "testdata/multiple"
			path = "path.not.found"
		}`,
		ExpectError: regexp.MustCompile(`Unexpected CUE Path Lookup Error`),
	})
}

func TestExportDataSourceConcurrency(t *testing.T) {
	UnitTestSteps(t, resource.TestStep{
		Config: `
			data "cue_export" "test" {
				count = 3
				dir   = "testdata/single"
			}
		`,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.cue_export.test.0", "id",
				"715eda0e975747591d5ed7b5d40c9d95183397598e42023fcc2eeb2ff8e69a24"),
			resource.TestCheckResourceAttr("data.cue_export.test.1", "id",
				"715eda0e975747591d5ed7b5d40c9d95183397598e42023fcc2eeb2ff8e69a24"),
			resource.TestCheckResourceAttr("data.cue_export.test.2", "id",
				"715eda0e975747591d5ed7b5d40c9d95183397598e42023fcc2eeb2ff8e69a24"),
		),
	})
}

func UnitTestSteps(t *testing.T, steps ...resource.TestStep) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps:                    steps,
	})
}

func UnitTestRendered(t *testing.T, config, want string) {
	UnitTestSteps(t, resource.TestStep{
		Config: fmt.Sprintf("data \"cue_export\" \"test\" %s", config),
		Check:  resource.TestCheckResourceAttr("data.cue_export.test", "rendered", want),
	})
}
