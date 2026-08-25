import { LoginPage } from '../../pages/LoginPage';
import { test, expect } from '../../fixtures/auth';
import { config, workspacePath } from '../../fixtures/config';

/**
 * When the workbench menu is unavailable, login must not hard-land on
 * /dashboard/analytics (404). The SPA should resolve the first accessible menu.
 */
test.describe('TC011 登录落地到首个可访问菜单', () => {
  test('TC011a: 无工作台菜单时登录进入系统用户页且非 404', async ({ page }) => {
    await page.route('**/api/v1/user/info', async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        status: 200,
        body: JSON.stringify({
          code: 0,
          message: 'ok',
          data: {
            avatar: '',
            // Stale/hardcoded workbench homePath must be corrected by the SPA.
            homePath: '/dashboard/analytics',
            permissions: ['*'],
            realName: 'Admin',
            roles: ['admin'],
            userId: 1,
            username: 'admin',
            organizationEnabled: false,
            tenantEnabled: false,
          },
        }),
      });
    });

    await page.route('**/api/v1/menus/all', async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        status: 200,
        body: JSON.stringify({
          code: 0,
          message: 'ok',
          data: {
            list: [
              {
                id: 10,
                parentId: 0,
                title: 'System',
                path: '/system',
                type: 'D',
                openMode: 'page',
                icon: 'lucide:settings',
                visible: 1,
                children: [
                  {
                    id: 11,
                    parentId: 10,
                    title: 'User',
                    path: 'user',
                    resource: 'system/user/index',
                    type: 'M',
                    openMode: 'page',
                    icon: 'lucide:user',
                    perms: '*',
                    visible: 1,
                  },
                ],
              },
            ],
          },
        }),
      });
    });

    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);

    await page.waitForURL((url) => !url.pathname.includes('/auth/login'), {
      timeout: 20_000,
    });

    expect(page.url()).not.toContain('/dashboard/analytics');
    expect(page.url()).toMatch(/\/system\/user/);

    await expect(page.locator('body')).not.toHaveText(/404|Not Found|页面不存在/i);

    const screenshotDir = `temp/${new Date().toISOString().slice(0, 10)}`;
    await page.screenshot({
      path: `${screenshotDir}/tc011-login-first-menu.png`,
      fullPage: false,
    });
  });

  test('TC011b: 已登录访问默认工作台路径时纠正到可访问菜单', async ({
    page,
  }) => {
    await page.route('**/api/v1/user/info', async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        status: 200,
        body: JSON.stringify({
          code: 0,
          message: 'ok',
          data: {
            avatar: '',
            homePath: '/dashboard/analytics',
            permissions: ['*'],
            realName: 'Admin',
            roles: ['admin'],
            userId: 1,
            username: 'admin',
            organizationEnabled: false,
            tenantEnabled: false,
          },
        }),
      });
    });

    await page.route('**/api/v1/menus/all', async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        status: 200,
        body: JSON.stringify({
          code: 0,
          message: 'ok',
          data: {
            list: [
              {
                id: 10,
                parentId: 0,
                title: 'System',
                path: '/system',
                type: 'D',
                openMode: 'page',
                icon: 'lucide:settings',
                visible: 1,
                children: [
                  {
                    id: 11,
                    parentId: 10,
                    title: 'User',
                    path: 'user',
                    resource: 'system/user/index',
                    type: 'M',
                    openMode: 'page',
                    icon: 'lucide:user',
                    perms: '*',
                    visible: 1,
                  },
                ],
              },
            ],
          },
        }),
      });
    });

    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);

    await page.goto(workspacePath('/dashboard/analytics'));
    await page.waitForURL(/\/system\/user/, { timeout: 20_000 });
    expect(page.url()).toMatch(/\/system\/user/);
    await expect(page.locator('body')).not.toHaveText(/404|Not Found|页面不存在/i);
  });
});
