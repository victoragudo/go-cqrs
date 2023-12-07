package gocqrs

import "context"

// Define interface types for command, query, and event handlers with generic types.
type (
	T any

	eventHandlersType struct {
		typeName     string
		eventHandler IEventHandler[T]
	}

	// ICommandHandler defines the interface for a command handler with a request and response of generic type.
	ICommandHandler[TCommand T, TResponse T] interface {
		Handle(ctx context.Context, command TCommand) (response TResponse, err error)
	}

	// IQueryHandler defines the interface for a query handler with a request and response of generic type.
	IQueryHandler[TQuery T, TResponse T] interface {
		Handle(ctx context.Context, query TQuery) (response TResponse, err error)
	}

	// IEventHandler defines the interface for an event handler with an event of generic type.
	IEventHandler[TEvent T] interface {
		Handle(ctx context.Context, event TEvent) (err error)
	}
)
