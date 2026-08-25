// This file defines lock host-service JSON payload types shared by guest
// clients and the WASM dispatcher. These structs travel inside the generic
// JSON envelope and must not grow dedicated binary marshal helpers.

package protocol

// HostServiceLockAcquireRequest carries one distributed lock acquire request.
type HostServiceLockAcquireRequest struct {
	// LeaseMillis is the requested lease duration in milliseconds.
	LeaseMillis int64 `json:"leaseMillis,omitempty"`
}

// HostServiceLockAcquireResponse carries one distributed lock acquire response.
type HostServiceLockAcquireResponse struct {
	// Acquired reports whether the lock was acquired successfully.
	Acquired bool `json:"acquired"`
	// Ticket is the opaque lock ticket when Acquired is true.
	Ticket string `json:"ticket,omitempty"`
	// ExpireAt is the next expiration time of the held lock.
	ExpireAt string `json:"expireAt,omitempty"`
}

// HostServiceLockRenewRequest carries one distributed lock renew request.
type HostServiceLockRenewRequest struct {
	// Ticket is the opaque lock ticket issued at acquire time.
	Ticket string `json:"ticket"`
}

// HostServiceLockRenewResponse carries one distributed lock renew response.
type HostServiceLockRenewResponse struct {
	// ExpireAt is the next expiration time of the renewed lock.
	ExpireAt string `json:"expireAt,omitempty"`
}

// HostServiceLockReleaseRequest carries one distributed lock release request.
type HostServiceLockReleaseRequest struct {
	// Ticket is the opaque lock ticket issued at acquire time.
	Ticket string `json:"ticket"`
}
