// Package cluster provides one topology abstraction for single-node and
// clustered deployments.
package cluster

import (
	"context"
	"time"

	"lina-core/internal/service/coordination"
)

// Default election timing constants keep standalone construction deterministic
// when config values are absent.
const (
	defaultElectionLease         = 30 * time.Second
	defaultElectionRenewInterval = 10 * time.Second
)

// Service defines the cluster service contract.
type Service interface {
	// Start starts clustered primary-election infrastructure when cluster mode is enabled.
	// The call is a no-op for nil services, standalone deployments, or services
	// constructed without a coordination lock backend. It does not return
	// startup errors; callers observe election state through IsPrimary.
	Start(ctx context.Context)
	// Stop stops clustered primary-election infrastructure when it is running.
	// The call is idempotent and only affects the local election worker.
	Stop(ctx context.Context)
	// IsEnabled reports whether clustered deployment mode is enabled from the
	// normalized host configuration. Disabled mode keeps all coordination local.
	IsEnabled() bool
	// IsPrimary reports whether the current node should behave as the primary
	// node. Standalone deployments are always primary; clustered deployments
	// without an election backend return false to avoid split-primary work.
	IsPrimary() bool
	// NodeID returns the stable identifier of the current host node. A fallback
	// local identifier is returned when the service is nil or uninitialized.
	NodeID() string
}

// Interface compliance assertion for the default cluster service
// implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	cfg         *ClusterConfig   // cfg stores the normalized cluster settings.
	nodeID      string           // nodeID is the stable identifier of the current node.
	electionSvc *electionService // electionSvc participates in primary election for clustered mode.
}

// New creates one cluster topology service bound to topology config and the
// optional coordination backend. A nil coordination backend keeps election
// inactive in clustered mode and is the standalone constructor path. Callers
// must pass the startup-owned instance; this constructor does not select or
// parse a concrete coordination implementation.
func New(cfg *ClusterConfig, coordinationSvc coordination.Service) Service {
	normalizedCfg := normalizeClusterConfig(cfg)
	service := &serviceImpl{
		cfg:    normalizedCfg,
		nodeID: generateNodeIdentifier(),
	}
	if !normalizedCfg.Enabled {
		return service
	}

	if coordinationSvc == nil || coordinationSvc.Lock() == nil {
		return service
	}
	service.electionSvc = newElectionService(coordinationSvc.Lock(), &normalizedCfg.Election, service.nodeID)
	return service
}
