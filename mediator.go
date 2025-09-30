package gocqrs

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// eventHandlerWrapper wraps a typed event handler to implement IEventHandler[any]
type eventHandlerWrapper[TEvent T] struct {
	handler IEventHandler[TEvent]
}

func (wrapper *eventHandlerWrapper[TEvent]) Handle(ctx context.Context, event any) error {
	if typedEvent, ok := event.(TEvent); ok {
		return wrapper.handler.Handle(ctx, typedEvent)
	}
	return fmt.Errorf("event type mismatch: expected %T, got %T", *new(TEvent), event)
}

var defaultMediator *Mediator
var once sync.Once

func GetDefaultMediator() *Mediator {
	once.Do(func() {
		defaultMediator = NewMediator()
	})
	return defaultMediator
}

func NewMediator() *Mediator {
	return &Mediator{
		handlerRegistry: &HandlerRegistry{
			handlers: make(map[reflect.Type]RegisteredHandler),
			mutex:    sync.RWMutex{},
		},
		eventRegistry: &EventRegistry{
			handlers: make(map[reflect.Type][]RegisteredEventHandler),
			mutex:    sync.RWMutex{},
		},
		middlewareBuilder: &MiddlewareBuilder{
			middlewareRegistry: &MiddlewareRegistry{
				preMiddlewares:  make(map[string]CompiledMiddleware),
				postMiddlewares: make(map[string]CompiledMiddleware),
				mutex:           sync.RWMutex{},
			},
		},
	}
}

// AddQueryHandler registers a query handler using the default mediator
func AddQueryHandler[Query T, QueryResponse T](handler IHandler[Query, QueryResponse]) *MiddlewareBuilder {
	mediator := GetDefaultMediator()
	requestType := reflect.TypeOf((*Query)(nil)).Elem()
	handlerTypeName := reflect.TypeOf(handler).String()

	registeredHandler := RegisteredHandler{
		handler:     handler,
		handlerType: reflect.TypeOf(handler),
		handlerName: handlerTypeName,
	}

	mediator.handlerRegistry.mutex.Lock()
	mediator.handlerRegistry.handlers[requestType] = registeredHandler
	mediator.handlerRegistry.mutex.Unlock()

	mediator.middlewareBuilder.currentHandlerName = handlerTypeName
	return mediator.middlewareBuilder
}

// AddCommandHandler registers a command handler using the default mediator
func AddCommandHandler[Command T, CommandResponse T](handler IHandler[Command, CommandResponse]) *MiddlewareBuilder {
	mediator := GetDefaultMediator()
	requestType := reflect.TypeOf((*Command)(nil)).Elem()
	handlerTypeName := reflect.TypeOf(handler).String()

	registeredHandler := RegisteredHandler{
		handler:     handler,
		handlerType: reflect.TypeOf(handler),
		handlerName: handlerTypeName,
	}

	mediator.handlerRegistry.mutex.Lock()
	mediator.handlerRegistry.handlers[requestType] = registeredHandler
	mediator.handlerRegistry.mutex.Unlock()

	mediator.middlewareBuilder.currentHandlerName = handlerTypeName
	return mediator.middlewareBuilder
}

// AddEventHandlers adds multiple event handlers for a given event type using the default mediator
func AddEventHandlers[TEvent T](eventHandlers ...IEventHandler[TEvent]) error {
	mediator := GetDefaultMediator()
	eventType := reflect.TypeOf((*TEvent)(nil)).Elem()

	mediator.eventRegistry.mutex.Lock()
	defer mediator.eventRegistry.mutex.Unlock()

	existingHandlers := mediator.eventRegistry.handlers[eventType]

	for _, handler := range eventHandlers {
		handlerName := reflect.TypeOf(handler).String()

		// Check if handler is already registered
		alreadyExists := false
		for _, existing := range existingHandlers {
			if existing.name == handlerName {
				alreadyExists = true
				break
			}
		}

		if !alreadyExists {
			// Create wrapper to convert typed handler to any handler
			anyHandler := &eventHandlerWrapper[TEvent]{handler: handler}
			registeredEventHandler := RegisteredEventHandler{
				handler: anyHandler,
				name:    handlerName,
			}
			existingHandlers = append(existingHandlers, registeredEventHandler)
		}
	}

	mediator.eventRegistry.handlers[eventType] = existingHandlers
	return nil
}

// SendCommand executes a command using the default mediator
func SendCommand[CommandResponse T](ctx context.Context, command any) (CommandResponse, error) {
	return sendRequest[CommandResponse](ctx, command)
}

// SendQuery executes a query using the default mediator
func SendQuery[QueryResponse T](ctx context.Context, query any) (QueryResponse, error) {
	return sendRequest[QueryResponse](ctx, query)
}

