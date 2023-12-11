package gocqrs

import (
	"context"
	"reflect"
	"runtime"
	"strings"
)

type (
	AddMiddlewareBuilder struct {
		currentHandlerName string
		commandMiddlewares map[string][]handlerMiddleware
	}
	handlerMiddleware struct {
		middlewareName string
		middleware     func(next IHandler[T, T]) IHandler[T, T]
	}
	reflectiveHandler[T1 T, T2 T] struct {
		method reflect.Value
	}
)

func (middlewareBuilder *AddMiddlewareBuilder) AddMiddleware(middleware func(next IHandler[T, T]) IHandler[T, T]) *AddMiddlewareBuilder {
	typedMiddlewareName := strings.TrimPrefix(runtime.FuncForPC(reflect.ValueOf(middleware).Pointer()).Name(), "*")

	middlewares, ok := middlewareBuilder.commandMiddlewares[middlewareBuilder.currentHandlerName]

	hMiddleware := handlerMiddleware{
		middlewareName: typedMiddlewareName,
		middleware:     middleware,
	}

	if !ok {
		middlewareBuilder.commandMiddlewares[middlewareBuilder.currentHandlerName] = []handlerMiddleware{hMiddleware}
		return middlewareBuilder
	}

	if !isMiddlewareRegisteredForHandler(&middlewares, typedMiddlewareName) {
		middlewareBuilder.commandMiddlewares[middlewareBuilder.currentHandlerName] =
			append(middlewareBuilder.commandMiddlewares[middlewareBuilder.currentHandlerName], hMiddleware)
	}
	return middlewareBuilder
}

func isMiddlewareRegisteredForHandler(middlewares *[]handlerMiddleware, middlewareName string) bool {
	for _, middleware := range *middlewares {
		if middleware.middlewareName == middlewareName {
			return true
		}
	}
	return false
}

func (r reflectiveHandler[T1, T2]) Handle(ctx context.Context, in T1) (out T, err error) {
	reflectResults := r.method.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(in)})
	result := reflectResults[0].Interface().(T2)
	if !reflectResults[1].IsNil() {
		err = reflectResults[1].Interface().(error)
	}

	return result, err
}

func createReflectiveHandler[T1 T](method reflect.Value) IHandler[T, T] {
	return reflectiveHandler[T, T1]{method: method}
}
