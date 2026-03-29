package httppath

import (
	"errors"
	"net/http"
	"reflect"
	"strconv"
)

func Decode[T any](r *http.Request, v T) error {
	val := reflect.ValueOf(v)
	t := reflect.TypeFor[T]()

	structValue := val.Elem()

	for i := range val.NumField() {
		field := structValue.Field(i)
		fieldType := t.Field(i)

		if !field.IsValid() || !field.CanSet() {
			return errors.New("unable to set")
		}

		tag := fieldType.Tag.Get("path")
		if tag == "" || tag == "-" {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(r.PathValue(tag))
		case reflect.Int64:
			value, err := strconv.ParseInt(r.PathValue(tag), 10, 64)
			if err != nil {
				return err
			}
			field.SetInt(value)
		}
	}

	return nil
}
