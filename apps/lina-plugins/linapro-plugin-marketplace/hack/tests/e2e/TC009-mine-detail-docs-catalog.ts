/**
 * TC009: Mine detail modal docs catalog lists all marketplace docs and hides README.
 *
 * Validates opening plugin detail from "My Plugins" shows every navigable
 * manifest/docs path in the Docs tab, and keeps package-root README hidden
 * while manifest docs exist.
 */
import { expect, test } from "@host-tests/fixtures/auth";

import { MarketplacePage } from "../pages/MarketplacePage";
import {
  captureMarketplaceScreenshot,
  installMarketplaceApiMocks,
  marketplaceSourcePluginId,
  openMarketplaceWorkbench,
} from "../support/marketplace-fixtures";

test.describe("TC-9 mine detail docs catalog completeness", () => {
  test("TC-9a: mine detail modal docs tab lists all docs and hides README", async ({
    authenticatedPage,
  }) => {
    const page = authenticatedPage;
    await page.setViewportSize({ height: 960, width: 1440 });
    await installMarketplaceApiMocks(page, {
      menuRole: "both",
    });
    await openMarketplaceWorkbench(page, "简体中文");

    const marketplace = new MarketplacePage(page);
    await marketplace.gotoMine();
    await marketplace.openMineDetail(marketplaceSourcePluginId());

    const detailDialog = page
      .locator('[role="dialog"]:visible')
      .filter({ has: marketplace.detailShell() })
      .last();
    await expect(detailDialog).toBeVisible();

    await marketplace.openDocsForVersion("v1.0.0");
    await marketplace.expectDocumentLayoutSeparated();
    await marketplace.expectDocumentCatalogEntries([
      "源码演示指南",
      "配置说明",
      "更新日志",
    ]);
    await expect(
      marketplace.documentCatalogNav().getByText("README", { exact: true }),
    ).toHaveCount(0);
    await expect(marketplace.documentPanel()).not.toContainText("README.md");
    await marketplace.expectRenderedMarkdownHeading(/源码演示指南/u);

    // Locale switcher chrome (label + Segmented) when multiple locales exist;
    // resolved locale must not also appear as a standalone toolbar Tag.
    await marketplace.expectDocumentLocaleOptions(["zh-CN", "en-US"]);
    await marketplace.expectNoDuplicateDocumentLocaleTag("zh-CN");
    await marketplace.expectNoDuplicateDocumentLocaleTag("en-US");
    await expect(
      marketplace.documentToolbar().locator(".marketplace-doc-path-tag"),
    ).toBeVisible();

    await marketplace.selectDocumentCatalogEntry("更新日志");
    await marketplace.expectRenderedMarkdownHeading(/更新日志/u);
    await expect(marketplace.documentPanel()).toContainText("changelog.md");

    await marketplace.expectNoRawI18nKeys();
    await captureMarketplaceScreenshot(page, "mine-detail-docs-catalog-all");
  });
});
