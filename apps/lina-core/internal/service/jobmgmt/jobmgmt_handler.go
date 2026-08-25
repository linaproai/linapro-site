// This file stores in-memory handler-registry state and adapts that registry
// to the internal persistent scheduler without creating an import cycle.

package jobmgmt

import (
	"context"
	"encoding/json"
	"sync"

	"lina-core/pkg/bizerr"
)

// handlerRegistry implements the in-memory handler registry.
type handlerRegistry struct {
	mu         sync.RWMutex           // mu protects handler and callback state.
	handlers   map[string]HandlerDef  // handlers stores definitions by stable ref.
	callbackID int                    // callbackID allocates deterministic observer keys.
	callbacks  map[int]ChangeCallback // callbacks stores registry observers.
}

// schedulerHandlerRuntime adapts Registry to the internal scheduler lookup
// surface. The nested scheduler package cannot import jobmgmt.
type schedulerHandlerRuntime struct {
	registry Registry // registry is the production handler registry.
}

// Lookup returns the invocation callback and parameter schema for one ref.
func (a schedulerHandlerRuntime) Lookup(
	ref string,
) (invoke func(ctx context.Context, params json.RawMessage) (any, error), paramsSchema string, ok bool) {
	if a.registry == nil {
		return nil, "", false
	}
	def, found := a.registry.Lookup(ref)
	if !found {
		return nil, "", false
	}
	return def.Invoke, def.ParamsSchema, true
}

// ValidateHandlerParams validates one parameter payload against a handler schema.
func (a schedulerHandlerRuntime) ValidateHandlerParams(schemaText string, paramsJSON json.RawMessage) error {
	return ValidateParams(schemaText, paramsJSON)
}

// HandlerNotFoundError returns the stable missing-handler business error.
func (a schedulerHandlerRuntime) HandlerNotFoundError() error {
	return bizerr.NewCode(CodeJobHandlerNotFound)
}
