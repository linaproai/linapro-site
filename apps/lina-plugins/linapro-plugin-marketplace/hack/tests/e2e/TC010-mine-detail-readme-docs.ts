/**
 * TC010: My plugin detail shows README-only Git repository docs.
 *
 * Validates a Git-sourced plugin release without manifest/docs still renders
 * its package-root README in the Docs tab as the current version document.
 */
import { expect, test } from "@host-tests/fixtures/auth";

import { MarketplacePage } from "../pages/MarketplacePage";
import {
  captureMarketplaceScreenshot,
  installMarketplaceApiMocks,
  marketplaceReadmeOnlyGitPluginId,
  openMarketplaceWorkbench,
} from "../support/marketplace-fixtures";

test.describe("TC-10 mine detail README-only Git docs", () => {
  test("TC-10a: docs tab renders README fallback for Git release", async ({
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
    await marketplace.openMineDetail(marketplaceReadmeOnlyGitPluginId());

    const detailDialog = page
      .locator('[role="dialog"]:visible')
      .filter({ has: marketplace.detailShell() })
      .last();
    await expect(detailDialog).toBeVisible();
    await expect(detailDialog).toContainText(
      "https://github.com/linaproai/official-plugins.git",
    );

    await marketplace.openDocsForVersion("v0.1.0");
    await marketplace.expectDocumentCatalogEntries(["LinaPro 租户核心"]);
    await marketplace.expectRenderedMarkdownHeading(/LinaPro 租户核心/u);
    await expect(marketplace.documentPanel()).toContainText("index.md");
    await expect(marketplace.documentPanel()).not.toContainText(
      "所选版本暂无可用文档",
    );
    await expect(marketplace.documentMarkdownBody()).toContainText(
      "租户上下文 API",
    );

    await marketplace.expectNoRawI18nKeys();
    await captureMarketplaceScreenshot(page, "mine-detail-readme-docs");
  });
});
