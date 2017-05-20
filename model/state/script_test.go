/*
Copyright 2017 Ankyra

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package state

import (
	. "github.com/ankyra/escape-client/model/state/types"
	"github.com/ankyra/escape-core"
	"github.com/ankyra/escape-core/script"
	"github.com/ankyra/escape-core/variables"
	. "gopkg.in/check.v1"
	"testing"
)

type deplSuite struct{}

var _ = Suite(&deplSuite{})

func Test(t *testing.T) { TestingT(t) }

var depl *DeploymentState
var deplWithDeps *DeploymentState
var fullDepl *DeploymentState

func (s *deplSuite) SetUpTest(c *C) {
	var err error
	p, err := NewProjectStateFromFile("prj", "testdata/project.json")
	c.Assert(err, IsNil)
	env := p.GetEnvironmentStateOrMakeNew("dev")
	depl, err = env.GetDeploymentState([]string{"archive-release"})
	c.Assert(err, IsNil)

	deplWithDeps, err = env.GetDeploymentState([]string{"archive-release-with-deps", "archive-release"})
	c.Assert(err, IsNil)

	fullDepl, err = env.GetDeploymentState([]string{"archive-full"})
	c.Assert(err, IsNil)
}

func (s *deplSuite) Test_ToScript(c *C) {
	metadata := core.NewReleaseMetadata("test", "1.0")
	metadata.Metadata["value"] = "yo"
	input := variables.NewVariableFromString("user_level", "string")
	metadata.AddInputVariable(input)
	metadata.AddOutputVariable(input)
	unit := toScript(depl, metadata, "deploy")
	dicts := map[string][]string{
		"inputs":   []string{"user_level"},
		"outputs":  []string{"user_level"},
		"metadata": []string{"value"},
	}
	test_helper_check_script_environment(c, unit, dicts)
}

func (s *deplSuite) Test_ToScript_doesnt_include_variable_that_are_not_defined_in_release_metadata(c *C) {
	metadata := core.NewReleaseMetadata("test", "1.0")
	unit := toScript(depl, metadata, "deploy")
	dicts := map[string][]string{
		"inputs":   []string{},
		"outputs":  []string{},
		"metadata": []string{},
	}
	test_helper_check_script_environment(c, unit, dicts)
}

func test_helper_check_script_environment(c *C, unit script.Script, dicts map[string][]string) {
	c.Assert(script.IsDictAtom(unit), Equals, true)
	dict := script.ExpectDictAtom(unit)
	strings := map[string]string{
		"version":     "1.0",
		"description": "",
		"logo":        "",
		"id":          "test-v1.0",
		"name":        "test",
		"branch":      "",
		"revision":    "",
		"project":     "project_name",
		"environment": "dev",
		"deployment":  "archive-release",
	}
	for key, val := range strings {
		c.Assert(script.IsStringAtom(dict[key]), Equals, true, Commentf("Expecting %s to be of type string, but was %T", key, dict[key]))
		c.Assert(script.ExpectStringAtom(dict[key]), Equals, val)
	}
	for key, keys := range dicts {
		c.Assert(script.IsDictAtom(dict[key]), Equals, true, Commentf("Expecting %s to be of type dict, but was %T", key, dict[key]))
		d := script.ExpectDictAtom(dict[key])
		c.Assert(d, HasLen, len(keys), Commentf("Expecting %d values in %s dict.", len(keys), key))
		for _, k := range keys {
			c.Assert(script.IsStringAtom(d[k]), Equals, true, Commentf("Expecting %s to be of type string, but was %T", k, d[k]))
		}
	}
}
