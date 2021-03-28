package hw09structvalidator

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type validationLogic = func(fieldName string, fieldValue reflect.Value) *ValidationError

type TypeValidationRule interface {
	Validate(fieldName string, fieldValue reflect.Value, validationLogic validationLogic) ValidationErrors
}

type typeRule struct {
	typeKind reflect.Kind
}

func (rule typeRule) Validate(fieldName string, fieldValue reflect.Value, validate validationLogic) ValidationErrors {
	switch fieldValue.Kind() {
	case rule.typeKind:
		validationError := validate(fieldName, fieldValue)
		if validationError != nil {
			return []ValidationError{*validationError}
		}
		return nil
	case reflect.Slice, reflect.Array:
		errors := ValidationErrors{}
		if fieldValue.IsNil() {
			return errors
		}
		for i := 0; i < fieldValue.Len(); i++ {
			elem := fieldValue.Index(i)
			if elem.Kind() != rule.typeKind {
				errors = append(errors, ValidationError{fieldName, fmt.Errorf("%w: one of field values type is not a %s type", ErrTypeRuleIsInvalid, rule.typeKind.String())})
			}
			if err := validate(fieldName, elem); err != nil {
				errors = append(errors, *err)
			}
		}
		return errors
	default:
		return []ValidationError{{fieldName, fmt.Errorf("%w: field is not a %s or %s array", ErrTypeRuleIsInvalid, rule.typeKind.String(), rule.typeKind.String())}}
	}
}

func parseRuleValue(ruleStr string, ruleStrPrefix string) string {
	if len(ruleStr) == 0 {
		return ""
	}
	if strings.HasPrefix(ruleStr, ruleStrPrefix) {
		ruleStrParts := strings.SplitN(ruleStr, ":", 2)
		if len(ruleStrParts) > 1 {
			return ruleStrParts[1]
		}
	}
	return ""
}

type ValidationRule interface {
	Validate(fieldName string, fieldValue reflect.Value) ValidationErrors
}

type stringLenRule struct {
	typeRule TypeValidationRule
	len      int
}

func (rule stringLenRule) Validate(fieldName string, fieldValue reflect.Value) ValidationErrors {
	return rule.typeRule.Validate(fieldName, fieldValue, func(name string, value reflect.Value) *ValidationError {
		if len(value.String()) != rule.len {
			return &ValidationError{name, fmt.Errorf("%w: field length not equals %d", ErrStrLengthRuleIsInvalid, rule.len)}
		}
		return nil
	})
}

func NewStringLenRule(ruleStr string) (ValidationRule, error) {
	ruleValue := parseRuleValue(ruleStr, "len:")
	length, err := strconv.Atoi(ruleValue)
	if err != nil {
		return nil, ErrStrLengthRuleWrongFormat
	}
	return stringLenRule{typeRule{reflect.String}, length}, nil
}

type stringRegexpRule struct {
	typeRule TypeValidationRule
	regexp   *regexp.Regexp
}

func (rule stringRegexpRule) Validate(fieldName string, fieldValue reflect.Value) ValidationErrors {
	return rule.typeRule.Validate(fieldName, fieldValue, func(name string, value reflect.Value) *ValidationError {
		if !rule.regexp.MatchString(value.String()) {
			return &ValidationError{name, fmt.Errorf("%w: field is not matching regular expression", ErrStrRegexpRuleIsInvalid)}
		}
		return nil
	})
}

func NewStringRegexpRule(ruleStr string) (ValidationRule, error) {
	ruleValue := parseRuleValue(ruleStr, "regexp:")
	r, err := regexp.Compile(ruleValue)
	if err != nil {
		return nil, ErrStrRegexpRuleWrongFormat
	}
	return stringRegexpRule{typeRule{reflect.String}, r}, nil
}

type stringInRule struct {
	typeRule            TypeValidationRule
	availableValuesList []string
}

