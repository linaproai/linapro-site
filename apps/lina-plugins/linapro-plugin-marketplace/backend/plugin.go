// Package backend wires the marketplace source plugin into the host registry.
// It keeps compile-time registration in the backend package because plugin-full
// builds import each source plugin's backend registration package.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/bizctxcap"
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
		panic(gerror.Wrap(err, "register marketplace source plugin backend"))
	}
	if err := pluginhost.RegisterSourcePlugin(declarations); err != nil {
		panic(gerror.Wrap(err, "register marketplace source plugin"))
	}
}

// Register attaches marketplace backend contributions to the source-plugin
// declaration, including HTTP routes for marketplace catalog, publish, review,
// document, risk, and download endpoints.
func Register(declarations pluginhost.Declarations) error {
	if declarations == nil {
		return gerror.New("marketplace source plugin declarations cannot be nil")
	}
	return declarations.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerMarketplaceRoutes,
	)
}

// registerMarketplaceRoutes binds marketplace controllers under the plugin API prefix.
func registerMarketplaceRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	if registrar == nil {
		return gerror.New("marketplace HTTP registrar cannot be nil")
	}
	services := registrar.Services()
	marketSvc, err := marketplacesvc.New(nil)
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
