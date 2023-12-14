package gocqrs

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockFunction is a sample function to be used with reflectiveHandler.
func mockFunction(ctx context.Context, input string) (string, error) {
	return "Processed: " + input, nil
}

// TestHandleValidInput tests the Handle method with valid input.
func TestHandleValidInput(t *testing.T) {
	method := reflect.ValueOf(mockFunction)
	handler := createReflectiveHandler[string](method)

	result, err := handler.Handle(context.Background(), "test")
	assert.NoError(t, err, "Handle should not return an error")
	assert.Equal(t, "Processed: test", result, "Result should match the expected output")
}

// TestHandleInvalidMethod tests the Handle method with an uninitialized method.
func TestHandleInvalidMethod(t *testing.T) {
	var handler reflectiveHandler[string, string]

	_, err := handler.Handle(context.Background(), "test")
	assert.Error(t, err, "Handle should return an error for uninitialized method")
}

// TestHandleValidContext tests the Handle method with and valid context.
func TestHandleValidContext(t *testing.T) {
	method := reflect.ValueOf(mockFunction)
	handler := createReflectiveHandler[string](method)

	_, err := handler.Handle(context.TODO(), "test")
	assert.Nil(t, err, "Handle should return a nil error")
}

// TestHandleInvalidInput tests the Handle method with an invalid input.
func TestHandleInvalidInput(t *testing.T) {
	method := reflect.ValueOf(mockFunction)
	handler := createReflectiveHandler[string](method)

	_, err := handler.Handle(context.Background(), nil)
	assert.Error(t, err, "Handle should return an error for invalid input")
}