func (rule stringInRule) Validate(fieldName string, fieldValue reflect.Value) ValidationErrors {
	return rule.typeRule.Validate(fieldName, fieldValue, func(name string, value reflect.Value) *ValidationError {
		for _, availableStr := range rule.availableValuesList {
			if availableStr == value.String() {
				return nil
			}
		}
		return &ValidationError{name, fmt.Errorf("%w field value is not matching any of %v", ErrStrInRuleIsInvalid, rule.availableValuesList)}
	})
}

func NewStringInRule(ruleStr string) (ValidationRule, error) {
	ruleValue := parseRuleValue(ruleStr, "in:")
	valuesList := strings.Split(ruleValue, ValidationRulesValuesSeparator)
	if len(valuesList) == 0 {
		return nil, ErrStrInRuleWrongFormat
	}
	return stringInRule{typeRule{reflect.String}, valuesList}, nil
}

type intMinRule struct {
	typeRule TypeValidationRule
	min      int64
}

func (rule intMinRule) Validate(fieldName string, fieldValue reflect.Value) ValidationErrors {
	return rule.typeRule.Validate(fieldName, fieldValue, func(name string, value reflect.Value) *ValidationError {
		if value.Int() < rule.min {
			return &ValidationError{name, fmt.Errorf("%w: field value is lower than min", ErrIntMinRuleIsInvalid)}
		}
		return nil
	})
}

func NewIntMinRule(ruleStr string) (ValidationRule, error) {
	ruleValue := parseRuleValue(ruleStr, "min:")
	min, err := strconv.ParseInt(ruleValue, 10, 0)
	if err != nil {
		return nil, ErrIntMinRuleWrongFormat
	}
	return intMinRule{typeRule{reflect.Int}, min}, nil
}

type intMaxRule struct {
	typeRule TypeValidationRule
	max      int64
}

func (rule intMaxRule) Validate(fieldName string, fieldValue reflect.Value) ValidationErrors {
	return rule.typeRule.Validate(fieldName, fieldValue, func(name string, value reflect.Value) *ValidationError {
		if value.Int() > rule.max {
			return &ValidationError{name, fmt.Errorf("%w: field value is bigger than max", ErrIntMaxRuleIsInvalid)}
		}
		return nil
	})
}

func NewIntMaxRule(ruleStr string) (ValidationRule, error) {
	ruleValue := parseRuleValue(ruleStr, "max:")
	max, err := strconv.ParseInt(ruleValue, 10, 0)
	if err != nil {
		return nil, ErrIntMaxRuleWrongFormat
	}
	return intMaxRule{typeRule{reflect.Int}, max}, nil
}

type intInRule struct {
	typeRule            TypeValidationRule
	availableValuesList []int64
}

func (rule intInRule) Validate(fieldName string, fieldValue reflect.Value) ValidationErrors {
	return rule.typeRule.Validate(fieldName, fieldValue, func(name string, value reflect.Value) *ValidationError {
		for _, availableValue := range rule.availableValuesList {
			if availableValue == value.Int() {
				return nil
			}
		}
		return &ValidationError{name, fmt.Errorf("%w: field value is not matching any of %v", ErrIntInRuleIsInvalid, rule.availableValuesList)}
	})
}

func NewIntInRule(ruleStr string) (ValidationRule, error) {
	ruleValue := parseRuleValue(ruleStr, "in:")
	strValues := strings.Split(ruleValue, ValidationRulesValuesSeparator)
	valuesList := make([]int64, 0, len(strValues))
	for _, strValue := range strValues {
		parsedInt, err := strconv.ParseInt(strValue, 10, 0)
		if err != nil {
			return nil, ErrIntInRuleWrongFormat
		}
		valuesList = append(valuesList, parsedInt)
	}
	if len(valuesList) == 0 {
		return nil, ErrIntInRuleWrongFormat
	}
	return intInRule{typeRule{reflect.Int}, valuesList}, nil
}
