// This file collects plugin-owned job definitions into stable projection
// records so the host can surface them in scheduled-job management.

package integration

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/errors/gerror"
	pluginv1 "lina-core/api/plugin/v1"
	"lina-core/internal/service/jobmeta"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/plugin/pluginhost"
	"lina-core/pkg/statusflag"
	"strings"
	"time"
)

const (
	managedJobDefaultTimeout = 5 * time.Minute
)

// managedJobCollector captures plugin-owned job registrations instead of
// registering them directly into gcron.
type managedJobCollector struct {
	pluginID string
	services capability.Services
	items    []ManagedJob
}

// Ensure managedJobCollector satisfies the published registrar contract.
var _ pluginhost.JobsRegistrar = (*managedJobCollector)(nil)

// Add records one plugin-owned scheduled job definition for sys_job projection.
func (c *managedJobCollector) Add(ctx context.Context, spec pluginhost.JobSpec) error {
	if spec.Handler == nil {
		return gerror.New("plugin job handler cannot be nil")
	}

	var (
		trimmedPattern     = strings.TrimSpace(spec.Pattern)
		trimmedName        = strings.TrimSpace(spec.Name)
		trimmedDisplayName = strings.TrimSpace(spec.DisplayName)
		trimmedDescription = strings.TrimSpace(spec.Description)
		scope              = pluginhost.JobScope(strings.TrimSpace(spec.Scope.String()))
		concurrency        = pluginhost.JobConcurrency(strings.TrimSpace(spec.Concurrency.String()))
	)
	if trimmedPattern == "" {
		return gerror.New("plugin job expression cannot be empty")
	}
	if trimmedName == "" {
		return gerror.New("plugin job name cannot be empty")
	}
	if scope != "" && !scope.IsValid() {
		return gerror.Newf("plugin job scope is invalid: %s", scope)
	}
	if concurrency != "" && !concurrency.IsValid() {
		return gerror.Newf("plugin job concurrency is invalid: %s", concurrency)
	}
	if trimmedDisplayName == "" {
		trimmedDisplayName = trimmedName
	}
	if trimmedDescription == "" {
		trimmedDescription = fmt.Sprintf("Built-in scheduled job registered by plugin %s.", strings.TrimSpace(c.pluginID))
	}

	maxConcurrency := spec.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}

	// Preserve empty enum zero-values so buildPluginBuiltinJobs applies host
	// defaults; only pass non-empty validated constants into Normalize.
	projectedScope := jobmeta.JobScope("")
	if scope != "" {
		projectedScope = jobmeta.NormalizeJobScope(scope.String())
	}
	projectedConcurrency := jobmeta.JobConcurrency("")
	if concurrency != "" {
		projectedConcurrency = jobmeta.NormalizeJobConcurrency(concurrency.String())
	}

	c.items = append(c.items, ManagedJob{
		PluginID:       strings.TrimSpace(c.pluginID),
		Name:           trimmedName,
		DisplayName:    trimmedDisplayName,
		Description:    trimmedDescription,
		Pattern:        trimmedPattern,
		Timezone:       "Asia/Shanghai",
		Scope:          projectedScope,
		Concurrency:    projectedConcurrency,
		MaxConcurrency: maxConcurrency,
		Timeout:        managedJobDefaultTimeout,
		Handler:        spec.Handler,
	})
	return nil
}

// IsPrimaryNode reports a stable true value while collecting definitions so
// source plugins do not accidentally hide jobs from the unified registry view.
func (c *managedJobCollector) IsPrimaryNode() bool {
	return true
}

// Services returns the host-published runtime services for source-plugin construction.
func (c *managedJobCollector) Services() capability.Services {
	if c == nil {
		return nil
	}
	return c.services
}

// collectManagedJobs gathers source-plugin job definitions without
// registering them into gcron.
func (s *serviceImpl) collectManagedJobs(
	ctx context.Context,
	pluginID string,
) ([]ManagedJob, error) {
	return s.collectManagedJobsWithOptions(ctx, pluginID, collectManagedJobOptions{})
}

