package gocqrs

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type handlerMap map[string]any

// Declare global variables for storing handlers and their mutexes for synchronization.
var (
	handlers          handlerMap
	handlerMutex      sync.RWMutex
	eventHandlerMutex sync.RWMutex
	eventHandlers     map[string][]eventHandlersType
	middlewareBuilder AddMiddlewareBuilder
)

// init initializes variables
func init() {
	handlers = make(map[string]any)
	handlerMutex = sync.RWMutex{}
	eventHandlerMutex = sync.RWMutex{}
	eventHandlers = make(map[string][]eventHandlersType)
	middlewareBuilder = AddMiddlewareBuilder{
		preMiddlewares:  make(map[string][]middlewareStruct),
		postMiddlewares: make(map[string][]middlewareStruct),
	}
}

// AddQueryHandler registers a command handler.
func AddQueryHandler[Query T, QueryResponse T](handler IHandler[Query, QueryResponse]) *AddMiddlewareBuilder {
	return addRequest[Query, QueryResponse](handler)
}

// AddCommandHandler registers a command handler.
func AddCommandHandler[Command T, CommandResponse T](handler IHandler[Command, CommandResponse]) *AddMiddlewareBuilder {
	return addRequest[Command, CommandResponse](handler)
}

func addRequest[T1 T, T2 T](handler IHandler[T1, T2]) *AddMiddlewareBuilder {
	// Determine the type name of the TCommand generic parameter, removing the pointer symbol if present.
	typed := reflect.TypeOf(new(T1)).Elem().String()

	// Determine the type name of the handler parameter, removing the pointer symbol if present.
	typedHandlerName := reflect.TypeOf(handler).String()

	// Store command handler for a specific command as a wrapper
	storeMapValue(handlers, typed, newHandlerWrapper[T1, T2](handler, typedHandlerName), &handlerMutex)

	middlewareBuilder.currentHandlerName = typedHandlerName
	return &middlewareBuilder
}

// AddEventHandlers adds multiple event handlers for a given event type.
// It uses generics to allow any event type and ensures type safety for handlers.
func AddEventHandlers[TEvent T](handlers ...IEventHandler[TEvent]) error {
	// Get the type name of the event, removing the pointer prefix if present.
	typedEvent := reflect.TypeOf(new(TEvent)).Elem().String()

	// Load the registered handlers for this event type, if any.
	registeredHandlers := loadOrStoreEventHandlers(eventHandlers, typedEvent, &eventHandlerMutex)

	// Iterate through the provided handlers and add them to the registered handlers.
	for _, handler := range handlers {
		typedHandlerName := reflect.TypeOf(handler).String()
		hWrapper := newEventHandlerWrapper[TEvent](handler, typedHandlerName)

		if !checkTypeNameInEventHandlers(typedHandlerName, registeredHandlers) {
			evtHandler := eventHandlersType{
				typeName:     typedHandlerName,
				eventHandler: hWrapper,
			}
			registeredHandlers = append(registeredHandlers, evtHandler)
		}
	}

	// Update the eventHandlers map with the newly added handlers.
	eventHandlers[typedEvent] = registeredHandlers
	return nil
}

// SendCommand executes a command by finding the appropriate handler.
// It is a generic function parameterized by 'CommandResponse T', where 'T' is the expected response type for the command.
func SendCommand[CommandResponse T](ctx context.Context, command any) (CommandResponse, error) {
	return send[CommandResponse](ctx, command)
}

// SendQuery executes a query by finding the appropriate handler.
// It is a generic function parameterized by 'QueryResponse T', where 'T' is the expected response type.
func SendQuery[QueryResponse T](ctx context.Context, query any) (QueryResponse, error) {
	return send[QueryResponse](ctx, query)
}

func send[Response T](ctx context.Context, in any) (Response, error) {
	// Retrieve the type of the request as a string
	typedIn := reflect.TypeOf(in).String()

	var zero Response
	var value any
	var ok bool

	tType := reflect.TypeOf(zero)

	if tType.Kind() == reflect.Ptr {
		zero = reflect.New(tType.Elem()).Interface().(Response)
	} else {
		zero = reflect.Zero(tType).Interface().(Response)
	}

	value, ok = getMapValue(handlers, typedIn, &handlerMutex)

	// If no handler is found for the command, return the zero value and an error.
	if !ok {
		return zero, fmt.Errorf("handler not found for this request: %v", typedIn)
	}

	handlerField, ok := getField(value, "Handler")
	if !ok {
		return zero, fmt.Errorf("[Handler] field found: %T", value)
	}

	handleMethod, ok := getMethodByName(handlerField, "Handle")
	if !ok {
		return zero, fmt.Errorf("[Handle] method not found for handler: %T", handlerField)
	}

	handlerNameField, ok := getField(value, "Name")
	if !ok {
		return zero, fmt.Errorf("[Name] field not found found: %T", value)
	}

	handlerName := (handlerNameField.Interface()).(string)

	in = middlewareBuilder.executePreMiddlewares(ctx, in, handlerName)
	response, err := createReflectiveHandler[Response](handleMethod).Handle(ctx, in)
	middlewareBuilder.executePostMiddlewares(ctx, in, handlerName)
	return response, err
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
	registeredEventHandlers, ok := eventHandlers[typedEvent]
	// If no event handlers are found for the type, return an error.
	if !ok {
		return fmt.Errorf("no handlers found for this event: %v", typedEvent)
	}

	// Initialize a slice to collect errors from the event handlers.
	handlerErrors := make([]error, 0)

	// Iterate over the registered event handlers.
	for _, eventHandler := range registeredEventHandlers {
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
