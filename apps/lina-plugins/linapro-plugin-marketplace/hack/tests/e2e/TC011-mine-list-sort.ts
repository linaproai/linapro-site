/**
 * TC011: My Plugins list remote sort.
 *
 * Validates that pluginId / marketStatus / downloadCount / updatedAt headers
 * are sortable, default order is pluginId ascending, and clicking downloadCount
 * re-queries with the corresponding order parameters.
 */
import { expect, test } from "@host-tests/fixtures/auth";

import { MarketplacePage } from "../pages/MarketplacePage";
import {
  captureMarketplaceScreenshot,
  installMarketplaceApiMocks,
  openMarketplaceWorkbench,
} from "../support/marketplace-fixtures";

test.describe("TC-11 marketplace mine list sort", () => {
  test("TC-11a: default pluginId asc and downloadCount remote sort", async ({
    authenticatedPage,
  }) => {
    const page = authenticatedPage;
    await page.setViewportSize({ height: 960, width: 1440 });
    const mockState = await installMarketplaceApiMocks(page, {
      menuRole: "both",
    });
    await openMarketplaceWorkbench(page, "简体中文");

    const marketplace = new MarketplacePage(page);
    await marketplace.gotoMine();

    await marketplace.expectMineColumnSortable("插件标识");
    await marketplace.expectMineColumnSortable("状态");
    await marketplace.expectMineColumnSortable("下载量");
    await marketplace.expectMineColumnSortable("更新时间");

    // Default remote sort: pluginId ascending.
    const defaultList = mockState.listRequests.at(-1);
    expect(defaultList?.get("orderBy") ?? "pluginId").toMatch(/pluginId/u);
    expect(defaultList?.get("orderDirection") ?? "asc").toMatch(/asc/u);

    const idsBefore = await marketplace.mineBodyPluginIds();
    expect(idsBefore.length).toBeGreaterThan(1);
    const sortedAsc = [...idsBefore].sort((a, b) => a.localeCompare(b));
    expect(idsBefore).toEqual(sortedAsc);

    const beforeCount = mockState.listRequests.length;
    await marketplace.clickMineColumnSort("下载量");
    await expect
      .poll(() => mockState.listRequests.length)
      .toBeGreaterThan(beforeCount);

    const downloadSort = mockState.listRequests.at(-1);
    expect(downloadSort?.get("orderBy")).toBe("downloadCount");
    expect(downloadSort?.get("orderDirection")).toMatch(/asc|desc/u);

    await marketplace.expectNoRawI18nKeys();
    await captureMarketplaceScreenshot(page, "mine-list-sort-download");
  });
});
