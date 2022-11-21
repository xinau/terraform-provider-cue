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
		"single argument": {
			config: `{ args = ["testdata/multiple/example_1.cue"] }`,
			want:   `{"Alice":"Bob"}`,
		},
		"multiple arguments": {
			config: `{ args = ["testdata/multiple/example_1.cue", "testdata/multiple/example_3.cue"] }`,
			want:   `{"Alice":"Bob","Hello":", World!"}`,
		},
		"lookup path": {
			config: `{
				dir  = "testdata/multiple"
				path = "Foo.Bar"
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
				"715eda0e975747591d5ed7b5d40c9d95183397598e42023fcc2eeb2ff8e69a24"),
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
		"cue instances loading error": {
			config: `{ dir = "testdata/missing" }`,
			want:   regexp.MustCompile(`Unexpected CUE Loading Error`),
		},
		"cue value validation error": {
			config: `{ dir = "testdata/incomplete" }`,
			want:   regexp.MustCompile(`Unexpected CUE Validation Error`),
		},
		"path parser error": {
			config: `{ path = "path,not,found" }`,
			want:   regexp.MustCompile(`Invalid Attribute Value`),
		},
		"path not found": {
			config: `{ 
				dir  = "testdata/multiple"
				path = "path.not.found"
			}`,
			want: regexp.MustCompile(`Unexpected CUE Path Lookup Error`),
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
				"715eda0e975747591d5ed7b5d40c9d95183397598e42023fcc2eeb2ff8e69a24"),
			resource.TestCheckResourceAttr("data.cue_export.test.1", "id",
				"715eda0e975747591d5ed7b5d40c9d95183397598e42023fcc2eeb2ff8e69a24"),
			resource.TestCheckResourceAttr("data.cue_export.test.2", "id",
				"715eda0e975747591d5ed7b5d40c9d95183397598e42023fcc2eeb2ff8e69a24"),
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
