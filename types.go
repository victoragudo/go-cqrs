package gocqrs

import (
	"context"
)

// Define interface types for command, query, and event handlers with generic types.
type (
	T any

	eventHandlersType struct {
		typeName     string
		eventHandler IHandler[T, T]
	}

	// IHandler defines the interface for a handler with a request and response of generic type.
	IHandler[TIn T, TOut T] interface {
		Handle(ctx context.Context, in TIn) (out TOut, err error)
	}

	// IEventHandler defines the interface for a handler with a request and response of generic type.
	IEventHandler[TEvent T] interface {
		Handle(ctx context.Context, event TEvent) (T, error)
	}

	IHandlerFunc func(ctx context.Context, in T) (T, error)
)

func (handlerFunc IHandlerFunc) Handle(ctx context.Context, in T) (T, error) {
	return handlerFunc(ctx, in)
}
