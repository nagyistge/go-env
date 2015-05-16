package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func populate(reflectValue reflect.Value, populateOptions PopulateOptions, recursive bool) error {
	restrictTo, err := getRestrictTo(populateOptions.RestrictTo)
	if err != nil {
		return err
	}
	decoderEnv, err := readDecoders(populateOptions.Decoders)
	if err != nil {
		return err
	}
	if reflectValue.Type().Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	} else if !recursive {
		return fmt.Errorf("expected pointer, got %v", reflectValue)
	}
	if reflectValue.Type().Kind() != reflect.Struct {
		return fmt.Errorf("expected struct pointer, got %v", reflectValue)
	}
	numField := reflectValue.NumField()
	for i := 0; i < numField; i++ {
		structField := reflectValue.Type().Field(i)
		envTag, err := getEnvTag(structField, restrictTo)
		if err != nil {
			return err
		}
		value, ok := decoderEnv[envTag.key]
		if !ok {
			value = os.Getenv(envTag.key)
			if value == "" {
				if envTag.required {
					return fmt.Errorf("%s is not set and is required for %v", envTag.key, reflectValue)
				}
				continue
			}
		}
		switch structField.Type.Kind() {
		case reflect.Struct:
			if value == envTag.value {
				if err := populate(reflectValue.Field(i), populateOptions, true); err != nil {
					return err
				}
			}
		default:
			parsedValue, err := parseField(structField, value)
			if err != nil {
				return err
			}
			reflectValue.Field(i).Set(reflect.ValueOf(parsedValue))
		}
	}
	return nil
}

func getRestrictTo(restrictTo []string) (map[string]bool, error) {
	if restrictTo == nil || len(restrictTo) == 0 {
		return nil, nil
	}
	restrictToMap := make(map[string]bool)
	for _, envKey := range restrictTo {
		if _, ok := restrictToMap[envKey]; ok {
			return nil, fmt.Errorf("duplicate restrict to env key %s", envKey)
		}
		restrictToMap[envKey] = true
	}
	return restrictToMap, nil
}

func readDecoders(decoders []Decoder) (map[string]string, error) {
	env := make(map[string]string)
	if decoders == nil || len(decoders) == 0 {
		return env, nil
	}
	for _, decoder := range decoders {
		subEnv, err := decoder.Decode()
		if err != nil {
			return nil, err
		}
		for key, value := range subEnv {
			env[key] = value
		}
	}
	return env, nil
}

type envTag struct {
	key      string
	required bool
	value    string
}

func getEnvTag(structField reflect.StructField, restrictTo map[string]bool) (*envTag, error) {
	tag := structField.Tag.Get("env")
	if tag == "" {
		return nil, fmt.Errorf("must have `env` tag on every field")
	}
	split := strings.Split(tag, ",")
	if len(split) != 2 && len(split) != 3 {
		return nil, fmt.Errorf("invalid tag %s, must be key,{required,optional} or key,{required,optional},value", tag)
	}
	key := split[0]
	if restrictTo != nil {
		if _, ok := restrictTo[key]; !ok {
			return nil, fmt.Errorf("invalid tag %s, key must be one of %v", tag, restrictTo)
		}
	}
	required := false
	switch split[1] {
	case "required":
		required = true
	case "optional":
		required = false
	default:
		return nil, fmt.Errorf("invalid tag %s, second field must be one of required,optionl", tag)
	}
	value := ""
	switch structField.Type.Kind() {
	case reflect.Struct:
		if len(split) == 3 {
			return nil, fmt.Errorf("invalid tag %s, must have value for struct fields", tag)
		}
		value = split[2]
	default:
		if len(split) == 3 {
			return nil, fmt.Errorf("invalid tag %s, can only have value for struct fields", tag)
		}
	}
	return &envTag{
		key:      key,
		required: required,
		value:    value,
	}, nil
}

func parseField(structField reflect.StructField, value string) (interface{}, error) {
	fieldKind := structField.Type.Kind()
	switch fieldKind {
	case reflect.Bool:
		return value != "" && value != "false", nil
	case reflect.Int:
		parsedValue, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			return nil, err
		}
		return int(parsedValue), nil
	case reflect.Int8:
		parsedValue, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return nil, err
		}
		return int8(parsedValue), nil
	case reflect.Int16:
		parsedValue, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return nil, err
		}
		return int16(parsedValue), nil
	case reflect.Int32:
		parsedValue, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, err
		}
		return int32(parsedValue), nil

	case reflect.Int64:
		parsedValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		return int64(parsedValue), nil
	case reflect.Uint:
		parsedValue, err := strconv.ParseUint(value, 10, 0)
		if err != nil {
			return nil, err
		}
		return uint(parsedValue), nil
	case reflect.Uint8:
		parsedValue, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return nil, err
		}
		return uint8(parsedValue), nil
	case reflect.Uint16:
		parsedValue, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return nil, err
		}
		return uint16(parsedValue), nil
	case reflect.Uint32:
		parsedValue, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return nil, err
		}
		return uint32(parsedValue), nil

	case reflect.Uint64:
		parsedValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, err
		}
		return uint64(parsedValue), nil
	case reflect.Float32:
		parsedValue, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return nil, err
		}
		return float32(parsedValue), nil
	case reflect.Float64:
		parsedValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		return float64(parsedValue), nil
	case reflect.String:
		return value, nil
	default:
		return nil, fmt.Errorf("field type %v not allowed", fieldKind)
	}
}