// sendRequest executes a request by finding the appropriate handler
func sendRequest[Response T](ctx context.Context, request any) (Response, error) {
	mediator := GetDefaultMediator()
	var zeroResponse Response
	requestType := reflect.TypeOf(request)

	mediator.handlerRegistry.mutex.RLock()
	registeredHandler, exists := mediator.handlerRegistry.handlers[requestType]
	mediator.handlerRegistry.mutex.RUnlock()

	if !exists {
		return zeroResponse, fmt.Errorf("no handler found for type: %v", requestType)
	}

	// Execute pre-middlewares
	modifiedContext, modifiedRequest, shouldContinue := mediator.executePreMiddlewares(ctx, request, registeredHandler.handlerName)

	if !shouldContinue {
		return zeroResponse, fmt.Errorf("middleware chain interrupted for handler: %s", registeredHandler.handlerName)
	}

	// Execute the handler using reflection
	handlerValue := reflect.ValueOf(registeredHandler.handler)
	handleMethod := handlerValue.MethodByName("Handle")

	if !handleMethod.IsValid() {
		return zeroResponse, fmt.Errorf("handler does not have Handle method: %v", requestType)
	}

	results := handleMethod.Call([]reflect.Value{
		reflect.ValueOf(modifiedContext),
		reflect.ValueOf(modifiedRequest),
	})

	if len(results) != 2 {
		return zeroResponse, fmt.Errorf("handler returned unexpected number of values")
	}

	// Extract response
	var response Response
	if results[0].IsValid() && !results[0].IsZero() {
		if typedResponse, ok := results[0].Interface().(Response); ok {
			response = typedResponse
		}
	}

	// Extract error
	var handlerError error
	if results[1].IsValid() && !results[1].IsZero() {
		if err, ok := results[1].Interface().(error); ok {
			handlerError = err
		}
	}

	// Execute post-middlewares
	mediator.executePostMiddlewares(modifiedContext, modifiedRequest, registeredHandler.handlerName)

	return response, handlerError
}

// PublishEvent publishes an event to all registered event handlers using the default mediator
func PublishEvent(ctx context.Context, event any) error {
	return GetDefaultMediator().publishEvent(ctx, event)
}

// publishEvent publishes an event to all registered event handlers
func (mediator *Mediator) publishEvent(ctx context.Context, event any) error {
	eventType := reflect.TypeOf(event)

	mediator.eventRegistry.mutex.RLock()
	registeredHandlers, exists := mediator.eventRegistry.handlers[eventType]
	mediator.eventRegistry.mutex.RUnlock()

	if !exists || len(registeredHandlers) == 0 {
		return fmt.Errorf("no handlers found for event type: %v", eventType)
	}

	var handlerErrors []error
	var errorMutex sync.Mutex

	// Execute handlers in parallel
	var waitGroup sync.WaitGroup
	for _, handler := range registeredHandlers {
		waitGroup.Add(1)
		go func(h RegisteredEventHandler) {
			defer waitGroup.Done()
			if err := h.handler.Handle(ctx, event); err != nil {
				errorMutex.Lock()
				handlerErrors = append(handlerErrors, err)
				errorMutex.Unlock()
			}
		}(handler)
	}

	waitGroup.Wait()

	if len(handlerErrors) > 0 {
		return errors.Join(handlerErrors...)
	}

	return nil
}

// executePreMiddlewares executes pre-middlewares for a request
func (mediator *Mediator) executePreMiddlewares(ctx context.Context, request any, handlerName string) (context.Context, any, bool) {
	mediator.middlewareBuilder.middlewareRegistry.mutex.RLock()
	compiledMiddleware, exists := mediator.middlewareBuilder.middlewareRegistry.preMiddlewares[handlerName]
	mediator.middlewareBuilder.middlewareRegistry.mutex.RUnlock()

	if !exists {
		return ctx, request, true
	}

	modifiedContext := ctx
	modifiedRequest := request

	for _, middlewareFunc := range compiledMiddleware.preChain {
		var shouldContinue bool
		modifiedContext, modifiedRequest, shouldContinue = middlewareFunc(modifiedContext, modifiedRequest)
		if !shouldContinue {
			return modifiedContext, modifiedRequest, false
		}
	}

	return modifiedContext, modifiedRequest, true
}

// executePostMiddlewares executes post-middlewares for a request
func (mediator *Mediator) executePostMiddlewares(ctx context.Context, request any, handlerName string) {
	mediator.middlewareBuilder.middlewareRegistry.mutex.RLock()
	compiledMiddleware, exists := mediator.middlewareBuilder.middlewareRegistry.postMiddlewares[handlerName]
	mediator.middlewareBuilder.middlewareRegistry.mutex.RUnlock()

	if !exists {
		return
	}

	modifiedContext := ctx
	modifiedRequest := request

	for _, middlewareFunc := range compiledMiddleware.postChain {
		var shouldContinue bool
		modifiedContext, modifiedRequest, shouldContinue = middlewareFunc(modifiedContext, modifiedRequest)
		if !shouldContinue {
			return
		}
	}
}
