// This file owns the source-plugin scheduled-job registrar implementation used
// by integration startup. Keeping it in integration lets runtime guards reuse
// plugin enablement and cluster topology services directly.

package integration

import (
	"context"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gcron"

	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/pluginhost"
)

// sourceJobRegistrar registers source-plugin scheduled jobs into GoFrame cron.
type sourceJobRegistrar struct {
	pluginID string
	service  *serviceImpl
	services capability.Services

	// runningByName implements process-local singleton guards for jobs that
	// declare JobConcurrencySingleton (or leave concurrency empty, which
	// defaults to singleton).
	runningMu     sync.Mutex
	runningByName map[string]bool
}

// Ensure sourceJobRegistrar satisfies the published registrar contract.
var _ pluginhost.JobsRegistrar = (*sourceJobRegistrar)(nil)

// newSourceJobRegistrar creates one host-owned jobs registrar for a source plugin.
func newSourceJobRegistrar(pluginID string, service *serviceImpl) pluginhost.JobsRegistrar {
	normalizedPluginID := strings.TrimSpace(pluginID)
	var services capability.Services
	if service != nil {
		services = service.sourceServicesForPlugin(normalizedPluginID)
	}
	return &sourceJobRegistrar{
		pluginID:      normalizedPluginID,
		service:       service,
		services:      services,
		runningByName: make(map[string]bool),
	}
}

// Add registers one guarded scheduled job described by JobSpec on the direct
// gcron path, applying master-only and singleton guards when requested.
func (r *sourceJobRegistrar) Add(ctx context.Context, spec pluginhost.JobSpec) error {
	if spec.Handler == nil {
		return gerror.New("pluginhost: job handler is nil")
	}

	jobName := strings.TrimSpace(spec.Name)
	pattern := strings.TrimSpace(spec.Pattern)
	if pattern == "" {
		return gerror.New("pluginhost: job pattern is empty")
	}
	if jobName == "" {
		return gerror.New("pluginhost: job name is empty")
	}

	scope := pluginhost.JobScope(strings.TrimSpace(spec.Scope.String()))
	concurrency := pluginhost.JobConcurrency(strings.TrimSpace(spec.Concurrency.String()))
	if scope != "" && !scope.IsValid() {
		return gerror.Newf("pluginhost: job scope is invalid: %s", scope)
	}
	if concurrency != "" && !concurrency.IsValid() {
		return gerror.Newf("pluginhost: job concurrency is invalid: %s", concurrency)
	}
	// Empty concurrency defaults to singleton so long-running ticks cannot stack.
	if concurrency == "" {
		concurrency = pluginhost.JobConcurrencySingleton
	}
	handler := spec.Handler

	_, err := gcron.Add(ctx, pattern, func(jobCtx context.Context) {
		if !r.canRun(jobCtx) {
			return
		}
		if scope == pluginhost.JobScopeMasterOnly && !r.IsPrimaryNode() {
			return
		}
		if concurrency == pluginhost.JobConcurrencySingleton {
			if !r.tryBeginSingleton(jobName) {
				return
			}
			defer r.endSingleton(jobName)
		}
		// Protect every scheduled-job callback at runtime so disabling a plugin
		// immediately stops future executions without requiring host restart or
		// plugin re-registration.
		if runErr := handler(jobCtx); runErr != nil {
			logger.Warningf(jobCtx, "plugin job failed plugin=%s name=%s err=%v", r.pluginID, jobName, runErr)
		}
	}, jobName)
	return err
}

// tryBeginSingleton marks one job as running when no prior run is active.
func (r *sourceJobRegistrar) tryBeginSingleton(name string) bool {
	if r == nil {
		return false
	}
	r.runningMu.Lock()
	defer r.runningMu.Unlock()
	if r.runningByName == nil {
		r.runningByName = make(map[string]bool)
	}
	if r.runningByName[name] {
		return false
	}
	r.runningByName[name] = true
	return true
}

// endSingleton clears the process-local running mark for one job.
func (r *sourceJobRegistrar) endSingleton(name string) {
	if r == nil {
		return
	}
	r.runningMu.Lock()
	defer r.runningMu.Unlock()
	if r.runningByName == nil {
		return
	}
	delete(r.runningByName, name)
}

// IsPrimaryNode reports whether the current host node is the primary node.
func (r *sourceJobRegistrar) IsPrimaryNode() bool {
	if r == nil || r.service == nil {
		return false
	}
	return r.service.isPrimaryNode()
}

// Services returns the host-published runtime services for source-plugin construction.
func (r *sourceJobRegistrar) Services() capability.Services {
	if r == nil {
		return nil
	}
	return r.services
}

// canRun reports whether the owning plugin may execute background work.
func (r *sourceJobRegistrar) canRun(ctx context.Context) bool {
	if r == nil || r.service == nil {
		return false
	}
	return r.service.canExposePluginBusinessEntries(ctx, r.pluginID)
}
