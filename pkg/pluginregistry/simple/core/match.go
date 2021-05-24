package core

import (
	"fmt"
	"regexp"
	"strconv"

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
	default:
		return false, fmt.Errorf("unknown operator: %s", operator)
	}
}
