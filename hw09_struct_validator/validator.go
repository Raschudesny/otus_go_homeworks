package hw09structvalidator

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const ValidationTag = "validate"

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	panic("implement me")
}

type ValidationRule interface {
	IsValid(value reflect.Value) error
}

type StringLenRule struct {
	length int
}

// TODO pointers ???
func (s StringLenRule) IsValid(value reflect.Value) error {
	switch value.Kind() {
	case reflect.String:
		if len(value.String()) != s.length {
			return fmt.Errorf("field length != %d", s.length)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			elem := value.Index(i)
			if elem.Kind() != reflect.String {
				return fmt.Errorf("one of field values is not a string")
			}
			if len(elem.String()) != s.length {
				return fmt.Errorf("one of field values has wrong length != %d", s.length)
			}
		}
	default:
		return fmt.Errorf("field is not a string or string array")
	}
	return nil
}

type StringRegexpRule struct {
	regexp string
}

func (s StringRegexpRule) IsValid(value reflect.Value) error {
	switch value.Kind() {
	case reflect.String:
		ruleRegexp, err := regexp.Compile(value.String())
		if err != nil {
			return fmt.Errorf("regexp rule for field has incorrect format")
		}
		if !ruleRegexp.MatchString(value.String()) {
			return fmt.Errorf("field is not matching it's regular expression")
		}
	case reflect.Slice, reflect.Array:
	default:
		return fmt.Errorf("field is not a string or string array")
	}
}

func extractValidationRuleValue(ruleStr string, ruleStrPrefix string) string {
	if len(ruleStr) <= 0 {
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

func extractValidationRules(tagValue string) []ValidationRule {
	rulesList := strings.Split(tagValue, "|")
	validationRules := make([]ValidationRule, 0, len(rulesList))
	for _, ruleStr := range rulesList {
		switch {
		case strings.HasPrefix(ruleStr, "len:"):
			length, err := strconv.Atoi(extractValidationRuleValue(ruleStr, "len:"))
			if err == nil {
				validationRules = append(validationRules, StringLenRule{length})
			}
		case strings.HasPrefix(ruleStr, "regex:"):
			extractValidationRuleValue()
		}

	}
}

func validateField(fieldName string, value reflect.Value, rules []ValidationRule) ValidationErrors {
	errors := ValidationErrors{}
	for _, rule := range rules {
		if err := rule.IsValid(value); err != nil {
			errors = append(errors, ValidationError{fieldName, err})
		}
	}
	return errors
}

func ValidateStruct(vi interface{}) error {
	errors := ValidationErrors{}
	v := reflect.ValueOf(vi)
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("validate struct error: expected a struct but recieved %T", vi)
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldType := t.Field(i)
		if tagValue, ok := fieldType.Tag.Lookup(ValidationTag); ok && len(tagValue) > 0 {
			validationRules := make([]ValidationRule, 0)
			errors = append(errors, validateField(fieldType.Name, fieldValue, validationRules)...)
		}
	}
	return nil
}
