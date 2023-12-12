package gocqrs

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

type (
	MiddlewareFunc func(ctx context.Context, command any) (context.Context, any, bool)
	// AddMiddlewareBuilder is a struct used for building middlewares
	// for a specific command handler.
	AddMiddlewareBuilder struct {
		currentHandlerName string
		t                  requestType
		preMiddlewares     map[string][]middlewareStruct
		postMiddlewares    map[string][]middlewareStruct
	}
	// middlewareStruct represents a middleware with its name and the function itself.
	middlewareStruct struct {
		middlewareName string
		middlewareFunc MiddlewareFunc
	}
	// reflectiveHandler is a struct that allows the invocation of a method using reflection.
	// T1 and T2 are generic types for input and output respectively.
	reflectiveHandler[T1 T, T2 T] struct {
		method reflect.Value
	}
)

// executePreMiddlewares runs pre-middlewares for a given command and context.
// If any middleware returns false, the chain is stopped.
func (middlewareBuilder *AddMiddlewareBuilder) executePreMiddlewares(ctx context.Context, request T, handlerName string) T {
	if middlewares, ok := middlewareBuilder.preMiddlewares[handlerName]; ok {
		for _, m := range middlewares {
			var chain bool
			ctx, request, chain = m.middlewareFunc(ctx, request)
			if !chain {
				// Middleware has stopped the chain.
				return request
			}
		}
	}
	return request
}

// executePostMiddlewares runs post-middlewares for a given command and context.
// If any middleware returns false, the chain is stopped.
func (middlewareBuilder *AddMiddlewareBuilder) executePostMiddlewares(ctx context.Context, request T, handlerName string) T {
	if middlewares, ok := middlewareBuilder.postMiddlewares[handlerName]; ok {
		for _, m := range middlewares {
			var chain bool
			ctx, request, chain = m.middlewareFunc(ctx, request)
			if !chain {
				// Middleware has stopped the chain.
				return request
			}
		}
	}
	return request
}

// PreMiddleware adds a pre-middleware to the current handler.
func (middlewareBuilder *AddMiddlewareBuilder) PreMiddleware(m MiddlewareFunc) *AddMiddlewareBuilder {
	typedMiddlewareName := strings.TrimPrefix(runtime.FuncForPC(reflect.ValueOf(m).Pointer()).Name(), "*")

	middleware := middlewareStruct{
		middlewareName: typedMiddlewareName,
		middlewareFunc: m,
	}
	middlewares, ok := middlewareBuilder.preMiddlewares[middlewareBuilder.currentHandlerName]
	if !ok {
		middlewareBuilder.preMiddlewares[middlewareBuilder.currentHandlerName] = []middlewareStruct{
			middleware,
		}
		return middlewareBuilder
	}

	// Add the middleware to the handler if it's not already registered.
	if !isMiddlewareRegisteredForHandler(&middlewares, typedMiddlewareName) {
		middlewareBuilder.preMiddlewares[middlewareBuilder.currentHandlerName] =
			append(middlewareBuilder.preMiddlewares[middlewareBuilder.currentHandlerName], middleware)
	}
	return middlewareBuilder
}

// PostMiddleware adds a post-middleware to the current handler.
func (middlewareBuilder *AddMiddlewareBuilder) PostMiddleware(m MiddlewareFunc) *AddMiddlewareBuilder {
	typedMiddlewareName := strings.TrimPrefix(runtime.FuncForPC(reflect.ValueOf(m).Pointer()).Name(), "*")

	middleware := middlewareStruct{
		middlewareName: typedMiddlewareName,
		middlewareFunc: m,
	}
	middlewares, ok := middlewareBuilder.postMiddlewares[middlewareBuilder.currentHandlerName]
	if !ok {
		middlewareBuilder.postMiddlewares[middlewareBuilder.currentHandlerName] = []middlewareStruct{
			middleware,
		}
		return middlewareBuilder
	}

	// Add the middleware to the handler if it's not already registered.
	if !isMiddlewareRegisteredForHandler(&middlewares, typedMiddlewareName) {
		middlewareBuilder.postMiddlewares[middlewareBuilder.currentHandlerName] =
			append(middlewareBuilder.postMiddlewares[middlewareBuilder.currentHandlerName], middleware)
	}
	return middlewareBuilder
}

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
func (r reflectiveHandler[T1, T2]) Handle(ctx context.Context, in T1) (out T, err error) {
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
func createReflectiveHandler[T1 T](method reflect.Value) IHandler[T, T] {
	return reflectiveHandler[T, T1]{method: method}
}
