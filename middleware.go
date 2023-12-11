package gocqrs

import (
	"context"
	"reflect"
	"runtime"
	"strings"
)

type (
	MiddlewareFunc func(ctx context.Context, command T) (chain bool)
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

// executePreMiddlewares runs the pre-middlewares chain with the given command and context
// It returns err if any error occurs.
func (middlewareBuilder *AddMiddlewareBuilder) executePreMiddlewares(ctx context.Context, command T, handlerName string) {
	if middlewares, ok := middlewareBuilder.preMiddlewares[handlerName]; ok {
		for _, m := range middlewares {
			chain := m.middlewareFunc(ctx, command)
			if !chain {
				// Middleware has stopped the chain.
				return
			}
		}
	}
	return
}

// executePreMiddlewares runs the pre-middlewares chain with the given command and context
// It returns err if any error occurs.
func (middlewareBuilder *AddMiddlewareBuilder) executePostMiddlewares(ctx context.Context, command T, handlerName string) {
	if middlewares, ok := middlewareBuilder.postMiddlewares[handlerName]; ok {
		for _, m := range middlewares {
			chain := m.middlewareFunc(ctx, command)
			if !chain {
				// Middleware has stopped the chain.
				return
			}
		}
	}
	return
}

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

func (middlewareBuilder *AddMiddlewareBuilder) PostMiddleware(m MiddlewareFunc) *AddMiddlewareBuilder {
	typedMiddlewareName := strings.TrimPrefix(runtime.FuncForPC(reflect.ValueOf(m).Pointer()).Name(), "*")

	middleware := middlewareStruct{
		middlewareName: typedMiddlewareName,
		middlewareFunc: m,
	}
	middlewares, ok := middlewareBuilder.preMiddlewares[middlewareBuilder.currentHandlerName]
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
	reflectResults := r.method.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(in)})
	result := reflectResults[0].Interface().(T2)
	if !reflectResults[1].IsNil() {
		err = reflectResults[1].Interface().(error)
	}

	return result, err
}

// createReflectiveHandler creates a new reflectiveHandler with the provided method.
func createReflectiveHandler[T1 T](method reflect.Value) IHandler[T, T] {
	return reflectiveHandler[T, T1]{method: method}
}
