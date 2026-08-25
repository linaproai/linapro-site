// This file binds catalog-published dispatcher methods onto the per-service
// handler graph. The catalog owns the method list; this package only supplies
// existing domain handlers. Missing or orphan handlers fail registry construction.

package wasm

import (
	"context"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"

	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/plugin/pluginbridge/protocol/hostservices"
)

var (
	hostServiceDispatchRegistryOnce sync.Once
	hostServiceDispatchRegistryMemo *hostServiceDispatchRegistry
	errHostServiceDispatchRegistry  error
)

// hostServiceDispatchContext carries one authorized host-service invocation
// into a registered handler.
type hostServiceDispatchContext struct {
	hostContext *hostCallContext
	service     string
	method      string
	resourceRef string
	table       string
	payload     []byte
}

// hostServiceDispatchHandler dispatches one authorized host-service invocation.
type hostServiceDispatchHandler func(context.Context, hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope

// hostServiceDispatchRegistry stores explicitly registered host-service handlers.
type hostServiceDispatchRegistry struct {
	handlers map[string]hostServiceDispatchHandler
	methods  []hostServiceDispatchMethod
}

// hostServiceDispatchMethod describes one registered service/method pair.
type hostServiceDispatchMethod struct {
	service string
	method  string
}

func newEmptyHostServiceDispatchRegistry() *hostServiceDispatchRegistry {
	return &hostServiceDispatchRegistry{handlers: make(map[string]hostServiceDispatchHandler)}
}

func (r *hostServiceDispatchRegistry) register(service string, method string, handler hostServiceDispatchHandler) error {
	if r == nil {
		return gerror.New("host service dispatch registry is nil")
	}
	service = strings.TrimSpace(service)
	method = strings.TrimSpace(method)
	if service == "" || method == "" {
		return gerror.New("host service dispatch registration requires service and method")
	}
	if handler == nil {
		return gerror.Newf("host service dispatch handler is nil: %s.%s", service, method)
	}
	key := hostServiceDispatchRegistryKey(service, method)
	if _, ok := r.handlers[key]; ok {
		return gerror.Newf("host service dispatch handler already registered: %s.%s", service, method)
	}
	r.handlers[key] = handler
	r.methods = append(r.methods, hostServiceDispatchMethod{service: service, method: method})
	return nil
}

func (r *hostServiceDispatchRegistry) lookup(service string, method string) (hostServiceDispatchHandler, bool) {
	if r == nil {
		return nil, false
	}
	handler, ok := r.handlers[hostServiceDispatchRegistryKey(service, method)]
	return handler, ok
}

func (r *hostServiceDispatchRegistry) dispatch(
	ctx context.Context,
	input hostServiceDispatchContext,
) *bridgehostcall.HostCallResponseEnvelope {
	handler, ok := r.lookup(input.service, input.method)
	if !ok {
		return hostServiceDispatchNotFound(input.service, input.method)
	}
	return handler(ctx, input)
}

func hostServiceDispatchNotFound(service string, method string) *bridgehostcall.HostCallResponseEnvelope {
	return bridgehostcall.NewHostCallErrorResponse(
		bridgehostcall.HostCallStatusNotFound,
		"host service method not registered: "+strings.TrimSpace(service)+"."+strings.TrimSpace(method),
	)
}

func hostServiceDispatchRegistryKey(service string, method string) string {
	return strings.TrimSpace(service) + "\x00" + strings.TrimSpace(method)
}

func defaultHostServiceDispatchRegistry() (*hostServiceDispatchRegistry, error) {
	hostServiceDispatchRegistryOnce.Do(func() {
		hostServiceDispatchRegistryMemo, errHostServiceDispatchRegistry = newHostServiceDispatchRegistry()
	})
	return hostServiceDispatchRegistryMemo, errHostServiceDispatchRegistry
}

func newHostServiceDispatchRegistry() (*hostServiceDispatchRegistry, error) {
	return buildHostServiceDispatchRegistry(hostservices.PublishedDispatcherMethods(), coreHostServiceAdapters())
}

// buildHostServiceDispatchRegistry binds catalog-published dispatcher methods
// onto the per-service handler graph. A published method without an adapter, or
// an adapter whose service is not in the catalog, must fail.
func buildHostServiceDispatchRegistry(
	methods []hostservices.MethodDescriptor,
	adapters map[string]hostServiceDispatchAdapter,
) (*hostServiceDispatchRegistry, error) {
	registry := newEmptyHostServiceDispatchRegistry()
	used := make(map[string]struct{}, len(adapters))
	for _, method := range methods {
		if !method.Published || !method.Dispatcher {
			continue
		}
		adapter, ok := adapters[method.Service]
		if !ok {
			return nil, gerror.Newf("host service dispatch handler missing: %s.%s", method.Service, method.Method)
		}
		used[method.Service] = struct{}{}
		if err := registerHostServiceMethod(registry, method.Service, method.Method, adapter); err != nil {
			return nil, err
		}
	}
	for service := range adapters {
		if _, ok := used[service]; !ok {
			return nil, gerror.Newf("host service dispatch handler is orphan: %s", service)
		}
	}
	return registry, nil
}

func dispatchRegisteredHostService(
	ctx context.Context,
	hcc *hostCallContext,
	request *bridgehostservice.HostServiceRequestEnvelope,
) *bridgehostcall.HostCallResponseEnvelope {
	registry, err := defaultHostServiceDispatchRegistry()
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	if request == nil {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusInvalidRequest,
			"host service request is nil",
		)
	}
	return registry.dispatch(ctx, hostServiceDispatchContext{
		hostContext: hcc,
		service:     request.Service,
		method:      request.Method,
		resourceRef: request.ResourceRef,
		table:       request.Table,
		payload:     request.Payload,
	})
}

