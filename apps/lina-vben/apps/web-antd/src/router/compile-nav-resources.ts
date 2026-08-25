import type { RouteRecordStringComponent } from '@vben/types';

import { getPluginPageByRoute } from '#/plugins/page-registry';

import {
  dynamicAccessModeEmbeddedMount,
  dynamicAccessModeQueryKey,
  dynamicEmbeddedSourceQueryKey,
  dynamicPageComponentPath,
} from './workbench-plugin-pages';

export type NavOpenMode = 'page' | 'embedded' | 'iframe' | 'external';

export interface NavResourceItem {
  cache?: number;
  children?: NavResourceItem[];
  icon?: string;
  id: number;
  i18nKey?: string;
  menuKey?: string;
  openMode: NavOpenMode;
  parentId: number;
  path: string;
  perms?: string;
  query?: Record<string, string>;
  resource?: string;
  sort?: number;
  status?: number;
  target?: string;
  title: string;
  type: string;
  visible?: number;
}

const directoryType = 'D';
const menuType = 'M';
const buttonType = 'B';

export function compileNavResources(
  items: NavResourceItem[] | undefined,
): RouteRecordStringComponent[] {
  if (!items?.length) {
    return [];
  }
  const compiled: RouteRecordStringComponent[] = [];
  for (const item of items) {
    if (!item || item.type === buttonType) {
      continue;
    }
    const route = compileNavResource(item);
    const children = compileNavResources(item.children);
    if (item.type === directoryType && children.length === 0) {
      continue;
    }
    if (children.length > 0) {
      route.children = children;
      if (item.type === directoryType) {
        route.redirect = String(children[0]?.path ?? '');
      }
    }
    compiled.push(route);
  }
  return compiled;
}

function compileNavResource(item: NavResourceItem): RouteRecordStringComponent {
  const query = { ...(item.query ?? {}) };
  const route: RouteRecordStringComponent = {
    name: compileRouteName(item),
    path: compileRoutePath(item),
    meta: {
      title: item.title,
      icon: item.icon,
      i18nKey: item.i18nKey,
      hideInMenu: item.visible === 0,
      keepAlive: item.cache === 1,
      order: item.sort ?? 0,
      authority: item.perms,
      ignoreAccess: false,
    },
  };

  if (item.type !== menuType) {
    return route;
  }

  switch (item.openMode) {
    case 'embedded': {
      route.component = viewComponentPath(dynamicPageComponentPath);
      route.meta = {
        ...route.meta,
        query: {
          ...query,
          [dynamicEmbeddedSourceQueryKey]: item.target ?? '',
          [dynamicAccessModeQueryKey]: dynamicAccessModeEmbeddedMount,
        },
      };
      break;
    }
    case 'external': {
      route.component = 'BasicLayout';
      route.meta = {
        ...route.meta,
        link: item.target ?? '',
        openInNewWindow: true,
      };
      break;
    }
    case 'iframe': {
      route.component = 'IFrameView';
      route.meta = {
        ...route.meta,
        iframeSrc: item.target ?? '',
      };
      break;
    }
    default: {
      route.component = compilePageComponent(item);
      if (Object.keys(query).length > 0) {
        route.meta = { ...route.meta, query };
      }
    }
  }

  return route;
}

function compileRouteName(item: NavResourceItem): string {
  if (item.openMode !== 'page') {
    return toPascalCase(buildLinkSlug(item));
  }
  if (item.path) {
    return toPascalCase(item.path);
  }
  return toPascalCase(item.title);
}

function compileRoutePath(item: NavResourceItem): string {
  if (item.openMode !== 'page') {
    const slug = buildLinkSlug(item);
    return item.parentId === 0 ? `/${slug}` : slug;
  }
  if (!item.path) {
    return '';
  }
  if (item.parentId !== 0) {
    return item.path;
  }
  return item.path.startsWith('/') ? item.path : `/${item.path}`;
}

function compilePageComponent(item: NavResourceItem): string {
  if (isRegisteredPluginPage(item.path)) {
    return viewComponentPath(dynamicPageComponentPath);
  }
  return viewComponentPath(item.resource ?? '');
}

function isRegisteredPluginPage(path: string): boolean {
  for (const candidate of pluginRouteCandidates(path)) {
    if (getPluginPageByRoute(candidate)) {
      return true;
    }
  }
  return false;
}

function pluginRouteCandidates(path: string): string[] {
  const normalized = path.trim().replace(/^\/+/u, '').replace(/\/+$/u, '');
  if (!normalized) {
    return [];
  }
  const lastSegment = normalized.split('/').filter(Boolean).at(-1) ?? '';
  return [...new Set([normalized, lastSegment].filter(Boolean))];
}

function viewComponentPath(resource: string): string {
  const trimmed = resource.trim();
  if (!trimmed) {
    return '';
  }
  return `#/views/${trimmed}`;
}

function buildLinkSlug(item: NavResourceItem): string {
  let slug = `link-${item.id}-`;
  const source = (item.target || item.path || '').toLowerCase();
  for (const current of source) {
    if (/[a-z0-9]/i.test(current)) {
      slug += current;
      continue;
    }
    slug += '-';
  }
  slug = slug.replace(/-+/g, '-').replace(/^-|-$/g, '');
  return slug || `link-${item.id}`;
}

function toPascalCase(value: string): string {
  let result = '';
  let upperNext = true;
  for (const current of value) {
    if (current === '-' || current === '_' || current === '/' || current === ' ') {
      upperNext = true;
      continue;
    }
    result += upperNext ? current.toUpperCase() : current;
    upperNext = false;
  }
  return result;
}
