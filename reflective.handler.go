package gocqrs

import (
	"context"
	"fmt"
	"reflect"
)

// isMiddlewareRegisteredForHandler checks if a middleware is already registered for a handler.
func isMiddlewareRegisteredForHandler(middlewares *[]middlewareStruct, middlewareName string) bool {
	for _, middleware := range *middlewares {
		if middleware.middlewareName == middlewareName {
			return true
		}
	}
	return false
}

// Handle executes the method associated with the reflectiveHandler,
// passing in the context and input, and returns the result and any error.
func (r reflectiveHandler[T1, T2]) Handle(ctx context.Context, in T1) (out T2, err error) {

	// Check if the method is properly initialized
	if !r.method.IsValid() {
		return out, fmt.Errorf("reflectiveHandler: method not initialized")
	}

	// Check if the context and input are properly initialized
	ctxVal := reflect.ValueOf(ctx)
	inVal := reflect.ValueOf(in)

	if !ctxVal.IsValid() {
		return out, fmt.Errorf("reflectiveHandler: invalid or nil context")
	}

	if !inVal.IsValid() {
		return out, fmt.Errorf("reflectiveHandler: invalid or nil request")
	}

	// Perform the reflective call
	reflectResults := r.method.Call([]reflect.Value{ctxVal, inVal})

	// Handle the results of the reflective call
	if len(reflectResults) > 0 {
		result, ok := reflectResults[0].Interface().(T2)
		if ok {
			out = result
		} else {
			return out, fmt.Errorf("reflectiveHandler: error in result conversion")
		}
	}

	// Handle potential errors returned by the reflective call
	if len(reflectResults) > 1 {
		errVal := reflectResults[1].Interface()
		if errVal != nil {
			var ok bool
			err, ok = errVal.(error)
			if !ok {
				return out, fmt.Errorf("reflectiveHandler: error type assertion failed")
			}
		}
	}

	return out, err
}

// createReflectiveHandler creates a new reflectiveHandler with the provided method.
func createReflectiveHandler[TResponse T](method reflect.Value) IHandler[T, TResponse] {
	return reflectiveHandler[T, TResponse]{method: method}
}
