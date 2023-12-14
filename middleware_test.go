package gocqrs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockMiddlewareFunc creates a middleware function for testing.
func MockMiddlewareFunc(continueChain bool) MiddlewareFunc {
	return func(ctx context.Context, request any) (context.Context, any, bool) {
		return ctx, request, continueChain
	}
}

// TestPreMiddleware tests the addition of pre-middlewares.
func TestPreMiddleware(t *testing.T) {
	builder := AddMiddlewareBuilder{
		currentHandlerName: "testHandler",
		preMiddlewares:     make(map[string][]middlewareStruct),
	}

	middlewareFunc := MockMiddlewareFunc(true)
	builder.PreMiddleware(middlewareFunc)

	assert.Len(t, builder.preMiddlewares["testHandler"], 1, "preMiddlewares should contain one middleware for testHandler")
}

// TestPostMiddleware tests the addition of post-middlewares.
func TestPostMiddleware(t *testing.T) {
	builder := AddMiddlewareBuilder{
		currentHandlerName: "testHandler",
		postMiddlewares:    make(map[string][]middlewareStruct),
	}

	middlewareFunc := MockMiddlewareFunc(true)
	builder.PostMiddleware(middlewareFunc)

	assert.Len(t, builder.postMiddlewares["testHandler"], 1, "postMiddlewares should contain one middleware for testHandler")
}

// TestExecutepreMiddlewares tests the execution of pre-middlewares.
func TestExecutepreMiddlewares(t *testing.T) {
	builder := AddMiddlewareBuilder{
		currentHandlerName: "testHandler",
		preMiddlewares:     make(map[string][]middlewareStruct),
	}

	// Add middlewares
	builder.PreMiddleware(MockMiddlewareFunc(true))
	builder.PreMiddleware(MockMiddlewareFunc(false)) // This should stop the chain

	request := "original"
	modifiedRequest := builder.executePreMiddlewares(context.Background(), request, "testHandler")

	assert.Equal(t, request, modifiedRequest, "Request should not be modified as the chain is stopped by the second middleware")
}

// TestExecutepostMiddlewares tests the execution of post-middlewares.
func TestExecutepostMiddlewares(t *testing.T) {
	builder := AddMiddlewareBuilder{
		currentHandlerName: "testHandler",
		postMiddlewares:    make(map[string][]middlewareStruct),
	}

	// Add middlewares
	builder.PostMiddleware(MockMiddlewareFunc(true))
	builder.PostMiddleware(MockMiddlewareFunc(false)) // This should stop the chain

	request := "original"
	builder.executePostMiddlewares(context.Background(), request, "testHandler")

	// No assertion needed as we are testing the flow, not the output
}

// TestIsMiddlewareRegisteredForHandler tests if a middleware is correctly identified as registered.
func TestIsMiddlewareRegisteredForHandler(t *testing.T) {
	middlewares := []middlewareStruct{
		{middlewareName: "Middleware1", middlewareFunc: MockMiddlewareFunc(true)},
		{middlewareName: "Middleware2", middlewareFunc: MockMiddlewareFunc(true)},
	}

	assert.True(t, isMiddlewareRegisteredForHandler(&middlewares, "Middleware1"), "Middleware1 should be registered")
	assert.False(t, isMiddlewareRegisteredForHandler(&middlewares, "Middleware3"), "Middleware3 should not be registered")
}

// TestMultipleMiddlewareRegistration tests if adding the same middleware multiple times is handled correctly.
func TestMultipleMiddlewareRegistration(t *testing.T) {
	builder := AddMiddlewareBuilder{
		currentHandlerName: "testHandler",
		preMiddlewares:     make(map[string][]middlewareStruct),
	}

	middlewareFunc := MockMiddlewareFunc(true)
	builder.PreMiddleware(middlewareFunc)
	builder.PreMiddleware(middlewareFunc) // Add the same middleware again

	assert.Len(t, builder.preMiddlewares["testHandler"], 1, "Middleware should only be registered once")
}

// TestMiddlewareFunctionality tests the actual functionality of the middleware.
func TestMiddlewareFunctionality(t *testing.T) {
	builder := AddMiddlewareBuilder{
		currentHandlerName: "testHandler",
		preMiddlewares:     make(map[string][]middlewareStruct),
	}

	// Middleware that modifies the request
	modifyingMiddleware := func(ctx context.Context, request any) (context.Context, any, bool) {
		return ctx, "modified", true
	}

	builder.PreMiddleware(modifyingMiddleware)

	modifiedRequest := builder.executePreMiddlewares(context.Background(), "original", "testHandler")
	assert.Equal(t, "modified", modifiedRequest, "Request should be modified by the middleware")
}
