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

package compiler

import (
	"fmt"
	"github.com/ankyra/escape-core/script"
	"github.com/ankyra/escape-core/variables"
)

func (c *Compiler) compileInputs(inputs []interface{}) error {
	for _, input := range inputs {
		v, err := c.compileVariable(input)
		if err != nil {
			return fmt.Errorf("Error compiling input variable: %s", err.Error())
		}
		c.metadata.AddInputVariable(v)
	}
	return nil
}

func (c *Compiler) compileOutputs(outputs []interface{}) error {
	for _, output := range outputs {
		v, err := c.compileVariable(output)
		if err != nil {
			return fmt.Errorf("Error compiling output variable: %s", err.Error())
		}
		c.metadata.AddOutputVariable(v)
	}
	return nil
}

func (c *Compiler) compileVariable(v interface{}) (result *variables.Variable, err error) {
	switch v.(type) {
	case string:
		result, err = variables.NewVariableFromString(v.(string), "string")
		if err != nil {
			return nil, err
		}
	case map[interface{}]interface{}:
		result, err = variables.NewVariableFromDict(v.(map[interface{}]interface{}))
		if err != nil {
			return nil, err
		}
	default:
		fmt.Errorf("Unexpected type")
	}
	if result.Default != nil {
		return c.compileDefault(result)
	}
	return result, nil
}

func (c *Compiler) compileDefault(v *variables.Variable) (*variables.Variable, error) {
	switch v.Default.(type) {
	case int, float64, bool:
		return v, nil
	case string:
		defaultValue := v.Default.(string)
		_, err := script.ParseScript(defaultValue)
		if err != nil {
			return nil, fmt.Errorf("Couldn't parse expression '%s' in default field: %s", defaultValue, err.Error())
		}
		str, err := RunScriptForCompileStep(defaultValue, c.VariableCtx)
		if err == nil {
			v.Default = &str
		}
		return v, nil
	case []interface{}:
		values := []interface{}{}
		for _, k := range v.Default.([]interface{}) {
			switch k.(type) {
			case string:
				_, err := script.ParseScript(k.(string))
				if err != nil {
					return nil, fmt.Errorf("Couldn't parse expression '%s' in default field: %s", k.(string), err.Error())
				}
				str, err := RunScriptForCompileStep(k.(string), c.VariableCtx)
				if err == nil {
					values = append(values, str)
				} else {
					values = append(values, k)
				}
			}
		}
		v.Default = values
		return v, nil
	}
	return nil, fmt.Errorf("Unexpected type '%T' for default field of variable '%s'", v.Default, v.Id)
}