// This file defines structured data host-service JSON payload types shared by
// guest clients and the WASM dispatcher. These structs travel inside the
// generic JSON envelope and must not grow dedicated binary marshal helpers.

package protocol

// HostServiceDataListRequest carries one governed paged list request.
type HostServiceDataListRequest struct {
	// PlanJSON is the JSON-encoded typed query plan used by pluginbridge/recordstore.
	PlanJSON []byte `json:"planJson,omitempty"`
}

// HostServiceDataListResponse carries one governed paged list response.
type HostServiceDataListResponse struct {
	// Records contains one JSON document per returned row.
	Records [][]byte `json:"records,omitempty"`
	// Total is the total number of matching rows before pagination.
	Total int32 `json:"total,omitempty"`
}

// HostServiceDataGetRequest carries one governed detail query by key.
type HostServiceDataGetRequest struct {
	// PlanJSON is the JSON-encoded typed query plan used by pluginbridge/recordstore.
	PlanJSON []byte `json:"planJson,omitempty"`
}

// HostServiceDataGetResponse carries one governed detail response.
type HostServiceDataGetResponse struct {
	// Found reports whether one matching row exists inside the current governance boundary.
	Found bool `json:"found"`
	// RecordJSON is the JSON-encoded record when Found is true.
	RecordJSON []byte `json:"recordJson,omitempty"`
}

// HostServiceDataBatchGetRequest carries one governed detail query by key set.
type HostServiceDataBatchGetRequest struct {
	// KeyJSON contains JSON-encoded key values.
	KeyJSON [][]byte `json:"keyJson,omitempty"`
	// Fields contains an optional logical field projection.
	Fields []string `json:"fields,omitempty"`
}

// HostServiceDataBatchGetResponse carries one governed detail batch response.
type HostServiceDataBatchGetResponse struct {
	// Records contains one JSON document per found row.
	Records [][]byte `json:"records,omitempty"`
	// MissingKeyJSON contains requested keys with no visible returned record.
	MissingKeyJSON [][]byte `json:"missingKeyJson,omitempty"`
}

// HostServiceDataMutationRequest carries one governed create/update/delete request.
type HostServiceDataMutationRequest struct {
	// KeyJSON is the JSON-encoded key value for update/delete.
	KeyJSON []byte `json:"keyJson,omitempty"`
	// RecordJSON is the JSON-encoded input document for create/update.
	RecordJSON []byte `json:"recordJson,omitempty"`
}

// HostServiceDataMutationResponse carries one governed mutation response.
type HostServiceDataMutationResponse struct {
	// AffectedRows is the number of rows affected by the mutation.
	AffectedRows int64 `json:"affectedRows,omitempty"`
	// KeyJSON is the JSON-encoded resource key returned after create/update when available.
	KeyJSON []byte `json:"keyJson,omitempty"`
	// RecordJSON is the optional JSON-encoded record snapshot returned by the host.
	RecordJSON []byte `json:"recordJson,omitempty"`
}

// HostServiceDataTransactionOperation carries one structured mutation step inside a transaction.
type HostServiceDataTransactionOperation struct {
	// Method is one structured data mutation method such as create/update/delete.
	Method string `json:"method"`
	// KeyJSON is the JSON-encoded resource key used by update/delete.
	KeyJSON []byte `json:"keyJson,omitempty"`
	// RecordJSON is the JSON-encoded input document used by create/update.
	RecordJSON []byte `json:"recordJson,omitempty"`
}

// HostServiceDataTransactionRequest carries one governed transaction request.
type HostServiceDataTransactionRequest struct {
	// Operations is the ordered list of mutation steps executed atomically.
	Operations []*HostServiceDataTransactionOperation `json:"operations,omitempty"`
}

// HostServiceDataTransactionResponse carries one governed transaction result summary.
type HostServiceDataTransactionResponse struct {
	// Results is the ordered list of per-step mutation results.
	Results []*HostServiceDataMutationResponse `json:"results,omitempty"`
	// AffectedRows is the aggregate affected row count across all steps.
	AffectedRows int64 `json:"affectedRows,omitempty"`
}
