// This file defines cache host-service JSON payload types shared by guest
// clients and the WASM dispatcher. These structs travel inside the generic
// JSON envelope and must not grow dedicated binary marshal helpers.

package protocol

// Cache value kind constants describe the concrete payload representation
// carried in cache response snapshots.
const (
	// HostServiceCacheValueKindString identifies string cache values.
	HostServiceCacheValueKindString = 1
	// HostServiceCacheValueKindInt identifies integer cache values.
	HostServiceCacheValueKindInt = 2
)

// HostServiceCacheValue describes one governed cache value snapshot.
type HostServiceCacheValue struct {
	// ValueKind identifies whether the cache value is string or integer based.
	ValueKind int32 `json:"valueKind"`
	// Value is the canonical string representation of the cache value.
	Value string `json:"value"`
	// IntValue is the integer payload when ValueKind is integer.
	IntValue int64 `json:"intValue,omitempty"`
	// ExpireAt is the expiration time when the backend can report it.
	ExpireAt string `json:"expireAt,omitempty"`
}

// HostServiceCacheGetRequest carries one cache read request.
type HostServiceCacheGetRequest struct {
	// Key is the logical cache key inside the authorized namespace.
	Key string `json:"key"`
}

// HostServiceCacheGetResponse carries one cache read response.
type HostServiceCacheGetResponse struct {
	// Found reports whether the cache entry exists.
	Found bool `json:"found"`
	// Value is the cache value snapshot when Found is true.
	Value *HostServiceCacheValue `json:"value,omitempty"`
}

// HostServiceCacheSetRequest carries one cache write request.
type HostServiceCacheSetRequest struct {
	// Key is the logical cache key inside the authorized namespace.
	Key string `json:"key"`
	// Value is the string payload to store.
	Value string `json:"value"`
	// ExpireSeconds is the positive expiration duration in seconds.
	ExpireSeconds int64 `json:"expireSeconds,omitempty"`
}

// HostServiceCacheSetResponse carries one cache write response.
type HostServiceCacheSetResponse struct {
	// Value is the resulting cache value snapshot.
	Value *HostServiceCacheValue `json:"value,omitempty"`
}

// HostServiceCacheDeleteRequest carries one cache delete request.
type HostServiceCacheDeleteRequest struct {
	// Key is the logical cache key inside the authorized namespace.
	Key string `json:"key"`
}

// HostServiceCacheIncrRequest carries one cache integer increment request.
type HostServiceCacheIncrRequest struct {
	// Key is the logical cache key inside the authorized namespace.
	Key string `json:"key"`
	// Delta is the increment delta applied to the current integer value.
	Delta int64 `json:"delta,omitempty"`
	// ExpireSeconds is the positive expiration duration in seconds.
	ExpireSeconds int64 `json:"expireSeconds,omitempty"`
}

// HostServiceCacheIncrResponse carries one cache integer increment response.
type HostServiceCacheIncrResponse struct {
	// Value is the resulting cache value snapshot.
	Value *HostServiceCacheValue `json:"value,omitempty"`
}

// HostServiceCacheExpireRequest carries one cache expiration update request.
type HostServiceCacheExpireRequest struct {
	// Key is the logical cache key inside the authorized namespace.
	Key string `json:"key"`
	// ExpireSeconds is the positive new expiration duration in seconds.
	ExpireSeconds int64 `json:"expireSeconds,omitempty"`
}

// HostServiceCacheExpireResponse carries one cache expiration update response.
type HostServiceCacheExpireResponse struct {
	// Found reports whether the cache entry exists.
	Found bool `json:"found"`
	// ExpireAt is the updated expiration time when the entry exists.
	ExpireAt string `json:"expireAt,omitempty"`
}
