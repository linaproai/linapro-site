import {expect, type Locator, type Page} from '@playwright/test';
import {preferZhHans} from './locale';

export class HomePage {
  readonly page: Page;
  readonly modulesTitle: Locator;
  readonly modulesSection: Locator;
  readonly quickstartSection: Locator;

  constructor(page: Page) {
    this.page = page;
    this.modulesTitle = page
      .getByRole('heading', {name: '开箱即用的管理工作台'})
      .first();
    this.modulesSection = page
      .locator('section.home-section--modules')
      .first();
    this.quickstartSection = page
      .locator('section.home-section--quickstart')
      .first();
  }

  async gotoZhHome(): Promise<void> {
    await preferZhHans(this.page);
    await this.page.goto('/zh/');
    await this.page.waitForLoadState('domcontentloaded');
  }

  async expectZhHomeUsesDarkWorkspacePreviewWithoutQuickstart(): Promise<void> {
    await expect(this.modulesTitle).toBeVisible();
    await expect(this.modulesSection).toContainText('开箱即用的管理工作台');
    await expect(this.modulesSection).toHaveCSS(
      'background-image',
      /linear-gradient/,
    );
    await expect(this.modulesSection.locator('.modules-preview-card')).toHaveCount(4);
    await expect(this.modulesSection.locator('.modules-preview-card').first())
      .toHaveCSS('background-color', 'rgba(0, 0, 0, 0)');
    await expect(this.quickstartSection).toHaveCount(0);
    await expect(this.page.getByText('三条命令，立即上手')).toHaveCount(0);
  }
}