type hostServiceDispatchAdapter func(context.Context, *hostCallContext, hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope

func registerHostServiceMethod(
	registry *hostServiceDispatchRegistry,
	service string,
	method string,
	adapter hostServiceDispatchAdapter,
) error {
	if adapter == nil {
		return gerror.Newf("host service dispatch adapter is nil: %s.%s", service, method)
	}
	return registry.register(service, method, func(ctx context.Context, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		if input.hostContext == nil {
			return bridgehostcall.NewHostCallErrorResponse(
				bridgehostcall.HostCallStatusInternalError,
				"host service call context is missing",
			)
		}
		return adapter(ctx, input.hostContext, input)
	})
}

// coreHostServiceAdapters is the handler graph keyed by catalog service ID.
func coreHostServiceAdapters() map[string]hostServiceDispatchAdapter {
	return map[string]hostServiceDispatchAdapter{
		bridgehostservice.HostServiceRuntime: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchRuntimeHostService(ctx, hcc, input.method, input.payload)
		},
		bridgehostservice.HostServiceStorage: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchStorageHostService(ctx, hcc, input.resourceRef, input.method, input.payload)
		},
		bridgehostservice.HostServiceNetwork: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchNetworkHostService(ctx, hcc, input.resourceRef, input.method, input.payload)
		},
		bridgehostservice.HostServiceData: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchDataHostService(ctx, hcc, input.table, input.method, input.payload)
		},
		bridgehostservice.HostServiceCache: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchCacheHostService(ctx, hcc, input.resourceRef, input.method, input.payload)
		},
		bridgehostservice.HostServiceLock: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchLockHostService(ctx, hcc, input.resourceRef, input.method, input.payload)
		},
		bridgehostservice.HostServiceHostConfig: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchHostConfigService(ctx, hcc, input.resourceRef, input.method, input.payload)
		},
		bridgehostservice.HostServiceManifest: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchManifestHostService(ctx, hcc, input.resourceRef, input.method, input.payload)
		},
		bridgehostservice.HostServiceAPIDoc: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchAPIDocHostService(ctx, hcc, input.method, input.payload)
		},
		bridgehostservice.HostServiceAuth: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchAuthHostService(ctx, hcc, input.method, input.payload)
		},
		bridgehostservice.HostServiceUsers: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchUsersHostService(ctx, hcc, input.method, input.payload)
		},
		bridgehostservice.HostServiceBizCtx: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchBizCtxHostService(ctx, hcc, input.method, input.payload)
		},
		bridgehostservice.HostServiceDict: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchDictHostService(ctx, hcc, input.method, input.payload)
		},
		bridgehostservice.HostServiceFiles: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchFilesHostService(ctx, hcc, input.method, input.payload)
		},
		bridgehostservice.HostServiceJobs: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchJobsHostService(ctx, hcc, input.method, input.payload)
		},
		bridgehostservice.HostServiceNotifications: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchNotificationsHostService(ctx, hcc, input.resourceRef, input.method, input.payload)
		},
		bridgehostservice.HostServicePlugins: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchPluginsHostService(ctx, hcc, input.method, input.payload)
		},
		bridgehostservice.HostServiceRoute: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchRouteHostService(ctx, hcc, input.method, input.payload)
		},
		bridgehostservice.HostServiceSessions: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchSessionsHostService(ctx, hcc, input.method, input.payload)
		},
		bridgehostservice.HostServiceOrg: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchOrgHostService(ctx, hcc, input.method, input.payload)
		},
		bridgehostservice.HostServiceTenant: func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchTenantHostService(ctx, hcc, input.method, input.payload)
		},
	}
}
