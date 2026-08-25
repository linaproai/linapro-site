// Package backend wires the marketplace source plugin into the host registry.
// It keeps compile-time registration in the backend package because plugin-full
// builds import each source plugin's backend registration package.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/pluginhost"
	marketplace "linapro-plugin-marketplace"
	marketctrl "linapro-plugin-marketplace/backend/internal/controller/market"
	marketplacesvc "linapro-plugin-marketplace/backend/internal/service/marketplace"
)

// init registers the marketplace plugin at compile time for builtin source
// plugin discovery. The panic boundary is limited to static registration, where
// a duplicate or invalid declaration must stop the host build/startup.
func init() {
	declarations := pluginhost.NewDeclarations(marketplace.PluginID)
	declarations.Assets().UseEmbeddedFiles(marketplace.EmbeddedFiles())

	if err := Register(declarations); err != nil {
		err = gerror.Wrap(err, "register marketplace source plugin backend")
		panic(err)
	}
	if err := pluginhost.RegisterSourcePlugin(declarations); err != nil {
		err = gerror.Wrap(err, "register marketplace source plugin")
		panic(err)
	}
}

// Register attaches marketplace backend contributions to the source-plugin
// declaration, including HTTP routes for marketplace catalog, publish, review,
// document, risk, and download endpoints, plus Git metadata discovery jobs.
func Register(declarations pluginhost.Declarations) error {
	if declarations == nil {
		return gerror.New("marketplace source plugin declarations cannot be nil")
	}
	if err := declarations.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerMarketplaceRoutes,
	); err != nil {
		return err
	}
	return declarations.Jobs().RegisterJobs(
		pluginhost.ExtensionPointJobsRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerMarketplaceJobs,
	)
}

// registerMarketplaceRoutes binds marketplace controllers under the plugin API prefix.
func registerMarketplaceRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	if registrar == nil {
		return gerror.New("marketplace HTTP registrar cannot be nil")
	}
	services := registrar.Services()
	marketSvc, err := marketplacesvc.New(nil, resolvePluginConfig(services))
	if err != nil {
		return gerror.Wrap(err, "create marketplace service")
	}
	controller := marketctrl.NewV1(marketSvc, resolveBizCtx(services))
	routes := registrar.Routes()
	middlewares := routes.Middlewares()
	routes.Group(routes.APIPrefix()+"/api/v1", func(group pluginhost.RouteGroup) {
		if middlewares != nil {
			group.Middleware(
				middlewares.NeverDoneCtx(),
				middlewares.HandlerResponse(),
				middlewares.CORS(),
				middlewares.RequestBodyLimit(),
				middlewares.Ctx(),
				middlewares.Auth(),
				middlewares.Tenancy(),
				middlewares.Permission(),
			)
		}
		group.Bind(controller)
	})
	return routes.Err()
}

// resolveBizCtx returns the host business-context service when available.
func resolveBizCtx(services capability.Services) bizctxcap.Service {
	if services == nil {
		return nil
	}
	return services.BizCtx()
}

// resolvePluginConfig returns the current plugin's scoped config reader when
// the host capability directory is bound. Nil is allowed for incomplete test
// registrars; Git discovery then skips platform token fallback.
func resolvePluginConfig(services capability.Services) plugincap.ConfigService {
	if services == nil || services.Plugins() == nil {
		return nil
	}
	return services.Plugins().Config()
}

// registerMarketplaceJobs registers async process pipeline and Git metadata jobs.
func registerMarketplaceJobs(ctx context.Context, registrar pluginhost.JobsRegistrar) error {
	if registrar == nil {
		return gerror.New("marketplace jobs registrar cannot be nil")
	}
	marketSvc, err := marketplacesvc.New(nil, resolvePluginConfig(registrar.Services()))
	if err != nil {
		return gerror.Wrap(err, "create marketplace service for jobs")
	}
	// Every 10 seconds on the primary node only. Singleton skips overlapping ticks
	// when a previous monorepo discovery batch is still running.
	if err = registrar.Add(ctx, pluginhost.JobSpec{
		Pattern:     "@every 10s",
		Name:        "linapro-plugin-marketplace-process-pipeline",
		DisplayName: "Marketplace process pipeline",
		Description: "Discover, verify, and auto-submit marketplace plugins waiting in the async process queue",
		Scope:       pluginhost.JobScopeMasterOnly,
		Concurrency: pluginhost.JobConcurrencySingleton,
		Handler: func(jobCtx context.Context) error {
			_, processErr := marketSvc.ProcessMarketplacePipeline(jobCtx)
			return processErr
		},
	}); err != nil {
		return err
	}
	// Every 20 minutes: poll already registered Git sources for new tags.
	// Master-only + singleton avoid multi-node Git API storms.
	return registrar.Add(ctx, pluginhost.JobSpec{
		Pattern:     "0 */20 * * * *",
		Name:        "linapro-plugin-marketplace-git-metadata-sync",
		DisplayName: "Marketplace Git metadata sync",
		Description: "Discover marketplace Git source tags and plugin.yaml metadata without cloning full repositories",
		Scope:       pluginhost.JobScopeMasterOnly,
		Concurrency: pluginhost.JobConcurrencySingleton,
		Handler: func(jobCtx context.Context) error {
			_, syncErr := marketSvc.DiscoverAllGitSources(jobCtx)
			return syncErr
		},
	})
}
