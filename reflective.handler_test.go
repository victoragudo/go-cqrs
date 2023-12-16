package gocqrs

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockFunction is a sample function to be used with reflectiveHandler.
func mockFunction(ctx context.Context, input string) (string, error) {
	return "Processed: " + input, nil
}

type anotherMock struct {
	err error
}

// anotherMockFunction is a sample function to be used with reflectiveHandler.
func (a *anotherMock) anotherMockFunction(ctx context.Context, input string) (string, error) {
	return "Processed: " + input, a.err
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
	assert.Panics(t, func() {
		var handler reflectiveHandler[string, string]
		_, _ = handler.Handle(context.Background(), "test")
	})
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
	mockFunc := anotherMock{err: fmt.Errorf("error")}
	method := reflect.ValueOf(mockFunc.anotherMockFunction)
	handler := createReflectiveHandler[string](method)

	_, err := handler.Handle(context.Background(), "test")
	assert.Error(t, err, "Handle should return an error")
}
