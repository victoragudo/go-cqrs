package gocqrs

import (
	"context"
	"fmt"
)

type (
	// handlerWrapper is a generic struct that wraps ICommandHandler.
	// It uses T1 and T2 as generic types.
	handlerWrapper[TRequest T, TResponse T] struct {
		Handler IHandler[TRequest, TResponse]
		Name    string
	}
)

// Handle method for commandHandlerWrapper.
// It takes a context and a commandRequest of generic type T,
// and returns a response of type T and an error if any.
func (handlerWrapper *handlerWrapper[T1, T2]) Handle(ctx context.Context, in T) (T, error) {
	// Assert the type of commandRequest to CommandRequest.
	typedIn, ok := in.(T1)
	if !ok {
		// Return an error if the assertion fails.
		return nil, fmt.Errorf("incorrect request type: %T", in)
	}
	// Call the wrapped command handler's Handle method.
	return handlerWrapper.Handler.Handle(ctx, typedIn)
}

// newHandlerWrapper creates a new handlerWrapper instance.
func newHandlerWrapper[T1 T, T2 T](handler IHandler[T1, T2], handlerName string) *handlerWrapper[T1, T2] {
	return &handlerWrapper[T1, T2]{
		Handler: handler,
		Name:    handlerName,
	}
}

// v creates a new handlerWrapper instance.
func newEventHandlerWrapper[TEvent T](handler IEventHandler[TEvent]) *handlerWrapper[TEvent, T] {
	return &handlerWrapper[TEvent, T]{
		Handler: handler,
	}
}
