package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"testing"
)

type UserRole string

type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
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

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		in          interface{}
		expectedErr error
	}{
		{
			name: "valid user",
			in: User{
				ID:     "123e4567-e89b-12d3-a456-426614174000",
				Age:    25,
				Email:  "test@example.com",
				Role:   "admin",
				Phones: []string{"12345678901"},
			},
			expectedErr: nil,
		},
		{
			name: "invalid user - age too low",
			in: User{
				ID:     "123e4567-e89b-12d3-a456-426614174000",
				Age:    15,
				Email:  "test@example.com",
				Role:   "admin",
				Phones: []string{"12345678901"},
			},
			expectedErr: ValidationErrors{},
		},
		{
			name:        "not a struct",
			in:          "not a struct",
			expectedErr: ErrNotStruct,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.in)

			if tt.expectedErr == nil {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				return
			}

			if errors.Is(tt.expectedErr, ErrNotStruct) {
				if !errors.Is(err, ErrNotStruct) {
					t.Errorf("expected ErrNotStruct, got %v", err)
				}
				return
			}

			var valErrs ValidationErrors
			if !errors.As(err, &valErrs) {
				t.Errorf("expected ValidationErrors, got %v", err)
			}
		})
	}
}
func TestValidate_DeveloperErrors(t *testing.T) {
	type BadStruct struct {
		Data map[string]string `validate:"len:5"`
	}
	err := Validate(BadStruct{Data: map[string]string{"a": "b"}})
	if !errors.Is(err, ErrInvalidTag) {
		t.Errorf("expected ErrInvalidTag for map, got %v", err)
	}

	type BadTag struct {
		Name string `validate:"unknown:5"`
	}
	err = Validate(BadTag{Name: "test"})
	if !errors.Is(err, ErrInvalidTag) {
		t.Errorf("expected ErrInvalidTag for unknown rule, got %v", err)
	}

	type BadRegexp struct {
		Email string `validate:"regexp:([)"`
	}
	err = Validate(BadRegexp{Email: "test"})
	if !errors.Is(err, ErrInvalidRegexp) {
		t.Errorf("expected ErrInvalidRegexp, got %v", err)
	}
}
