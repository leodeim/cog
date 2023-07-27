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
		if val := sf.Tag.Get(tag); val != "" {
			return val
		}

		return ""
	}
}

func setNested(v reflect.Value) {
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Kind() == reflect.Struct {
			setNested(v.Field(i))
		} else {
			t := v.Type()
			for i := 0; i < t.NumField(); i++ {
				setField(t.Field(i), v.Field(i))
			}
		}
	}
}

func setField(sf reflect.StructField, f reflect.Value) {
	for _, getValue := range tagHandlers {
		setValue(f, getValue(sf))
	}
}

func setValue(field reflect.Value, val string) {
	if val == "" || !isEmpty(field) || !field.CanSet() {
		return
	}

	switch field.Kind() {
	case reflect.Int:
		if val, err := strconv.Atoi(val); err == nil {
			field.Set(reflect.ValueOf(int(val)).Convert(field.Type()))
		}
	case reflect.String:
		field.Set(reflect.ValueOf(val).Convert(field.Type()))
	case reflect.Bool:
		if val, err := strconv.ParseBool(val); err == nil {
			field.Set(reflect.ValueOf(bool(val)).Convert(field.Type()))
		}
	}
}

func isEmpty(v reflect.Value) bool {
	return !v.IsValid() || reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
