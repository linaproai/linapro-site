// This file defines storage host-service JSON payload types shared by guest
// clients and the WASM dispatcher. These structs travel inside the generic
// JSON envelope and must not grow dedicated binary marshal helpers.

package protocol

// HostServiceStorageObject describes one governed storage object snapshot.
type HostServiceStorageObject struct {
	// Path is the logical object path relative to the storage resource root.
	Path string `json:"path"`
	// Size is the current object size in bytes.
	Size int64 `json:"size"`
	// ContentType is the normalized MIME type for the object.
	ContentType string `json:"contentType,omitempty"`
	// UpdatedAt is the host-side last update timestamp.
	UpdatedAt string `json:"updatedAt,omitempty"`
	// Visibility records the configured resource visibility policy.
	Visibility string `json:"visibility,omitempty"`
}

// HostServiceStoragePutRequest carries one governed storage write request.
type HostServiceStoragePutRequest struct {
	// Path is the logical target path relative to the resource root.
	Path string `json:"path"`
	// Body is the raw object payload.
	Body []byte `json:"body,omitempty"`
	// ContentType is the optional MIME type hint supplied by the guest.
	ContentType string `json:"contentType,omitempty"`
	// Overwrite requests replacement of an existing object at the same path.
	Overwrite bool `json:"overwrite,omitempty"`
}

// HostServiceStoragePutResponse carries storage metadata after a successful write.
type HostServiceStoragePutResponse struct {
	// Object is the resulting object metadata snapshot.
	Object *HostServiceStorageObject `json:"object,omitempty"`
}

// HostServiceStoragePutInitRequest starts one governed storage upload session.
type HostServiceStoragePutInitRequest struct {
	// Path is the logical target path relative to the resource root.
	Path string `json:"path"`
	// ContentType is the optional MIME type hint supplied by the guest.
	ContentType string `json:"contentType,omitempty"`
	// Overwrite requests replacement of an existing object at the same path.
	Overwrite bool `json:"overwrite,omitempty"`
}

// HostServiceStoragePutInitResponse identifies the started upload session.
type HostServiceStoragePutInitResponse struct {
	// UploadID is the opaque host-issued upload session identifier.
	UploadID string `json:"uploadId"`
}

// HostServiceStoragePutChunkRequest appends one chunk to an upload session.
type HostServiceStoragePutChunkRequest struct {
	// Path is the final logical target path bound to the upload session.
	Path string `json:"path"`
	// UploadID identifies the upload session created by put.init.
	UploadID string `json:"uploadId"`
	// Offset is the zero-based byte offset expected for this chunk.
	Offset int64 `json:"offset"`
	// Body is the chunk payload.
	Body []byte `json:"body,omitempty"`
}

// HostServiceStoragePutChunkResponse acknowledges the next expected offset.
type HostServiceStoragePutChunkResponse struct {
	// NextOffset is the next byte offset expected by the host.
	NextOffset int64 `json:"nextOffset"`
}

// HostServiceStoragePutCommitRequest commits one upload session.
type HostServiceStoragePutCommitRequest struct {
	// Path is the final logical target path bound to the upload session.
	Path string `json:"path"`
	// UploadID identifies the upload session created by put.init.
	UploadID string `json:"uploadId"`
	// Size is the total object size observed by the guest.
	Size int64 `json:"size"`
}

// HostServiceStoragePutCommitResponse carries storage metadata after commit.
type HostServiceStoragePutCommitResponse struct {
	// Object is the resulting object metadata snapshot.
	Object *HostServiceStorageObject `json:"object,omitempty"`
}

// HostServiceStoragePutAbortRequest aborts one upload session.
type HostServiceStoragePutAbortRequest struct {
	// Path is the final logical target path bound to the upload session.
	Path string `json:"path"`
	// UploadID identifies the upload session created by put.init.
	UploadID string `json:"uploadId"`
}

// HostServiceStorageGetRequest carries one governed storage read request.
type HostServiceStorageGetRequest struct {
	// Path is the logical object path relative to the resource root.
	Path string `json:"path"`
}

// HostServiceStorageGetResponse carries one governed storage read response.
type HostServiceStorageGetResponse struct {
	// Found reports whether the requested object currently exists.
	Found bool `json:"found"`
	// Object is the metadata snapshot when the object exists.
	Object *HostServiceStorageObject `json:"object,omitempty"`
	// Body is the raw object payload when the object exists.
	Body []byte `json:"body,omitempty"`
}

// HostServiceStorageDeleteRequest carries one governed storage delete request.
type HostServiceStorageDeleteRequest struct {
	// Path is the logical object path relative to the resource root.
	Path string `json:"path"`
}

// HostServiceStorageListRequest carries one governed storage list request.
type HostServiceStorageListRequest struct {
	// Prefix restricts the result set to one logical object prefix.
	Prefix string `json:"prefix,omitempty"`
	// Limit caps the number of returned objects.
	Limit uint32 `json:"limit,omitempty"`
}

// HostServiceStorageListResponse carries one governed storage list response.
type HostServiceStorageListResponse struct {
	// Objects is the ordered list of visible object metadata snapshots.
	Objects []*HostServiceStorageObject `json:"objects,omitempty"`
}

// HostServiceStorageStatRequest carries one governed storage stat request.
type HostServiceStorageStatRequest struct {
	// Path is the logical object path relative to the resource root.
	Path string `json:"path"`
}

// HostServiceStorageStatResponse carries one governed storage stat response.
type HostServiceStorageStatResponse struct {
	// Found reports whether the requested object currently exists.
	Found bool `json:"found"`
	// Object is the metadata snapshot when the object exists.
	Object *HostServiceStorageObject `json:"object,omitempty"`
}
