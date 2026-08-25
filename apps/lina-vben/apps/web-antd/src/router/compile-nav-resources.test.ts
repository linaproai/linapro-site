import { describe, expect, it, vi } from 'vitest';

vi.mock('#/plugins/page-registry', () => ({
  getPluginPageByRoute: (routePath: string) => {
    const normalized = routePath.replace(/^\//u, '');
    if (normalized === 'plugin-marketplace-mine') {
      return { routePath: 'plugin-marketplace-mine' };
    }
    return null;
  },
}));

import { compileNavResources } from './compile-nav-resources';

describe('compileNavResources', () => {
  it('compiles hosted HTML as IFrameView', () => {
    const routes = compileNavResources([
      {
        id: 101,
        parentId: 0,
        title: 'Runtime Iframe Entry',
        path: '/x-assets/plugin-runtime-demo/v0.1.0/index.html',
        type: 'M',
        openMode: 'iframe',
        target: '/x-assets/plugin-runtime-demo/v0.1.0/index.html',
        visible: 1,
        status: 1,
      },
    ]);
    expect(routes).toHaveLength(1);
    expect(routes[0]?.component).toBe('IFrameView');
    expect(routes[0]?.meta?.iframeSrc).toBe(
      '/x-assets/plugin-runtime-demo/v0.1.0/index.html',
    );
    expect(routes[0]?.path).not.toBe(
      '/x-assets/plugin-runtime-demo/v0.1.0/index.html',
    );
  });

  it('compiles external mode as BasicLayout link', () => {
    const routes = compileNavResources([
      {
        id: 102,
        parentId: 0,
        title: 'Runtime New Window Entry',
        path: '/x-assets/plugin-runtime-demo/v0.1.0/index.html',
        type: 'M',
        openMode: 'external',
        target: '/x-assets/plugin-runtime-demo/v0.1.0/index.html',
        visible: 1,
      },
    ]);
    expect(routes[0]?.component).toBe('BasicLayout');
    expect(routes[0]?.meta?.link).toBe(
      '/x-assets/plugin-runtime-demo/v0.1.0/index.html',
    );
    expect(routes[0]?.meta?.openInNewWindow).toBe(true);
  });

  it('compiles embedded mode onto the workbench dynamic page', () => {
    const routes = compileNavResources([
      {
        id: 103,
        parentId: 0,
        title: 'Runtime Embedded Entry',
        path: '/x-assets/plugin-runtime-demo/v0.1.0/mount.js',
        type: 'M',
        openMode: 'embedded',
        target: '/x-assets/plugin-runtime-demo/v0.1.0/mount.js',
        query: { tab: 'overview' },
        visible: 1,
      },
    ]);
    expect(routes[0]?.component).toBe('#/views/system/plugin/dynamic-page');
    expect(routes[0]?.meta?.query).toMatchObject({
      tab: 'overview',
      pluginAccessMode: 'embedded-mount',
      embeddedSrc: '/x-assets/plugin-runtime-demo/v0.1.0/mount.js',
    });
  });

  it('compiles in-app pages with opaque resources', () => {
    const routes = compileNavResources([
      {
        id: 104,
        parentId: 0,
        title: 'User',
        path: '/system/user',
        resource: 'system/user/index',
        type: 'M',
        openMode: 'page',
        visible: 1,
      },
    ]);
    expect(routes[0]?.component).toBe('#/views/system/user/index');
    expect(routes[0]?.path).toBe('/system/user');
  });

  it('compiles registered source plugin pages by path', () => {
    const routes = compileNavResources([
      {
        id: 105,
        parentId: 0,
        title: 'My Plugins',
        path: 'plugin-marketplace-mine',
        type: 'M',
        openMode: 'page',
        visible: 1,
      },
    ]);
    expect(routes[0]?.component).toBe('#/views/system/plugin/dynamic-page');
  });

  it('does not treat hash-prefixed resources as already compiled', () => {
    const routes = compileNavResources([
      {
        id: 106,
        parentId: 0,
        title: 'User',
        path: '/system/user',
        resource: '#/views/system/user/index',
        type: 'M',
        openMode: 'page',
        visible: 1,
      },
    ]);
    expect(routes[0]?.component).not.toBe('#/views/system/user/index');
  });

  it('hides directories that only contained buttons', () => {
    const routes = compileNavResources([
      {
        id: 301,
        parentId: 0,
        title: '系统监控',
        path: 'monitor',
        type: 'D',
        openMode: 'page',
        visible: 1,
        children: [
          {
            id: 302,
            parentId: 301,
            title: '操作日志查看',
            path: 'linapro-monitor-operlog-view',
            type: 'B',
            openMode: 'page',
            visible: 1,
          },
        ],
      },
    ]);
    expect(routes).toHaveLength(0);
  });
});
