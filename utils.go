package gocqrs

import (
	"reflect"
)

// getField obtiene un campo por nombre de un struct y devuelve su valor como interfaz.
func getField(object T, fieldName string) (reflect.Value, bool) {
	val := reflect.ValueOf(object)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		fieldVal := val.FieldByName(fieldName)
		if fieldVal.IsValid() {
			return fieldVal, true
		}
	}
	return reflect.Value{}, false
}

func getMethodByName(value reflect.Value, methodName string) (reflect.Value, bool) {
	method := value.MethodByName(methodName)
	if !method.IsValid() {
		return reflect.Value{}, false
	}
	return method, true
}

// storeMapValue stores a value with a string key in the given map.
func storeMapValue(m map[string]any, key string, value any) {
	m[key] = value
}

// getMapValue retrieves a value by key from the given map.
func getMapValue(m map[string]any, key string) (any, bool) {
	v, ok := m[key]
	return v, ok
}
