import {expect, type Locator, type Page} from '@playwright/test';
import {preferZhHans} from './locale';

export class DocsPage {
  readonly page: Page;
  readonly mainHeading: Locator;
  readonly staticAssetsIntroHeading: Locator;
  readonly crashMessage: Locator;
  readonly colorModeError: Locator;

  constructor(page: Page) {
    this.page = page;
    this.mainHeading = page
      .getByRole('heading', {name: /^项目介绍/})
      .first();
    this.staticAssetsIntroHeading = page
      .getByRole('heading', {name: /^基本介绍/})
      .first();
    this.crashMessage = page.getByText('页面已崩溃。');
    this.colorModeError = page.getByText('Hook useColorMode');
  }

  async gotoZhQuickOverview(): Promise<void> {
    await preferZhHans(this.page);
    await this.page.goto('/zh/quick/overview');
    await this.page.waitForLoadState('domcontentloaded');
  }

  async expectZhQuickOverviewRendered(): Promise<void> {
    await expect(this.mainHeading).toBeVisible();
    await expect(this.crashMessage).toHaveCount(0);
    await expect(this.colorModeError).toHaveCount(0);
  }

  async gotoZhStaticAssets(): Promise<void> {
    await preferZhHans(this.page);
    await this.page.goto('/zh/docs/static-assets');
    await this.page.waitForLoadState('domcontentloaded');
  }

  async expectZhStaticAssetsRendered(): Promise<void> {
    await expect(this.staticAssetsIntroHeading).toBeVisible();
    await expect(this.crashMessage).toHaveCount(0);
  }
}
