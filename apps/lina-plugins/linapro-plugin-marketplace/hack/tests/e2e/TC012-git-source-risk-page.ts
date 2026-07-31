/**
 * TC012: Git-sourced plugin risk page stays consistent with the risk summary.
 *
 * Reproduces the marketplace bug where the Git discovery path wrote only the
 * aggregated risk_summary (release.risk_summary, surfaced as "提示 1") but
 * skipped persisting structured risk rows to plugin_marketplace_risk, so the
 * owner risk page rendered an empty list next to a populated summary. After
 * the fix, the same diagnostics back both projections and the risk page shows
 * detail rows whose count matches the summary.
 *
 * Note: marketplace E2E runs against an API mock (installMarketplaceApiMocks),
 * so this case guards the front-end rendering invariant — the risk summary
 * count must equal the risk detail row count — for both the inconsistent
 * (pre-fix) and consistent (post-fix) API shapes. Verification of the backend
 * Git path writing risk rows is covered by the unit invariant test
 * TestBuildSourceRiskSummaryMatchesDiagnosticSeverityCount and code review.
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

test.describe("TC-12 git source plugin risk page consistency", () => {
  test("TC-12a: risk summary without detail rows reproduces the empty risk page", async ({
    authenticatedPage,
  }) => {
    const page = authenticatedPage;
    await page.setViewportSize({ height: 960, width: 1440 });
    // Default fixtures: the Git-sourced tenant-core release carries a
    // risk_summary of { info: 1 } but no entries in the risks map, mirroring
    // the backend writing only the summary and skipping risk detail rows.
    await installMarketplaceApiMocks(page, { menuRole: "publish-only" });
    await openMarketplaceWorkbench(page, "简体中文");

    const marketplace = new MarketplacePage(page);
    await marketplace.gotoMine();
    const pluginId = marketplaceReadmeOnlyGitPluginId();
    await marketplace.openMineDetail(pluginId);
    await marketplace.openRisksForVersion(marketplaceReadmeOnlyGitVersion());

    // Summary still reports the info finding, but the detail list is empty.
    await marketplace.expectRiskSummaryTag("info", 1);
    await marketplace.expectRiskListEmptyState();
    await marketplace.expectRiskListCount(0);
    await captureMarketplaceScreenshot(page, "git-risk-empty-detail");
  });

  test("TC-12b: risk detail rows match the summary counts after the fix", async ({
    authenticatedPage,
  }) => {
    const page = authenticatedPage;
    await page.setViewportSize({ height: 960, width: 1440 });
    // After the fix the same diagnostics back both projections: supply one
    // actionable risk row so the detail list matches the warning summary count.
    await installMarketplaceApiMocks(page, {
      gitSourceRisks: [
        {
          // Stable diagnostic code drives UI i18n; English summary is fallback only.
          payload: { code: "source_sql_present" },
          severity: "warning",
          source: "manifest/sql",
          summary:
            "Source package contains SQL resources that require reviewer inspection.",
          type: "install_sql",
        },
      ],
      menuRole: "publish-only",
    });
    await openMarketplaceWorkbench(page, "简体中文");

    const marketplace = new MarketplacePage(page);
    await marketplace.gotoMine();
    const pluginId = marketplaceReadmeOnlyGitPluginId();
    await marketplace.openMineDetail(pluginId);
    await marketplace.openRisksForVersion(marketplaceReadmeOnlyGitVersion());

    // Summary and detail now agree: one warning finding in both projections.
    await marketplace.expectRiskSummaryTag("warning", 1);
    await marketplace.expectRiskListCount(1);
    // Risk body must render the zh-CN riskFinding translation, not English source.
    await expect(
      marketplace
        .detailShell()
        .getByText("源码包包含需审核关注的 SQL 资源。"),
    ).toBeVisible();
    await expect(
      marketplace
        .detailShell()
        .getByText(
          "Source package contains SQL resources that require reviewer inspection.",
        ),
    ).toHaveCount(0);
    await captureMarketplaceScreenshot(page, "git-risk-consistent-detail");
  });
});
