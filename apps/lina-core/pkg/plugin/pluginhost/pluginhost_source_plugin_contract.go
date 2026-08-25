// This file defines the published source-plugin interfaces and grouped
// registration facades exposed to source-plugin authors.

package pluginhost

import (
	"io/fs"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/authcap/extlogin/extidspi"
	"lina-core/pkg/plugin/capability/capregistry"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// ID returns the stable plugin identifier declared by the source plugin.
func (p *sourcePlugin) ID() string {
	if p == nil {
		return ""
	}
	return p.id
}

// Assets returns the plugin asset registration facade.
func (p *sourcePlugin) Assets() AssetDeclarations {
	if p == nil {
		return nil
	}
	return p
}

// Lifecycle returns the plugin lifecycle callback registration facade.
func (p *sourcePlugin) Lifecycle() LifecycleDeclarations {
	if p == nil {
		return nil
	}
	return p
}

// Hooks returns the event-hook registration facade.
func (p *sourcePlugin) Hooks() HookDeclarations {
	if p == nil {
		return nil
	}
	return p
}

// HTTP returns the HTTP registration facade.
func (p *sourcePlugin) HTTP() HTTPDeclarations {
	if p == nil {
		return nil
	}
	return p
}

// Jobs returns the scheduled-job registration facade.
func (p *sourcePlugin) Jobs() JobDeclarations {
	if p == nil {
		return nil
	}
	return p
}

// Providers returns the framework capability provider declaration facade.
func (p *sourcePlugin) Providers() ProviderDeclarations {
	if p == nil {
		return nil
	}
	return p
}

// Access returns the menu and permission access-control registration facade.
func (p *sourcePlugin) Access() AccessDeclarations {
	if p == nil {
		return nil
	}
	return p
}

// ProvideTenant declares one source-plugin tenant provider factory.
func (p *sourcePlugin) ProvideTenant(factory tenantspi.ProviderFactory) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin provider facade is nil")
	}
	return p.registerTenantProvider(factory)
}

// ProvideOrg declares one source-plugin organization provider factory.
func (p *sourcePlugin) ProvideOrg(factory orgspi.ProviderFactory) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin provider facade is nil")
	}
	return p.registerOrgProvider(factory)
}

// ProvideCapability declares one plugin-owned capability descriptor.
func (p *sourcePlugin) ProvideCapability(descriptor capregistry.Descriptor) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin provider facade is nil")
	}
	return p.registerCapabilityDescriptor(descriptor)
}

// ProvideExternalIdentity declares one source-plugin external-identity provider ID.
func (p *sourcePlugin) ProvideExternalIdentity(providerID string) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin provider facade is nil")
	}
	return p.registerExternalIdentityProvider(providerID)
}

// ProvideExternalIdentityProvider declares this source plugin's external-identity
// provider engine factory (linapro-extlogin-core). It is distinct from
// ProvideExternalIdentity, which stamps provider-ID ownership for calling
// plugins. A plugin that supplies the engine need not own any provider ID.
func (p *sourcePlugin) ProvideExternalIdentityProvider(factory extidspi.ProviderFactory) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin provider facade is nil")
	}
	return p.registerExternalIdentityProviderFactory(factory)
}

// UseEmbeddedFiles binds one plugin-owned embedded filesystem.
func (p *sourcePlugin) UseEmbeddedFiles(fileSystem fs.FS) {
	if p == nil {
		return
	}
	p.useEmbeddedFiles(fileSystem)
}

// RegisterUninstallHandler registers one uninstall cleanup callback.
func (p *sourcePlugin) RegisterUninstallHandler(handler SourcePluginUninstallHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerUninstallHandler(handler)
}

// RegisterBeforeInstallHandler registers one pre-install veto callback.
func (p *sourcePlugin) RegisterBeforeInstallHandler(handler SourcePluginBeforeLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerBeforeInstallHandler(handler)
}

// RegisterAfterInstallHandler registers one post-install callback.
func (p *sourcePlugin) RegisterAfterInstallHandler(handler SourcePluginAfterLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerAfterInstallHandler(handler)
}

// RegisterBeforeEnableHandler registers one pre-enable veto callback.
func (p *sourcePlugin) RegisterBeforeEnableHandler(handler SourcePluginBeforeLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerBeforeEnableHandler(handler)
}

// RegisterAfterEnableHandler registers one post-enable callback.
func (p *sourcePlugin) RegisterAfterEnableHandler(handler SourcePluginAfterLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerAfterEnableHandler(handler)
}

// RegisterGlobalBeforeInstallHandler registers one global pre-install veto callback.
func (p *sourcePlugin) RegisterGlobalBeforeInstallHandler(handler SourcePluginGlobalLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerGlobalBeforeInstallHandler(handler)
}

// RegisterGlobalBeforeEnableHandler registers one global pre-enable veto callback.
func (p *sourcePlugin) RegisterGlobalBeforeEnableHandler(handler SourcePluginGlobalLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerGlobalBeforeEnableHandler(handler)
}

