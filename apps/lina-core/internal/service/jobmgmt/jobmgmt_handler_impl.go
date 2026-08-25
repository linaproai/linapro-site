// This file contains the in-memory job handler registry implementation,
// including normalization, duplicate detection, and change notifications.

package jobmgmt

import (
	"sort"
	"strings"

	jobhandlerv1 "lina-core/api/jobhandler/v1"
	"lina-core/pkg/bizerr"
)

// Register stores one handler definition and rejects duplicate refs.
func (r *handlerRegistry) Register(def HandlerDef) error {
	def.Ref = strings.TrimSpace(def.Ref)
	def.DisplayName = strings.TrimSpace(def.DisplayName)
	def.Description = strings.TrimSpace(def.Description)
	def.PluginID = strings.TrimSpace(def.PluginID)
	if def.Ref == "" {
		return bizerr.NewCode(CodeJobHandlerRefRequired)
	}
	if def.DisplayName == "" {
		return bizerr.NewCode(CodeJobHandlerDisplayNameRequired)
	}
	if def.Invoke == nil {
		return bizerr.NewCode(CodeJobHandlerCallbackRequired)
	}
	if !def.Source.IsValid() {
		return bizerr.NewCode(CodeJobHandlerSourceUnsupported)
	}
	if def.Source == jobhandlerv1.SourcePlugin && def.PluginID == "" {
		return bizerr.NewCode(CodeJobHandlerPluginIDRequired)
	}
	if def.Source == jobhandlerv1.SourceHost {
		def.PluginID = ""
	}

	schemaText, err := normalizeSchema(def.ParamsSchema)
	if err != nil {
		return err
	}
	def.ParamsSchema = schemaText

	r.mu.Lock()
	if _, exists := r.handlers[def.Ref]; exists {
		r.mu.Unlock()
		return bizerr.NewCode(CodeJobHandlerExists, bizerr.P("ref", def.Ref))
	}
	r.handlers[def.Ref] = def
	callbacks := r.snapshotCallbacksLocked()
	r.mu.Unlock()

	notifyCallbacks(callbacks, def.Ref, true)
	return nil
}

// Unregister removes one handler definition and notifies change observers.
func (r *handlerRegistry) Unregister(ref string) {
	trimmedRef := strings.TrimSpace(ref)
	if trimmedRef == "" {
		return
	}

	r.mu.Lock()
	if _, exists := r.handlers[trimmedRef]; !exists {
		r.mu.Unlock()
		return
	}
	delete(r.handlers, trimmedRef)
	callbacks := r.snapshotCallbacksLocked()
	r.mu.Unlock()

	notifyCallbacks(callbacks, trimmedRef, false)
}

// Lookup returns one registered handler definition by ref.
func (r *handlerRegistry) Lookup(ref string) (HandlerDef, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	def, ok := r.handlers[strings.TrimSpace(ref)]
	return def, ok
}

// List returns all registered handlers sorted by ref.
func (r *handlerRegistry) List() []HandlerInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]HandlerInfo, 0, len(r.handlers))
	for _, def := range r.handlers {
		items = append(items, HandlerInfo{
			Ref:          def.Ref,
			DisplayName:  def.DisplayName,
			Description:  def.Description,
			ParamsSchema: def.ParamsSchema,
			Source:       def.Source,
			PluginID:     def.PluginID,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Ref < items[j].Ref
	})
	return items
}

// SubscribeChanges registers one change callback and returns its unsubscribe function.
func (r *handlerRegistry) SubscribeChanges(callback ChangeCallback) func() {
	if callback == nil {
		return func() {}
	}

	r.mu.Lock()
	r.callbackID++
	currentID := r.callbackID
	r.callbacks[currentID] = callback
	r.mu.Unlock()

	return func() {
		r.mu.Lock()
		delete(r.callbacks, currentID)
		r.mu.Unlock()
	}
}

// snapshotCallbacksLocked clones all callbacks while the write lock is held.
func (r *handlerRegistry) snapshotCallbacksLocked() []ChangeCallback {
	callbacks := make([]ChangeCallback, 0, len(r.callbacks))
	for _, callback := range r.callbacks {
		callbacks = append(callbacks, callback)
	}
	return callbacks
}

// notifyCallbacks executes all registry change observers outside the registry lock.
func notifyCallbacks(callbacks []ChangeCallback, ref string, exists bool) {
	for _, callback := range callbacks {
		callback(ref, exists)
	}
}
