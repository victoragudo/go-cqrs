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
	commandHandlers   map[string]any
	queryHandlers     sync.Map
	eventHandlers     sync.Map
	middlewareBuilder AddMiddlewareBuilder
)

// init initializes variables
func init() {
	commandHandlers = make(map[string]any)
	queryHandlers = sync.Map{}
	eventHandlers = sync.Map{}
	middlewareBuilder = AddMiddlewareBuilder{
		commandMiddlewares: make(map[string][]handlerMiddleware),
	}
}

// AddQueryHandler registers a query handler.
func AddQueryHandler[TQuery T, QueryResponse T](handler IHandler[TQuery, QueryResponse]) *AddMiddlewareBuilder {
	// Determine the type name of the TQuery generic parameter, removing the pointer symbol if present.
	typedQueryRequest := strings.TrimPrefix(reflect.TypeOf(new(TQuery)).String(), "*")

	queryHandlers.Store(typedQueryRequest, newHandlerWrapper[TQuery, QueryResponse](handler, ""))

	middlewareBuilder.currentHandlerName = typedQueryRequest
	return &middlewareBuilder
}

// AddCommandHandler registers a command handler.
func AddCommandHandler[TCommand T, TResponse T](handler IHandler[TCommand, TResponse]) *AddMiddlewareBuilder {
	// Determine the type name of the TCommand generic parameter, removing the pointer symbol if present.
	typedCommandRequest := strings.TrimPrefix(reflect.TypeOf(new(TCommand)).String(), "*")

	// Determine the type name of the handler parameter, removing the pointer symbol if present.
	typedHandlerName := strings.TrimPrefix(reflect.TypeOf(handler).String(), "*")

	// Store command handler for a specific command as a wrapper
	storeMapValue(commandHandlers, typedCommandRequest, newHandlerWrapper[TCommand, TResponse](handler, typedHandlerName))

	middlewareBuilder.currentHandlerName = typedHandlerName
	return &middlewareBuilder
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
		hWrapper := newEventHandlerWrapper[TEvent](handler)

		if !checkTypeNameInEventHandlers(handlerTypeName, registeredHandlers) {
			evtHandler := eventHandlersType{
				typeName:     handlerTypeName,
				eventHandler: hWrapper,
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
func SendCommand[TResponse T](ctx context.Context, command any) (TResponse, error) {
	// Retrieve the type of the command request as a string, removing the pointer symbol (*) if present.
	typedCommand := strings.TrimPrefix(reflect.TypeOf(command).String(), "*")

	// Create a zero value instance of CommandResponse.
	zero := *new(TResponse)

	value, ok := getMapValue(commandHandlers, typedCommand)
	// If no handler is found for the command, return the zero value and an error.
	if !ok {
		return zero, fmt.Errorf("no handler found for this command: %v", typedCommand)
	}

	handlerField, ok := getField(value, "Handler")
	if !ok {
		return zero, fmt.Errorf("no Handler field found: %T", value)
	}

	handleMethod, ok := getMethodByName(handlerField, "Handle")
	if !ok {
		return zero, fmt.Errorf("no Handle method found for handler: %T", handlerField)
	}

	params := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(command)}

	handlerNameField, ok := getField(value, "Name")
	if !ok {
		return zero, fmt.Errorf("no handler name found: %T", value)
	}

	handlerName := (handlerNameField.Interface()).(string)
	middlewares, ok := middlewareBuilder.commandMiddlewares[handlerName]

	if len(middlewares) > 0 {
		// Call the handle method with the registered middlewares of the command handler, passing the context and command request.
		handler := createReflectiveHandler[TResponse](handleMethod)
		var h IHandler[T, T]
		for _, middleware := range middlewares {
			h = middleware.middleware(handler)
		}
		response, err := h.Handle(ctx, command)
		return response.(TResponse), err
	} else {
		results := handleMethod.Call(params)

		if len(results) >= 2 {
			var err error
			if results[1].IsNil() {
				err = nil
			} else {
				err, _ = results[1].Interface().(error)
			}
			return (results[0].Interface()).(TResponse), err
		}
		// If the assertion fails, return the zero value and no error.
		return zero, nil
	}
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
	queryHandler := v.(IHandler[T, T])

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
		_, err := eventHandler.eventHandler.Handle(ctx, event)
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
