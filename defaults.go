package cog

import (
	"os"
	"reflect"
	"strconv"
)

type getValue func(reflect.StructField) string

var tagHandlers = []getValue{
	environmentVariable("env"),
	defaultValue("default"),
}

// SetDefaults fills zero-value fields in data using struct tags
// ("default" for literal values, "env" for environment variables).
func SetDefaults[T any](data *T) {
	setNested(reflect.ValueOf(data).Elem())
}

func environmentVariable(tag string) getValue {
	return func(sf reflect.StructField) string {
		if env := sf.Tag.Get(tag); env != "" {
			return os.Getenv(env)
		}
		return ""
	}
}

func defaultValue(tag string) getValue {
	return func(sf reflect.StructField) string {
		return sf.Tag.Get(tag)
	}
}

func setNested(v reflect.Value) {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.Struct {
			setNested(field)
		} else {
			setField(t.Field(i), field)
		}
	}
}

func setField(sf reflect.StructField, f reflect.Value) {
	for _, get := range tagHandlers {
		setValue(f, get(sf))
	}
}

func setValue(field reflect.Value, val string) {
	if val == "" || !isEmpty(field) || !field.CanSet() {
		return
	}

	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v, err := strconv.ParseInt(val, 10, 64); err == nil {
			field.SetInt(v)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v, err := strconv.ParseUint(val, 10, 64); err == nil {
			field.SetUint(v)
		}
	case reflect.Float32, reflect.Float64:
		if v, err := strconv.ParseFloat(val, 64); err == nil {
			field.SetFloat(v)
		}
	case reflect.String:
		field.SetString(val)
	case reflect.Bool:
		if v, err := strconv.ParseBool(val); err == nil {
			field.SetBool(v)
		}
	}
}

func isEmpty(v reflect.Value) bool {
	return !v.IsValid() || v.IsZero()
}
