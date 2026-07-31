/**
 * TC015: Marketplace risk list orders by blocking/disposition then severity.
 *
 * Owner detail must surface need_fix (blocking) before need_attention and
 * still localize finding titles. info_only tips are hidden from the risk tab.
 * Overview risk summary Tags remain severity-count ordered high → warning → info.
 */
import { expect, test } from "@host-tests/fixtures/auth";

import { MarketplacePage } from "../pages/MarketplacePage";
import {
  captureMarketplaceScreenshot,
  installMarketplaceApiMocks,
  marketplaceReadmeOnlyGitPluginId,
  marketplaceReadmeOnlyGitVersion,
  openMarketplaceWorkbench,
} from "../support/marketplace-fixtures";

test.describe("TC-15 risk severity order", () => {
  test("TC-15a: risk tab sorts fixable first; summary keeps severity order", async ({
    authenticatedPage,
  }) => {
    const page = authenticatedPage;
    await page.setViewportSize({ height: 960, width: 1440 });
    // Deliberately reverse presentation order so UI sort is what assertions exercise.
    await installMarketplaceApiMocks(page, {
      gitSourceRisks: [
        {
          // Framework compatibility is non-blocking need_attention.
          payload: { code: "framework_dependency_missing" },
          severity: "warning",
          source: "plugin.yaml",
          summary: "Framework compatibility dependency is not declared.",
          type: "dependency",
        },
        {
          // Blocking need_fix item used for sort priority.
          payload: { code: "i18n_files_missing" },
          severity: "warning",
          source: "manifest/i18n",
          summary:
            "Plugin declares i18n.enabled but no manifest i18n JSON files were detected.",
          type: "dependency",
        },
        {
          payload: { code: "dynamic_host_services_present" },
          severity: "high",
          source: "hostServices",
          summary: "Dynamic package requests host service authorization.",
          type: "host_service",
        },
        // info_only tip must be hidden from the risk tab list.
        {
          payload: { code: "source_docs_indexed" },
          severity: "info",
          source: "manifest/docs",
          summary: "Marketplace documentation entries were detected.",
          type: "docs",
        },
      ],
      menuRole: "publish-only",
    });
    await openMarketplaceWorkbench(page, "简体中文");

    const marketplace = new MarketplacePage(page);
    await marketplace.gotoMine();
    await marketplace.openMineDetail(marketplaceReadmeOnlyGitPluginId());

    // Overview risk summary Tags still reflect API severity counts (including hidden info_only).
    await marketplace.expectRiskSummaryTag("high", 1);
    await marketplace.expectRiskSummaryTag("warning", 2);
    await marketplace.expectRiskSummaryTag("info", 1);
    await marketplace.expectRiskSummarySeverityOrder([
      "high",
      "warning",
      "info",
    ]);
    await captureMarketplaceScreenshot(page, "risk-summary-severity-order");

    await marketplace.openRisksForVersion(marketplaceReadmeOnlyGitVersion());
    // Risk tab hides info_only; only need_fix + need_attention remain.
    await marketplace.expectRiskListCount(3);
    await marketplace.expectRiskListDispositionOrder([
      "need_fix",
      "need_attention",
      "need_attention",
    ]);
    await marketplace.expectRiskListSeverityOrder([
      "warning",
      "high",
      "warning",
    ]);

    await marketplace.expectRiskFindingText(
      "插件声明了 i18n.enabled，但未检测到 manifest i18n JSON 文件。",
    );
    await marketplace.expectRiskFindingText("未声明框架兼容性依赖。");
    await marketplace.expectRiskFindingText("动态包请求宿主服务授权。");
    await expect(
      marketplace
        .detailShell()
        .getByText("已检测到可用于市场展示的文档条目。"),
    ).toHaveCount(0);
    await expect(
      marketplace.detailShell().getByRole("button", { name: /仅提示/u }),
    ).toHaveCount(0);
    // Framework compatibility must not show the blocking-submit tag.
    // Use exact match: the intro line also contains the substring "阻塞提交".
    await expect(
      marketplace.detailShell().getByText("阻塞提交", { exact: true }),
    ).toHaveCount(1);
    await expect(
      marketplace
        .detailShell()
        .getByText("Dynamic package requests host service authorization."),
    ).toHaveCount(0);

    await captureMarketplaceScreenshot(page, "risk-list-severity-order");
  });
});
