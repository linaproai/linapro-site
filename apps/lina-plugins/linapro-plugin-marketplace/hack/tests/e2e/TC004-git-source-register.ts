/**
 * TC004: Git source registration is available inside the Add Plugin drawer.
 *
 * Validates the simplified add surface exposes Git repository submission
 * as a distribution mode of "Add Plugin", without visibility selection.
 */
import { expect, test } from "@host-tests/fixtures/auth";

import { MarketplacePage } from "../pages/MarketplacePage";
import {
  captureMarketplaceScreenshot,
  installMarketplaceApiMocks,
  openMarketplaceWorkbench,
} from "../support/marketplace-fixtures";

test.describe("TC004 marketplace git source register entry", () => {
  test("shows git distribution mode inside add plugin drawer", async ({
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
    await marketplace.expectNoRegisterGitToolbarAction();

    await marketplace.openPublishDrawer();
    await expect(
      marketplace.publishDrawer().locator("h2").filter({ hasText: "添加插件" }),
    ).toBeVisible();
    await marketplace.selectPublishSourceKind("git");
    await expect(
      marketplace
        .publishDrawer()
        .locator("label")
        .filter({ hasText: "仓库地址" }),
    ).toBeVisible();
    await expect(
      marketplace
        .publishDrawer()
        .locator(".ant-form-item:visible")
        .filter({ hasText: "插件 ID" }),
    ).toHaveCount(0);
    await expect(
      marketplace
        .publishDrawer()
        .locator(".ant-form-item:visible")
        .filter({ hasText: "可见性" }),
    ).toHaveCount(0);
    await marketplace.expectAddPluginDrawerLayout();
    await expect(
      marketplace
        .publishDrawer()
        .locator(".mine-drawer-actions")
        .getByRole("button", { name: "添加插件" }),
    ).toBeVisible();
    await captureMarketplaceScreenshot(page, "mine-git-distribution-mode");
  });
});
