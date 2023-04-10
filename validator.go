package go_course_for_students_hw06_validator

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var ErrNotStruct = errors.New("wrong argument given, should be a struct")
var ErrInvalidValidatorSyntax = errors.New("invalid validator syntax")
var ErrValidateForUnexportedFields = errors.New("validation for unexported field is not allowed")

type ValidationError struct {
	Err error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if v == nil {
		return ""
	}
	var errStrings []string
	for _, err := range v {
		errStrings = append(errStrings, err.Err.Error())
	}
	return strings.Join(errStrings, "\n")
}

func Validate(v any) error {
	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	var errs ValidationErrors
	for i := 0; i < value.NumField(); i++ {
		fieldValue := value.Field(i)
		fieldType := value.Type().Field(i)
		tagValue := fieldType.Tag.Get("validate")
		if tagValue == "" {
			continue // skip fields without `validate` tag
		}

		if !fieldType.IsExported() {
			errs = append(errs, ValidationError{ErrValidateForUnexportedFields})
			break
		}

		var fieldErr error
		switch fieldType.Type.Kind() {
		case reflect.String:
			fieldErr = validateFieldString(fieldValue.String(), tagValue)
		case reflect.Int:
			fieldErr = validateFieldInt(fieldValue.Int(), tagValue)
		case reflect.Slice:
			fieldErr = validateFieldSlice(fieldValue, tagValue)
		default:
			panic(fmt.Sprintf("unsupported field type %s", fieldType.Type.Kind()))
		}
		if fieldErr != nil {
			errs = append(errs, ValidationError{Err: fieldErr})
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func validateFieldString(fieldValue string, tagValue string) error {
	for _, v := range strings.Split(tagValue, ",") {
		parts := strings.Split(v, ":")
		validator := parts[0]
		switch validator {
		case "len":
			expectedLen, err := strconv.Atoi(parts[1])
			if err != nil {
				return ErrInvalidValidatorSyntax
			}
			if len(fieldValue) != expectedLen {
				errMsg := fmt.Sprintf("expected string of length %d, go string of length %d", expectedLen, len(fieldValue))
				return errors.New(errMsg)
			}
		case "min":
			expectedMin, err := strconv.Atoi(parts[1])
			if err != nil {
				return ErrInvalidValidatorSyntax
			}
			if len(fieldValue) < expectedMin {
				errMsg := fmt.Sprintf("string of length %d less than expected min %d", len(fieldValue), expectedMin)
				return errors.New(errMsg)
			}
		case "max":
			expectedMax, err := strconv.Atoi(parts[1])
			if err != nil {
				return ErrInvalidValidatorSyntax
			}
			if len(fieldValue) > expectedMax {
				errMsg := fmt.Sprintf("string of length %d greater than expected max %d", len(fieldValue), expectedMax)
				return errors.New(errMsg)
			}
		case "in":
			var inValues []string
			if len(parts[1]) == 0 {
				inValues = []string{} // empty set
			} else {
				inValues = strings.Split(parts[1], ",")
			}
			for _, v := range inValues {
				if v == fieldValue {
					return nil
				}
			}
			errMsg := fmt.Sprintf("value %s not found in %v", fieldValue, inValues)
			return errors.New(errMsg)
		default:
			panic(fmt.Sprintf("unsupported validator type %s", validator))
		}
	}
	return nil
}

func validateFieldInt(fieldValue int64, tagValue string) error {
	parts := strings.Split(tagValue, ":")
	validator := parts[0]
	switch validator {
	case "min":
		expectedMin, err := strconv.Atoi(parts[1])
		if err != nil {
			return ErrInvalidValidatorSyntax
		}
		if fieldValue < int64(expectedMin) {
			errMsg := fmt.Sprintf("field value %d less than expected min %d", fieldValue, expectedMin)
			return errors.New(errMsg)
		}
	case "max":
		expectedMax, err := strconv.Atoi(parts[1])
		if err != nil {
			return ErrInvalidValidatorSyntax
		}
		if fieldValue > int64(expectedMax) {
			errMsg := fmt.Sprintf("field value %d greater than expected max %d", fieldValue, expectedMax)
			return errors.New(errMsg)
		}
	case "in":
		inValues := strings.Split(parts[1], ",")
		for _, v := range inValues {
			vInt, err := strconv.Atoi(v)
			if err != nil {
				return ErrInvalidValidatorSyntax
			}
			if int64(vInt) == fieldValue {
				return nil
			}
		}
		errMsg := fmt.Sprintf("value %d not found in %v", fieldValue, inValues)
		return errors.New(errMsg)
	default:
		panic(fmt.Sprintf("unsupported validator type %s", validator))
	}
	return nil
}

func validateFieldSlice(fieldValue reflect.Value, tagValue string) error {
	for i := 0; i < fieldValue.Len(); i++ {
		sliceElem := fieldValue.Index(i)
		var err error
		switch sliceElem.Kind() {
		case reflect.String:
			err = validateFieldString(sliceElem.String(), tagValue)
		case reflect.Int:
			err = validateFieldInt(sliceElem.Int(), tagValue)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
