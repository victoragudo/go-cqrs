package gocqrs

import (
	"context"
	"fmt"
)

// Define a set of generic types for different types of handlers.
type (
	// commandHandlerWrapper is a generic struct that wraps ICommandHandler.
	// It uses CommandRequest and CommandResponse as generic types.
	commandHandlerWrapper[TCommand T, TResponse T] struct {
		commandHandler ICommandHandler[TCommand, TResponse]
	}

	// queryHandlerWrapper is a generic struct that wraps IQueryHandler.
	// It uses QueryRequest and QueryResponse as generic types.
	queryHandlerWrapper[TQuery T, TResponse T] struct {
		queryHandler IQueryHandler[TQuery, TResponse]
	}

	// eventHandlerWrapper is a generic struct that wraps IEventHandler.
	// It uses Event as a generic type.
	eventHandlerWrapper[TEvent T] struct {
		eventHandler IEventHandler[TEvent]
	}
)

// Handle method for commandHandlerWrapper.
// It takes a context and a commandRequest of generic type T,
// and returns a response of type T and an error if any.
func (cmdHandlerWrapper *commandHandlerWrapper[TCommand, TResponse]) Handle(ctx context.Context, command T) (T, error) {
	// Assert the type of commandRequest to CommandRequest.
	typedCommandRequest, ok := command.(TCommand)
	if !ok {
		// Return an error if the assertion fails.
		return nil, fmt.Errorf("incorrect command request type: %T", command)
	}
	// Call the wrapped command handler's Handle method.
	return cmdHandlerWrapper.commandHandler.Handle(ctx, typedCommandRequest)
}

// Handle method for queryHandlerWrapper.
// Similar to commandHandlerWrapper's Handle, but for queries.
func (queryHandlerWrapper *queryHandlerWrapper[TQuery, TResponse]) Handle(ctx context.Context, query T) (T, error) {
	// Assert the type of queryRequest to QueryRequest.
	typedQueryRequest, ok := query.(TQuery)
	if !ok {
		// Return an error if the assertion fails.
		return nil, fmt.Errorf("incorrect query request type: %T", query)
	}
	// Call the wrapped query handler's Handle method.
	return queryHandlerWrapper.queryHandler.Handle(ctx, typedQueryRequest)
}

// Handle method for eventHandlerWrapper.
// Similar to commandHandlerWrapper's Handle, but for events.
func (eventHandlerWrapper *eventHandlerWrapper[TEvent]) Handle(ctx context.Context, event T) error {
	// Assert the type of event to Event.
	typedEvent, ok := event.(TEvent)
	if !ok {
		// Return an error if the assertion fails.
		return fmt.Errorf("incorrect event type: %T", event)
	}
	// Call the wrapped event handler's Handle method.
	return eventHandlerWrapper.eventHandler.Handle(ctx, typedEvent)
}

// newCommandHandlerWrapper creates a new commandHandlerWrapper instance.
func newCommandHandlerWrapper[TCommand T, TResponse T](handler ICommandHandler[TCommand, TResponse]) *commandHandlerWrapper[TCommand, TResponse] {
	return &commandHandlerWrapper[TCommand, TResponse]{
		commandHandler: handler,
	}
}

// newQueryHandlerWrapper creates a new queryHandlerWrapper instance.
func newQueryHandlerWrapper[TQuery T, TResponse T](handler IQueryHandler[TQuery, TResponse]) *queryHandlerWrapper[TQuery, TResponse] {
	return &queryHandlerWrapper[TQuery, TResponse]{
		queryHandler: handler,
	}
}

// newEventHandlerWrapper creates a new eventHandlerWrapper instance.
func newEventHandlerWrapper[TEvent T](handler IEventHandler[TEvent]) *eventHandlerWrapper[TEvent] {
	return &eventHandlerWrapper[TEvent]{
		eventHandler: handler,
	}
}
