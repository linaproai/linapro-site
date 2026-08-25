// This file implements the governed structured data host service dispatcher.

package wasm

import (
	"context"
	"strings"

	"lina-core/internal/service/plugin/internal/datahost"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchDataHostService routes governed data service methods to the structured data host layer.
func dispatchDataHostService(
	ctx context.Context,
	hcc *hostCallContext,
	table string,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	if hcc == nil {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusInternalError,
			"host call context not available",
		)
	}
	serviceSpec := hcc.hostServiceSpec(bridgehostservice.HostServiceData)
	if serviceSpec == nil {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusCapabilityDenied,
			"data host service authorization snapshot not found",
		)
	}
	if !dataHostServiceTableOwnedByPlugin(hcc.pluginID, table) {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusCapabilityDenied,
			"data host service table is outside plugin namespace",
		)
	}
	resource, err := datahost.BuildCachedAuthorizedTableContract(ctx, datahost.AuthorizedTableContractInput{
		PluginID: hcc.pluginID,
		Table:    table,
		Methods:  serviceSpec.Methods,
	})
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}

	orgSvc := orgServiceForHostCall(hcc)
	switch method {
	case bridgehostservice.HostServiceMethodDataList:
		var request bridgehostservice.HostServiceDataListRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		response, callErr := datahost.ExecuteList(
			ctx,
			hcc.pluginID,
			table,
			hcc.executionSource,
			hcc.identity,
			orgSvc,
			resource,
			&request,
		)
		if callErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, callErr)
		}
		return capabilityJSONResponse(response)
	case bridgehostservice.HostServiceMethodDataGet:
		var request bridgehostservice.HostServiceDataGetRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		response, callErr := datahost.ExecuteGet(
			ctx,
			hcc.pluginID,
			table,
			hcc.executionSource,
			hcc.identity,
			orgSvc,
			resource,
			&request,
		)
		if callErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, callErr)
		}
		return capabilityJSONResponse(response)
	case bridgehostservice.HostServiceMethodDataBatchGet:
		var request bridgehostservice.HostServiceDataBatchGetRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		response, callErr := datahost.ExecuteBatchGet(
			ctx,
			hcc.pluginID,
			table,
			hcc.executionSource,
			hcc.identity,
			orgSvc,
			resource,
			&request,
		)
		if callErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, callErr)
		}
		return capabilityJSONResponse(response)
	case bridgehostservice.HostServiceMethodDataCreate:
		var request bridgehostservice.HostServiceDataMutationRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		response, callErr := datahost.ExecuteCreate(
			ctx,
			hcc.pluginID,
			table,
			hcc.executionSource,
			hcc.identity,
			orgSvc,
			resource,
			&request,
		)
		if callErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, callErr)
		}
		return capabilityJSONResponse(response)
	case bridgehostservice.HostServiceMethodDataUpdate:
		var request bridgehostservice.HostServiceDataMutationRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		response, callErr := datahost.ExecuteUpdate(
			ctx,
			hcc.pluginID,
			table,
			hcc.executionSource,
			hcc.identity,
			orgSvc,
			resource,
			&request,
		)
		if callErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, callErr)
		}
		return capabilityJSONResponse(response)
	case bridgehostservice.HostServiceMethodDataDelete:
		var request bridgehostservice.HostServiceDataMutationRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		response, callErr := datahost.ExecuteDelete(
			ctx,
			hcc.pluginID,
			table,
			hcc.executionSource,
			hcc.identity,
			orgSvc,
			resource,
			&request,
		)
		if callErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, callErr)
		}
		return capabilityJSONResponse(response)
	case bridgehostservice.HostServiceMethodDataTransaction:
		var request bridgehostservice.HostServiceDataTransactionRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		response, callErr := datahost.ExecuteTransaction(
			ctx,
			hcc.pluginID,
			table,
			hcc.executionSource,
			hcc.identity,
			orgSvc,
			resource,
			&request,
		)
		if callErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, callErr)
		}
		return capabilityJSONResponse(response)
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"unsupported data host service method: "+method,
		)
	}
}

func dataHostServiceTableOwnedByPlugin(pluginID string, table string) bool {
	normalizedPluginID := strings.NewReplacer("-", "_", ".", "_").Replace(strings.ToLower(strings.TrimSpace(pluginID)))
	normalizedTable := strings.ToLower(strings.TrimSpace(table))
	if normalizedPluginID == "" || normalizedTable == "" || strings.HasPrefix(normalizedTable, "sys_") {
		return false
	}
	ownedTable := "plugin_" + normalizedPluginID
	return normalizedTable == ownedTable || strings.HasPrefix(normalizedTable, ownedTable+"_")
}
