package gocqrs

import (
	"context"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

// MockCommandHandler implements ICommandHandler for testing purposes.
type MockCommandHandler struct{}

func (m *MockCommandHandler) Handle(ctx context.Context, command string) (string, error) {
	return "handled: " + command, nil
}

// Helper functions for assertions
func assertNilError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
}

func assertNotNilError(t *testing.T, err error) {
	if err == nil {
		t.Errorf("Expected an error, but got none")
	}
}

func assertEqual(t *testing.T, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}

// TestCommandHandlerWrapper_Handle tests the Handle method of commandHandlerWrapper.
func TestCommandHandlerWrapper_Handle(t *testing.T) {
	ctx := context.Background()
	command := "test command"
	mockHandler := &MockCommandHandler{}
	// Determine the type name of the handler parameter, removing the pointer symbol if present.
	typedHandlerName := strings.TrimPrefix(reflect.TypeOf(mockHandler).String(), "*")
	wrapper := newHandlerWrapper[string, string](mockHandler, typedHandlerName)

	// Success case
	got, err := wrapper.Handle(ctx, command)
	assertNilError(t, err)
	assertEqual(t, "handled: "+command, got)

	// Error case: incorrect type
	_, err = wrapper.Handle(ctx, 123) // 123 is int, not string
	assertNotNilError(t, err)
}

// TestAddCommandHandler tests the AddCommandHandler function.
func TestAddCommandHandler(t *testing.T) {
	mockHandler := &MockCommandHandler{}
	AddCommandHandler[string, string](mockHandler)

	// Verify if the handler was added correctly
	handler, ok := handlers["string"]
	if !ok {
		t.Fatal("Handler not found in commandHandlers")
	}

	// Assert that it's the correct type
	_, ok = handler.(*handlerWrapper[string, string])
	if !ok {
		t.Errorf("Handler is not of type *commandHandlerWrapper[string, string]")
	}
}

// TestSendCommand tests the SendCommand function.
func TestSendCommand(t *testing.T) {
	ctx := context.Background()
	command := "test command"
	AddCommandHandler[string, string](&MockCommandHandler{})

	// Success case
	response, err := SendCommand[string](ctx, command)
	assertNilError(t, err)
	assertEqual(t, "handled: "+command, response)

	// Error case: no registered handler
	assert.Panics(t, func() {
		_, err = SendCommand[int](ctx, 123)
	})
}

// MockEventHandler for events
type MockEventHandler struct{}

func newMockEventHandler() *MockEventHandler {
	return &MockEventHandler{}
}

func (m *MockEventHandler) Handle(ctx context.Context, event string) error {
	return nil
}

// TestPublishEvent tests the PublishEvent function.
func TestPublishEvent(t *testing.T) {
	ctx := context.Background()
	event := "test event"
	err := AddEventHandlers[string](newMockEventHandler())
	assertNilError(t, err)

	// Success case
	err = PublishEvent(ctx, event)
	assertNilError(t, err)

	// Error case: no registered handlers
	assert.Panics(t, func() {
		err = PublishEvent(ctx, 123) // 123 is int, a different type
	})
}

// TestSendCommand_Concurrency tests the SendCommand function for concurrent access.
func TestSendCommand_Concurrency(t *testing.T) {
	ctx := context.Background()
	command := "test command"
	AddCommandHandler[string, string](&MockCommandHandler{})

	// Number of goroutines to simulate concurrency.
	numGoroutines := 100

	// Channel to receive responses and errors from each goroutine.
	responses := make(chan string, numGoroutines)
	errors := make(chan error, numGoroutines)

	// Execute SendCommand in multiple goroutines concurrently.
	for i := 0; i < numGoroutines; i++ {
		go func() {
			response, err := SendCommand[string](ctx, command)
			if err != nil {
				errors <- err
				return
			}
			responses <- response
		}()
	}

	// Check the responses and errors.
	for i := 0; i < numGoroutines; i++ {
		select {
		case response := <-responses:
			expectedResponse := "handled: " + command
			if response != expectedResponse {
				t.Errorf("Expected response %v, got %v", expectedResponse, response)
			}
		case err := <-errors:
			t.Errorf("Received error: %v", err)
		}
	}
}

// TestSendQuery_Concurrency tests the SendQuery function for concurrent access.
func TestSendQuery_Concurrency(t *testing.T) {
	ctx := context.Background()
	query := "test query"
	AddQueryHandler[string, string](&MockQueryHandler{})

	// Number of goroutines to simulate concurrency.
	numGoroutines := 100

	// Channel to receive responses and errors from each goroutine.
	responses := make(chan string, numGoroutines)
	errors := make(chan error, numGoroutines)

	// Execute SendQuery in multiple goroutines concurrently.
	for i := 0; i < numGoroutines; i++ {
		go func() {
			response, err := SendQuery[string](ctx, query)
			if err != nil {
				errors <- err
				return
			}
			responses <- response
		}()
	}

	// Check the responses and errors.
	for i := 0; i < numGoroutines; i++ {
		select {
		case response := <-responses:
			expectedResponse := "handled: " + query
			if response != expectedResponse {
				t.Errorf("Expected response %v, got %v", expectedResponse, response)
			}
		case err := <-errors:
			t.Errorf("Received error: %v", err)
		}
	}
}

// MockQueryHandler implements IQueryHandler for testing purposes.
type MockQueryHandler struct{}

func (m *MockQueryHandler) Handle(ctx context.Context, query string) (string, error) {
	return "handled: " + query, nil
}

// TestPublishEvent_Concurrency tests the PublishEvent function with multiple handlers under concurrent conditions.
func TestPublishEvent_Concurrency(t *testing.T) {
	ctx := context.Background()
	event := "test event"

	// Number of event handlers and goroutines to simulate concurrency.
	numHandlers := 10
	numGoroutines := 100

	// Create and register multiple event handlers
	for i := 0; i < numHandlers; i++ {
		err := AddEventHandlers[string](newMockEventHandler())
		assertNilError(t, err)
	}

	// Channel to receive errors from each goroutine.
	errors := make(chan error, numGoroutines)

	// Execute PublishEvent in multiple goroutines concurrently.
	for i := 0; i < numGoroutines; i++ {
		go func() {
			err := PublishEvent(ctx, event)
			if err != nil {
				errors <- err
			} else {
				errors <- nil
			}
		}()
	}

	// Check for any errors received.
	for i := 0; i < numGoroutines; i++ {
		err := <-errors
		if err != nil {
			t.Errorf("Received error: %v", err)
		}
	}
}
