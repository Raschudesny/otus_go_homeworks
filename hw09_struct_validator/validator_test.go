package hw09structvalidator

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int      `validate:"min:18|max:50"`
		Email  string   `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole `validate:"in:admin,stuff"`
		Phones []string `validate:"len:11"`
		meta   json.RawMessage
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

func TestValidateStruct(t *testing.T) {
	for _, testData := range [...]struct {
		name           string
		input          interface{}
		expectedErrors []error
	}{
		{
			"str length rule all valid",
			App{"12345"},
			nil,
		},
		{
			"str length rule field invalid",
			App{"123456"},
			[]error{ErrStrLengthRuleIsInvalid},
		},
		{
			"int in rule with multiple values all valid",
			Response{200, "123"},
			nil,
		},
		{
			"int in rule with multiple values not valid",
			Response{123123123213, "yo"},
			[]error{ErrIntInRuleIsInvalid},
		},
		{
			"no validation rules provided",
			Token{
				[]byte{1, 2, 3, 4, 5},
				[]byte{1, 2, 3, 4, 5},
				[]byte{1, 2, 3, 4, 5},
			},
			nil,
		},
		{
			"str length and str in rule is invalid at the same time",
			User{
				"123",
				"name",
				20,
				"kek@mail.ru",
				"admine",
				[]string{"89261232323", "89263212121", "89261242424"},
				[]byte{1, 2, 3, 4, 5},
			},
			[]error{ErrStrLengthRuleIsInvalid, ErrStrInRuleIsInvalid},
		},
		{
			"len and max rule not valid",
			User{
				"123",
				"name",
				600,
				"kek@mail.ru",
				"admin",
				[]string{"89261232323", "89263212121", "89261242424"},
				[]byte{1, 2, 3, 4, 5},
			},
			[]error{ErrStrLengthRuleIsInvalid, ErrIntMaxRuleIsInvalid},
		},
	} {
		t.Run(testData.name, func(t *testing.T) {
			testData := testData
			t.Parallel()
			var validationErrors ValidationErrors
			errs := ValidateStruct(testData.input)
			if len(testData.expectedErrors) == 0 {
				require.NoError(t, errs)
				return
			}
			require.ErrorAs(t, errs, &validationErrors)
			for i, e := range validationErrors {
				require.ErrorIs(t, e.Err, testData.expectedErrors[i])
			}
		})
	}
}
