import { existsSync } from 'node:fs';
import path from 'node:path';

import { test, expect } from '../../../fixtures/auth';
import { createAdminApiContext, listPlugins } from '../../../fixtures/plugin';

function officialPluginWorkspacePresent() {
  const candidates = [
    path.resolve(process.cwd(), 'apps/lina-plugins'),
    path.resolve(process.cwd(), '..', 'apps/lina-plugins'),
    path.resolve(process.cwd(), '..', '..', 'apps/lina-plugins'),
  ];
  return candidates.some((dir) =>
    existsSync(path.join(dir, 'linapro-plugin-marketplace', 'plugin.yaml')),
  );
}

test.describe('TC-5 Host-only plugin workspace', () => {
  test.skip(
    process.env.E2E_HOST_ONLY_PLUGINS !== '1',
    'Host-only plugin workspace assertion runs only in host-only validation.',
  );
  test.skip(
    officialPluginWorkspacePresent(),
    'Official plugin workspace is present; empty source-plugin list does not apply.',
  );

  test('TC009a: source plugin list is empty without the official plugin workspace', async () => {
    const adminApi = await createAdminApiContext();
    try {
      const plugins = await listPlugins(adminApi);
      const sourcePlugins = plugins.filter((plugin) => plugin.type === 'source');
      expect(sourcePlugins).toEqual([]);
    } finally {
      await adminApi.dispose();
    }
  });
});
