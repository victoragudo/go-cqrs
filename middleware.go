package gocqrs

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

type (
	// MiddlewareFunc defines a function type used for middleware.
	// It takes a context and a request (of any type), and returns a potentially modified context,
	// a result (of any type), and a boolean to indicate whether to proceed with further processing.
	MiddlewareFunc func(ctx context.Context, request any) (context.Context, any, bool)

	// AddMiddlewareBuilder is a struct used for building middleware chains
	// for a specific command/query/event handler. It stores the name of the current handler
	// and maps of pre- and post-middlewares associated with that handler.
	AddMiddlewareBuilder struct {
		currentHandlerName string                        // Name of the handler for which middlewares are being added.
		preMiddlewares     map[string][]middlewareStruct // Map of pre-middlewares for each handler.
		postMiddlewares    map[string][]middlewareStruct // Map of post-middlewares for each handler.
	}

	// middlewareStruct represents a middleware with its name and the function itself.
	// It is used to store individual middleware functions along with their names.
	middlewareStruct struct {
		middlewareName string         // Name of the middleware.
		middlewareFunc MiddlewareFunc // The middleware function.
	}

	// reflectiveHandler is a struct that allows the invocation of a method using reflection.
	// It is generic and can handle methods with different input (T1) and output (T2) types.
	// This structure is useful for creating flexible and dynamic handler functions.
	reflectiveHandler[T1 T, T2 T] struct {
		method reflect.Value // The method to be invoked, stored as a reflect.Value.
	}
)

// executePreMiddlewares runs pre-middlewares for a given request and context.
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

// executePostMiddlewares runs post-middlewares for a given request and context.
// If any middleware returns false, the chain is stopped.
func (middlewareBuilder *AddMiddlewareBuilder) executePostMiddlewares(ctx context.Context, request T, handlerName string) {
	if middlewares, ok := middlewareBuilder.postMiddlewares[handlerName]; ok {
		for _, m := range middlewares {
			var chain bool
			ctx, request, chain = m.middlewareFunc(ctx, request)
			if !chain {
				// Middleware has stopped the chain.
				return
			}
		}
	}
}

// PreMiddleware adds a pre-middleware to the current handler.
func (middlewareBuilder *AddMiddlewareBuilder) PreMiddleware(m MiddlewareFunc) *AddMiddlewareBuilder {

	// Extract the name of the middleware function using reflection and strip the pointer indicator.
	typedMiddlewareName := strings.TrimPrefix(runtime.FuncForPC(reflect.ValueOf(m).Pointer()).Name(), "*")

	// Create a middlewareStruct instance with the middleware name and function.
	middleware := middlewareStruct{
		middlewareName: typedMiddlewareName,
		middlewareFunc: m,
	}

	// Retrieve the slice of pre-middlewares associated with the current handler from the middlewareBuilder.
	middlewares, ok := middlewareBuilder.preMiddlewares[middlewareBuilder.currentHandlerName]

	// If the current handler does not have any pre-middlewares, initialize it with the new middleware.
	if !ok {
		middlewareBuilder.preMiddlewares[middlewareBuilder.currentHandlerName] = []middlewareStruct{
			middleware,
		}
		return middlewareBuilder
	}

	// Add the middleware to the handler if it's not already registered.
	// This is a check to avoid registering the same middleware multiple times for a handler.
	if !isMiddlewareRegisteredForHandler(&middlewares, typedMiddlewareName) {
		middlewareBuilder.preMiddlewares[middlewareBuilder.currentHandlerName] =
			append(middlewareBuilder.preMiddlewares[middlewareBuilder.currentHandlerName], middleware)
	}

	// Return the middlewareBuilder to allow method chaining.
	return middlewareBuilder
}

// PostMiddleware adds a post-middleware to the current handler.
func (middlewareBuilder *AddMiddlewareBuilder) PostMiddleware(m MiddlewareFunc) *AddMiddlewareBuilder {

	// Extract the name of the middleware function using reflection and strip the pointer indicator.
	typedMiddlewareName := strings.TrimPrefix(runtime.FuncForPC(reflect.ValueOf(m).Pointer()).Name(), "*")

	// Create a middlewareStruct instance with the middleware name and function.
	middleware := middlewareStruct{
		middlewareName: typedMiddlewareName,
		middlewareFunc: m,
	}

	// Retrieve the slice of post-middlewares associated with the current handler from the middlewareBuilder.
	middlewares, ok := middlewareBuilder.postMiddlewares[middlewareBuilder.currentHandlerName]

	// If the current handler does not have any post-middlewares, initialize it with the new middleware.
	if !ok {
		middlewareBuilder.postMiddlewares[middlewareBuilder.currentHandlerName] = []middlewareStruct{
			middleware,
		}
		return middlewareBuilder
	}

	// Add the middleware to the handler if it's not already registered.
	// This is a check to avoid registering the same middleware multiple times for a handler.
	if !isMiddlewareRegisteredForHandler(&middlewares, typedMiddlewareName) {
		middlewareBuilder.postMiddlewares[middlewareBuilder.currentHandlerName] =
			append(middlewareBuilder.postMiddlewares[middlewareBuilder.currentHandlerName], middleware)
	}

	// Return the middlewareBuilder to allow method chaining.
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
func createReflectiveHandler[T1 T](method reflect.Value) IHandler[T, T] {
	return reflectiveHandler[T, T1]{method: method}
}
