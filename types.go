package gocqrs

import (
	"context"
)

type (
	// T is a generic type alias for any type.
	T any
	// eventHandlersType is a struct that holds the type name of an event
	// and its corresponding event handler.
	eventHandlersType struct {
		typeName     string
		eventHandler IHandler[T, T]
	}
	// IHandler is an interface representing a generic handler
	// with input and output of generic types T1 and T2.
	// It requires implementing a Handle method.
	IHandler[T1 T, T2 T] interface {
		Handle(ctx context.Context, in T1) (out T2, err error)
	}
	// IEventHandler is an interface for event handlers
	// that handle events of a generic type TEvent.
	// It defines a Handle method for processing events.
	IEventHandler[TEvent T] interface {
		Handle(ctx context.Context, event TEvent) error
	}
)
