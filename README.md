# A thread-safe Golang CQRS library using mediator pattern

[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](https://choosealicense.com/licenses/mit/)

GoCQRS is a Go package designed to facilitate the implementation of the Command Query Responsibility Segregation (CQRS) pattern in Go applications. This package provides a straightforward and type-safe way to handle commands, queries, and events within your application.

## Features

- **Type-Safe Handlers**: Utilizes Go generics to ensure type safety across commands, queries, and event handlers.
- **Concurrent Handler Management**: Uses `sync.Map` for managing handlers, ensuring safe concurrent access.
- **Easy Registration of Handlers**: Simplified functions to register command handlers, query handlers, and event handlers.
- **Generic Command and Query Processing**: Provides generic functions `SendCommand` and `SendQuery` for processing commands and queries, ensuring return types match the expected response types.
- **Event Publishing**: Facilitates the publishing of events to all registered handlers, handling errors gracefully.

## Usage

### Adding Handlers

- **Command Handlers**: Register command handlers using `AddCommandHandler`, specifying the command and response types.
- **Query Handlers**: Register query handlers using `AddQueryHandler`, specifying the query and response types.
- **Event Handlers**: Register one or more event handlers for a specific event type using `AddEventHandlers`.

### Sending Commands and Queries

- **SendCommand**: Execute a command and receive a response of the expected type.
- **SendQuery**: Execute a query and receive a response of the expected type.

### Publishing Events

- **PublishEvent**: Publish an event to all registered handlers, handling any errors that occur during the process.

## Installation

To use GoCQRS in your project, you can install it by running:

```
go get -u github.com/victoragudo/gocqrs
```

## Example of adding a command handler and dispatching the command
```golang
// Example of adding a command handler and sending a command
AddCommandHandler[YourCommandType, YourResponseType](yourCommandHandler)
response, err := SendCommand[YourResponseType](context.Background(), yourCommand)
```