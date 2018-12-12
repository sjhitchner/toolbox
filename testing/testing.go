package testing

import (
	"math"
	"strings"

	gocheck "gopkg.in/check.v1"
)

// IsTrue Checker
// c.Assert(value, IsTrue)
var IsTrue gocheck.Checker = &isTrueChecker{
	&gocheck.CheckerInfo{
		Name:   "IsTrue",
		Params: []string{"obtained"},
	},
}

type isTrueChecker struct {
	*gocheck.CheckerInfo
}

func (c *isTrueChecker) Check(params []interface{}, names []string) (bool, string) {
	obtained, ok := params[0].(bool)
	if !ok {
		return false, "First parameter must be a boolean"
	}

	return obtained == true, ""
}

// IsFalse Checker
// c.Assert(value, IsFalse)
var IsFalse gocheck.Checker = &isFalseChecker{
	&gocheck.CheckerInfo{
		Name:   "IsFalse",
		Params: []string{"obtained"},
	},
}

type isFalseChecker struct {
	*gocheck.CheckerInfo
}

func (c *isFalseChecker) Check(params []interface{}, names []string) (bool, string) {
	obtained, ok := params[0].(bool)
	if !ok {
		return false, "First parameter must be a boolean"
	}

	return obtained == false, ""
}

// EqualsWithin Checker
// For testing equality for floats
// c.Assert(value, EqualsWithin, expected, torerance)
var EqualsWithin gocheck.Checker = &equalsWithinChecker{
	&gocheck.CheckerInfo{
		Name:   "EqualsWithin",
		Params: []string{"obtained", "expected", "tolerance"},
	},
}

type equalsWithinChecker struct {
	*gocheck.CheckerInfo
}

func (c *equalsWithinChecker) Check(params []interface{}, names []string) (bool, string) {
	obtained, ok := params[0].(float64)
	if !ok {
		return false, "Obtained value must be a float64"
	}

	expected, ok := params[1].(float64)
	if !ok {
		return false, "Expected value must be a float64"
	}

	tolerance, ok := params[2].(float64)
	if !ok {
		return false, "Tolerance value is not a float64"
	}

	return EqualsWithinTolerance(obtained, expected, tolerance), ""
}

func EqualsWithinTolerance(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= math.Abs(tolerance)
}

// Between Checker
// Check a value is between two values
// c.Assert(value Between, lowerBound, upperBound)
var Between gocheck.Checker = &betweenChecker{
	&gocheck.CheckerInfo{
		Name:   "Between",
		Params: []string{"obtained", "lower", "upper"},
	},
}

type betweenChecker struct {
	*gocheck.CheckerInfo
}

func (c *betweenChecker) Check(params []interface{}, names []string) (bool, string) {
	switch obtained := params[0].(type) {
	case int:
		lower, ok := params[1].(int)
		if !ok {
			return false, "Lower value must be an int"
		}

		upper, ok := params[2].(int)
		if !ok {
			return false, "Upper value must be an int"
		}

		return WithinBound(float64(obtained), float64(lower), float64(upper)), ""

	case int64:
		lower, ok := params[1].(int64)
		if !ok {
			return false, "Lower value must be an int64"
		}

		upper, ok := params[2].(int64)
		if !ok {
			return false, "Upper value must be an int64"
		}

		return WithinBound(float64(obtained), float64(lower), float64(upper)), ""

	case float32:
		lower, ok := params[1].(float32)
		if !ok {
			return false, "Lower value must be an float32"
		}

		upper, ok := params[2].(float32)
		if !ok {
			return false, "Upper value must be an float32"
		}

		return WithinBound(float64(obtained), float64(lower), float64(upper)), ""

	case float64:
		lower, ok := params[1].(float64)
		if !ok {
			return false, "Lower value must be an float64"
		}

		upper, ok := params[2].(float64)
		if !ok {
			return false, "Upper value must be an float64"
		}

		return WithinBound(obtained, lower, upper), ""

	default:
		return false, "Obtained value not supported [int, int64, float32, float64]"
	}
}

func WithinBound(value, lower, upper float64) bool {
	return value >= lower && value <= upper
}

// Contains Checker
// Check whether a value is contained in string or slice
var Contains gocheck.Checker = &containsChecker{
	&gocheck.CheckerInfo{
		Name:   "Contains",
		Params: []string{"obtained", "expected"},
	},
}

type containsChecker struct {
	*gocheck.CheckerInfo
}

func (checker *containsChecker) Check(params []interface{}, names []string) (bool, string) {
	switch obtained := params[0].(type) {
	case string:
		expected, ok := params[1].(string)
		if !ok {
			return false, "Expected value must be a string if obtained is string"
		}

		return strings.Contains(obtained, expected), ""

	case []string:
		expected, ok := params[1].(string)
		if !ok {
			return false, "Expected value must be a string if obtained is []string"
		}

		return StringInSlice(obtained, expected), ""

	case []int:
		expected, ok := params[1].(int)
		if !ok {
			return false, "Expected value must be an int if obtained is []int"
		}

		return IntInSlice(obtained, expected), ""

	case []int64:
		expected, ok := params[1].(int64)
		if !ok {
			return false, "Expected value must be an int64 if obtained is []int64"
		}

		return Int64InSlice(obtained, expected), ""

	case []float32:
		expected, ok := params[1].(float32)
		if !ok {
			return false, "Expected value must be a float32 if obtained is []float32"
		}

		return Float32InSlice(obtained, expected), ""

	case []float64:
		expected, ok := params[1].(float64)
		if !ok {
			return false, "Expected value must be a float64 if obtained is []float64"
		}

		return Float64InSlice(obtained, expected), ""

	default:
		return false, "Obtained value not supported [int, int64, float32, float64]"
	}
}

func StringInSlice(list []string, a string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func IntInSlice(list []int, a int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func Int64InSlice(list []int64, a int64) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func Float32InSlice(list []float32, a float32) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func Float64InSlice(list []float64, a float64) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
