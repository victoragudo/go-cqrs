package gocqrs

// PreMiddleware adds a pre-middleware to the current handler
func (middlewareBuilder *MiddlewareBuilder) PreMiddleware(middlewareFunc MiddlewareFunction) *MiddlewareBuilder {
	middlewareBuilder.middlewareRegistry.mutex.Lock()
	defer middlewareBuilder.middlewareRegistry.mutex.Unlock()

	// Get existing compiled middleware or create new one
	existingPreMiddleware, exists := middlewareBuilder.middlewareRegistry.preMiddlewares[middlewareBuilder.currentHandlerName]
	if !exists {
		existingPreMiddleware = CompiledMiddleware{
			preChain:  make([]MiddlewareFunction, 0),
			postChain: make([]MiddlewareFunction, 0),
		}
	}

	// Add the middleware function to the pre-chain
	existingPreMiddleware.preChain = append(existingPreMiddleware.preChain, middlewareFunc)
	middlewareBuilder.middlewareRegistry.preMiddlewares[middlewareBuilder.currentHandlerName] = existingPreMiddleware

	return middlewareBuilder
}

// PreMiddlewares adds multiple pre-middlewares to the current handler
func (middlewareBuilder *MiddlewareBuilder) PreMiddlewares(middlewareFunctions ...MiddlewareFunction) *MiddlewareBuilder {
	for _, middlewareFunc := range middlewareFunctions {
		middlewareBuilder.PreMiddleware(middlewareFunc)
	}
	return middlewareBuilder
}

// PostMiddleware adds a post-middleware to the current handler
func (middlewareBuilder *MiddlewareBuilder) PostMiddleware(middlewareFunc MiddlewareFunction) *MiddlewareBuilder {
	middlewareBuilder.middlewareRegistry.mutex.Lock()
	defer middlewareBuilder.middlewareRegistry.mutex.Unlock()

	// Get existing compiled middleware or create new one
	existingPostMiddleware, exists := middlewareBuilder.middlewareRegistry.postMiddlewares[middlewareBuilder.currentHandlerName]
	if !exists {
		existingPostMiddleware = CompiledMiddleware{
			preChain:  make([]MiddlewareFunction, 0),
			postChain: make([]MiddlewareFunction, 0),
		}
	}

	// Add the middleware function to the post-chain
	existingPostMiddleware.postChain = append(existingPostMiddleware.postChain, middlewareFunc)
	middlewareBuilder.middlewareRegistry.postMiddlewares[middlewareBuilder.currentHandlerName] = existingPostMiddleware

	return middlewareBuilder
}

// PostMiddlewares adds multiple post-middlewares to the current handler
func (middlewareBuilder *MiddlewareBuilder) PostMiddlewares(middlewareFunctions ...MiddlewareFunction) *MiddlewareBuilder {
	for _, middlewareFunc := range middlewareFunctions {
		middlewareBuilder.PostMiddleware(middlewareFunc)
	}
	return middlewareBuilder
}
