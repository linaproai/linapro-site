/**
 * TC014: Marketplace lastSyncMessage renders localized sync diagnostics.
 *
 * Git discovery persists English source text in lastSyncMessage. The owner
 * detail overview must map known patterns through plugin runtime i18n so zh-CN
 * users never see phrases like "discovered 0 new draft releases...".
 */
import { expect, test } from "@host-tests/fixtures/auth";

import { MarketplacePage } from "../pages/MarketplacePage";
import {
  captureMarketplaceScreenshot,
  installMarketplaceApiMocks,
  marketplaceReadmeOnlyGitPluginId,
  openMarketplaceWorkbench,
} from "../support/marketplace-fixtures";

test.describe("TC-14 lastSyncMessage i18n", () => {
  test("TC-14a: zh-CN detail shows translated sync message", async ({
    authenticatedPage,
  }) => {
    const page = authenticatedPage;
    await page.setViewportSize({ height: 960, width: 1440 });
    await installMarketplaceApiMocks(page, {
      menuRole: "publish-only",
    });
    await openMarketplaceWorkbench(page, "简体中文");

    const marketplace = new MarketplacePage(page);
    await marketplace.gotoMine();
    await marketplace.openMineDetail(marketplaceReadmeOnlyGitPluginId());

    // Field label stays localized.
    await expect(
      marketplace.detailShell().getByText("同步信息").first(),
    ).toBeVisible();

    // Body uses zh-CN syncMessage translation, not English source.
    await marketplace.expectLastSyncMessageText(
      "未发现新草稿版本（已有 1 个不可变版本）",
    );
    await marketplace.expectLastSyncMessageAbsent(
      "discovered 0 new draft releases (1 existing immutable version(s))",
    );
    await marketplace.expectLastSyncMessageAbsent(
      "discovered 0 new draft releases",
    );

    // Risk summary tags remain Chinese severity counts (adjacent field).
    await marketplace.expectRiskSummaryTag("info", 1);

    // Sync message value must share Descriptions body font-size with peer fields
    // (not marketplace-muted 12px).
    const syncContent = marketplace.detailDescriptionsContentByLabel(
      /同步信息|Sync Message/u,
    );
    const licenseContent = marketplace.detailDescriptionsContentByLabel(
      /许可证|License/u,
    );
    await expect(syncContent).toBeVisible();
    await expect(licenseContent).toBeVisible();
    const syncFontSize = await syncContent.evaluate(
      (el) => getComputedStyle(el).fontSize,
    );
    const licenseFontSize = await licenseContent.evaluate(
      (el) => getComputedStyle(el).fontSize,
    );
    expect(syncFontSize).toBe(licenseFontSize);
    expect(syncFontSize).not.toBe("12px");

    await captureMarketplaceScreenshot(page, "last-sync-message-i18n-zh");
  });
});
