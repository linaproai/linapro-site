import { Buffer } from 'node:buffer';
import { existsSync, readdirSync, readFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import { join, relative } from 'node:path';

import { parse as parseVueSfc } from 'vue/compiler-sfc';

export const appRequire = createRequire(import.meta.url);
export const sourcePluginsEnabled = process.env.LINAPRO_SOURCE_PLUGINS === '1';

export function collectPluginSourceFiles(pluginRoot: string) {
  const pageFiles: string[] = [];
  const slotFiles: string[] = [];

  if (!sourcePluginsEnabled || !existsSync(pluginRoot)) {
    return { pageFiles, slotFiles };
  }

  const walk = (currentPath: string) => {
    for (const entry of readdirSync(currentPath, { withFileTypes: true })) {
      const fullPath = join(currentPath, entry.name);
      if (entry.isDirectory()) {
        walk(fullPath);
        continue;
      }
      if (!entry.isFile() || !entry.name.endsWith('.vue')) {
        continue;
      }
      const normalizedPath = normalizeFsPath(fullPath);
      if (normalizedPath.includes('/frontend/pages/')) {
        pageFiles.push(fullPath);
      }
      if (normalizedPath.includes('/frontend/slots/')) {
        slotFiles.push(fullPath);
      }
    }
  };

  walk(pluginRoot);
  return { pageFiles, slotFiles };
}

export function normalizeFsPath(filePath: string) {
  return filePath.replaceAll('\\', '/');
}

export function normalizeImporterPath(filePath: string) {
  const normalized = normalizeFsPath(
    filePath.split('?')[0]?.split('#')[0] || filePath,
  );
  if (!normalized.startsWith('/@fs/')) {
    return normalized;
  }

  const fsPath = normalized.slice('/@fs/'.length);
  return /^[A-Za-z]:\//.test(fsPath) ? fsPath : `/${fsPath}`;
}

export function toViteFsPath(filePath: string) {
  const normalizedPath = normalizeFsPath(filePath);
  if (normalizedPath.startsWith('/@fs/')) {
    return normalizedPath;
  }
  return normalizedPath.startsWith('/')
    ? `/@fs${normalizedPath}`
    : `/@fs/${normalizedPath}`;
}

type VueStyleBlockRequest = {
  filePath: string;
  index: number;
};

const missingVueStyleBlockModulePrefix = '\0lina-missing-vue-style-block:';

export function parseVueStyleBlockRequest(
  id: string,
): null | VueStyleBlockRequest {
  const queryStart = id.indexOf('?');
  if (queryStart < 0) {
    return null;
  }

  const rawPath = id.slice(0, queryStart);
  if (!rawPath.endsWith('.vue')) {
    return null;
  }

  const rawQuery = id.slice(queryStart + 1).split('#')[0] || '';
  const query = new URLSearchParams(rawQuery);
  if (!query.has('vue') || query.get('type') !== 'style') {
    return null;
  }

  const rawIndex = query.get('index') ?? '0';
  if (!/^\d+$/.test(rawIndex)) {
    return null;
  }

  return {
    filePath: normalizeImporterPath(rawPath),
    index: Number(rawIndex),
  };
}

function isMissingVueStyleBlockRequest(id: string) {
  const request = parseVueStyleBlockRequest(id);
  if (!request || !existsSync(request.filePath)) {
    return false;
  }

  try {
    const source = readFileSync(request.filePath, 'utf8');
    const { descriptor } = parseVueSfc(source, { filename: request.filePath });
    return request.index >= descriptor.styles.length;
  } catch {
    return false;
  }
}

export function resolveMissingVueStyleBlockId(id: string) {
  return isMissingVueStyleBlockRequest(id)
    ? `${missingVueStyleBlockModulePrefix}${Buffer.from(id).toString('base64url')}`
    : null;
}

export function loadMissingVueStyleBlock(id: string) {
  return id.startsWith(missingVueStyleBlockModulePrefix) ? '' : null;
}

export function isPluginFrontendEntryFile(
  pluginRoot: string,
  filePath: string,
) {
  const normalizedPluginRoot = normalizeFsPath(pluginRoot);
  const normalizedFilePath = normalizeImporterPath(filePath);

  if (!normalizedFilePath.startsWith(normalizedPluginRoot)) {
    return false;
  }

  if (!normalizedFilePath.endsWith('.vue')) {
    return false;
  }

  return (
    normalizedFilePath.includes('/frontend/pages/') ||
    normalizedFilePath.includes('/frontend/slots/')
  );
}

export function isPluginFrontendModuleFile(
  pluginRoot: string,
  filePath: string,
) {
  const normalizedPluginRoot = normalizeFsPath(pluginRoot);
  const normalizedFilePath = normalizeImporterPath(filePath);

  if (!normalizedFilePath.startsWith(normalizedPluginRoot)) {
    return false;
  }

  if (!normalizedFilePath.includes('/frontend/')) {
    return false;
  }

  return /\.(?:[jt]sx?|vue)$/.test(normalizedFilePath);
}

export function resolvePluginFrontendPackageJson(
  pluginRoot: string,
  importer: string,
) {
  const normalizedFilePath = normalizeImporterPath(importer);
  const relativePath = normalizeFsPath(
    relative(pluginRoot, normalizedFilePath),
  );
  const match = relativePath.match(/^([^/]+)\/frontend\//);
  if (!match?.[1]) {
    return null;
  }

  const packageJsonPath = join(
    pluginRoot,
    match[1],
    'frontend',
    'package.json',
  );
  return existsSync(packageJsonPath) ? packageJsonPath : null;
}

function getBarePackageName(source: string) {
  if (source.startsWith('@')) {
    const [scope, name] = source.split('/');
    return scope && name ? `${scope}/${name}` : source;
  }
  return source.split('/')[0] || source;
}

export function isPluginFrontendDependencyDeclared(
  pluginRoot: string,
  importer: string,
  source: string,
) {
  const packageJsonPath = resolvePluginFrontendPackageJson(
    pluginRoot,
    importer,
  );
  if (!packageJsonPath) {
    return false;
  }

  try {
    const parsed = JSON.parse(readFileSync(packageJsonPath, 'utf8')) as {
      dependencies?: Record<string, string>;
      optionalDependencies?: Record<string, string>;
    };
    const packageName = getBarePackageName(source);
    return (
      packageName in (parsed.dependencies ?? {}) ||
      packageName in (parsed.optionalDependencies ?? {})
    );
  } catch {
    return false;
  }
}

export function shouldResolveFromAppFirst(source: string) {
  return (
    source === 'vue' ||
    source.startsWith('vue/') ||
    source === 'vue-router' ||
    source.startsWith('vue-router/') ||
    source === 'pinia' ||
    source.startsWith('pinia/') ||
    source === 'ant-design-vue' ||
    source.startsWith('ant-design-vue/') ||
    source.startsWith('@vben/') ||
    source === '@vueuse/core' ||
    source.startsWith('@vueuse/')
  );
}

export function collectPluginFrontendDependencies(pluginRoot: string) {
  if (!sourcePluginsEnabled || !existsSync(pluginRoot)) {
    return [];
  }

  const dependencies = new Set<string>();
  for (const entry of readdirSync(pluginRoot, { withFileTypes: true })) {
    if (!entry.isDirectory()) {
      continue;
    }
    const packageJsonPath = join(
      pluginRoot,
      entry.name,
      'frontend',
      'package.json',
    );
    if (!existsSync(packageJsonPath)) {
      continue;
    }

    try {
      const parsed = JSON.parse(readFileSync(packageJsonPath, 'utf8')) as {
        dependencies?: Record<string, string>;
        optionalDependencies?: Record<string, string>;
      };
      Object.keys(parsed.dependencies ?? {}).forEach((name) => {
        if (!shouldResolveFromAppFirst(name)) {
          dependencies.add(name);
        }
      });
      Object.keys(parsed.optionalDependencies ?? {}).forEach((name) => {
        if (!shouldResolveFromAppFirst(name)) {
          dependencies.add(name);
        }
      });
    } catch {
      continue;
    }
  }

  return [...dependencies].toSorted();
}

export function isBareModuleImport(source: string) {
  return (
    !!source &&
    !source.startsWith('.') &&
    !source.startsWith('/') &&
    !source.startsWith('#') &&
    !source.startsWith('\0') &&
    !source.startsWith('virtual:')
  );
}

export const vitePluginSourceDepsTestHooks = {
  collectPluginFrontendDependencies,
  isBareModuleImport,
  isPluginFrontendDependencyDeclared,
  isPluginFrontendEntryFile,
  isPluginFrontendModuleFile,
  loadMissingVueStyleBlock,
  normalizeImporterPath,
  parseVueStyleBlockRequest,
  resolveMissingVueStyleBlockId,
  resolvePluginFrontendPackageJson,
  shouldResolveFromAppFirst,
  toViteFsPath,
};
