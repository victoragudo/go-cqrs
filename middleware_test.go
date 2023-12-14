package gocqrs

import (
	"context"
	"reflect"
	"testing"
)

func TestAddMiddlewareBuilder_PostMiddleware(t *testing.T) {
	type fields struct {
		currentHandlerName string
		preMiddlewares     map[string][]middlewareStruct
		postMiddlewares    map[string][]middlewareStruct
	}
	type args struct {
		m MiddlewareFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *AddMiddlewareBuilder
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middlewareBuilder := &AddMiddlewareBuilder{
				currentHandlerName: tt.fields.currentHandlerName,
				preMiddlewares:     tt.fields.preMiddlewares,
				postMiddlewares:    tt.fields.postMiddlewares,
			}
			if got := middlewareBuilder.PostMiddleware(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PostMiddleware() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddMiddlewareBuilder_PreMiddleware(t *testing.T) {
	type fields struct {
		currentHandlerName string
		preMiddlewares     map[string][]middlewareStruct
		postMiddlewares    map[string][]middlewareStruct
	}
	type args struct {
		m MiddlewareFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *AddMiddlewareBuilder
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middlewareBuilder := &AddMiddlewareBuilder{
				currentHandlerName: tt.fields.currentHandlerName,
				preMiddlewares:     tt.fields.preMiddlewares,
				postMiddlewares:    tt.fields.postMiddlewares,
			}
			if got := middlewareBuilder.PreMiddleware(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PreMiddleware() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddMiddlewareBuilder_executePostMiddlewares(t *testing.T) {
	type fields struct {
		currentHandlerName string
		preMiddlewares     map[string][]middlewareStruct
		postMiddlewares    map[string][]middlewareStruct
	}
	type args struct {
		ctx         context.Context
		request     T
		handlerName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middlewareBuilder := &AddMiddlewareBuilder{
				currentHandlerName: tt.fields.currentHandlerName,
				preMiddlewares:     tt.fields.preMiddlewares,
				postMiddlewares:    tt.fields.postMiddlewares,
			}
			middlewareBuilder.executePostMiddlewares(tt.args.ctx, tt.args.request, tt.args.handlerName)
		})
	}
}

func TestAddMiddlewareBuilder_executePreMiddlewares(t *testing.T) {
	type fields struct {
		currentHandlerName string
		preMiddlewares     map[string][]middlewareStruct
		postMiddlewares    map[string][]middlewareStruct
	}
	type args struct {
		ctx         context.Context
		request     T
		handlerName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   T
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middlewareBuilder := &AddMiddlewareBuilder{
				currentHandlerName: tt.fields.currentHandlerName,
				preMiddlewares:     tt.fields.preMiddlewares,
				postMiddlewares:    tt.fields.postMiddlewares,
			}
			if got := middlewareBuilder.executePreMiddlewares(tt.args.ctx, tt.args.request, tt.args.handlerName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("executePreMiddlewares() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isMiddlewareRegisteredForHandler(t *testing.T) {
	type args struct {
		middlewares    *[]middlewareStruct
		middlewareName string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMiddlewareRegisteredForHandler(tt.args.middlewares, tt.args.middlewareName); got != tt.want {
				t.Errorf("isMiddlewareRegisteredForHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