// collectDeclaredJobs gathers plugin-owned job declarations for management
// review without checking runtime business-entry enablement.
func (s *serviceImpl) collectDeclaredJobs(
	ctx context.Context,
	pluginID string,
) ([]ManagedJob, error) {
	return s.collectManagedJobsWithOptions(ctx, pluginID, collectManagedJobOptions{
		includeDisabledDynamic: true,
	})
}

// collectManagedJobsWithOptions gathers plugin-owned job definitions from
// matching source and dynamic plugins. Dynamic plugins declare built-in jobs
// through the Jobs-domain discovery entry point. When includeDisabledDynamic is
// true, dynamic discovery uses the current manifest before enablement; when
// installedOnly is true, uninstalled plugins are skipped so management previews
// do not leak into scheduled-job projections.
func (s *serviceImpl) collectManagedJobsWithOptions(
	ctx context.Context,
	pluginID string,
	options collectManagedJobOptions,
) ([]ManagedJob, error) {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return nil, err
	}

	result := make([]ManagedJob, 0)
	trimmedPluginID := strings.TrimSpace(pluginID)
	for _, manifest := range manifests {
		if manifest == nil {
			continue
		}
		if trimmedPluginID != "" && manifest.ID != trimmedPluginID {
			continue
		}
		registry, err := s.storeSvc.GetRegistry(ctx, manifest.ID)
		if err != nil {
			return nil, err
		}
		if options.installedOnly &&
			(registry == nil || registry.Installed != statusflag.Installed.Int()) {
			continue
		}
		sourceItems, err := s.collectSourceManagedJobs(ctx, manifest)
		if err != nil {
			return nil, err
		}
		result = append(result, sourceItems...)

		dynamicItems, err := s.collectDynamicManagedJobs(ctx, manifest, options)
		if err != nil {
			return nil, err
		}
		result = append(result, dynamicItems...)
	}
	return result, nil
}

// collectManagedJobOptions controls the declaration and executable discovery
// modes used by management previews, job projections, and handler publication.
type collectManagedJobOptions struct {
	includeDisabledDynamic bool
	installedOnly          bool
}

// collectSourceManagedJobs gathers source-plugin managed job registrations
// for one manifest.
func (s *serviceImpl) collectSourceManagedJobs(
	ctx context.Context,
	manifest *catalog.Manifest,
) ([]ManagedJob, error) {
	if manifest == nil {
		return nil, nil
	}
	sourcePlugin, ok := pluginhost.GetSourcePlugin(manifest.ID)
	if !ok || sourcePlugin == nil {
		return nil, nil
	}

	collector := &managedJobCollector{
		pluginID: manifest.ID,
		services: s.sourceServicesForPlugin(manifest.ID),
		items:    make([]ManagedJob, 0),
	}
	for _, registration := range sourcePlugin.GetJobRegistrars() {
		if registration == nil || registration.Handler == nil {
			continue
		}
		if err := registration.Handler(ctx, collector); err != nil {
			return nil, err
		}
	}
	return collector.items, nil
}

