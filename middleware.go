package gocqrs

import (
	"context"
	"reflect"
	"runtime"
	"strings"
)

type (
	// MiddlewareFunc defines a function type used for middleware.
	// It receives a context and a request (of any type). The function returns three values:
	// 1. A potentially modified context, which is the chained context after processing.
	// 2. A result (of any type), which is the chained request parameter after processing.
	// 3. A boolean indicating whether to continue with the chain of middlewares or not.
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
		middlewareName string                                                              // Name of the middleware.
		middlewareFunc func(ctx context.Context, request any) (context.Context, any, bool) // The middleware function.
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
// middlewareFunc parameter defines a function type used for middleware.
// It receives a context and a request (of any type). The function returns three values:
// 1. A potentially modified context, which is the chained context after processing.
// 2. A result (of any type), which is the chained request parameter after processing.
// 3. A boolean indicating whether to continue with the chain of middlewares or not.
func (middlewareBuilder *AddMiddlewareBuilder) PreMiddleware(middlewareFunc func(ctx context.Context, request any) (context.Context, any, bool)) *AddMiddlewareBuilder {

	// Extract the name of the middleware function using reflection and strip the pointer indicator.
	typedMiddlewareName := strings.TrimPrefix(runtime.FuncForPC(reflect.ValueOf(middlewareFunc).Pointer()).Name(), "*")

	// Create a middlewareStruct instance with the middleware name and function.
	middleware := middlewareStruct{
		middlewareName: typedMiddlewareName,
		middlewareFunc: middlewareFunc,
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

// PreMiddlewares adds a list of middleware functions to be executed before a primary action.
// The function is a method of the AddMiddlewareBuilder type.
func (middlewareBuilder *AddMiddlewareBuilder) PreMiddlewares(middlewaresFunc ...func(ctx context.Context, request any) (context.Context, any, bool)) *AddMiddlewareBuilder {
	// Iterate over the provided list of middleware functions.
	for _, middleware := range middlewaresFunc {
		// For each middleware, add it to the list of pre-execution middlewares
		// using the PreMiddleware method of middlewareBuilder.
		middlewareBuilder.PreMiddleware(middleware)
	}
	// Return the middlewareBuilder to allow for method chaining.
	return middlewareBuilder
}

// PostMiddlewares adds a list of middleware functions to be executed after a primary action.
// This function is also a method of the AddMiddlewareBuilder type.
func (middlewareBuilder *AddMiddlewareBuilder) PostMiddlewares(middlewaresFunc ...func(ctx context.Context, request any) (context.Context, any, bool)) *AddMiddlewareBuilder {
	// Iterate over the provided list of middleware functions.
	for _, middleware := range middlewaresFunc {
		// For each middleware, add it to the list of post-execution middlewares
		// using the PostMiddleware method of middlewareBuilder.
		middlewareBuilder.PostMiddleware(middleware)
	}
	// Return the middlewareBuilder to allow for method chaining.
	return middlewareBuilder
}

// PostMiddleware adds a post-middleware to the current handler.
// middlewareFunc parameter defines a function type used for middleware.
// It receives a context and a request (of any type). The function returns three values:
// 1. A potentially modified context, which is the chained context after processing.
// 2. A result (of any type), which is the chained request parameter after processing.
// 3. A boolean indicating whether to continue with the chain of middlewares or not.
func (middlewareBuilder *AddMiddlewareBuilder) PostMiddleware(middlewareFunc func(ctx context.Context, request any) (context.Context, any, bool)) *AddMiddlewareBuilder {

	// Extract the name of the middleware function using reflection and strip the pointer indicator.
	typedMiddlewareName := strings.TrimPrefix(runtime.FuncForPC(reflect.ValueOf(middlewareFunc).Pointer()).Name(), "*")

	// Create a middlewareStruct instance with the middleware name and function.
	middleware := middlewareStruct{
		middlewareName: typedMiddlewareName,
		middlewareFunc: middlewareFunc,
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
