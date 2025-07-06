package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrNotStruct      = errors.New("input is not a struct")
	ErrInvalidTag     = errors.New("invalid validation tag")
	ErrInvalidRegexp  = errors.New("invalid regexp pattern")
	ErrStringLength   = errors.New("string length validation failed")
	ErrStringRegexp   = errors.New("string regexp validation failed")
	ErrStringNotInSet = errors.New("string not in allowed set")
	ErrNumberMin      = errors.New("number below minimum")
	ErrNumberMax      = errors.New("number above maximum")
	ErrNumberNotInSet = errors.New("number not in allowed set")
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var msgs []string
	for _, err := range v {
		msgs = append(msgs, fmt.Sprintf("%s: %s", err.Field, err.Err.Error()))
	}
	return strings.Join(msgs, "; ")
}

// isDeveloperError определяет, является ли ошибка ошибкой программиста.
func isDeveloperError(err error) bool {
	return errors.Is(err, ErrInvalidTag) || errors.Is(err, ErrInvalidRegexp)
}

// Validate проверяет структуру на соответствие правилам валидации.
func Validate(v interface{}) error {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	rt := reflect.TypeOf(v)
	var validationErrors ValidationErrors

	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		if !field.IsExported() {
			continue
		}

		validateTag := field.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		// Вложенные структуры
		if validateTag == "nested" {
			if fieldValue.Kind() == reflect.Struct {
				err := Validate(fieldValue.Interface())
				if err != nil {
					var valErrs ValidationErrors
					if errors.As(err, &valErrs) {
						for _, valErr := range valErrs {
							validationErrors = append(validationErrors, ValidationError{
								Field: field.Name + "." + valErr.Field,
								Err:   valErr.Err,
							})
						}
					} else {
						return err // Ошибка программиста
					}
				}
			}
			continue
		}

		err := validateField(field.Name, fieldValue, validateTag)
		if err != nil {
			var valErrs ValidationErrors
			if errors.As(err, &valErrs) {
				validationErrors = append(validationErrors, valErrs...)
			} else {
				return err // Ошибка программиста
			}
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}

	return nil
}

func validateField(fieldName string, fieldValue reflect.Value, tag string) error {
	var validationErrors ValidationErrors

	rules := strings.Split(tag, "|")

	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		var err error

		switch fieldValue.Kind() {
		case reflect.String:
			err = validateString(fieldValue.String(), rule)
			if err != nil {
				if isDeveloperError(err) {
					return err
				}
				validationErrors = append(validationErrors, ValidationError{
					Field: fieldName,
					Err:   err,
				})
			}
		case reflect.Int:
			err = validateInt(int(fieldValue.Int()), rule)
			if err != nil {
				if isDeveloperError(err) {
					return err
				}
				validationErrors = append(validationErrors, ValidationError{
					Field: fieldName,
					Err:   err,
				})
			}
		case reflect.Slice:
			err = validateSlice(fieldName, fieldValue, rule)
			if err != nil {
				if isDeveloperError(err) {
					return err
				}
				var valErrs ValidationErrors
				if errors.As(err, &valErrs) {
					validationErrors = append(validationErrors, valErrs...)
				} else {
					validationErrors = append(validationErrors, ValidationError{
						Field: fieldName,
						Err:   err,
					})
				}
			}
		default:
			return fmt.Errorf("%w: unsupported field type %s", ErrInvalidTag, fieldValue.Kind())
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}

	return nil
}

func validateString(value, rule string) error {
	parts := strings.SplitN(rule, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("%w: rule format should be 'type:value'", ErrInvalidTag)
	}

	ruleType, ruleValue := parts[0], parts[1]

	switch ruleType {
	case "len":
		expectedLen, err := strconv.Atoi(ruleValue)
		if err != nil {
			return fmt.Errorf("%w: invalid length value", ErrInvalidTag)
		}
		if len(value) != expectedLen {
			return fmt.Errorf("%w: expected length %d, got %d", ErrStringLength, expectedLen, len(value))
		}

	case "regexp":
		pattern := strings.ReplaceAll(ruleValue, "\\\\", "\\")
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidRegexp, err)
		}
		if !regex.MatchString(value) {
			return fmt.Errorf("%w: value '%s' doesn't match pattern '%s'", ErrStringRegexp, value, pattern)
		}

	case "in":
		allowedValues := strings.Split(ruleValue, ",")
		found := false
		for _, allowed := range allowedValues {
			if strings.TrimSpace(allowed) == value {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%w: value '%s' not in allowed set %v", ErrStringNotInSet, value, allowedValues)
		}

	default:
		return fmt.Errorf("%w: unknown string validation rule '%s'", ErrInvalidTag, ruleType)
	}

	return nil
}

func validateInt(value int, rule string) error {
	parts := strings.SplitN(rule, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("%w: rule format should be 'type:value'", ErrInvalidTag)
	}

	ruleType, ruleValue := parts[0], parts[1]

	switch ruleType {
	case "min":
		minVal, err := strconv.Atoi(ruleValue)
		if err != nil {
			return fmt.Errorf("%w: invalid min value", ErrInvalidTag)
		}
		if value < minVal {
			return fmt.Errorf("%w: value %d is less than minimum %d", ErrNumberMin, value, minVal)
		}

	case "max":
		maxVal, err := strconv.Atoi(ruleValue)
		if err != nil {
			return fmt.Errorf("%w: invalid max value", ErrInvalidTag)
		}
		if value > maxVal {
			return fmt.Errorf("%w: value %d is greater than maximum %d", ErrNumberMax, value, maxVal)
		}

	case "in":
		allowedStrs := strings.Split(ruleValue, ",")
		found := false
		for _, allowedStr := range allowedStrs {
			allowedVal, err := strconv.Atoi(strings.TrimSpace(allowedStr))
			if err != nil {
				return fmt.Errorf("%w: invalid number in 'in' rule", ErrInvalidTag)
			}
			if value == allowedVal {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%w: value %d not in allowed set %s", ErrNumberNotInSet, value, ruleValue)
		}

	default:
		return fmt.Errorf("%w: unknown int validation rule '%s'", ErrInvalidTag, ruleType)
	}

	return nil
}

func validateSlice(fieldName string, sliceValue reflect.Value, rule string) error {
	var validationErrors ValidationErrors

	for i := 0; i < sliceValue.Len(); i++ {
		elem := sliceValue.Index(i)
		var err error

		switch elem.Kind() {
		case reflect.String:
			err = validateString(elem.String(), rule)
		case reflect.Int:
			err = validateInt(int(elem.Int()), rule)
		default:
			return fmt.Errorf("%w: unsupported slice element type %s", ErrInvalidTag, elem.Kind())
		}

		if err != nil {
			if isDeveloperError(err) {
				return err
			}
			validationErrors = append(validationErrors, ValidationError{
				Field: fmt.Sprintf("%s[%d]", fieldName, i),
				Err:   err,
			})
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}

	return nil
}
