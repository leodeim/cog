package defaults

import (
	"fmt"
	"reflect"
	"strconv"
)

const defaultTag = "default"

func Set[T any](data *T) error {
	v := reflect.ValueOf(data).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		if defaultVal := t.Field(i).Tag.Get(defaultTag); defaultVal != "" {
			if err := setField(v.Field(i), defaultVal); err != nil {
				return err
			}

		}
	}
	return nil
}

func setField(field reflect.Value, defaultVal string) error {

	if !field.CanSet() {
		return fmt.Errorf("can't set value")
	}

	if !isEmpty(field) {
		// field already set.
		return nil
	}

	switch field.Kind() {
	case reflect.Int:
		if val, err := strconv.Atoi(defaultVal); err == nil {
			field.Set(reflect.ValueOf(int(val)).Convert(field.Type()))
		}
	case reflect.String:
		field.Set(reflect.ValueOf(defaultVal).Convert(field.Type()))
	case reflect.Bool:
		if val, err := strconv.ParseBool(defaultVal); err == nil {
			field.Set(reflect.ValueOf(bool(val)).Convert(field.Type()))
		}
	}

	return nil
}

func isEmpty(v reflect.Value) bool {
	return !v.IsValid() || reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
