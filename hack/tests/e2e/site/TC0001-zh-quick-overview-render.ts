import {expect, test} from '../../fixtures/auth';
import {DocsPage} from '../../pages/DocsPage';

test.describe('TC-1 Chinese quick overview render', () => {
  test('TC-1a: renders document page without color mode crash', async ({
    page,
  }) => {
    const consoleErrors: string[] = [];

    page.on('console', (message) => {
      if (message.type() === 'error') {
        consoleErrors.push(message.text());
      }
    });
    page.on('pageerror', (error) => consoleErrors.push(error.message));

    await page.route('https://giscus.app/**', (route) =>
      route.fulfill({status: 204, body: ''}),
    );

    const docsPage = new DocsPage(page);
    await docsPage.gotoZhQuickOverview();
    await docsPage.expectZhQuickOverviewRendered();

    const errors = consoleErrors.join('\n');
    expect(errors).not.toContain('useColorMode');
    expect(errors).not.toContain('ColorModeProvider');
  });
});
