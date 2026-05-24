import type {Page} from '@playwright/test';

export async function preferZhHans(page: Page): Promise<void> {
  await page.addInitScript(() => {
    window.localStorage.setItem('user-locale-preference', 'zh-Hans');
  });
}
