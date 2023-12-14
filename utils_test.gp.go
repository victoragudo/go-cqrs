package gocqrs

import (
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetField tests the getField function.
func TestGetField(t *testing.T) {
	type TestStruct struct {
		Field string
	}
	obj := TestStruct{Field: "value"}

	fieldValue, found := getField(obj, "Field")
	assert.True(t, found, "Field should be found")
	assert.Equal(t, "value", fieldValue.String(), "Field value should match")

	_, notFound := getField(obj, "NonExistentField")
	assert.False(t, notFound, "Non-existent field should not be found")
}

// TestgetMethodByName tests the getMethodByName function.
func TestgetMethodByName(t *testing.T) {
	obj := reflect.ValueOf(&sync.Mutex{})
	method, found := getMethodByName(obj, "Lock")
	assert.True(t, found, "Method should be found")
	assert.NotNil(t, method, "Method should not be nil")

	_, notFound := getMethodByName(obj, "NonExistentMethod")
	assert.False(t, notFound, "Non-existent method should not be found")
}

// TestStoreAndGetMapValue tests the storeMapValue and getMapValue functions.
func TestStoreAndGetMapValue(t *testing.T) {
	var mutex sync.RWMutex
	m := make(map[string]any)

	storeMapValue(m, "key", "value", &mutex)
	value, found := getMapValue(m, "key", &mutex)
	assert.True(t, found, "Value should be found")
	assert.Equal(t, "value", value, "Stored value should match")

	_, notFound := getMapValue(m, "NonExistentKey", &mutex)
	assert.False(t, notFound, "Non-existent key should not be found")
}

// TestCheckTypeNameInEventHandlers tests the checkTypeNameInEventHandlers function.
func TestCheckTypeNameInEventHandlers(t *testing.T) {
	handlers := []eventHandlersType{
		{typeName: "Handler1"},
		{typeName: "Handler2"},
	}

	found := checkTypeNameInEventHandlers("Handler1", handlers)
	assert.True(t, found, "Handler1 should be found")

	notFound := checkTypeNameInEventHandlers("NonExistentHandler", handlers)
	assert.False(t, notFound, "Non-existent handler should not be found")
}

// TestLoadOrStoreEventHandlers tests the loadOrStoreEventHandlers function.
func TestLoadOrStoreEventHandlers(t *testing.T) {
	var mutex sync.RWMutex
	m := make(map[string][]eventHandlersType)

	handlers := loadOrStoreEventHandlers(m, "Event1", &mutex)
	assert.Empty(t, handlers, "Handlers should be empty for a new event")

	existingHandlers := []eventHandlersType{{typeName: "ExistingHandler"}}
	m["Event1"] = existingHandlers

	handlers = loadOrStoreEventHandlers(m, "Event1", &mutex)
	assert.Equal(t, existingHandlers, handlers, "Should return existing handlers for an existing event")
}
