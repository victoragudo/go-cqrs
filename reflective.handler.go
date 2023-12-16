package gocqrs

import (
	"context"
	"reflect"
)

// reflectiveHandler is a struct that allows the invocation of a method using reflection.
// It is generic and can handle methods with different input (T1) and output (T2) types.
// This structure is useful for creating flexible and dynamic handler functions.
type reflectiveHandler[T1 T, T2 T] struct {
	method reflect.Value // The method to be invoked, stored as a reflect.Value.
}

// Handle executes the method associated with the reflectiveHandler,
// passing in the context and input, and returns the result and any error.
func (r reflectiveHandler[T1, T2]) Handle(ctx context.Context, in T1) (out T2, err error) {

	// Check if the method is properly initialized
	if !r.method.IsValid() {
		panic("Handle method not initialized")
	}

	// Check if the context and input are properly initialized
	ctxVal := reflect.ValueOf(ctx)
	inVal := reflect.ValueOf(in)

	if !ctxVal.IsValid() || !inVal.IsValid() {
		return out, err
	}

	// Perform the reflective call
	reflectResults := r.method.Call([]reflect.Value{ctxVal, inVal})

	// Handle the results of the reflective call
	if len(reflectResults) > 0 {
		result, ok := reflectResults[0].Interface().(T2)
		if ok {
			out = result
		}
	}

	// Handle potential errors returned by the reflective call
	if len(reflectResults) > 1 {
		errVal := reflectResults[1].Interface()
		if errVal != nil {
			var ok bool
			retErr, ok := errVal.(error)
			if ok {
				err = retErr
			}
		}
	}

	return out, err
}

// createReflectiveHandler creates a new reflectiveHandler with the provided method.
func createReflectiveHandler[TResponse T](method reflect.Value) IHandler[T, TResponse] {
	return reflectiveHandler[T, TResponse]{method: method}
}
