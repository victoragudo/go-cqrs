package gocqrs

import (
	"context"
	"reflect"
	"sync"
)

type (
	T any

	IHandler[RequestType T, ResponseType T] interface {
		Handle(context context.Context, request RequestType) (response ResponseType, error error)
	}

	IEventHandler[EventType T] interface {
		Handle(context context.Context, event EventType) error
	}

	HandlerRegistry struct {
		handlers map[reflect.Type]RegisteredHandler
		mutex    sync.RWMutex
	}

	RegisteredHandler struct {
		handler     interface{}
		handlerType reflect.Type
		handlerName string
	}

	EventRegistry struct {
		handlers map[reflect.Type][]RegisteredEventHandler
		mutex    sync.RWMutex
	}

	RegisteredEventHandler struct {
		handler IEventHandler[any]
		name    string
	}

	CompiledMiddleware struct {
		preChain  []MiddlewareFunction
		postChain []MiddlewareFunction
	}

	MiddlewareFunction func(context context.Context, request any) (modifiedContext context.Context, modifiedRequest any, shouldContinue bool)

	MiddlewareRegistry struct {
		preMiddlewares  map[string]CompiledMiddleware
		postMiddlewares map[string]CompiledMiddleware
		mutex           sync.RWMutex
	}

	Mediator struct {
		handlerRegistry   *HandlerRegistry
		eventRegistry     *EventRegistry
		middlewareBuilder *MiddlewareBuilder
	}

	MiddlewareBuilder struct {
		currentHandlerName string
		middlewareRegistry *MiddlewareRegistry
	}
)
