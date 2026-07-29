import { expect, test } from "@host-tests/fixtures/auth";

import { MarketplacePage } from "../pages/MarketplacePage";
import {
  captureMarketplaceScreenshot,
  installMarketplaceApiMocks,
  marketplaceDynamicPluginId,
  marketplaceExternalPluginId,
  marketplaceSourcePluginId,
  openMarketplaceWorkbench,
} from "../support/marketplace-fixtures";

test.describe("TC-2 marketplace reviewer workspace", () => {
  test("TC-2a: review-only users isolate inspections and decide the default queue", async ({
    authenticatedPage,
  }) => {
    const page = authenticatedPage;
    await page.setViewportSize({ height: 960, width: 1440 });
    const mockState = await installMarketplaceApiMocks(page, {
      inspectionDelayMsByRelease: {
        [`${marketplaceExternalPluginId()}@v0.9.0`]: 600,
      },
      menuRole: "review-only",
    });
    await openMarketplaceWorkbench(page, "简体中文");

    await page.getByText("插件市场", { exact: true }).click();
    await expect(page.getByText("我的插件", { exact: true })).toHaveCount(0);
    await expect(page.getByText("插件列表", { exact: true })).toBeVisible();
    await expect(page.getByText("插件审核", { exact: true })).toBeVisible();

    const marketplace = new MarketplacePage(page);
    await marketplace.gotoAdminList();
    await expect(
      page.getByText("插件列表", { exact: true }).last(),
    ).toBeVisible();
    await expect(
      marketplace.adminRow(marketplaceSourcePluginId()),
    ).toContainText("LinaPro 源码演示");
    await expect(
      marketplace.adminRow(marketplaceDynamicPluginId()),
    ).toContainText("LinaPro 动态插件演示");
    await expect(
      marketplace.adminRow(marketplaceExternalPluginId()),
    ).toContainText("Acme 可观测性");
    // Table chrome matches My Plugins: separate pluginId/name, status early.
    // Updated At may sit behind horizontal scroll at mid widths.
    await marketplace.expectColumn("admin", "插件标识");
    await marketplace.expectColumn("admin", "名称");
    await marketplace.expectColumn("admin", "状态");
    await marketplace.expectColumnBeforeFixedActions("admin", "状态");
    await marketplace.expectColumnBeforeFixedActions("admin", "审核");
    await captureMarketplaceScreenshot(page, "admin-list-first-load");

    await marketplace.filterAdminByKeyword("source");
    await expect(
      marketplace.adminRow(marketplaceSourcePluginId()),
    ).toBeVisible();
    await expect(
      marketplace.adminRow(marketplaceDynamicPluginId()),
    ).toHaveCount(0);
    await captureMarketplaceScreenshot(page, "admin-list-filtered");

    await marketplace.openAdminDetail(marketplaceSourcePluginId());
    await expect(marketplace.detailShell()).toContainText("LinaPro 源码演示");
    await marketplace.backFromDetail("admin-list");
    await expect(
      marketplace.adminRow(marketplaceSourcePluginId()),
    ).toBeVisible();
    await marketplace.expectWorkspaceTabCount("插件列表", 1);

    await page.setViewportSize({ height: 768, width: 1024 });
    await marketplace.gotoReview();
    await expect(
      page.getByText("待审版本", { exact: true }).last(),
    ).toBeVisible();
    await expect(
      marketplace.reviewRow(marketplaceDynamicPluginId(), "v2.1.0"),
    ).toContainText("LinaPro 动态插件演示");
    await expect(
      marketplace.reviewRow(marketplaceExternalPluginId(), "v0.9.0"),
    ).toContainText("Acme 可观测性");
    await marketplace.expectVisibleCellContentFits(
      "review",
      "review-submitted-at-column",
    );
    await marketplace.expectColumnBeforeFixedActions("review", "提交时间");
    await marketplace.expectColumnBeforeFixedActions("review", "版本");
    await marketplace.expectNoHorizontalPageOverflow();
    expect(mockState.reviewQueueRequests.length).toBeGreaterThan(0);
    expect(mockState.reviewQueueRequests[0]?.get("pluginId")).toBeNull();
    await captureMarketplaceScreenshot(page, "review-default-queue");

    await marketplace.inspectReviewRelease(
      marketplaceExternalPluginId(),
      "v0.9.0",
    );
    await marketplace.closeReviewPanel();
    await marketplace.inspectReviewRelease(
      marketplaceDynamicPluginId(),
      "v2.1.0",
    );
    await expect(marketplace.reviewPanel()).toBeVisible();
    await expect(marketplace.reviewPanel()).toContainText("清单一致性");
    await expect(marketplace.reviewPanel()).toContainText("宿主服务");
    await expect(marketplace.reviewPanel()).toContainText("需审核");
    await marketplace.expectReviewPanelContains("发现 2 条风险");
    await marketplace.expectReviewPanelContains("plugin.wasm");
    await marketplace.expectReviewPanelContains(
      "Requests host service access for reviewer approval.",
    );
    await marketplace.expectReviewPanelContains(
      "Registers a dynamic route for reviewer inspection.",
    );
    await expect
      .poll(
        () =>
          mockState.inspectionResponses.filter(
            (item) => item.pluginId === marketplaceExternalPluginId(),
          ).length,
      )
      .toBe(2);
    await marketplace.expectReviewPanelNotContains("发现 1 条风险");
    await captureMarketplaceScreenshot(page, "review-inspection");
    await marketplace.closeReviewPanel();

    await marketplace.inspectReviewRelease(
      marketplaceExternalPluginId(),
      "v0.9.0",
    );
    await marketplace.expectReviewPanelContains("发现 1 条风险");
    await marketplace.expectReviewPanelContains(
      "Requests a menu permission that requires reviewer approval.",
    );
    await marketplace.submitReviewDecision("已拒绝", "E2E 审核拒绝");
    await expect(page.locator(".ant-message-notice").last()).toContainText(
      "审核结论已保存",
    );
    expect(mockState.reviewRequests).toContainEqual({
      message: "E2E 审核拒绝",
      pluginId: marketplaceExternalPluginId(),
      releaseStatus: "draft",
      status: "rejected",
      version: "v0.9.0",
    });
    expect(
      mockState.releaseSnapshot(marketplaceExternalPluginId(), "v0.9.0"),
    ).toEqual({
      releaseStatus: "draft",
      reviewMessage: "E2E 审核拒绝",
      reviewStatus: "rejected",
    });
    await expect(
      marketplace.reviewRow(marketplaceExternalPluginId(), "v0.9.0"),
    ).toHaveCount(0);
    await expect(marketplace.reviewPanel()).toBeHidden();

    await marketplace.inspectReviewRelease(
      marketplaceDynamicPluginId(),
      "v2.1.0",
    );
    await marketplace.expectReviewPanelContains("发现 2 条风险");
    await marketplace.submitReviewDecision("已通过", "E2E 审核通过");
    await expect(page.locator(".ant-message-notice").last()).toContainText(
      "审核结论已保存",
    );
    expect(mockState.reviewRequests).toContainEqual({
      message: "E2E 审核通过",
      pluginId: marketplaceDynamicPluginId(),
      releaseStatus: "published",
      status: "approved",
      version: "v2.1.0",
    });
    await expect(
      marketplace.reviewRow(marketplaceDynamicPluginId(), "v2.1.0"),
    ).toHaveCount(0);
    await expect(marketplace.reviewPanel()).toBeHidden();
    await marketplace.expectReviewEmptyState("当前没有待审核版本。");
    await marketplace.expectNoRawI18nKeys();
    await captureMarketplaceScreenshot(page, "review-approved");
  });
});
