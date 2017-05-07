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

package variable_types

import (
	"errors"
	"fmt"
	. "github.com/ankyra/escape-client/model/interfaces"
)

type listVarType struct {
}

func NewListVariableType() VariableType {
	return &listVarType{}
}

func (s *listVarType) Validate(value interface{}, options map[string]interface{}) (interface{}, error) {
	result := []interface{}{}
	valueType, ok := options["type"]
	if !ok {
		valueType = "string"
	}
	valueType = valueType.(string)
	switch value.(type) {
	case []interface{}:
		for _, val := range value.([]interface{}) {
			switch val.(type) {
			case string:
				if valueType != "string" {
					return nil, errors.New("Unexpected 'string' value in list, expecting '" + valueType.(string) + "'")
				}
				str, err := NewStringVariableType().Validate(val, nil)
				if err != nil {
					return nil, err
				}
				result = append(result, str)
			case int:
				if valueType != "integer" {
					return nil, errors.New("Unexpected 'integer' value in list, expecting '" + valueType.(string) + "'")
				}
				str, err := NewIntegerVariableType().Validate(val, nil)
				if err != nil {
					return nil, err
				}
				result = append(result, str)
			}
		}
		return result, nil
	}
	return nil, fmt.Errorf("Expecting 'list' value, got '%T' (value: %v)", value, value)
}
