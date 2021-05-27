// Copyright 2021 bilibili-base
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"fmt"
)

// CheckErrors is used to check multi errors
func CheckErrors(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

// PluggableConfig defines the pluggable Config
type PluggableConfig interface {
	// IsEnabled is used to return whether the current component is enabled
	// This attribute is required in pluggable components
	IsEnabled() bool
}

// ValidatableConfig defines the validatable Config
type ValidatableConfig interface {
	// Validate is used to validate config and returns error on failure
	Validate() error
}

// ValidateConfigs is used to validate validatable configs
func ValidateConfigs(configs ...ValidatableConfig) error {
	for _, config := range configs {
		if config == nil {
			return fmt.Errorf("config(%T) is nil", config)
		}
		if err := config.Validate(); err != nil {
			return fmt.Errorf("%T: %s", config, err)
		}
	}
	return nil
}
