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

package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// Float64 is used to convert a string to float64
func Float64(val string) float64 {
	ret, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0
	}
	return ret
}

// Match is used to perform the operation of operator (operandX, operandY)
func Match(operandX string, operator string, operandY string) (bool, error) {
	switch operator {
	case "=", "==", "===":
		return operandX == operandY, nil
	case "!=":
		return operandX != operandY, nil
	case ">":
		return Float64(operandX) > Float64(operandY), nil
	case ">=":
		return Float64(operandX) >= Float64(operandY), nil
	case "<":
		return Float64(operandX) < Float64(operandY), nil
	case "<=":
		return Float64(operandX) <= Float64(operandY), nil
	case "regex":
		expr, err := regexp.Compile(operandY)
		if err != nil {
			logrus.Warningln(err)
			return false, nil
		}
		return expr.MatchString(operandX), nil
	case "in":
		return strings.Contains(operandY, operandX), nil
	default:
		return false, fmt.Errorf("unknown operator: %s", operator)
	}
}
