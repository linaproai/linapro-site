/**
 * TC007: Plugin detail docs catalog and Markdown rendering.
 *
 * Validates the marketplace detail Docs tab lists every release document path
 * and renders Markdown through the plugin markdown-it helper with syntax
 * highlighting, tables, images, and Mermaid diagrams (VS Code–style preview).
 */
import { expect, test } from "@host-tests/fixtures/auth";

import { MarketplacePage } from "../pages/MarketplacePage";
import {
  captureMarketplaceScreenshot,
  installMarketplaceApiMocks,
  marketplaceSourcePluginId,
  openMarketplaceWorkbench,
} from "../support/marketplace-fixtures";

test.describe("TC-7 marketplace detail docs catalog and markdown", () => {
  test("TC-7a: docs tab lists catalog entries and renders markdown tables", async ({
    authenticatedPage,
  }) => {
    const page = authenticatedPage;
    await page.setViewportSize({ height: 960, width: 1440 });
    await installMarketplaceApiMocks(page, {
      menuRole: "both",
    });
    await openMarketplaceWorkbench(page, "简体中文");

    const marketplace = new MarketplacePage(page);
    await marketplace.gotoDetail(marketplaceSourcePluginId());
    await marketplace.openDocsForVersion("v1.0.0");

    // Catalog navigation must expose every manifest/docs path (not only index.md).
    await marketplace.expectDocumentCatalogEntries([
      "源码演示指南",
      "配置说明",
      "更新日志",
    ]);
    // When manifest docs exist, package-root README stays out of catalog/content.
    await expect(
      marketplace.documentCatalogNav().getByText("README", { exact: true }),
    ).toHaveCount(0);
    await expect(marketplace.documentPanel()).not.toContainText("README.md");
    await marketplace.expectRenderedMarkdownHeading(/源码演示指南|Source Demo Guide/u);
    await marketplace.expectRenderedMarkdownTable();
    await marketplace.expectRenderedMarkdownImage();
    await marketplace.expectRenderedMermaidDiagram();
    await expect(marketplace.documentMarkdownBody()).toContainText(
      "apps/lina-plugins",
    );
    // Must not surface raw Markdown fence markers as the only content.
    await expect(marketplace.documentMarkdownBody()).not.toHaveText(
      /^#\s+源码演示指南/u,
    );
    await expect(marketplace.documentMarkdownBody()).not.toContainText(
      "```mermaid",
    );

    await marketplace.expectNoRawI18nKeys();
    await captureMarketplaceScreenshot(page, "detail-docs-catalog-index");
  });

  test("TC-7b: switching catalog path loads configuration markdown", async ({
    authenticatedPage,
  }) => {
    const page = authenticatedPage;
    await page.setViewportSize({ height: 960, width: 1440 });
    await installMarketplaceApiMocks(page, {
      menuRole: "both",
    });
    await openMarketplaceWorkbench(page, "简体中文");

    const marketplace = new MarketplacePage(page);
    await marketplace.gotoDetail(marketplaceSourcePluginId());
    await marketplace.openDocsForVersion("v1.0.0");
    await marketplace.selectDocumentCatalogEntry("配置说明");

    await marketplace.expectRenderedMarkdownHeading(/配置说明|Configuration/u);
    await expect(
      marketplace.documentMarkdownBody().locator("pre code.language-yaml, pre code").first(),
    ).toContainText("enabled: true");
    await marketplace.expectRenderedMarkdownCodeHighlight();
    await expect(
      marketplace.documentMarkdownBody().locator("pre code.language-ts, pre code.hljs").nth(1),
    ).toContainText("export const demo");
    await expect(marketplace.documentPanel()).toContainText("configuration.md");

    await marketplace.expectNoRawI18nKeys();
    await captureMarketplaceScreenshot(page, "detail-docs-catalog-configuration");
  });
});
