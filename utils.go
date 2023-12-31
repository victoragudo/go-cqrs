package gocqrs

import (
	"reflect"
	"sync"
)

// getField extracts a field's value from a struct (or pointer to a struct) based on the field name.
// It returns the field value and a boolean indicating if the field was found.
func getField(object T, fieldName string) (reflect.Value, bool) {
	val := reflect.ValueOf(object)

	// If the object is a pointer, dereference it to get the actual value.
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Check if the value is a struct, then try to get the field by name.
	if val.Kind() == reflect.Struct {
		fieldVal := val.FieldByName(fieldName)
		if fieldVal.IsValid() {
			return fieldVal, true // Return the field value if it is valid.
		}
	}
	// Return a zero reflect.Value and false if the field is not found.
	return reflect.Value{}, false
}

// getMethodByName retrieves a method by its name with reflection.
// It returns the method as reflect.Value and a boolean indicating if the method was found.
func getMethodByName(value reflect.Value, methodName string) (reflect.Value, bool) {
	method := value.MethodByName(methodName)
	if !method.IsValid() {
		return reflect.Value{}, false // Return a zero reflect.Value and false if the method is not found.
	}
	return method, true // Return the method if it is valid.
}

// storeMapValue stores a value with a string key in the given map.
func storeMapValue(m map[string]any, key string, value any, mutex *sync.RWMutex) {
	mutex.Lock()
	defer mutex.Unlock()
	m[key] = value
}

// getMapValue retrieves a value by key from the given map.
// It returns the value and a boolean indicating if the key was found in the map.
func getMapValue(m map[string]any, key string, mutex *sync.RWMutex) (any, bool) {
	mutex.RLock()
	defer mutex.RUnlock()
	v, ok := m[key]
	return v, ok // Return the value and a boolean indicating if the key was found.
}

// checkTypeNameInEventHandlers checks if the string is in the slice of structs.
// It returns true if the string is found, false otherwise.
func checkTypeNameInEventHandlers(typeName string, eventHandlers []eventHandlersType) bool {
	for _, v := range eventHandlers {
		if v.typeName == typeName {
			return true
		}
	}
	return false
}

func loadOrStoreEventHandlers(m map[string][]eventHandlersType, typedEvent string, eventHandlerMutex *sync.RWMutex) []eventHandlersType {
	eventHandlerMutex.Lock()
	defer eventHandlerMutex.Unlock()

	if handlers, exists := m[typedEvent]; exists {
		return handlers
	}

	m[typedEvent] = make([]eventHandlersType, 0)
	return m[typedEvent]
}