// collectDynamicManagedJobs gathers dynamic-plugin Jobs declarations from the
// runtime registration entry point and binds them to the shared executor.
func (s *serviceImpl) collectDynamicManagedJobs(
	ctx context.Context,
	manifest *catalog.Manifest,
	options collectManagedJobOptions,
) ([]ManagedJob, error) {
	if manifest == nil {
		return nil, nil
	}
	if plugintypes.NormalizeType(manifest.Type) != pluginv1.PluginTypeDynamic {
		return nil, nil
	}
	if !manifestDeclaresJobsRegister(manifest) {
		return nil, nil
	}
	if s.runtimeSvc == nil {
		return nil, gerror.Newf("dynamic plugin Jobs executor is not injected: %s", manifest.ID)
	}
	registry, err := s.storeSvc.GetRegistry(ctx, manifest.ID)
	if err != nil {
		return nil, err
	}
	if !options.includeDisabledDynamic {
		enabled, err := s.registryBusinessEntryEnabledForTenant(ctx, registry, manifest)
		if err != nil {
			return nil, err
		}
		if !enabled {
			return nil, nil
		}
	}

	contracts, err := s.runtimeSvc.DiscoverJobContracts(ctx, manifest)
	if err != nil {
		return nil, err
	}
	if len(contracts) == 0 {
		return nil, nil
	}

	items := make([]ManagedJob, 0, len(contracts))
	for _, contract := range contracts {
		if contract == nil {
			continue
		}
		contractSnapshot := *contract
		manifestSnapshot := manifest
		items = append(items, ManagedJob{
			PluginID:       strings.TrimSpace(manifest.ID),
			Name:           strings.TrimSpace(contractSnapshot.Name),
			DisplayName:    strings.TrimSpace(contractSnapshot.DisplayName),
			Description:    strings.TrimSpace(contractSnapshot.Description),
			Pattern:        strings.TrimSpace(contractSnapshot.Pattern),
			Timezone:       strings.TrimSpace(contractSnapshot.Timezone),
			Scope:          jobmeta.NormalizeJobScope(contractSnapshot.Scope.String()),
			Concurrency:    jobmeta.NormalizeJobConcurrency(contractSnapshot.Concurrency.String()),
			MaxConcurrency: contractSnapshot.MaxConcurrency,
			Timeout:        time.Duration(contractSnapshot.TimeoutSeconds) * time.Second,
			Handler: func(ctx context.Context) error {
				return s.runtimeSvc.ExecuteDeclaredJob(ctx, manifestSnapshot, &contractSnapshot)
			},
		})
	}
	return items, nil
}

// manifestDeclaresJobsRegister reports whether the manifest explicitly
// authorizes the Jobs-domain declaration method for dynamic job discovery.
func manifestDeclaresJobsRegister(manifest *catalog.Manifest) bool {
	if manifest == nil {
		return false
	}
	for _, service := range manifest.HostServices {
		if service == nil {
			continue
		}
		if strings.TrimSpace(service.Service) != protocol.HostServiceJobs {
			continue
		}
		methods := service.Methods
		if len(methods) == 0 {
			methods = []string{protocol.HostServiceMethodJobsBatchGet}
		}
		for _, method := range methods {
			if strings.TrimSpace(method) == protocol.HostServiceMethodJobsRegister {
				return true
			}
		}
	}
	return false
}

// ListExecutableJobs returns all plugin-owned job definitions that can be
// bound to executable handlers.
func (s *serviceImpl) ListExecutableJobs(ctx context.Context) ([]ManagedJob, error) {
	return s.collectManagedJobs(ctx, "")
}

// ListExecutableJobsByPlugin returns executable job definitions owned by
// one plugin. It is the narrow handler-publication path used by plugin
// lifecycle observers.
func (s *serviceImpl) ListExecutableJobsByPlugin(ctx context.Context, pluginID string) ([]ManagedJob, error) {
	return s.collectManagedJobs(ctx, pluginID)
}

// ListJobDeclarationsByPlugin returns job declarations owned by one plugin for
// management review, including dynamic plugins that are not yet installed or
// enabled.
func (s *serviceImpl) ListJobDeclarationsByPlugin(ctx context.Context, pluginID string) ([]ManagedJob, error) {
	return s.collectDeclaredJobs(ctx, pluginID)
}

// ListInstalledJobDeclarations returns job declarations from installed plugins
// so scheduled-job management can display disabled plugin jobs as paused while
// avoiding uninstalled authorization-preview-only plugins.
func (s *serviceImpl) ListInstalledJobDeclarations(ctx context.Context) ([]ManagedJob, error) {
	return s.collectManagedJobsWithOptions(ctx, "", collectManagedJobOptions{
		includeDisabledDynamic: true,
		installedOnly:          true,
	})
}
