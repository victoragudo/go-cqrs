package gocqrs

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// Declare global variables for storing handlers and their mutexes for synchronization.
var (
	commandHandlers sync.Map
	queryHandlers   sync.Map
	eventHandlers   sync.Map
)

// Initialize the handler maps and mutexes.
func init() {
	commandHandlers = sync.Map{}
	queryHandlers = sync.Map{}
	eventHandlers = sync.Map{}
}

// AddQueryHandler registers a query handler.
func AddQueryHandler[TQuery T, QueryResponse T](handler IQueryHandler[TQuery, QueryResponse]) {
	// Determine the type name of the TQuery generic parameter, removing the pointer symbol if present.
	typedQueryRequest := strings.TrimPrefix(reflect.TypeOf(new(TQuery)).String(), "*")

	// Store the handler in a global concurrent map with the type name as the key.
	// Wraps the handler in a queryHandlerWrapper to maintain type safety.
	queryHandlers.Store(typedQueryRequest, newQueryHandlerWrapper[TQuery, QueryResponse](handler))
}

// AddCommandHandler registers a command handler.
func AddCommandHandler[TCommand T, TResponse T](handler ICommandHandler[TCommand, TResponse]) {
	// Determine the type name of the TCommand generic parameter, removing the pointer symbol if present.
	typedCommandRequest := strings.TrimPrefix(reflect.TypeOf(new(TCommand)).String(), "*")

	// Store the handler in a global concurrent map with the type name as the key.
	// Wraps the handler in a commandHandlerWrapper to maintain type safety.
	commandHandlers.Store(typedCommandRequest, newCommandHandlerWrapper[TCommand, TResponse](handler))
}

// AddEventHandlers adds multiple event handlers for a given event type.
// It uses generics to allow any event type and ensures type safety for handlers.
func AddEventHandlers[TEvent T](handlers ...IEventHandler[TEvent]) error {
	// Get the type name of the event, removing the pointer prefix if present.
	typedEvent := strings.TrimPrefix(reflect.TypeOf(new(TEvent)).String(), "*")

	// Load the registered handlers for this event type, if any.
	value, _ := eventHandlers.LoadOrStore(typedEvent, make([]eventHandlersType, 0, len(handlers)))
	registeredHandlers, ok := value.([]eventHandlersType)
	if !ok {
		return fmt.Errorf("unable to cast registered handlers")
	}

	// Iterate through the provided handlers and add them to the registered handlers.
	for _, handler := range handlers {
		handlerTypeName := strings.TrimPrefix(reflect.TypeOf(handler).String(), "*")

		if !checkTypeNameInEventHandlers(handlerTypeName, registeredHandlers) {
			evtHandler := eventHandlersType{
				typeName:     handlerTypeName,
				eventHandler: newEventHandlerWrapper[TEvent](handler),
			}
			registeredHandlers = append(registeredHandlers, evtHandler)
		}
	}

	// Update the eventHandlers map with the newly added handlers.
	eventHandlers.Store(typedEvent, registeredHandlers)
	return nil
}

// SendCommand executes a command by finding the appropriate handler.
// It is a generic function parameterized by 'CommandResponse T', where 'T' is the expected response type for the command.
func SendCommand[TResponse T](ctx context.Context, command T) (TResponse, error) {
	// Retrieve the type of the command request as a string, removing the pointer symbol (*) if present.
	typedCommand := strings.TrimPrefix(reflect.TypeOf(command).String(), "*")

	// Look up the command handler in a concurrent map using the command's type.
	v, ok := commandHandlers.Load(typedCommand)

	// Create a zero value instance of CommandResponse.
	zeroValue := *new(TResponse)

	// If no handler is found for the command, return the zero value and an error.
	if !ok {
		return zeroValue, fmt.Errorf("no handler found for this command: %v", typedCommand)
	}

	// Assert that the retrieved handler matches the ICommandHandler interface.
	cmdHandler := v.(ICommandHandler[T, T])

	// Call the handle method of the command handler, passing the context and command request.
	commandResponse, err := cmdHandler.Handle(ctx, command)

	// Attempt to assert that the response is of the expected type (CommandResponse).
	response, ok := commandResponse.(TResponse)

	// If the assertion is successful, return the response along with any error that occurred.
	if ok {
		return response, err
	}

	// If the assertion fails, return the zero value and no error.
	return zeroValue, nil
}

