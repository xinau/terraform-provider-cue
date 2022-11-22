package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestExportDataSourceRendered(t *testing.T) {
	t.Parallel()

	type testCase struct {
		config string
		want   string
	}

	tests := map[string]testCase{
		"inside dir": {
			config: `{ dir = "testdata/multiple" }`,
			want:   `{"Alice":"Bob","Foo":{"Bar":"Baz"},"Hello":", World!"}`,
		},
		"single path": {
			config: `{ paths = ["testdata/multiple/example_1.cue"] }`,
			want:   `{"Alice":"Bob"}`,
		},
		"multiple paths": {
			config: `{ paths = ["testdata/multiple/example_1.cue", "testdata/multiple/example_3.cue"] }`,
			want:   `{"Alice":"Bob","Hello":", World!"}`,
		},
		"lookup expr": {
			config: `{
				dir  = "testdata/multiple"
				expr = "Foo.Bar"
			}`,
			want: `"Baz"`,
		},
		"load package": {
			config: `{
				dir = "testdata/packages"
				pkg = "example"
			}`,
			want: `{"Hello":", World!"}`,
		},
		"inject tags": {
			config: `{
				dir  = "testdata/single"
				tags = ["name=Alice"]
			}`,
			want: `{"Hello":", Alice!"}`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			UnitTestSteps(t, resource.TestStep{
				Config: fmt.Sprintf("data \"cue_export\" \"test\" %s", test.config),
				Check:  resource.TestCheckResourceAttr("data.cue_export.test", "rendered", test.want),
			})
		})
	}
}

func TestExportDataSourceFull(t *testing.T) {
	UnitTestSteps(t, resource.TestStep{
		Config: `data "cue_export" "test" { dir = "testdata/single" }`,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.cue_export.test", "id",
				"92c9a4b5349b70f09b04dc59628ac36c"),
			resource.TestCheckResourceAttr("data.cue_export.test", "rendered",
				"{\"Hello\":\", World!\"}"),
		),
	})
}

func TestExportDataSourceErrors(t *testing.T) {
	t.Parallel()

	type testCase struct {
		config string
		want   *regexp.Regexp
	}

	tests := map[string]testCase{
		"loading instances error": {
			config: `{ dir = "testdata/missing" }`,
			want:   regexp.MustCompile(`Loading Instances Error`),
		},
		"value validation error": {
			config: `{ dir = "testdata/incomplete" }`,
			want:   regexp.MustCompile(`Value Validation Error`),
		},
		"expression parsing error": {
			config: `{ expr = "path,not,found" }`,
			want:   regexp.MustCompile(`Expression Parsing Error`),
		},
		"lookup expr error": {
			config: `{ 
				dir  = "testdata/multiple"
				expr = "path.not.found"
			}`,
			want: regexp.MustCompile(`Lookup Path Error`),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			UnitTestSteps(t, resource.TestStep{
				Config:      fmt.Sprintf("data \"cue_export\" \"test\" %s", test.config),
				ExpectError: test.want,
			})
		})
	}
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
				"92c9a4b5349b70f09b04dc59628ac36c"),
			resource.TestCheckResourceAttr("data.cue_export.test.1", "id",
				"92c9a4b5349b70f09b04dc59628ac36c"),
			resource.TestCheckResourceAttr("data.cue_export.test.2", "id",
				"92c9a4b5349b70f09b04dc59628ac36c"),
		),
	})
}

func UnitTestSteps(t *testing.T, steps ...resource.TestStep) {
	t.Helper()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps:                    steps,
	})
}
