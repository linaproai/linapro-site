import {expect, test} from '../../fixtures/auth';
import {DocsPage} from '../../pages/DocsPage';

test.describe('TC-3 Chinese static assets Mermaid render', () => {
  test('TC-3a: renders backend route boundary diagram without Mermaid parse errors', async ({
    page,
  }) => {
    const runtimeErrors: string[] = [];

    page.on('console', (message) => {
      if (message.type() === 'error') {
        runtimeErrors.push(message.text());
      }
    });
    page.on('pageerror', (error) => runtimeErrors.push(error.message));

    await page.route('https://giscus.app/**', (route) =>
      route.fulfill({status: 204, body: ''}),
    );

    const docsPage = new DocsPage(page);
    await docsPage.gotoZhStaticAssets();
    await docsPage.expectZhStaticAssetsRendered();

    const errors = runtimeErrors.join('\n');
    expect(errors).not.toContain('Parse error');
    expect(errors).not.toContain('DIAMOND_START');
  });
});
