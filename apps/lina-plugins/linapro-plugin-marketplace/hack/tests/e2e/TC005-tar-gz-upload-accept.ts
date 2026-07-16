/**
 * TC005: Upload control accepts zip and tar.gz packages.
 *
 * Confirms the publisher upload surface advertises .zip/.tar.gz/.tgz.
 */
import { expect, test } from "@host-tests/fixtures/auth";

import { MarketplacePage } from "../pages/MarketplacePage";
import {
  installMarketplaceApiMocks,
  openMarketplaceWorkbench,
} from "../support/marketplace-fixtures";

test.describe("TC005 marketplace tar.gz upload accept", () => {
  test("upload dragger accepts tar.gz packages", async ({
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
    await marketplace.openPublishDrawer();
    await marketplace.expectAddPluginDrawerLayout();

    const accept = page
      .getByTestId("mine-package-field")
      .locator('input[type="file"]')
      .first();
    await expect(accept).toBeAttached();
    await expect(accept).toHaveAttribute("accept", /\.tar\.gz/);
    await expect(accept).toHaveAttribute("accept", /\.zip/);
    await expect(accept).toHaveAttribute("accept", /\.tgz/);
  });
});
