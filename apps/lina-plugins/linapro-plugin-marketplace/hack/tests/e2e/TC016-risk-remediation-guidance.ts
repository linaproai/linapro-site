/**
 * TC016: Risk tab shows disposition, expandable remediation, and evidence.
 *
 * Publisher detail risks must present fixable findings with reason/remediation
 * acceptance text and file evidence when the API payload includes them.
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

test.describe("TC-16 risk remediation guidance", () => {
  test("TC-16a: disposition filter, blocking banner, and expanded guidance", async ({
    authenticatedPage,
  }) => {
    const page = authenticatedPage;
    await page.setViewportSize({ height: 960, width: 1440 });
    await installMarketplaceApiMocks(page, {
      gitSourceRisks: [
        {
          blocking: true,
          disposition: "need_fix",
          payload: {
            blocking: true,
            code: "i18n_files_missing",
            disposition: "need_fix",
            example: "manifest/i18n/zh-CN/plugin.json",
            expectedField: "i18n.enabled / locale JSON bundles",
            expectedPath: "manifest/i18n",
          },
          severity: "warning",
          source: "manifest/i18n",
          summary:
            "Plugin declares i18n.enabled but no manifest i18n JSON files were detected.",
          type: "dependency",
        },
        {
          blocking: false,
          disposition: "need_attention",
          payload: {
            blocking: false,
            code: "framework_dependency_missing",
            disposition: "need_attention",
            example: ">=1.0.0 <2.0.0",
            expectedField: "dependencies.framework.version",
            expectedPath: "plugin.yaml",
          },
          severity: "warning",
          source: "plugin.yaml",
          summary: "Framework compatibility dependency is not declared.",
          type: "dependency",
        },
        {
          blocking: false,
          disposition: "need_attention",
          payload: {
            blocking: false,
            code: "source_sql_present",
            disposition: "need_attention",
            files: ["manifest/sql/001-init.sql"],
            totalCount: 1,
          },
          severity: "warning",
          source: "manifest/sql",
          summary:
            "Source package contains SQL resources that require reviewer inspection.",
          type: "install_sql",
        },
        {
          blocking: false,
          disposition: "info_only",
          payload: {
            blocking: false,
            code: "source_docs_indexed",
            disposition: "info_only",
            files: ["manifest/docs/zh-CN/index.md"],
            totalCount: 1,
          },
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

    await marketplace.expectRiskListCount(4);
    await expect(
      marketplace.detailShell().getByText("存在 1 条阻塞项，须修复后才能提交审核。"),
    ).toBeVisible();
    await expect(
      marketplace.detailShell().getByText("阻塞提交", { exact: true }).first(),
    ).toBeVisible();
    // Only the true blocking finding shows the tag (framework must not).
    // Use exact match: the intro line also contains the substring "阻塞提交".
    await expect(
      marketplace.detailShell().getByText("阻塞提交", { exact: true }),
    ).toHaveCount(1);
    await expect(
      marketplace.detailShell().getByText("需修复").first(),
    ).toBeVisible();

    // Localized titles (not English summary fallback).
    await marketplace.expectRiskFindingText(
      "插件声明了 i18n.enabled，但未检测到 manifest i18n JSON 文件。",
    );
    await marketplace.expectRiskFindingText("未声明框架兼容性依赖。");
    await marketplace.expectRiskFindingText(
      "源码包包含需审核关注的 SQL 资源。",
    );
    await expect(
      marketplace
        .detailShell()
        .getByText(
          "Source package contains SQL resources that require reviewer inspection.",
        ),
    ).toHaveCount(0);

    // need_fix findings auto-expand guidance.
    await marketplace.expectRiskGuidanceVisible("原因与影响");
    await marketplace.expectRiskGuidanceVisible("manifest/i18n");
    await marketplace.expectRiskGuidanceVisible(
      "manifest/i18n/zh-CN/plugin.json",
    );

    // Filter to attention-only findings (framework + SQL).
    await marketplace
      .detailShell()
      .getByRole("button", { name: /需说明\(2\)/u })
      .click();
    await marketplace.expectRiskListCount(2);
    await marketplace.expectRiskFindingText("未声明框架兼容性依赖。");
    await marketplace.expectRiskFindingText(
      "源码包包含需审核关注的 SQL 资源。",
    );
    await marketplace
      .detailShell()
      .getByRole("button", { name: /查看处理指引|收起指引/u })
      .nth(1)
      .click();
    await marketplace.expectRiskGuidanceVisible("manifest/sql/001-init.sql");

    await captureMarketplaceScreenshot(page, "risk-remediation-guidance");
  });
});
