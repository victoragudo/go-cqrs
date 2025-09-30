package gocqrs

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockMiddlewareFunc creates a middleware function for testing.
func MockMiddlewareFunc(continueChain bool) MiddlewareFunction {
	return func(ctx context.Context, request T) (context.Context, T, bool) {
		return ctx, request, continueChain
	}
}

// TestPreMiddleware tests the addition of pre-middlewares.
func TestPreMiddleware(t *testing.T) {
	mediator := NewMediator()
	AddCommandHandler[string, string](&MockCommandHandler{})

	middlewareFunc := MockMiddlewareFunc(true)
	builder := mediator.middlewareBuilder
	builder.PreMiddleware(middlewareFunc)

	// Test that middleware was added by checking the registry
	builder.middlewareRegistry.mutex.RLock()
	compiledMiddleware, exists := builder.middlewareRegistry.preMiddlewares[builder.currentHandlerName]
	builder.middlewareRegistry.mutex.RUnlock()

	assert.True(t, exists, "Pre-middleware should be registered")
	assert.Len(t, compiledMiddleware.preChain, 1, "Should have one pre-middleware")
}

// TestPostMiddleware tests the addition of post-middlewares.
func TestPostMiddleware(t *testing.T) {
	mediator := NewMediator()
	AddCommandHandler[string, string](&MockCommandHandler{})

	middlewareFunc := MockMiddlewareFunc(true)
	builder := mediator.middlewareBuilder
	builder.PostMiddleware(middlewareFunc)

	// Test that middleware was added by checking the registry
	builder.middlewareRegistry.mutex.RLock()
	compiledMiddleware, exists := builder.middlewareRegistry.postMiddlewares[builder.currentHandlerName]
	builder.middlewareRegistry.mutex.RUnlock()

	assert.True(t, exists, "Post-middleware should be registered")
	assert.Len(t, compiledMiddleware.postChain, 1, "Should have one post-middleware")
}

// TestPreMiddlewares tests adding multiple pre-middlewares.
func TestPreMiddlewares(t *testing.T) {
	mediator := NewMediator()
	AddCommandHandler[string, string](&MockCommandHandler{})

	middleware1 := MockMiddlewareFunc(true)
	middleware2 := MockMiddlewareFunc(true)

	builder := mediator.middlewareBuilder
	builder.PreMiddlewares(middleware1, middleware2)

	// Test that middlewares were added
	builder.middlewareRegistry.mutex.RLock()
	compiledMiddleware, exists := builder.middlewareRegistry.preMiddlewares[builder.currentHandlerName]
	builder.middlewareRegistry.mutex.RUnlock()

	assert.True(t, exists, "Pre-middlewares should be registered")
	assert.Len(t, compiledMiddleware.preChain, 2, "Should have two pre-middlewares")
}

// TestPostMiddlewares tests adding multiple post-middlewares.
func TestPostMiddlewares(t *testing.T) {
	mediator := NewMediator()
	AddCommandHandler[string, string](&MockCommandHandler{})

	middleware1 := MockMiddlewareFunc(true)
	middleware2 := MockMiddlewareFunc(true)

	builder := mediator.middlewareBuilder
	builder.PostMiddlewares(middleware1, middleware2)

	// Test that middlewares were added
	builder.middlewareRegistry.mutex.RLock()
	compiledMiddleware, exists := builder.middlewareRegistry.postMiddlewares[builder.currentHandlerName]
	builder.middlewareRegistry.mutex.RUnlock()

	assert.True(t, exists, "Post-middlewares should be registered")
	assert.Len(t, compiledMiddleware.postChain, 2, "Should have two post-middlewares")
}

// TestMiddlewareExecution tests that middlewares execute correctly with commands.
func TestMiddlewareExecution(t *testing.T) {
	// Reset default mediator for clean test
	defaultMediator = nil
	once = sync.Once{}

	AddCommandHandler[string, string](&MockCommandHandler{}).
		PreMiddleware(func(ctx context.Context, request T) (context.Context, T, bool) {
			// Modify request
			if str, ok := request.(string); ok {
				return ctx, "pre-" + str, true
			}
			return ctx, request, true
		})

	ctx := context.Background()
	response, err := SendCommand[string](ctx, "test")

	assert.NoError(t, err)
	assert.Equal(t, "handled: pre-test", response, "Middleware should modify the request")
}

// TestMiddlewareChainStop tests that middleware can stop the chain.
func TestMiddlewareChainStop(t *testing.T) {
	// Reset default mediator for clean test
	defaultMediator = nil
	once = sync.Once{}

	AddCommandHandler[string, string](&MockCommandHandler{}).
		PreMiddleware(func(ctx context.Context, request T) (context.Context, T, bool) {
			return ctx, request, false // Stop the chain
		})

	ctx := context.Background()
	_, err := SendCommand[string](ctx, "test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "middleware chain interrupted")
}
