/**
 * TC006: Historical releases remain installable and git pins use commit SHA.
 *
 * Validates marketplace detail keeps multiple published versions visible,
 * shows the pinned Git coordinates for main-fallback history, and that
 * distribution / download for a non-latest version do not drift to latest.
 */
import { expect, test } from "@host-tests/fixtures/auth";

import { MarketplacePage } from "../pages/MarketplacePage";
import {
  callMarketplaceApiFromPage,
  captureMarketplaceScreenshot,
  installMarketplaceApiMocks,
  marketplaceHistoricalSourceCommit,
  marketplaceLatestSourceCommit,
  marketplaceSourcePluginId,
  openMarketplaceWorkbench,
} from "../support/marketplace-fixtures";

test.describe("TC-6 marketplace history versions and git pin", () => {
  test("TC-6a: detail lists historical versions with pinned git coordinates", async ({
    authenticatedPage,
  }) => {
    const page = authenticatedPage;
    await page.setViewportSize({ height: 960, width: 1440 });
    await installMarketplaceApiMocks(page, {
      menuRole: "both",
    });
    await openMarketplaceWorkbench(page, "简体中文");

    const marketplace = new MarketplacePage(page);
    await marketplace.gotoDetail(marketplaceSourcePluginId());
    await expect(marketplace.detailShell()).toContainText("LinaPro 源码演示");

    const latestRow = marketplace.detailReleaseRow("v1.0.0");
    const historyRow = marketplace.detailReleaseRow("v0.9.0");
    await expect(latestRow).toBeVisible();
    await expect(historyRow).toBeVisible();

    // Latest git tag pin and historical main-fallback pin must both remain visible.
    // Avoid "@" in copy: vue-i18n treats "@" as linked-message syntax.
    // Full commit SHA is shown on its own line (no mid-string truncation).
    await expect(latestRow).toContainText("Git v1.0.0");
    await expect(
      latestRow.locator(".marketplace-source-commit"),
    ).toHaveText(marketplaceLatestSourceCommit);
    await expect(historyRow).toContainText("Git main");
    await expect(
      historyRow.locator(".marketplace-source-commit"),
    ).toHaveText(marketplaceHistoricalSourceCommit);
    await expect(historyRow.locator(".marketplace-source-pin")).toBeVisible();
    // Artifact column was removed; version tab must not reintroduce it as a header.
    await expect(
      marketplace.detailShell().getByRole("columnheader", { name: "产物" }),
    ).toHaveCount(0);

    await marketplace.expectNoRawI18nKeys();
    await captureMarketplaceScreenshot(page, "detail-history-versions-git-pin");
  });

  test("TC-6b: historical version distribution and download stay on that version", async ({
    authenticatedPage,
  }) => {
    const page = authenticatedPage;
    await page.setViewportSize({ height: 960, width: 1440 });
    const mockState = await installMarketplaceApiMocks(page, {
      menuRole: "both",
    });
    await openMarketplaceWorkbench(page, "简体中文");

    const marketplace = new MarketplacePage(page);
    // Use the dedicated detail route so the shell is always the standalone page.
    await marketplace.gotoDetail(marketplaceSourcePluginId());

    const historyDistribution = await callMarketplaceApiFromPage(
      page,
      `market/plugins/${marketplaceSourcePluginId()}/releases/v0.9.0/distribution`,
    );
    expect(historyDistribution.code).toBe(0);
    const historyPayload = historyDistribution.data as {
      distribution?: { mode?: string; ref?: string; version?: string };
    };
    expect(historyPayload.distribution?.version).toBe("v0.9.0");
    expect(historyPayload.distribution?.mode).toBe("git");
    // Install ref must be the pinned commit, not floating main or latest tip.
    expect(historyPayload.distribution?.ref).toBe(
      marketplaceHistoricalSourceCommit,
    );
    expect(historyPayload.distribution?.ref).not.toBe("main");
    expect(historyPayload.distribution?.ref).not.toBe(
      marketplaceLatestSourceCommit,
    );

    const latestDistribution = await callMarketplaceApiFromPage(
      page,
      `market/plugins/${marketplaceSourcePluginId()}/releases/v1.0.0/distribution`,
    );
    expect(latestDistribution.code).toBe(0);
    const latestPayload = latestDistribution.data as {
      distribution?: { mode?: string; ref?: string; version?: string };
    };
    expect(latestPayload.distribution?.version).toBe("v1.0.0");
    expect(latestPayload.distribution?.ref).toBe(marketplaceLatestSourceCommit);

    await marketplace.openDocsForVersion("v0.9.0");
    await expect(page.getByText("Source Demo History").first()).toBeVisible();

    await marketplace.confirmDownloadForVersion("v0.9.0");
    // Historical download must target the selected version only.
    expect(mockState.downloadRequests).toEqual([
      { pluginId: marketplaceSourcePluginId(), version: "v0.9.0" },
    ]);

    await marketplace.expectNoRawI18nKeys();
    await captureMarketplaceScreenshot(page, "detail-history-version-download");
  });
});
