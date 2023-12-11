package gocqrs

import (
	"context"
	"reflect"
	"runtime"
	"strings"
)

type (
	// AddMiddlewareBuilder is a struct used for building middlewares
	// for a specific command handler.
	AddMiddlewareBuilder struct {
		currentHandlerName string
		commandMiddlewares map[string][]handlerMiddleware
		queryMiddlewares   map[string][]handlerMiddleware
		requestType        requestType
	}
	// handlerMiddleware represents a middleware with its name and the function itself.
	handlerMiddleware struct {
		middlewareName string
		middleware     func(next IHandler[T, T]) IHandler[T, T]
	}
	// reflectiveHandler is a struct that allows the invocation of a method using reflection.
	// T1 and T2 are generic types for input and output respectively.
	reflectiveHandler[T1 T, T2 T] struct {
		method reflect.Value
	}
)

// AddMiddleware adds a middleware to the current handler.
// It extracts the middlewares name using reflection and runtime information.
func (middlewareBuilder *AddMiddlewareBuilder) AddMiddleware(middleware func(next IHandler[T, T]) IHandler[T, T]) *AddMiddlewareBuilder {
	if middlewareBuilder.requestType == commandType {
		return middlewareBuilder.addCommandMiddleware(middleware)
	}
	return middlewareBuilder.addQueryMiddleware(middleware)
}

func (middlewareBuilder *AddMiddlewareBuilder) addCommandMiddleware(middleware func(next IHandler[T, T]) IHandler[T, T]) *AddMiddlewareBuilder {
	typedMiddlewareName := strings.TrimPrefix(runtime.FuncForPC(reflect.ValueOf(middleware).Pointer()).Name(), "*")

	var middlewares []handlerMiddleware
	var ok bool

	middlewares, ok = middlewareBuilder.commandMiddlewares[middlewareBuilder.currentHandlerName]

	hMiddleware := handlerMiddleware{
		middlewareName: typedMiddlewareName,
		middleware:     middleware,
	}

	// If the handler does not have any middleware yet, initialize its slice.
	if !ok {
		middlewareBuilder.commandMiddlewares[middlewareBuilder.currentHandlerName] = []handlerMiddleware{hMiddleware}
		return middlewareBuilder
	}

	// Add the middleware to the handler if it's not already registered.
	if !isMiddlewareRegisteredForHandler(&middlewares, typedMiddlewareName) {
		middlewareBuilder.commandMiddlewares[middlewareBuilder.currentHandlerName] =
			append(middlewareBuilder.commandMiddlewares[middlewareBuilder.currentHandlerName], hMiddleware)
	}
	return middlewareBuilder
}

func (middlewareBuilder *AddMiddlewareBuilder) addQueryMiddleware(middleware func(next IHandler[T, T]) IHandler[T, T]) *AddMiddlewareBuilder {
	typedMiddlewareName := strings.TrimPrefix(runtime.FuncForPC(reflect.ValueOf(middleware).Pointer()).Name(), "*")

	var middlewares []handlerMiddleware
	var ok bool

	middlewares, ok = middlewareBuilder.queryMiddlewares[middlewareBuilder.currentHandlerName]

	hMiddleware := handlerMiddleware{
		middlewareName: typedMiddlewareName,
		middleware:     middleware,
	}

	// If the handler does not have any middleware yet, initialize its slice.
	if !ok {
		middlewareBuilder.queryMiddlewares[middlewareBuilder.currentHandlerName] = []handlerMiddleware{hMiddleware}
		return middlewareBuilder
	}

	// Add the middleware to the handler if it's not already registered.
	if !isMiddlewareRegisteredForHandler(&middlewares, typedMiddlewareName) {
		middlewareBuilder.queryMiddlewares[middlewareBuilder.currentHandlerName] =
			append(middlewareBuilder.queryMiddlewares[middlewareBuilder.currentHandlerName], hMiddleware)
	}
	return middlewareBuilder
}

// isMiddlewareRegisteredForHandler checks if a middleware is already registered for a handler.
func isMiddlewareRegisteredForHandler(middlewares *[]handlerMiddleware, middlewareName string) bool {
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
