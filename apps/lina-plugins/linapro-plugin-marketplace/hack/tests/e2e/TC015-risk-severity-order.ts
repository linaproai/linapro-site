/**
 * TC015: Marketplace risk list and summary order by severity (high first).
 *
 * The risks API historically returned insertion/id order. Owner detail must
 * still present high → warning → info in both the overview risk summary Tags
 * and the risk Tab rows. This case feeds the mock API findings out of order
 * and asserts the UI reorders them for presentation.
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
  test("TC-15a: risk tab and summary sort high before warning and info", async ({
    authenticatedPage,
  }) => {
    const page = authenticatedPage;
    await page.setViewportSize({ height: 960, width: 1440 });
    // Deliberately reverse the desired presentation order so the UI sort is
    // what the assertions exercise (info → warning → high in the API payload).
    await installMarketplaceApiMocks(page, {
      gitSourceRisks: [
        {
          payload: { code: "source_docs_indexed" },
          severity: "info",
          source: "manifest/docs",
          summary: "Marketplace documentation entries were detected.",
          type: "docs",
        },
        {
          payload: { code: "framework_dependency_missing" },
          severity: "warning",
          source: "plugin.yaml",
          summary: "Framework compatibility dependency is not declared.",
          type: "dependency",
        },
        {
          payload: { code: "dynamic_host_services_present" },
          severity: "high",
          source: "hostServices",
          summary: "Dynamic package requests host service authorization.",
          type: "host_service",
        },
      ],
      menuRole: "publish-only",
    });
    await openMarketplaceWorkbench(page, "简体中文");

    const marketplace = new MarketplacePage(page);
    await marketplace.gotoMine();
    await marketplace.openMineDetail(marketplaceReadmeOnlyGitPluginId());

    // Overview risk summary Tags: high → warning → info (not API order).
    await marketplace.expectRiskSummaryTag("high", 1);
    await marketplace.expectRiskSummaryTag("warning", 1);
    await marketplace.expectRiskSummaryTag("info", 1);
    await marketplace.expectRiskSummarySeverityOrder([
      "high",
      "warning",
      "info",
    ]);
    await captureMarketplaceScreenshot(page, "risk-summary-severity-order");

    await marketplace.openRisksForVersion(marketplaceReadmeOnlyGitVersion());
    await marketplace.expectRiskListCount(3);
    await marketplace.expectRiskListSeverityOrder([
      "high",
      "warning",
      "info",
    ]);

    // Localized body text still renders for each severity row.
    await marketplace.expectRiskFindingText(
      "动态包请求宿主服务授权。",
    );
    await marketplace.expectRiskFindingText("未声明框架兼容性依赖。");
    await marketplace.expectRiskFindingText(
      "已检测到可用于市场展示的文档条目。",
    );
    await expect(
      marketplace
        .detailShell()
        .getByText("Dynamic package requests host service authorization."),
    ).toHaveCount(0);

    await captureMarketplaceScreenshot(page, "risk-list-severity-order");
  });
});
