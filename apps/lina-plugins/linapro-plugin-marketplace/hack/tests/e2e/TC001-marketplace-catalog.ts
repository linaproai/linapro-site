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

    const marketplace = new MarketplacePage(page);
    await marketplace.gotoMine();
    await expect(
      page.getByText("我的插件", { exact: true }).first(),
    ).toBeVisible();
    await expect(page.getByText("插件列表", { exact: true })).toHaveCount(0);
    await expect(page.getByText("插件审核", { exact: true })).toHaveCount(0);

    await expect(
      page.getByText("我的插件", { exact: true }).last(),
    ).toBeVisible();
    await marketplace.expectColumn("mine", "可见性");
    await marketplace.expectColumn("mine", "下载量");
    await marketplace.expectMineSourceColumn();
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
    await marketplace.expectMineSearchFormStatusOptions();
    await marketplace.expectNoRegisterGitToolbarAction();
    await expect(
      page.getByRole("button", { name: "登记发布者" }),
    ).toBeVisible();
    await expect(page.getByRole("button", { name: "添加插件" })).toBeVisible();
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
    // Download count is projected next to latest version; review status is no
    // longer a dedicated mine-table column.
    await expect(compactSourceRow).toContainText("128");
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

    await marketplace.openPublisherDrawer();
    await marketplace.expectPublisherKeyHidden();
    await marketplace.expectPublisherFormValues({
      contactEmail: "plugins@linapro.ai",
      homepage: "https://linapro.ai",
      name: "LinaPro",
      summary: "Official LinaPro publisher",
    });
    await expect(marketplace.publisherDrawer()).toContainText("编辑发布者");
    await captureMarketplaceScreenshot(page, "mine-publisher-drawer");
    await marketplace.closePublisherDrawer();

    await marketplace.openPublishDrawer();
    await expect(
      marketplace.publishDrawer().locator("h2").filter({ hasText: "添加插件" }),
    ).toBeVisible();
    // Drawer body must not repeat the same title under the chrome header.
    await expect(
      marketplace
        .publishDrawer()
        .locator("h3")
        .filter({ hasText: "添加插件" }),
    ).toHaveCount(0);
    await expect(
      marketplace.publishDrawer().getByText("分发方式"),
    ).toBeVisible();
    await expect(
      marketplace
        .publishDrawer()
        .locator("label.ant-radio-button-wrapper")
        .filter({ hasText: "上传包" }),
    ).toBeVisible();
    await expect(
      marketplace
        .publishDrawer()
        .locator("label.ant-radio-button-wrapper")
        .filter({ hasText: "Git 仓库" }),
    ).toBeVisible();
    await expect(
      marketplace.publishDrawer().getByText("发布者资料"),
    ).toHaveCount(0);
    await expect(marketplace.publishDrawer().getByText("可见性")).toHaveCount(
      0,
    );
    await marketplace.expectAddPluginDrawerLayout();
    // Package upload sits in the same form label column as distribution mode.
    await expect(
      marketplace
        .publishDrawer()
        .locator("label")
        .filter({ hasText: "上传压缩包" }),
    ).toBeVisible();
    await expect(
      marketplace.publishDrawer().getByTestId("mine-package-field"),
    ).toBeVisible();
    await marketplace.setUploadFile(sourceMarketplaceZipUpload());
    await marketplace.closePublishDrawer();
    await marketplace.openPublishDrawer();
    await marketplace.expectPluginDraftReset();

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
    await expect(
      marketplace.publishDrawer().getByTestId("mine-package-field"),
    ).toBeHidden();
    await expect(
      marketplace
        .publishDrawer()
        .locator(".mine-drawer-actions")
        .getByRole("button", { name: "添加插件" }),
    ).toBeVisible();
    await marketplace.expectAddPluginDrawerLayout();
    await captureMarketplaceScreenshot(page, "mine-publish-git-mode");
    await marketplace.selectPublishSourceKind("upload");
    await marketplace.expectAddPluginDrawerLayout();
    await expect(
      marketplace.publishDrawer().getByTestId("mine-package-field"),
    ).toBeVisible();

    const pageErrors: string[] = [];
    const consoleErrors: string[] = [];
    page.on("pageerror", (error) => pageErrors.push(error.message));
    page.on("console", (message) => {
      if (message.type() === "error") {
        consoleErrors.push(message.text());
      }
    });
    await marketplace.submitEmptyPluginDraft();
    await expect(page.locator(".ant-message-notice").last()).toContainText(
      /请先选择|ZIP|file/i,
    );
    expect(pageErrors).toEqual([]);
    await captureMarketplaceScreenshot(page, "mine-required-validation");
  });

  test("TC-1b: publishers add packages then publish versions from Mine", async ({
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
    await marketplace.setUploadFile({
      buffer: Buffer.from("e2e private marketplace zip placeholder"),
      mimeType: "application/zip",
      name: "linapro-e2e-private-v1.0.0.zip",
    });
    await marketplace.uploadDraft();
    await expect(page.locator(".ant-message-notice").last()).toContainText(
      "插件已添加",
    );
    expect(mockState.uploadRequests).toEqual([
      { pluginId, pluginType: "source", version: "v1.0.0" },
    ]);
    await expect(marketplace.mineRow(pluginId)).toContainText("v1.0.0");
    // Add-plugin lands in pending_verify; UI shows 待验证 instead of raw 草稿.
    // Row actions stay Detail / New Version / Delist only (no publish/more).
    await expect(marketplace.mineRow(pluginId)).toContainText("待验证");
    await marketplace.expectMineRowActions(pluginId);
    await marketplace.expectMineRowActions(marketplaceSourcePluginId());
    await captureMarketplaceScreenshot(page, "mine-private-pending-verify");

    await marketplace.openNewVersionDrawer(marketplaceSourcePluginId());
    await marketplace.expectReleaseTarget(
      marketplaceSourcePluginId(),
      "源码插件",
    );
    await marketplace.setUploadFile({
      ...sourceMarketplaceZipUpload(),
      name: "linapro-demo-source-v1.1.0.zip",
    });
    await marketplace.uploadDraft();
    await expect(page.locator(".ant-message-notice").last()).toContainText(
      "插件已添加",
    );
    await expect(
      marketplace.mineRow(marketplaceSourcePluginId()),
    ).toContainText("v1.1.0");
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
