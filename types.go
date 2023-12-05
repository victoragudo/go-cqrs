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
	ICommandHandler[CommandRequest T, CommandResponse T] interface {
		Handle(ctx context.Context, commandRequest CommandRequest) (commandResponse CommandResponse, err error)
	}

	// IQueryHandler defines the interface for a query handler with a request and response of generic type.
	IQueryHandler[QueryRequest T, QueryResponse T] interface {
		Handle(ctx context.Context, queryRequest QueryRequest) (queryResponse QueryResponse, err error)
	}

	// IEventHandler defines the interface for an event handler with an event of generic type.
	IEventHandler[Event T] interface {
		Handle(ctx context.Context, event Event) (err error)
	}
)
