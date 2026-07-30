import { expect, test } from "@host-tests/fixtures/auth";
import { config } from "@host-tests/fixtures/config";
import { LoginPage } from "@host-tests/pages/LoginPage";

import { MarketplacePage } from "../pages/MarketplacePage";
import {
  captureMarketplaceScreenshot,
  installMarketplaceApiMocks,
  marketplacePrivatePluginId,
  marketplaceSourcePluginId,
  openMarketplaceWorkbench,
} from "../support/marketplace-fixtures";

test.describe("TC-3 marketplace English workspace", () => {
  test("TC-3a: both-role users see English menus, details, risks, and localized download errors", async ({
    authenticatedPage,
  }) => {
    const page = authenticatedPage;
    const detailModalBridgeErrors: string[] = [];
    page.on("pageerror", (error) => {
      if (/setData is not a function/u.test(error.message)) {
        detailModalBridgeErrors.push(error.message);
      }
    });
    await page.setViewportSize({ height: 960, width: 1440 });
    const mockState = await installMarketplaceApiMocks(page, {
      failDownloadsFor: [marketplaceSourcePluginId()],
      menuRole: "both",
    });
    await openMarketplaceWorkbench(page, "English");

    await page.getByText("Plugin Marketplace", { exact: true }).click();
    await expect(page.getByText("My Plugins", { exact: true })).toBeVisible();
    await expect(page.getByText("Plugin List", { exact: true })).toBeVisible();
    await expect(
      page.getByText("Plugin Review", { exact: true }),
    ).toBeVisible();

    const marketplace = new MarketplacePage(page);
    await marketplace.gotoMine();
    await expect(
      page.getByText("My Plugins", { exact: true }).last(),
    ).toBeVisible();
    await marketplace.expectColumn("mine", "Status");
    await marketplace.expectColumn("mine", "Plugin Identifier");
    await marketplace.expectColumn("mine", "Plugin Name");
    // Visibility is not a user-facing mine column (status drives catalog).
    await expect(
      page
        .locator(
          ".plugin-marketplace-mine .vxe-table--main-wrapper .vxe-header--column",
        )
        .filter({ hasText: "Visibility" }),
    ).toHaveCount(0);
    await expect(
      marketplace.mineRow(marketplaceSourcePluginId()),
    ).toContainText("Published");
    await expect(
      marketplace.mineRow(marketplacePrivatePluginId()),
    ).toBeVisible();
    await marketplace.expectColumnHeaderFits("mine", "Latest Version");
    await marketplace.expectColumnHeaderFits("mine", "Downloads");
    await marketplace.expectColumnBeforeFixedActions("mine", "Status");
    await captureMarketplaceScreenshot(page, "english-my-plugins");

    await marketplace.gotoAdminList();
    await expect(
      page.getByText("Plugin List", { exact: true }).last(),
    ).toBeVisible();
    await marketplace.expectColumnHeaderFits("admin", "Latest Version");
    await marketplace.expectColumnHeaderFits("admin", "Downloads");
    await marketplace.expectColumnBeforeFixedActions("admin", "Review");
    await captureMarketplaceScreenshot(page, "english-plugin-list");
    await marketplace.gotoReview();
    await expect(
      page.getByText("Review Queue", { exact: true }).last(),
    ).toBeVisible();
    await marketplace.expectColumnHeaderFits("review", "Submitted At");
    await marketplace.expectVisibleCellContentFits(
      "review",
      "review-submitted-at-column",
    );
    await captureMarketplaceScreenshot(page, "english-workspaces");

    await marketplace.gotoDetail(marketplaceSourcePluginId(), "admin-list");
    await expect(marketplace.detailShell()).toContainText(
      "LinaPro Source Demo",
    );
    await marketplace.backFromDetail("admin-list");

    await marketplace.gotoDetail(marketplaceSourcePluginId(), "review");
    await expect(marketplace.detailShell()).toContainText(
      "LinaPro Source Demo",
    );
    await marketplace.backFromDetail("review");

    await page.setViewportSize({ height: 844, width: 390 });
    await marketplace.gotoDetail(marketplaceSourcePluginId(), "mine");
    await expect(marketplace.detailShell()).toContainText(
      "LinaPro Source Demo",
    );
    await expect(marketplace.detailShell()).toContainText("Latest Version");
    await marketplace.expectNoHorizontalPageOverflow();

    await marketplace.openDocsForVersion("v1.0.0");
    await expect(page.getByText("Source Demo Guide").first()).toBeVisible();
    await expect(
      page.getByText("Place the source under apps/lina-plugins"),
    ).toBeVisible();
    await marketplace.expectDocumentLocaleOptions(["en-US", "zh-CN"]);
    const docsRequestCount = mockState.inspectionResponses.filter(
      (response) => response.kind === "docs",
    ).length;
    await marketplace.switchDocumentLocale("zh-CN");
    await expect(page.getByText("源码演示指南").first()).toBeVisible();
    await expect(page.getByText("将源码放入 apps/lina-plugins")).toBeVisible();
    // Locale switches re-fetch the selected path so catalog titles and markdown
    // stay aligned with the preferred locale from the server.
    expect(
      mockState.inspectionResponses.filter(
        (response) => response.kind === "docs",
      ).length,
    ).toBe(docsRequestCount + 1);
    await marketplace.switchDocumentLocale("en-US");
    await expect(page.getByText("Source Demo Guide").first()).toBeVisible();
    expect(
      mockState.inspectionResponses.filter(
        (response) => response.kind === "docs",
      ).length,
    ).toBe(docsRequestCount + 2);
    await captureMarketplaceScreenshot(page, "english-detail-docs");

    await marketplace.openRisksForVersion("v1.0.0");
    await expect(page.getByText("Install SQL", { exact: true })).toBeVisible();
    await expect(
      page.getByText("Install SQL creates plugin-owned demo tables."),
    ).toBeVisible();
    await captureMarketplaceScreenshot(page, "english-detail-risks");

    await marketplace.confirmDownloadForVersion("v1.0.0");
    await expect(page.locator(".ant-message-notice").last()).toContainText(
      "Marketplace download session does not exist",
    );
    expect(mockState.downloadRequests).toEqual([
      { pluginId: marketplaceSourcePluginId(), version: "v1.0.0" },
    ]);
    expect(mockState.dynamicUploadRequests).toBe(0);
    await marketplace.expectNoRawI18nKeys();
    await captureMarketplaceScreenshot(page, "english-download-denied");

    await marketplace.backFromDetail("mine");
    await expect(
      marketplace.mineRow(marketplaceSourcePluginId()),
    ).toBeVisible();

    await marketplace.gotoDashboard();
    await marketplace.gotoDetail(marketplaceSourcePluginId());
    await marketplace.backFromDetail("dashboard/analytics");
    expect(detailModalBridgeErrors).toEqual([]);
  });

  test("TC-3b: users without download permission do not see the download action", async ({
    browser,
  }) => {
    const context = await browser.newContext({ baseURL: config.baseURL });
    const page = await context.newPage();
    try {
      await installMarketplaceApiMocks(page, { menuRole: "publish-only" });
      await page.route("**/api/v1/user/info", async (route) => {
        const response = await route.fetch();
        const body = (await response.json()) as {
          data?: { permissions?: string[] };
        };
        if (body.data) {
          body.data.permissions = [
            "market:plugin:publish",
            "market:plugin:view",
          ];
        }
        await route.fulfill({ json: body, response });
      });

      const loginPage = new LoginPage(page);
      await loginPage.goto();
      await loginPage.loginAndWaitForRedirect(
        config.adminUser,
        config.adminPass,
      );
      await openMarketplaceWorkbench(page, "English");

      const marketplace = new MarketplacePage(page);
      await marketplace.gotoDetail(marketplaceSourcePluginId(), "mine");
      await marketplace.expectDownloadActionHidden("v1.0.0");
      await marketplace.expectNoRawI18nKeys();
      await captureMarketplaceScreenshot(page, "download-action-hidden");
    } finally {
      await context.close();
    }
  });
});
