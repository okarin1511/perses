// Copyright The Perses Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lint

import (
	"errors"
	"testing"

	"github.com/perses/perses/internal/api/plugin/schema"
	cmdTest "github.com/perses/perses/internal/cli/test"
	modelV1 "github.com/perses/perses/pkg/model/api/v1"
	"github.com/perses/spec/go/dashboard"
	"github.com/stretchr/testify/assert"
)

func TestLintCMD(t *testing.T) {
	testSuite := []cmdTest.Suite{
		{
			Title:           "empty args",
			Args:            []string{},
			IsErrorExpected: true,
			ExpectedMessage: `at least one of the flags in the group [file directory] is required`,
		},
		{
			Title:           "use args",
			Args:            []string{"whatever", "-f", "file.json"},
			IsErrorExpected: true,
			ExpectedMessage: "no args are supported by the command 'lint'",
		},
		{
			Title:           "lint unknown document",
			Args:            []string{"-f", "../../test/sample_resources/unknown_resource.json"},
			IsErrorExpected: true,
			ExpectedMessage: `resource "game" from file "../../test/sample_resources/unknown_resource.json" not supported by the command`,
		},
		{
			Title:           "lint a single resource",
			Args:            []string{"-f", "../../test/sample_resources/single_resource.json"},
			IsErrorExpected: false,
			ExpectedMessage: `your resources look good
`,
		},
		{
			Title:           "lint multiples different resources",
			Args:            []string{"-f", "../../test/sample_resources/multiple_resources.json"},
			IsErrorExpected: false,
			ExpectedMessage: `your resources look good
`,
		},
		{
			Title:           "lint a directory reports every invalid file",
			Args:            []string{"-d", "testdata/lint_invalid"},
			IsErrorExpected: true,
			ExpectedMessage: `2 invalid resource(s) found:
- resource "game" from file "testdata/lint_invalid/unknown_kind.json" not supported by the command
- objects[0] from file "testdata/lint_invalid/zz_missing_kind.json" unable to find 'kind' field`,
		},
	}
	cmdTest.ExecuteSuiteTest(t, NewCMD, testSuite)
}

// failingPanelSchema is a schema that rejects every panel.
type failingPanelSchema struct {
	schema.Schema
}

func (s *failingPanelSchema) ValidateDashboardVariables([]dashboard.Variable) error {
	return nil
}

func (s *failingPanelSchema) ValidateDashboardAnnotations([]dashboard.AnnotationSpec) error {
	return nil
}

func (s *failingPanelSchema) ValidatePanels(map[string]*dashboard.Panel) error {
	return errors.New("invalid panel")
}

func newDashboard(project, name string) *modelV1.Dashboard {
	dash := &modelV1.Dashboard{Kind: modelV1.KindDashboard}
	dash.Metadata.Name = name
	dash.Metadata.Project = project
	return dash
}

func TestValidateReportsEveryInvalidResource(t *testing.T) {
	o := &option{sch: &failingPanelSchema{}}
	lintErrors := o.validate([]sourcedEntity{
		{entity: newDashboard("perses", "first"), file: "first.json"},
		{entity: newDashboard("perses", "second"), file: "second.json"},
	})
	assert.Equal(t, `2 invalid resource(s) found:
- Dashboard "perses/first" (file "first.json"): invalid panel
- Dashboard "perses/second" (file "second.json"): invalid panel`, buildError(lintErrors).Error())
}

func TestLintErrorMessage(t *testing.T) {
	errs := []*lintError{
		{kind: "Dashboard", project: "perses", name: "demo", file: "dashboards/demo.yaml", err: errors.New("invalid panel")},
		{kind: "GlobalDatasource", name: "prom", err: errors.New("invalid plugin")},
		{file: "broken.json", err: errors.New("unable to read")},
	}
	expected := `3 invalid resource(s) found:
- Dashboard "perses/demo" (file "dashboards/demo.yaml"): invalid panel
- GlobalDatasource "prom": invalid plugin
- unable to read`
	assert.Equal(t, expected, buildError(errs).Error())
}