// RegisterGlobalBeforeDisableHandler registers one global pre-disable veto callback.
func (p *sourcePlugin) RegisterGlobalBeforeDisableHandler(handler SourcePluginGlobalLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerGlobalBeforeDisableHandler(handler)
}

// RegisterGlobalBeforeUninstallHandler registers one global pre-uninstall veto callback.
func (p *sourcePlugin) RegisterGlobalBeforeUninstallHandler(handler SourcePluginGlobalLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerGlobalBeforeUninstallHandler(handler)
}

// RegisterBeforeUpgradeHandler registers one pre-upgrade veto callback.
func (p *sourcePlugin) RegisterBeforeUpgradeHandler(handler SourcePluginBeforeUpgradeHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerBeforeUpgradeHandler(handler)
}

// RegisterUpgradeHandler registers one source-plugin custom upgrade callback.
func (p *sourcePlugin) RegisterUpgradeHandler(handler SourcePluginUpgradeHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerUpgradeHandler(handler)
}

// RegisterAfterUpgradeHandler registers one post-upgrade event callback.
func (p *sourcePlugin) RegisterAfterUpgradeHandler(handler SourcePluginUpgradeHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerAfterUpgradeHandler(handler)
}

// RegisterBeforeDisableHandler registers one pre-disable veto callback.
func (p *sourcePlugin) RegisterBeforeDisableHandler(handler SourcePluginBeforeLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerBeforeDisableHandler(handler)
}

// RegisterAfterDisableHandler registers one post-disable callback.
func (p *sourcePlugin) RegisterAfterDisableHandler(handler SourcePluginAfterLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerAfterDisableHandler(handler)
}

// RegisterBeforeUninstallHandler registers one pre-uninstall veto callback.
func (p *sourcePlugin) RegisterBeforeUninstallHandler(handler SourcePluginBeforeLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerBeforeUninstallHandler(handler)
}

// RegisterAfterUninstallHandler registers one post-uninstall callback.
func (p *sourcePlugin) RegisterAfterUninstallHandler(handler SourcePluginAfterLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerAfterUninstallHandler(handler)
}

// RegisterBeforeTenantDisableHandler registers one tenant-disable precondition callback.
func (p *sourcePlugin) RegisterBeforeTenantDisableHandler(handler SourcePluginBeforeTenantLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerBeforeTenantDisableHandler(handler)
}

// RegisterAfterTenantDisableHandler registers one tenant-disable post callback.
func (p *sourcePlugin) RegisterAfterTenantDisableHandler(handler SourcePluginAfterTenantLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerAfterTenantDisableHandler(handler)
}

// RegisterBeforeTenantDeleteHandler registers one tenant-delete precondition callback.
func (p *sourcePlugin) RegisterBeforeTenantDeleteHandler(handler SourcePluginBeforeTenantLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerBeforeTenantDeleteHandler(handler)
}

// RegisterAfterTenantDeleteHandler registers one tenant-delete post callback.
func (p *sourcePlugin) RegisterAfterTenantDeleteHandler(handler SourcePluginAfterTenantLifecycleHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerAfterTenantDeleteHandler(handler)
}

// RegisterBeforeInstallModeChangeHandler registers one install-mode change precondition callback.
func (p *sourcePlugin) RegisterBeforeInstallModeChangeHandler(handler SourcePluginBeforeInstallModeChangeHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerBeforeInstallModeChangeHandler(handler)
}

// RegisterAfterInstallModeChangeHandler registers one install-mode change post callback.
func (p *sourcePlugin) RegisterAfterInstallModeChangeHandler(handler SourcePluginAfterInstallModeChangeHandler) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin lifecycle facade is nil")
	}
	return p.registerAfterInstallModeChangeHandler(handler)
}

// RegisterHook registers one callback-style host hook handler.
func (p *sourcePlugin) RegisterHook(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler HookHandler,
) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin hook facade is nil")
	}
	return p.registerHook(point, mode, handler)
}

// RegisterRoutes registers one callback that contributes plugin-owned HTTP routes.
func (p *sourcePlugin) RegisterRoutes(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler RouteRegisterHandler,
) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin http facade is nil")
	}
	return p.registerRoutes(point, mode, handler)
}

// RegisterJobs registers one callback that contributes plugin-owned scheduled jobs.
func (p *sourcePlugin) RegisterJobs(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler JobRegisterHandler,
) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin jobs facade is nil")
	}
	return p.registerJobs(point, mode, handler)
}

// RegisterMenuFilter registers one callback that filters host menus.
func (p *sourcePlugin) RegisterMenuFilter(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler MenuFilterHandler,
) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin access facade is nil")
	}
	return p.registerMenuFilter(point, mode, handler)
}

// RegisterPermissionFilter registers one callback that filters host permissions.
func (p *sourcePlugin) RegisterPermissionFilter(
	point ExtensionPoint,
	mode CallbackExecutionMode,
	handler PermissionFilterHandler,
) error {
	if p == nil {
		return gerror.New("pluginhost: source plugin access facade is nil")
	}
	return p.registerPermissionFilter(point, mode, handler)
}
