import type { RouteRecordStringComponent } from '@vben/types';

import { requestClient } from '#/api/request';
import {
  compileNavResources,
  type NavResourceItem,
} from '#/router/compile-nav-resources';

/**
 * 获取用户所有菜单，并把宿主通用导航资源编译为默认工作台路由。
 */
export async function getAllMenusApi(): Promise<RouteRecordStringComponent[]> {
  const response = await requestClient.get<{ list: NavResourceItem[] }>(
    '/menus/all',
  );
  return compileNavResources(response.list ?? []);
}
