import { expect, test } from "@host-tests/fixtures/auth";

import { MarketplacePage } from "../pages/MarketplacePage";
import {
  callMarketplaceApiFromPage,
  captureMarketplaceScreenshot,
  installMarketplaceApiMocks,
  marketplaceDynamicPluginId,
  marketplaceExternalPluginId,
  marketplacePrivatePluginId,
  marketplaceSourcePluginId,
  openMarketplaceWorkbench,
  sourceMarketplaceZipUpload,
} from "../support/marketplace-fixtures";

test.describe("TC-1 marketplace publisher workspace", () => {
  test("TC-1a: publish-only users enforce ownership and reset the publish form", async ({
    authenticatedPage,
  }) => {
    const page = authenticatedPage;
    await page.setViewportSize({ height: 960, width: 1440 });
    const mockState = await installMarketplaceApiMocks(page, {
      menuRole: "publish-only",
    });
    await openMarketplaceWorkbench(page, "简体中文");

    await page.getByText("插件市场", { exact: true }).click();
    await expect(page.getByText("我的插件", { exact: true })).toBeVisible();
    await expect(page.getByText("插件列表", { exact: true })).toHaveCount(0);
    await expect(page.getByText("插件审核", { exact: true })).toHaveCount(0);

    const marketplace = new MarketplacePage(page);
    await marketplace.gotoMine();

    await expect(
      page.getByText("我的插件", { exact: true }).last(),
    ).toBeVisible();
    await marketplace.expectColumn("mine", "可见性");
    await marketplace.expectNoHorizontalPageOverflow();
    await marketplace.expectNoHorizontalMineTableOverflow();
    await expect(
      marketplace.mineRow(marketplaceSourcePluginId()),
    ).toContainText("公开");
    await expect(
      marketplace.mineRow(marketplaceDynamicPluginId()),
    ).toContainText("公开");
    await expect(
      marketplace.mineRow(marketplacePrivatePluginId()),
    ).toContainText("私有");
    await expect(
      marketplace.mineRow(marketplaceExternalPluginId()),
    ).toHaveCount(0);
    await captureMarketplaceScreenshot(page, "mine-first-load");

    await page.setViewportSize({ height: 768, width: 1024 });
    await marketplace.expectColumn("mine", "更新时间");
    await marketplace.expectNoHorizontalMineTableOverflow();
    await marketplace.expectMineRowsVerticallySeparated();
    const compactSourceRow = marketplace.mineRow(marketplaceSourcePluginId());
    await expect(compactSourceRow).toContainText("源码插件");
    await expect(compactSourceRow).toContainText("已发布");
    await expect(compactSourceRow).toContainText("公开");
    await expect(compactSourceRow).toContainText("v1.0.0");
    await expect(compactSourceRow).toContainText("已通过");
    await expect(compactSourceRow).toContainText("2026-07-10 16:00:00");
    await captureMarketplaceScreenshot(page, "mine-table-1024");
    await page.setViewportSize({ height: 960, width: 1440 });

    const pendingBeforeDeniedReview = mockState.releaseSnapshot(
      marketplaceDynamicPluginId(),
      "v2.1.0",
    );
    const deniedManagedList = await callMarketplaceApiFromPage(
      page,
      "market/managed-plugins",
    );
    const deniedReviewQueue = await callMarketplaceApiFromPage(
      page,
      "market/review-queue",
    );
    const deniedReviewDecision = await callMarketplaceApiFromPage(
      page,
      `market/plugins/${marketplaceDynamicPluginId()}/releases/v2.1.0/review`,
      {
        body: {
          message: "publish-only users must not review",
          reviewStatus: "approved",
        },
        method: "PUT",
      },
    );
    for (const response of [
      deniedManagedList,
      deniedReviewQueue,
      deniedReviewDecision,
    ]) {
      expect(response).toMatchObject({ code: 403, data: null });
    }
    expect(mockState.deniedRequests).toEqual([
      { method: "GET", path: "market/managed-plugins" },
      { method: "GET", path: "market/review-queue" },
      {
        method: "PUT",
        path: `market/plugins/${marketplaceDynamicPluginId()}/releases/v2.1.0/review`,
      },
    ]);
    expect(
      mockState.releaseSnapshot(marketplaceDynamicPluginId(), "v2.1.0"),
    ).toEqual(pendingBeforeDeniedReview);
    expect(mockState.reviewRequests).toEqual([]);

    await marketplace.filterMineByKeyword("private");
    await expect(
      marketplace.mineRow(marketplacePrivatePluginId()),
    ).toBeVisible();
    await expect(marketplace.mineRow(marketplaceSourcePluginId())).toHaveCount(
      0,
    );
    await captureMarketplaceScreenshot(page, "mine-visibility-filtered");

    await marketplace.openPublishDrawer();
    await marketplace.expectExistingPublisher("LinaPro (linapro)");
    await marketplace.fillPluginDraft({
      name: "Reset Probe",
      pluginId: "linapro-reset-probe",
      pluginType: "源码插件",
      summary: "This value must not survive closing the drawer.",
    });
    await marketplace.closePublishDrawer();
    await marketplace.openPublishDrawer();
    await marketplace.expectExistingPublisher("LinaPro (linapro)");
    await marketplace.expectPluginDraftReset();

    const pageErrors: string[] = [];
    const consoleErrors: string[] = [];
    page.on("pageerror", (error) => pageErrors.push(error.message));
    page.on("console", (message) => {
      if (message.type() === "error") {
        consoleErrors.push(message.text());
      }
    });
    await marketplace.submitEmptyPluginDraft();
    await expect(marketplace.validationErrors()).toHaveCount(4);
    expect(pageErrors).toEqual([]);
    expect(consoleErrors).toHaveLength(1);
    expect(consoleErrors[0]).toContain("validate error");
    await captureMarketplaceScreenshot(page, "mine-required-validation");
  });

  test("TC-1b: publishers submit new plugins and versions from Mine", async ({
    authenticatedPage,
  }) => {
    const page = authenticatedPage;
    await page.setViewportSize({ height: 960, width: 1440 });
    const mockState = await installMarketplaceApiMocks(page, {
      menuRole: "publish-only",
    });
    await openMarketplaceWorkbench(page, "简体中文");

    const marketplace = new MarketplacePage(page);
    await marketplace.gotoMine();
    await marketplace.openPublishDrawer();

    const pageErrors: string[] = [];
    const consoleErrors: string[] = [];
    page.on("pageerror", (error) => pageErrors.push(error.message));
    page.on("console", (message) => {
      if (message.type() === "error") {
        consoleErrors.push(message.text());
      }
    });

    const pluginId = "linapro-e2e-private";
    await marketplace.fillPluginDraft({
      name: "LinaPro E2E Private",
      pluginId,
      pluginType: "源码插件",
      summary: "用于验证发布流程与私有可见性的 E2E 插件。",
      visibility: "私有",
    });
    await marketplace.savePluginDraft();
    await expect(page.locator(".ant-message-notice").last()).toContainText(
      "插件草稿已保存",
    );

    await marketplace.fillReleaseUpload({ version: "v1.0.0" });
    await marketplace.setUploadFile(sourceMarketplaceZipUpload());
    await marketplace.uploadDraft();
    await expect(page.locator(".ant-message-notice").last()).toContainText(
      "版本包已上传",
    );
    await marketplace.submitLatestDraft();
    await expect(page.locator(".ant-message-notice").last()).toContainText(
      "版本已提交审核",
    );
    expect(mockState.uploadRequests).toEqual([
      { pluginId, pluginType: "source", version: "v1.0.0" },
    ]);
    await marketplace.closePublishDrawer();
    await expect(marketplace.mineRow(pluginId)).toContainText("v1.0.0");
    await expect(marketplace.mineRow(pluginId)).toContainText("已提交");
    await captureMarketplaceScreenshot(page, "mine-private-submitted");

    await marketplace.openNewVersionDrawer(marketplaceSourcePluginId());
    await marketplace.expectReleaseTarget(
      marketplaceSourcePluginId(),
      "源码插件",
    );
    await marketplace.fillReleaseUpload({ version: "v1.1.0" });
    await marketplace.setUploadFile(sourceMarketplaceZipUpload());
    await marketplace.uploadDraft();
    await expect(page.locator(".ant-message-notice").last()).toContainText(
      "版本包已上传",
    );
    await marketplace.submitLatestDraft();
    await expect(page.locator(".ant-message-notice").last()).toContainText(
      "版本已提交审核",
    );
    await marketplace.closePublishDrawer();
    await expect(
      marketplace.mineRow(marketplaceSourcePluginId()),
    ).toContainText("v1.1.0");
    await expect(
      marketplace.mineRow(marketplaceSourcePluginId()),
    ).toContainText("已提交");
    expect(mockState.uploadRequests).toEqual([
      { pluginId, pluginType: "source", version: "v1.0.0" },
      {
        pluginId: marketplaceSourcePluginId(),
        pluginType: "source",
        version: "v1.1.0",
      },
    ]);
    expect(pageErrors).toEqual([]);
    expect(consoleErrors).toEqual([]);
    await marketplace.expectNoRawI18nKeys();
    await captureMarketplaceScreenshot(page, "mine-existing-version-submitted");
  });
});