// SendQuery executes a query by finding the appropriate handler.
// It is a generic function parameterized by 'QueryResponse T', where 'T' is the expected response type.
func SendQuery[TQuery T](ctx context.Context, query T) (TQuery, error) {
	// Get the type of the query request, removing the pointer symbol (*) if present.
	typedQuery := strings.TrimPrefix(reflect.TypeOf(query).String(), "*")

	// Attempt to find a query handler for the given type of query in a concurrent map.
	v, ok := queryHandlers.Load(typedQuery)

	// Create a zero value of the generic type QueryResponse.
	zeroValue := *new(TQuery)

	// If no handler is found for the query, return the zero value and an error.
	if !ok {
		return zeroValue, fmt.Errorf("no handler found for this query: %v", typedQuery)
	}

	// Assert that the retrieved value is of the expected interface type (IQueryHandler).
	queryHandler := v.(IQueryHandler[T, T])

	// Use the handler to process the query, passing the context and the query itself.
	queryResponse, err := queryHandler.Handle(ctx, query)

	// Attempt to assert that the response is of the expected type (QueryResponse).
	response, ok := queryResponse.(TQuery)

	// If the assertion is successful, return the response and any error encountered.
	if ok {
		return response, err
	}

	// If the assertion fails, return the zero value and no error.
	return zeroValue, nil
}

// PublishEvent publishes an event of a generic type T to all registered event handlers.
// This function performs the following steps:
//  1. It determines the type of the provided event using reflection and removes any pointer
//     indicator from the type name (if present).
//  2. It then attempts to retrieve the list of event handlers registered for this event type.
//  3. If no handlers are registered for this event type, it returns an error indicating this.
//  4. If handlers are found, the function iterates over each handler and calls its Handle
//     method, passing the current context and the event.
//  5. Any errors returned by the handlers are collected. If one or more handlers return errors,
//     these are combined into a single error that is then returned.
//  6. If all handlers process the event without error, the function returns nil, indicating
//     successful processing of the event.
//
// The function is designed to work within an event-driven architecture where different
// types of events can be handled by different handlers. This allows for a decoupled and
// scalable system where new event types and handlers can be added without modifying the
// core event publishing logic.
func PublishEvent(ctx context.Context, event T) error {
	// Obtain the type of the event as a string using reflection.
	// This strips the "*" prefix, which indicates a pointer type, to get the base type name.
	typedEvent := strings.TrimPrefix(reflect.TypeOf(event).String(), "*")

	// Attempt to load the registered event handlers for the specific event type.
	registeredEventHandlers, ok := eventHandlers.Load(typedEvent)

	// If no event handlers are found for the type, return an error.
	if !ok {
		return fmt.Errorf("no handlers found for this event: %v", typedEvent)
	}

	// Initialize a slice to collect errors from the event handlers.
	handlerErrors := make([]error, 0)

	// Iterate over the registered event handlers.
	for _, eventHandler := range registeredEventHandlers.([]eventHandlersType) {
		// Call the event handler and pass the context and the event.
		// If the handler returns an error, append it to the handlerErrors slice.
		err := eventHandler.eventHandler.Handle(ctx, event)
		if err != nil {
			handlerErrors = append(handlerErrors, err)
		}
	}

	// If there were any errors collected from the handlers, return them joined together.
	// This combines multiple errors into a single error.
	if len(handlerErrors) > 0 {
		return errors.Join(handlerErrors...)
	}

	// If execution reaches here, it means all handlers executed without error.
	// Return nil indicating successful execution.
	return nil
}

// checkTypeNameInEventHandlers checks if the string is in the slice of structs.
// It returns true if the string is found, false otherwise.
func checkTypeNameInEventHandlers(typeName string, eventHandlers []eventHandlersType) bool {
	for _, v := range eventHandlers {
		if v.typeName == typeName {
			return true
		}
	}
	return false
}
