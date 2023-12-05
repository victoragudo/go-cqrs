package gocqrs

import (
	"context"
	"errors"
	"testing"
)

type commandHandlerTest struct {
	callbackFunc func(cmdRequestTest commandRequestTest) (commandResponseTest, error)
}
type commandRequestTest struct {
	data string
}
type commandResponseTest struct {
	data string
}

func (cmdHandlerTest *commandHandlerTest) Handle(ctx context.Context, cmdRequestTest commandRequestTest) (commandResponseTest, error) {
	return cmdHandlerTest.callbackFunc(cmdRequestTest)
}

func newCommandHandlerWrapperTest(f func(cmdRequestTest commandRequestTest) (commandResponseTest, error)) *commandHandlerTest {
	return &commandHandlerTest{
		callbackFunc: f,
	}
}

func TestAddCommandHandler(t *testing.T) {

	type args[CommandRequest T, CommandResponse T] struct {
		handler        ICommandHandler[CommandRequest, CommandResponse]
		context        context.Context
		commandRequest commandRequestTest
	}
	type expected struct {
		err error
	}
	type testCase[CommandRequest T, CommandResponse T] struct {
		name     string
		args     args[CommandRequest, CommandResponse]
		expected expected
	}
	tests := []testCase[commandRequestTest, commandResponseTest]{
		{
			name: "CommandHandler_SendCommand",
			args: args[commandRequestTest, commandResponseTest]{
				handler: newCommandHandlerWrapperTest(func(cmdRequestTest commandRequestTest) (commandResponseTest, error) {
					return commandResponseTest{
						data: cmdRequestTest.data,
					}, nil
				}),
				context: context.Background(),
				commandRequest: commandRequestTest{
					data: "data",
				},
			},
			expected: expected{
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddCommandHandler(tt.args.handler)
			commandResponse, err := SendCommand[commandResponseTest](tt.args.context, tt.args.commandRequest)

			if commandResponse.data != tt.args.commandRequest.data {
				t.Errorf("data are not equal")
			}

			if !errors.Is(err, tt.expected.err) {
				t.Errorf("returned error is not equal")
			}
		})
	}
}
