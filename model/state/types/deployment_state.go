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

package types

import (
	"encoding/json"
	"fmt"
)

type DeploymentState struct {
	Name        string                 `json:"name"`
	Release     string                 `json:"release"`
	Stages      map[string]*StageState `json:"stages"`
	Inputs      map[string]interface{} `json:"inputs"`
	environment *EnvironmentState      `json:"-"`
	parent      *DeploymentState       `json:"-"`
	parentStage *StageState            `json:"-"`
}

func NewDeploymentState(env *EnvironmentState, name, release string) *DeploymentState {
	return &DeploymentState{
		Name:        name,
		Release:     release,
		Stages:      map[string]*StageState{},
		Inputs:      map[string]interface{}{},
		environment: env,
	}
}

func (d *DeploymentState) GetName() string {
	return d.Name
}

func (d *DeploymentState) GetRelease() string {
	return d.Release
}

func (d *DeploymentState) GetReleaseId(stage string) string {
	return d.GetRelease() + "-v" + d.GetVersion(stage)
}

func (d *DeploymentState) GetVersion(stage string) string {
	return d.getStage(stage).Version
}

func (d *DeploymentState) GetEnvironmentState() *EnvironmentState {
	return d.environment
}

func (d *DeploymentState) GetDeployment(stage, deploymentName string) *DeploymentState {
	st := d.getStage(stage)
	for _, val := range st.Deployments {
		if val.GetName() == deploymentName {
			val.parentStage = st
			return val
		}
	}
	newDepl := NewDeploymentState(d.environment, deploymentName, deploymentName)
	newDepl.parent = d
	newDepl.parentStage = st
	st.Deployments[deploymentName] = newDepl
	return newDepl
}

func (d *DeploymentState) GetUserInputs(stage string) map[string]interface{} {
	return d.getStage(stage).UserInputs
}

func (d *DeploymentState) GetCalculatedInputs(stage string) map[string]interface{} {
	return d.getStage(stage).Inputs
}

func (d *DeploymentState) GetCalculatedOutputs(stage string) map[string]interface{} {
	return d.getStage(stage).Outputs
}

func (d *DeploymentState) UpdateInputs(stage string, inputs map[string]interface{}) error {
	d.getStage(stage).setInputs(inputs)
	return d.Save()
}

func (d *DeploymentState) UpdateUserInputs(stage string, inputs map[string]interface{}) error {
	d.getStage(stage).setUserInputs(inputs)
	return d.Save()
}

func (d *DeploymentState) UpdateOutputs(stage string, outputs map[string]interface{}) error {
	d.getStage(stage).setOutputs(outputs)
	return d.Save()
}

func (d *DeploymentState) SetVersion(stage, version string) error {
	d.getStage(stage).setVersion(version)
	return nil
}

func (d *DeploymentState) IsDeployed(stage, version string) bool {
	return d.getStage(stage).Version == version
}

func (d *DeploymentState) Save() error {
	return d.GetEnvironmentState().Save(d)
}

func (p *DeploymentState) ToJson() string {
	str, err := json.MarshalIndent(p, "", "   ")
	if err != nil {
		panic(err)
	}
	return string(str)
}

func (d *DeploymentState) SetProvider(stage, name, deplName string) {
	d.getStage(stage).Providers[name] = deplName
}

func (d *DeploymentState) GetProviders(stage string) map[string]string {
	result := map[string]string{}
	for key, val := range d.getStage(stage).Providers {
		result[key] = val
	}
	current := d
	for current.parent != nil {
		current = current.parent
		for key, val := range current.getStage(stage).Providers {
			if _, alreadySet := result[key]; !alreadySet {
				result[key] = val
			}
		}
	}
	return result
}

func (d *DeploymentState) GetPreStepInputs(stage string) map[string]interface{} {
	result := map[string]interface{}{}
	for key, val := range d.environment.getInputs() {
		result[key] = val
	}
	// deps = { this, dep1, dep2, ...., root }
	deps := []*DeploymentState{d}
	stages := []*StageState{d.getStage(stage)}
	prev := d
	p := d.parent
	for p != nil {
		deps = append(deps, p)
		stages = append(stages, prev.parentStage)
		fmt.Println(stages)
		prev = p
		p = p.parent
	}
	// add dep inputs in reverse
	for i := len(deps) - 1; i >= 0; i-- {
		p = deps[i]
		if p.Inputs != nil {
			for key, val := range p.Inputs {
				result[key] = val
			}
		}
		st := p.getStage(stages[i].Name)
		if st.UserInputs != nil {
			for key, val := range st.UserInputs {
				result[key] = val
			}
		}
	}
	return result
}

func (d *DeploymentState) validateAndFix(name string, env *EnvironmentState) error {
	d.Name = name
	d.environment = env
	if d.Name == "" {
		return fmt.Errorf("Deployment name is missing from DeploymentState")
	}
	if d.Release == "" {
		d.Release = name
	}
	if d.Inputs == nil {
		d.Inputs = map[string]interface{}{}
	}
	if d.Stages == nil {
		d.Stages = map[string]*StageState{}
	}
	for name, st := range d.Stages {
		st.validateAndFix(name, env, d)
	}
	return nil
}

func (d *DeploymentState) validateAndFixSubDeployment(stage *StageState, env *EnvironmentState, parent *DeploymentState) error {
	d.parent = parent
	d.parentStage = stage
	return d.validateAndFix(d.Name, env)
}

func (d *DeploymentState) getStage(stage string) *StageState {
	st, ok := d.Stages[stage]
	if !ok || st == nil {
		st = newStage()
		d.Stages[stage] = st
	}
	st.validateAndFix(stage, d.environment, d)
	return st
}