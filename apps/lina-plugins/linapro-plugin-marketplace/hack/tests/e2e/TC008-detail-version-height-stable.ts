/**
 * TC008: Detail modal version tab body height stays content-sized.
 *
 * Regression for the fill-parent vxe-table height loop inside the marketplace
 * detail modal: opening "My Plugins" detail and staying on the Versions tab
 * must not keep expanding blank space under the release rows.
 */
import { expect, test } from "@host-tests/fixtures/auth";

import { MarketplacePage } from "../pages/MarketplacePage";
import {
  captureMarketplaceScreenshot,
  installMarketplaceApiMocks,
  marketplaceSourcePluginId,
  openMarketplaceWorkbench,
} from "../support/marketplace-fixtures";

test.describe("TC-8 marketplace detail version tab height stability", () => {
  test("TC-8a: mine detail modal versions tab does not expand blank body space", async ({
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
    await expect(marketplace.detailShell()).toContainText(
      marketplaceSourcePluginId(),
    );

    // Stay on the default Versions tab and assert body height is stable.
    await marketplace.expectVersionTableHeightStable({
      samples: 5,
      settleMs: 300,
    });
    await expect(marketplace.detailReleaseRow("v1.0.0")).toBeVisible();
    await expect(marketplace.detailReleaseRow("v0.9.0")).toBeVisible();

    await marketplace.expectNoRawI18nKeys();
    await captureMarketplaceScreenshot(
      page,
      "mine-detail-version-height-stable",
    );
  });
});
