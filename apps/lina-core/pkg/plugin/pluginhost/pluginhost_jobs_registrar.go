// This file defines the public jobs registrar contract exposed to source
// plugins. Host-owned implementations live in the integration layer where they
// can directly reuse the owning services for enablement and topology decisions.

package pluginhost

import (
	"context"

	"lina-core/pkg/plugin/capability"
)

// JobHandler defines one plugin-owned scheduled job callback.
type JobHandler func(ctx context.Context) error

// JobScope identifies multi-node placement for one plugin-owned scheduled job.
type JobScope string

// Supported JobScope values for JobSpec.Scope.
const (
	// JobScopeMasterOnly limits execution to the primary (master) node.
	JobScopeMasterOnly JobScope = "master_only"
	// JobScopeAllNode allows execution on every host node.
	JobScopeAllNode JobScope = "all_node"
)

// String returns the canonical scope value.
func (s JobScope) String() string {
	return string(s)
}

// IsValid reports whether the scope is a supported non-empty value.
// The zero value is intentionally not valid here; JobSpec treats empty Scope
// as "use host default" without requiring a concrete enum constant.
func (s JobScope) IsValid() bool {
	switch s {
	case JobScopeMasterOnly, JobScopeAllNode:
		return true
	default:
		return false
	}
}

// JobConcurrency identifies the in-process / in-node overlap policy for one job.
type JobConcurrency string

// Supported JobConcurrency values for JobSpec.Concurrency.
const (
	// JobConcurrencySingleton skips overlapping executions of the same job.
	JobConcurrencySingleton JobConcurrency = "singleton"
	// JobConcurrencyParallel allows overlapping executions up to MaxConcurrency.
	JobConcurrencyParallel JobConcurrency = "parallel"
)

// String returns the canonical concurrency value.
func (c JobConcurrency) String() string {
	return string(c)
}

// IsValid reports whether the concurrency is a supported non-empty value.
// The zero value is intentionally not valid here; JobSpec treats empty
// Concurrency as "use host default" (singleton on both registration paths).
func (c JobConcurrency) IsValid() bool {
	switch c {
	case JobConcurrencySingleton, JobConcurrencyParallel:
		return true
	default:
		return false
	}
}

// JobSpec is one complete plugin-owned scheduled job registration.
//
// Required fields must be non-empty (after trim) or non-nil; optional fields may
// be left zero-value so the host applies stable defaults. Callers should set
// only the fields they care about and rely on defaults for the rest.
type JobSpec struct {
	// Pattern is required. It is the raw gcron expression for the job, for
	// example "@every 10s" or a six-field cron string such as "0 */20 * * * *".
	// Empty or whitespace-only values are rejected.
	Pattern string

	// Name is required. It is the stable plugin-local job identity used to
	// build the handler reference (plugin:<id>/jobs:<name>) and must stay
	// unique within the plugin. Empty or whitespace-only values are rejected.
	Name string

	// DisplayName is optional. It is the human-readable title shown in the
	// unified scheduled-job management UI. When empty, the host uses Name.
	DisplayName string

	// Description is optional. It explains the job purpose for operators in
	// the management UI. When empty, the host synthesizes a default description
	// from the owning plugin ID.
	Description string

	// Scope is optional. It selects multi-node placement using JobScope:
	//   - JobScopeMasterOnly: run only on the primary (master) node
	//   - JobScopeAllNode: run on every host node
	// When empty (zero value), the managed/sys_job projection path defaults to
	// all_node. Prefer JobScopeMasterOnly for exclusive background work
	// (pipelines, remote metadata sync) so multi-node deployments do not
	// duplicate load. Non-empty values must pass JobScope.IsValid().
	Scope JobScope

	// Concurrency is optional. It selects overlap handling when a previous
	// tick is still running, using JobConcurrency:
	//   - JobConcurrencySingleton: skip overlapping ticks (one in-flight run)
	//   - JobConcurrencyParallel: allow overlaps up to MaxConcurrency
	// When empty (zero value), the host defaults to singleton so long-running
	// work cannot stack concurrent executions of the same job. Non-empty
	// values must pass JobConcurrency.IsValid().
	Concurrency JobConcurrency

	// MaxConcurrency is optional. It caps parallel in-flight runs when
	// Concurrency is JobConcurrencyParallel. Zero or negative values use the
	// host default (1). When Concurrency is singleton (including the empty
	// default), the host forces MaxConcurrency to 1 and ignores larger values.
	MaxConcurrency int

	// Handler is required. It is the callback invoked on each scheduled tick
	// after host enablement, scope, and concurrency guards. Nil is rejected.
	Handler JobHandler
}

// JobsRegistrar exposes host job registration and node-role inspection for one plugin.
type JobsRegistrar interface {
	// Add registers one guarded scheduled job described by JobSpec.
	// Required JobSpec fields: Pattern, Name, Handler. All other fields are
	// optional and use host defaults when left empty or zero. See JobSpec
	// field comments for per-field required/optional semantics.
	Add(ctx context.Context, spec JobSpec) error
	// IsPrimaryNode reports whether the current host node is the primary node.
	IsPrimaryNode() bool
	// Services returns the host-published runtime services for source-plugin construction.
	Services() capability.Services
}
