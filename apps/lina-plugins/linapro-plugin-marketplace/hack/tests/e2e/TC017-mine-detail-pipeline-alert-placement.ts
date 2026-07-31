/**
 * TC017: Mine detail modal pipeline alert sits between overview and tabs.
 *
 * When a publisher opens plugin detail for a pending_review plugin, the
 * "verification passed / waiting for admin review" banner must appear after
 * the overview Descriptions table and before the Versions/Risks/Docs tabs —
 * not above the meta table.
 */
import { expect, test } from "@host-tests/fixtures/auth";

import { MarketplacePage } from "../pages/MarketplacePage";
import {
  captureMarketplaceScreenshot,
  installMarketplaceApiMocks,
  marketplaceSourcePluginId,
  openMarketplaceWorkbench,
} from "../support/marketplace-fixtures";

test.describe("TC-17 mine detail pipeline alert placement", () => {
  test("TC-17a: pending_review banner is between descriptions and tabs", async ({
    authenticatedPage,
  }) => {
    const page = authenticatedPage;
    await page.setViewportSize({ height: 960, width: 1440 });
    await installMarketplaceApiMocks(page, {
      menuRole: "publish-only",
      pluginPatches: {
        [marketplaceSourcePluginId()]: {
          marketStatus: "draft",
          processStatus: "pending_review",
        },
      },
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

    await marketplace.expectPipelineAlertBetweenDescriptionsAndTabs(
      "验证已通过，插件正在等待管理员审核；通过后才会出现在公开市场。",
    );

    await marketplace.expectNoRawI18nKeys();
    await captureMarketplaceScreenshot(
      page,
      "mine-detail-pipeline-alert-between-table-and-tabs",
    );
  });
});
