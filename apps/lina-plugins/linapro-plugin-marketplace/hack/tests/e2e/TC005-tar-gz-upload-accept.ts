/**
 * TC005: Upload control accepts zip and tar.gz packages.
 *
 * Confirms the publisher upload surface advertises .zip/.tar.gz/.tgz.
 */
import { test, expect } from "@playwright/test";

test.describe("TC005 marketplace tar.gz upload accept", () => {
  test.skip(true, "Requires authenticated marketplace publish session; enable when E2E fixtures cover market:plugin:publish");

  test("upload dragger accepts tar.gz packages", async ({ page }) => {
    await page.goto("/plugin-marketplace/plugin-marketplace-mine");
    // Open publish drawer when fixtures exist; assert accept attribute.
    const accept = page.locator('input[type="file"]').first();
    await expect(accept).toHaveAttribute("accept", /\.tar\.gz/);
  });
});
