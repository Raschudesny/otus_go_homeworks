package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const (
	ValidationTag                  = "validate"
	ValidationRulesSeparator       = "|"
	ValidationRulesValuesSeparator = ","
)

var (
	ErrOnlyStructValidationAllowed = errors.New("non struct argument passed for validation")
	ErrStrLengthRuleIsInvalid      = errors.New("string length rule is invalid")
	ErrStrRegexpRuleIsInvalid      = errors.New("string regexp rule is invalid")
	ErrStrInRuleIsInvalid          = errors.New("string in rule is invalid")
	ErrIntMaxRuleIsInvalid         = errors.New("int max rule is invalid")
	ErrIntMinRuleIsInvalid         = errors.New("int min rule is invalid")
	ErrIntInRuleIsInvalid          = errors.New("int in rule is invalid")
	ErrTypeRuleIsInvalid           = errors.New("field type is invalid")
	ErrStrLengthRuleWrongFormat    = errors.New("string length rule wrong format")
	ErrStrRegexpRuleWrongFormat    = errors.New("string regexp rule wrong format")
	ErrStrInRuleWrongFormat        = errors.New("string in rule wrong format")
	ErrIntMaxRuleWrongFormat       = errors.New("int max rule wrong format")
	ErrIntMinRuleWrongFormat       = errors.New("int min rule wrong format")
	ErrIntInRuleWrongFormat        = errors.New("int in rule wrong format")
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var b strings.Builder
	for i, err := range v {
		b.WriteString(err.Field)
		b.WriteString(" field")
		b.WriteString(" - ")
		b.WriteString(err.Err.Error())
		if i != len(v)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func extractValidationRules(tagValue string, fieldKind reflect.Kind) ([]ValidationRule, error) {
	rulesList := strings.Split(tagValue, ValidationRulesSeparator)
	validationRules := make([]ValidationRule, 0, len(rulesList))
	for _, ruleStr := range rulesList {
		var err error
		var rule ValidationRule
		switch {
		case strings.HasPrefix(ruleStr, "len:"):
			rule, err = NewStringLenRule(ruleStr)
			validationRules = append(validationRules, rule)
		case strings.HasPrefix(ruleStr, "regexp:"):
			rule, err = NewStringRegexpRule(ruleStr)
			validationRules = append(validationRules, rule)
		case strings.HasPrefix(ruleStr, "min:"):
			rule, err = NewIntMinRule(ruleStr)
			validationRules = append(validationRules, rule)
		case strings.HasPrefix(ruleStr, "max:"):
			rule, err = NewIntMaxRule(ruleStr)
			validationRules = append(validationRules, rule)
		case strings.HasPrefix(ruleStr, "in:"):
			if fieldKind == reflect.Int {
				rule, err = NewIntInRule(ruleStr)
				validationRules = append(validationRules, rule)
			}
			if fieldKind == reflect.String {
				rule, err = NewStringInRule(ruleStr)
				validationRules = append(validationRules, rule)
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return validationRules, nil
}

func ValidateStruct(vi interface{}) error {
	var validationErrors ValidationErrors
	v := reflect.ValueOf(vi)
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("struct validation error: %w: expected a struct but received %T", ErrOnlyStructValidationAllowed, vi)
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldType := t.Field(i)
		if tagValue, ok := fieldType.Tag.Lookup(ValidationTag); ok && len(tagValue) > 0 && fieldValue.CanInterface() {
			rules, err := extractValidationRules(tagValue, fieldValue.Kind())
			if err != nil {
				return err
			}
			for _, rule := range rules {
				validationErrors = append(validationErrors, rule.Validate(fieldType.Name, fieldValue)...)
			}
		}
	}
	if len(validationErrors) == 0 {
		return nil
	}
	return validationErrors
}
