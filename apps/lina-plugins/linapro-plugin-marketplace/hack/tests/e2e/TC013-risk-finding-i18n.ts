/**
 * TC013: Marketplace risk findings render localized body text.
 *
 * Scanner APIs store English source summaries plus a stable diagnostic code in
 * payload.code. The owner risk tab and risk summary area must map known codes
 * through plugin runtime i18n so zh-CN users never see English finding prose.
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

test.describe("TC-13 risk finding i18n", () => {
  test("TC-13a: zh-CN risk tab shows translated finding messages", async ({
    authenticatedPage,
  }) => {
    const page = authenticatedPage;
    await page.setViewportSize({ height: 960, width: 1440 });
    await installMarketplaceApiMocks(page, {
      gitSourceRisks: [
        {
          payload: { code: "framework_dependency_missing" },
          severity: "warning",
          source: "plugin.yaml",
          summary: "Framework compatibility dependency is not declared.",
          type: "dependency",
        },
        {
          payload: { code: "source_sql_present" },
          severity: "warning",
          source: "manifest/sql",
          summary:
            "Source package contains SQL resources that require reviewer inspection.",
          type: "install_sql",
        },
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
    await marketplace.openRisksForVersion(marketplaceReadmeOnlyGitVersion());

    // Risk summary tags remain localized severity counts.
    await marketplace.expectRiskSummaryTag("warning", 2);
    await marketplace.expectRiskSummaryTag("info", 1);
    await marketplace.expectRiskListCount(3);

    // Body text uses zh-CN riskFinding translations, not English source.
    await marketplace.expectRiskFindingText("未声明框架兼容性依赖。");
    await marketplace.expectRiskFindingText(
      "源码包包含需审核关注的 SQL 资源。",
    );
    await marketplace.expectRiskFindingText(
      "已检测到可用于市场展示的文档条目。",
    );
    await expect(
      marketplace
        .detailShell()
        .getByText("Framework compatibility dependency is not declared."),
    ).toHaveCount(0);

    // Severity and type tags also stay in Chinese.
    await expect(marketplace.detailShell().getByText("警告").first()).toBeVisible();
    await expect(marketplace.detailShell().getByText("依赖").first()).toBeVisible();
    await expect(marketplace.detailShell().getByText("安装 SQL").first()).toBeVisible();
    await expect(marketplace.detailShell().getByText("文档").first()).toBeVisible();

    await captureMarketplaceScreenshot(page, "risk-finding-i18n-zh");
  });
});
