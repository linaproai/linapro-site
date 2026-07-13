/**
 * TC004: Git source registration entry is available on My Plugins.
 *
 * Validates the simplified publish surface exposes Git source registration
 * without requiring a full remote discovery integration in CI.
 */
import { test, expect } from "@playwright/test";

test.describe("TC004 marketplace git source register entry", () => {
  test.skip(true, "Requires authenticated marketplace publish session; enable when E2E fixtures cover market:plugin:publish");

  test("shows register git source action on my plugins page", async ({
    page,
  }) => {
    await page.goto("/plugin-marketplace/plugin-marketplace-mine");
    await expect(
      page.getByRole("button", { name: /登记 Git 源|Register Git Source/i }),
    ).toBeVisible();
  });
});
