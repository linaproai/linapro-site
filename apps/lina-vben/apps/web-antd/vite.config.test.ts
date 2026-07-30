import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

import { describe, expect, it } from 'vitest';

import { vitePluginSourceDepsTestHooks } from './vite.plugin-source-deps';

const pluginRoot = '/repo/apps/lina-plugins';

describe('source plugin Vite helpers', () => {
  it('treats all frontend TS/Vue files as dependency modules', () => {
    expect(
      vitePluginSourceDepsTestHooks.isPluginFrontendModuleFile(
        pluginRoot,
        '/repo/apps/lina-plugins/linapro-plugin-marketplace/frontend/utils/markdown.ts',
      ),
    ).toBe(true);
    expect(
      vitePluginSourceDepsTestHooks.isPluginFrontendModuleFile(
        pluginRoot,
        '/repo/apps/lina-plugins/linapro-plugin-marketplace/frontend/api/marketplace.ts?vue&type=script',
      ),
    ).toBe(true);
    expect(
      vitePluginSourceDepsTestHooks.isPluginFrontendModuleFile(
        pluginRoot,
        '/repo/apps/lina-plugins/linapro-plugin-marketplace/backend/plugin.go',
      ),
    ).toBe(false);
  });

  it('keeps virtual module entry invalidation scoped to page and slot Vue files', () => {
    expect(
      vitePluginSourceDepsTestHooks.isPluginFrontendEntryFile(
        pluginRoot,
        '/repo/apps/lina-plugins/linapro-plugin-marketplace/frontend/pages/detail/index.vue',
      ),
    ).toBe(true);
    expect(
      vitePluginSourceDepsTestHooks.isPluginFrontendEntryFile(
        pluginRoot,
        '/repo/apps/lina-plugins/linapro-plugin-marketplace/frontend/slots/header.vue',
      ),
    ).toBe(true);
    expect(
      vitePluginSourceDepsTestHooks.isPluginFrontendEntryFile(
        pluginRoot,
        '/repo/apps/lina-plugins/linapro-plugin-marketplace/frontend/utils/markdown.ts',
      ),
    ).toBe(false);
  });

  it('normalizes Vite fs importer paths before matching plugin sources', () => {
    const normalized = vitePluginSourceDepsTestHooks.normalizeImporterPath(
      '/@fs/repo/apps/lina-plugins/linapro-plugin-marketplace/frontend/utils/markdown.ts?import',
    );

    expect(normalized).toBe(
      '/repo/apps/lina-plugins/linapro-plugin-marketplace/frontend/utils/markdown.ts',
    );
    expect(
      vitePluginSourceDepsTestHooks.isPluginFrontendModuleFile(
        pluginRoot,
        '/@fs/repo/apps/lina-plugins/linapro-plugin-marketplace/frontend/utils/markdown.ts?import',
      ),
    ).toBe(true);
  });

  it('normalizes Vite fs paths for absolute and Windows-style paths', () => {
    expect(vitePluginSourceDepsTestHooks.toViteFsPath('/repo/app.vue')).toBe(
      '/@fs/repo/app.vue',
    );
    expect(
      vitePluginSourceDepsTestHooks.normalizeImporterPath(
        '/@fs/C:/repo/apps/lina-plugins/plugin/frontend/pages/index.vue?vue&type=script',
      ),
    ).toBe('C:/repo/apps/lina-plugins/plugin/frontend/pages/index.vue');
  });

  it('loads an empty module for missing Vue style block requests', () => {
    const root = mkdtempSync(join(tmpdir(), 'lina-vue-style-request-'));
    try {
      const componentPath = join(root, 'NoStyle.vue');
      writeFileSync(
        componentPath,
        [
          '<script setup lang="ts">',
          'const title = "No style";',
          '</script>',
          '<template><div>{{ title }}</div></template>',
        ].join('\n'),
      );

      const resolvedId =
        vitePluginSourceDepsTestHooks.resolveMissingVueStyleBlockId(
          `${componentPath}?vue&type=style&index=0&lang.css`,
        );

      expect(resolvedId).toContain('lina-missing-vue-style-block');
      expect(
        vitePluginSourceDepsTestHooks.loadMissingVueStyleBlock(resolvedId!),
      ).toBe('');
    } finally {
      rmSync(root, { force: true, recursive: true });
    }
  });

  it('keeps existing Vue style block requests on the Vue plugin path', () => {
    const root = mkdtempSync(join(tmpdir(), 'lina-vue-style-request-'));
    try {
      const componentPath = join(root, 'WithStyle.vue');
      writeFileSync(
        componentPath,
        [
          '<script setup lang="ts">',
          'const title = "Styled";',
          '</script>',
          '<template><div class="title">{{ title }}</div></template>',
          '<style scoped>',
          '.title { color: red; }',
          '</style>',
        ].join('\n'),
      );

      expect(
        vitePluginSourceDepsTestHooks.resolveMissingVueStyleBlockId(
          `${componentPath}?vue&type=style&index=0&scoped=true&lang.css`,
        ),
      ).toBeNull();

      const missingStyleId =
        vitePluginSourceDepsTestHooks.resolveMissingVueStyleBlockId(
          `${componentPath}?vue&type=style&index=1&scoped=true&lang.css`,
        );

      expect(missingStyleId).toContain('lina-missing-vue-style-block');
      expect(
        vitePluginSourceDepsTestHooks.loadMissingVueStyleBlock(missingStyleId!),
      ).toBe('');
    } finally {
      rmSync(root, { force: true, recursive: true });
    }
  });

  it('ignores non-style Vue module requests', () => {
    expect(
      vitePluginSourceDepsTestHooks.loadMissingVueStyleBlock(
        '/repo/apps/lina-plugins/plugin/frontend/pages/index.vue',
      ),
    ).toBeNull();
    expect(
      vitePluginSourceDepsTestHooks.resolveMissingVueStyleBlockId(
        '/repo/apps/lina-plugins/plugin/frontend/pages/index.vue?vue&type=script&lang.ts',
      ),
    ).toBeNull();
  });

  it('keeps host singleton packages on the host dependency graph', () => {
    expect(vitePluginSourceDepsTestHooks.shouldResolveFromAppFirst('vue')).toBe(
      true,
    );
    expect(
      vitePluginSourceDepsTestHooks.shouldResolveFromAppFirst(
        'ant-design-vue/es/button',
      ),
    ).toBe(true);
    expect(
      vitePluginSourceDepsTestHooks.shouldResolveFromAppFirst('markdown-it'),
    ).toBe(false);
  });

  it('detects plugin-declared private dependencies by package name', () => {
    const root = mkdtempSync(join(tmpdir(), 'lina-plugin-source-deps-'));
    try {
      const frontendDir = join(root, 'sample-plugin', 'frontend');
      mkdirSync(frontendDir, { recursive: true });
      writeFileSync(
        join(frontendDir, 'package.json'),
        JSON.stringify({
          dependencies: {
            'markdown-it': '14.1.1',
          },
          optionalDependencies: {
            '@scope/private-chart': '1.0.0',
          },
        }),
      );

      const importer = join(frontendDir, 'utils', 'markdown.ts');
      expect(
        vitePluginSourceDepsTestHooks.isPluginFrontendDependencyDeclared(
          root,
          importer,
          'markdown-it',
        ),
      ).toBe(true);
      expect(
        vitePluginSourceDepsTestHooks.isPluginFrontendDependencyDeclared(
          root,
          importer,
          'markdown-it/lib/index.mjs',
        ),
      ).toBe(true);
      expect(
        vitePluginSourceDepsTestHooks.isPluginFrontendDependencyDeclared(
          root,
          importer,
          '@scope/private-chart/vue',
        ),
      ).toBe(true);
      expect(
        vitePluginSourceDepsTestHooks.isPluginFrontendDependencyDeclared(
          root,
          importer,
          'vue',
        ),
      ).toBe(false);
    } finally {
      rmSync(root, { force: true, recursive: true });
    }
  });
});
