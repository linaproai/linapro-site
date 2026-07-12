import type { Locator, Page } from "@host-tests/fixtures/auth";

import { expect } from "@host-tests/fixtures/auth";

import { workspacePath } from "@host-tests/fixtures/config";
import { waitForRouteReady, waitForTableReady } from "@host-tests/support/ui";

type FieldScope = Locator | Page;
type PluginTypeLabel = "Dynamic" | "Source" | "动态插件" | "源码插件";
type ReviewDecisionLabel = "Approved" | "Rejected" | "已拒绝" | "已通过";
type VisibilityLabel =
  | "Private"
  | "Public"
  | "Reserved"
  | "公开"
  | "私有"
  | "预留授权";

const adminTableSelector = ".plugin-marketplace-admin-list .vxe-table";
const detailReleaseTableSelector = ".marketplace-detail-shell .vxe-table";
const mineTableSelector = ".plugin-marketplace-mine .vxe-table";
const reviewTableSelector = ".vxe-table:visible";
const rawMarketplaceKeyPattern =
  /plugin\.linapro-plugin-marketplace|error\.plugin\.marketplace/u;

export class MarketplacePage {
  constructor(readonly page: Page) {}

  async gotoMine() {
    await this.gotoTablePage(
      "/plugin-marketplace/plugin-marketplace-mine",
      mineTableSelector,
    );
  }

  async gotoAdminList() {
    await this.gotoTablePage(
      "/plugin-marketplace/plugin-marketplace-admin-list",
      adminTableSelector,
    );
  }

  async gotoReview() {
    await this.page.goto(
      workspacePath("/plugin-marketplace/plugin-marketplace-review"),
      { waitUntil: "domcontentloaded" },
    );
    await waitForRouteReady(this.page, 15_000);
    await waitForTableReady(this.page, reviewTableSelector);
    await expect(
      this.page.getByText(/待审版本|Review Queue/u, { exact: true }).last(),
    ).toBeVisible();
  }

  async gotoDashboard() {
    await this.page.goto(workspacePath("/dashboard/analytics"), {
      waitUntil: "domcontentloaded",
    });
    await waitForRouteReady(this.page, 15_000);
  }

  async gotoDetail(pluginId: string, from?: "admin-list" | "mine" | "review") {
    const query = new URLSearchParams({ pluginId });
    let routePath = "/plugin-marketplace/plugin-marketplace-detail";
    if (from) {
      query.set("from", from);
      query.set("view", "detail");
      routePath = `/plugin-marketplace/plugin-marketplace-${from}`;
    }
    query.set("pageKey", routePath);
    await this.page.goto(workspacePath(`${routePath}?${query.toString()}`), {
      waitUntil: "domcontentloaded",
    });
    await waitForRouteReady(this.page, 15_000);
    await expect(this.detailShell().getByText(pluginId).first()).toBeVisible();
  }

  mineRow(pluginId: string) {
    return this.tableRow(mineTableSelector, pluginId);
  }

  adminRow(pluginId: string) {
    return this.tableRow(adminTableSelector, pluginId);
  }

  reviewRow(pluginId: string, version: string) {
    return this.page
      .locator(`${reviewTableSelector} .vxe-table--main-wrapper .vxe-body--row`)
      .filter({ hasText: pluginId })
      .filter({ hasText: version })
      .first();
  }

  detailShell() {
    return this.page.locator(".marketplace-detail-shell").first();
  }

  publishDrawer() {
    return this.page
      .locator('[role="dialog"]:visible, .ant-drawer-content:visible')
      .filter({ hasText: /发布插件|Publish Plugin|新版本|New Version/u })
      .last();
  }

  publishSection(heading: string | RegExp) {
    return this.publishDrawer()
      .locator("section:visible")
      .filter({ has: this.page.getByRole("heading", { name: heading }) })
      .first();
  }

  reviewPanel() {
    return this.page
      .locator(".marketplace-review-drawer-content:visible")
      .last();
  }

  async filterMineByKeyword(keyword: string) {
    await this.fillField(this.page, /关键词|Keyword/u, keyword);
    await this.searchGrid(mineTableSelector, "/market/my-plugins");
  }

  async resetMineFilters() {
    await this.resetGrid(mineTableSelector, "/market/my-plugins");
  }

  async filterAdminByKeyword(keyword: string) {
    await this.fillField(this.page, /关键词|Keyword/u, keyword);
    await this.searchGrid(adminTableSelector, "/market/managed-plugins");
  }

  async filterReviewByPluginId(pluginId: string) {
    await this.fillField(this.page, /插件 ID|Plugin ID/u, pluginId);
    await this.searchGrid(reviewTableSelector, "/market/review-queue");
  }

  async expectColumn(table: "admin" | "mine" | "review", title: string) {
    const selector =
      table === "admin"
        ? adminTableSelector
        : table === "mine"
          ? mineTableSelector
          : reviewTableSelector;
    await expect(
      this.page
        .locator(`${selector} .vxe-header--column`)
        .filter({ hasText: title })
        .first(),
    ).toBeVisible();
  }

  async expectAdminTableNoHorizontalOverflow() {
    const body = this.page
      .locator(
        `${adminTableSelector} .vxe-table--main-wrapper .vxe-table--body-wrapper`,
      )
      .first();
    await expect(body).toBeVisible();
    await expect
      .poll(async () =>
        body.evaluate((element) => element.scrollWidth - element.clientWidth),
      )
      .toBeLessThanOrEqual(1);
  }

  async expectAdminColumnBeforeFixedActions(title: string) {
    const column = this.page
      .locator(
        `${adminTableSelector} .vxe-table--main-wrapper .vxe-header--column`,
      )
      .filter({ hasText: title })
      .first();
    const actionColumn = this.page
      .locator(`${adminTableSelector} .vxe-table--main-wrapper .vxe-header--column`)
      .filter({ hasText: /操作|Actions/u })
      .first();
    await expect(column).toBeVisible();
    await expect(actionColumn).toBeVisible();
    await expect
      .poll(async () => {
        const [columnBox, fixedBox] = await Promise.all([
          column.boundingBox(),
          actionColumn.boundingBox(),
        ]);
        if (!columnBox || !fixedBox) {
          return false;
        }
        return columnBox.x + columnBox.width <= fixedBox.x + 1;
      })
      .toBe(true);
  }

  async expectWorkspaceTabCount(title: string, count: number) {
    await expect(
      this.page.locator('[data-tab-item="true"]').filter({ hasText: title }),
    ).toHaveCount(count);
  }

  async openMineDetail(pluginId: string) {
    await this.openDetailFromTable(mineTableSelector, pluginId);
  }

  async openAdminDetail(pluginId: string) {
    await this.openDetailFromTable(adminTableSelector, pluginId);
  }

  async backFromDetail(
    expectedPath: "admin-list" | "dashboard/analytics" | "mine" | "review",
  ) {
    await this.detailShell()
      .getByRole("button", { name: /返\s*回|Back/u })
      .click();
    await waitForRouteReady(this.page, 15_000);
    await expect
      .poll(() => new URL(this.page.url()).pathname)
      .toContain(
        expectedPath === "dashboard/analytics"
          ? expectedPath
          : `plugin-marketplace-${expectedPath}`,
      );
    if (expectedPath === "mine") {
      await waitForTableReady(this.page, mineTableSelector, 15_000);
    } else if (expectedPath === "admin-list") {
      await waitForTableReady(this.page, adminTableSelector, 15_000);
    } else if (expectedPath === "review") {
      await waitForTableReady(this.page, reviewTableSelector, 15_000);
    }
  }

  async openPublishDrawer() {
    await this.page
      .getByRole("button", { name: /发布插件|Publish Plugin/u })
      .click();
    await expect(this.publishDrawer()).toBeVisible();
    await expect(
      this.publishDrawer().getByRole("heading", {
        name: /发布者资料|Publisher Profile/u,
      }),
    ).toBeVisible();
  }

  async openNewVersionDrawer(pluginId: string) {
    const moreButton = await this.rowActionButton(
      mineTableSelector,
      pluginId,
      /更\s*多|More/u,
    );
    await moreButton.click();
    const menu = this.page.locator(".ant-dropdown:visible").last();
    await expect(menu).toBeVisible();
    await menu.getByRole("menuitem", { name: /新版本|New Version/u }).click();
    await expect(this.publishDrawer()).toBeVisible();
    await expect(
      this.publishDrawer().getByText(/新版本|New Version/u, { exact: true }),
    ).toBeVisible();
  }

  async closePublishDrawer() {
    await this.publishDrawer().getByRole("button").first().click();
    await expect(this.publishDrawer()).toBeHidden();
  }

  async expectExistingPublisher(label: string) {
    const publisher = this.publishSection(/发布者资料|Publisher Profile/u)
      .locator(".ant-select-selection-item")
      .first();
    await expect(publisher).toHaveText(label);
  }

  async expectPluginDraftReset() {
    const section = this.publishSection(/插件草稿|Plugin Draft/u);
    await expect(
      this.formItem(section, /插件 ID|Plugin ID/u)
        .locator("input")
        .first(),
    ).toHaveValue("");
    await expect(
      this.formItem(section, /插件名称|Plugin Name/u)
        .locator("input")
        .first(),
    ).toHaveValue("");
    await expect(
      this.formItem(section, /摘要|Summary/u)
        .locator("textarea")
        .first(),
    ).toHaveValue("");
  }

  async expectReleaseTarget(pluginId: string, pluginType: PluginTypeLabel) {
    const section = this.publishSection(/版本包|Release Package/u);
    await expect(
      this.formItem(section, /插件 ID|Plugin ID/u)
        .locator("input")
        .first(),
    ).toHaveValue(pluginId);
    await expect(
      this.formItem(section, /类型|Type/u)
        .locator(".ant-select-selection-item")
        .first(),
    ).toHaveText(pluginType);
  }

  async submitEmptyPluginDraft() {
    await this.publishSection(/插件草稿|Plugin Draft/u)
      .getByRole("button", { name: /保存草稿|Save Draft/u })
      .click();
  }

  validationErrors() {
    return this.publishDrawer().getByRole("alert");
  }

  async fillPublisherProfile(values: {
    contactEmail?: string;
    homepage?: string;
    name: string;
    publisherKey: string;
    summary?: string;
  }) {
    const section = this.publishSection(/发布者资料|Publisher Profile/u);
    await this.fillField(
      section,
      /发布者 Key|Publisher Key/u,
      values.publisherKey,
    );
    await this.fillField(section, /发布者名称|Publisher Name/u, values.name);
    if (values.homepage) {
      await this.fillField(section, /主页|Homepage/u, values.homepage);
    }
    if (values.contactEmail) {
      await this.fillField(
        section,
        /联系邮箱|Contact Email/u,
        values.contactEmail,
      );
    }
    if (values.summary) {
      await this.fillField(section, /摘要|Summary/u, values.summary);
    }
  }

  async savePublisher() {
    await this.publishSection(/发布者资料|Publisher Profile/u)
      .getByRole("button", { name: /保存发布者|Save Publisher/u })
      .click();
  }

  async fillPluginDraft(values: {
    name: string;
    pluginId: string;
    pluginType: PluginTypeLabel;
    summary: string;
    visibility?: VisibilityLabel;
  }) {
    const section = this.publishSection(/插件草稿|Plugin Draft/u);
    await this.fillField(section, /插件 ID|Plugin ID/u, values.pluginId);
    await this.fillField(section, /插件名称|Plugin Name/u, values.name);
    await this.selectField(section, /类型|Type/u, values.pluginType);
    if (values.visibility) {
      await this.selectField(section, /可见性|Visibility/u, values.visibility);
    }
    await this.fillField(section, /摘要|Summary/u, values.summary);
  }

  async savePluginDraft() {
    await this.publishSection(/插件草稿|Plugin Draft/u)
      .getByRole("button", { name: /保存草稿|Save Draft/u })
      .click();
  }

  async fillReleaseUpload(values: { version: string }) {
    const section = this.publishSection(/版本包|Release Package/u);
    await this.fillField(section, /版本|Version/u, values.version);
  }

  async setUploadFile(file: Parameters<Locator["setInputFiles"]>[0]) {
    const input = this.publishSection(/版本包|Release Package/u)
      .locator('input[type="file"]')
      .first();
    await input.setInputFiles(file);
  }

  async uploadDraft() {
    await this.publishSection(/版本包|Release Package/u)
      .getByRole("button", { name: /上传草稿|Upload Draft/u })
      .click();
  }

  async submitLatestDraft() {
    await this.publishSection(/版本包|Release Package/u)
      .getByRole("button", { name: /提交审核|Submit Review/u })
      .click();
  }

  async inspectReviewRelease(pluginId: string, version: string) {
    const row = this.reviewRow(pluginId, version);
    await expect(row).toBeVisible();
    const inspectButton = await this.rowActionButton(
      reviewTableSelector,
      `${pluginId} ${version}`,
      /检\s*查|Inspect/u,
    );
    await inspectButton.click();
    await expect(this.reviewPanel()).toContainText(version);
  }

  async closeReviewPanel() {
    const drawer = this.page
      .locator('[role="dialog"]:visible')
      .filter({ has: this.reviewPanel() })
      .last();
    await drawer.getByRole("button").first().click();
    await expect(this.reviewPanel()).toBeHidden();
  }

  async expectReviewPanelContains(text: string) {
    await expect(this.reviewPanel()).toContainText(text);
  }

  async expectReviewPanelNotContains(text: string) {
    await expect(this.reviewPanel()).not.toContainText(text);
  }

  async expectReviewEmptyState(text: string) {
    await expect(
      this.page
        .locator(".vxe-table--empty-content:visible")
        .filter({ hasText: text })
        .first(),
    ).toBeVisible();
  }

  async submitReviewDecision(decision: ReviewDecisionLabel, message?: string) {
    const panel = this.reviewPanel();
    await this.selectField(panel, /审核|Review/u, decision);
    if (message) {
      await this.fillField(panel, /审核意见|Review Message/u, message);
    }
    await this.page
      .getByRole("button", { name: /提交结论|Submit Decision/u })
      .last()
      .click();
  }

  async openDocsForVersion(version: string) {
    await this.showVersionTab();
    const docsButton = await this.rowActionButton(
      detailReleaseTableSelector,
      version,
      /文\s*档|Docs/u,
    );
    await docsButton.click();
    await expect(
      this.page.getByRole("tab", { name: /文档|Docs/u }),
    ).toHaveAttribute("aria-selected", "true");
  }

  async openRisksForVersion(version: string) {
    await this.showVersionTab();
    const risksButton = await this.rowActionButton(
      detailReleaseTableSelector,
      version,
      /风\s*险|Risks/u,
    );
    await risksButton.click();
    await expect(
      this.page.getByRole("tab", { name: /风\s*险|Risks/u }),
    ).toHaveAttribute("aria-selected", "true");
  }

  async confirmDownloadForVersion(version: string) {
    await this.showVersionTab();
    const downloadButton = await this.rowActionButton(
      detailReleaseTableSelector,
      version,
      /^下\s*载$|^Download$/u,
    );
    await downloadButton.click();
    const dialog = this.page
      .locator(".ant-modal-confirm, .ant-modal-wrap")
      .filter({ hasText: /确认下载|Confirm Download/u })
      .last();
    await expect(dialog).toBeVisible();
    await dialog
      .getByRole("button", {
        name: /下载并导入|下\s*载|Download and Import|Download/u,
      })
      .last()
      .click();
    await dialog.waitFor({ state: "hidden", timeout: 10_000 }).catch(() => {});
  }

  async expectDownloadActionHidden(version: string) {
    await this.showVersionTab();
    const row = this.tableRow(detailReleaseTableSelector, version);
    await expect(row).toBeVisible();
    await expect(
      row.getByRole("button", { name: /^下\s*载$|^Download$/u }),
    ).toHaveCount(0);
  }

  async expectNoRawI18nKeys() {
    const bodyText = await this.page.locator("body").innerText();
    expect(bodyText).not.toMatch(rawMarketplaceKeyPattern);
  }

  async expectNoHorizontalPageOverflow() {
    await expect
      .poll(async () =>
        this.page.evaluate(() => {
          const root = document.documentElement;
          const body = document.body;
          return (
            Math.max(root.scrollWidth, body.scrollWidth) - root.clientWidth
          );
        }),
      )
      .toBeLessThanOrEqual(1);
  }

  async expectNoHorizontalMineTableOverflow() {
    await expect
      .poll(async () =>
        this.page
          .locator(
            `${mineTableSelector} .vxe-table--main-wrapper .vxe-table--body-wrapper`,
          )
          .first()
          .evaluate((element) => element.scrollWidth - element.clientWidth),
      )
      .toBeLessThanOrEqual(1);
  }

  async expectMineRowsVerticallySeparated() {
    const rows = this.page.locator(
      `${mineTableSelector} .vxe-table--main-wrapper .vxe-body--row`,
    );
    const firstRow = rows.first();
    const secondRow = rows.nth(1);
    const pluginCell = firstRow.locator(".mine-plugin-cell");
    await expect(secondRow).toBeVisible();
    await expect(pluginCell).toBeVisible();
    await expect
      .poll(async () => {
        const [firstBox, secondBox, cellBox] = await Promise.all([
          firstRow.boundingBox(),
          secondRow.boundingBox(),
          pluginCell.boundingBox(),
        ]);
        if (!firstBox || !secondBox || !cellBox) {
          return false;
        }
        return (
          cellBox.y >= firstBox.y - 1 &&
          cellBox.y + cellBox.height <= firstBox.y + firstBox.height + 1 &&
          firstBox.y + firstBox.height <= secondBox.y + 1
        );
      })
      .toBe(true);
  }

  private async gotoTablePage(path: string, tableSelector: string) {
    await this.page.goto(workspacePath(path), {
      waitUntil: "domcontentloaded",
    });
    await waitForRouteReady(this.page, 15_000);
    await waitForTableReady(this.page, tableSelector);
  }

  private async showVersionTab() {
    const versionTable = this.page.locator(detailReleaseTableSelector).first();
    if (await versionTable.isVisible().catch(() => false)) {
      return;
    }
    await this.page.getByRole("tab", { name: /版\s*本|Versions/u }).click();
    await waitForTableReady(this.page, detailReleaseTableSelector, 15_000);
  }

  private tableRow(tableSelector: string, rowText: string) {
    return this.page
      .locator(`${tableSelector} .vxe-table--main-wrapper .vxe-body--row`)
      .filter({ hasText: rowText })
      .first();
  }

  private async openDetailFromTable(tableSelector: string, pluginId: string) {
    const detailButton = await this.rowActionButton(
      tableSelector,
      pluginId,
      /详\s*情|Details/u,
    );
    await detailButton.click();
    await waitForRouteReady(this.page, 15_000);
    await expect(this.detailShell().getByText(pluginId).first()).toBeVisible();
  }

  private async rowActionButton(
    tableSelector: string,
    rowText: string,
    buttonName: RegExp,
  ) {
    await waitForRouteReady(this.page, 15_000);
    await waitForTableReady(this.page, tableSelector, 15_000);
    const mainRows = this.page.locator(
      `${tableSelector} .vxe-table--main-wrapper .vxe-body--row`,
    );
    const rowCount = await mainRows.count();
    const rowTextTokens = rowText.split(/\s+/u).filter(Boolean);
    for (let index = 0; index < rowCount; index += 1) {
      const row = mainRows.nth(index);
      const text = await row.innerText().catch(() => "");
      if (!rowTextTokens.every((token) => text.includes(token))) {
        continue;
      }

      const fixedRow = this.page
        .locator(
          `${tableSelector} .vxe-table--fixed-right-wrapper .vxe-body--row`,
        )
        .nth(index);
      if (await fixedRow.isVisible().catch(() => false)) {
        return fixedRow.getByRole("button", { name: buttonName }).first();
      }
      return row.getByRole("button", { name: buttonName }).first();
    }
    throw new Error(`Table row not found: ${rowText}`);
  }

  private formItem(scope: FieldScope, label: string | RegExp) {
    return scope
      .locator(".ant-form-item, [data-form-render-field]")
      .filter({
        has: this.page.getByText(label, { exact: typeof label === "string" }),
      })
      .first();
  }

  private async fillField(
    scope: FieldScope,
    label: string | RegExp,
    value: string,
  ) {
    const formItem = this.formItem(scope, label);
    await expect(formItem).toBeVisible();
    const control = formItem.locator("input, textarea").first();
    await expect(control).toBeVisible();
    await control.fill(value);
  }

  private async selectField(
    scope: FieldScope,
    label: string | RegExp,
    option: string,
  ) {
    const formItem = this.formItem(scope, label);
    await expect(formItem).toBeVisible();
    if (await this.selectHasCurrentValue(formItem, option)) {
      return;
    }

    await formItem.locator(".ant-select-selector").first().click();
    const dropdown = this.page.locator(".ant-select-dropdown:visible").last();
    await expect(dropdown).toBeVisible();
    const optionIndex = await this.visibleSelectOptionIndex(dropdown, option);
    const dropdownBox = await dropdown.boundingBox();
    if (!dropdownBox) {
      throw new Error(`Select dropdown not visible for option: ${option}`);
    }
    await this.page.mouse.click(
      dropdownBox.x + dropdownBox.width / 2,
      dropdownBox.y + Math.min(24 + optionIndex * 36, dropdownBox.height - 10),
    );
    await dropdown.waitFor({ state: "hidden", timeout: 5000 }).catch(() => {});
  }

  private async visibleSelectOptionIndex(dropdown: Locator, option: string) {
    const options = dropdown.getByRole("option");
    const aliases = this.selectOptionAliases(option);
    const count = await options.count();
    for (let index = 0; index < count; index += 1) {
      const item = options.nth(index);
      const ariaLabel = (await item.getAttribute("aria-label")) ?? "";
      const text = await item.innerText().catch(() => "");
      if (
        [ariaLabel, text].some((value) =>
          this.selectValueMatches(value, aliases),
        )
      ) {
        return index;
      }
    }
    throw new Error(`Select option not found: ${option}`);
  }

  private async selectHasCurrentValue(formItem: Locator, option: string) {
    const selectedItem = formItem.locator(".ant-select-selection-item").first();
    if (!(await selectedItem.isVisible().catch(() => false))) {
      return false;
    }

    const aliases = this.selectOptionAliases(option);
    const selectedText = await selectedItem.innerText().catch(() => "");
    const selectedTitle = (await selectedItem.getAttribute("title")) ?? "";
    return [selectedText, selectedTitle].some((value) =>
      this.selectValueMatches(value, aliases),
    );
  }

  private selectOptionAliases(option: string) {
    const normalized = option.trim().toLowerCase();
    const aliases = new Set([normalized]);
    const groups = [
      ["source", "source plugin", "源码插件"],
      ["dynamic", "dynamic plugin", "动态插件"],
      ["public", "公开"],
      ["private", "私有"],
      ["reserved", "预留授权"],
      ["approved", "已通过"],
      ["rejected", "已拒绝"],
    ];
    for (const group of groups) {
      if (group.includes(normalized)) {
        group.forEach((value) => aliases.add(value));
      }
    }
    return aliases;
  }

  private selectValueMatches(value: string, aliases: Set<string>) {
    return aliases.has(value.trim().toLowerCase());
  }

  private async searchGrid(tableSelector: string, endpoint: string) {
    const response = this.page.waitForResponse((candidate) => {
      const request = candidate.request();
      const pathname = new URL(candidate.url()).pathname;
      return request.method() === "GET" && pathname.endsWith(endpoint);
    });
    await this.page.getByRole("button", { name: /搜\s*索|Search/u }).click();
    await response;
    await waitForTableReady(this.page, tableSelector);
  }

  private async resetGrid(tableSelector: string, endpoint: string) {
    const response = this.page.waitForResponse((candidate) => {
      const request = candidate.request();
      const pathname = new URL(candidate.url()).pathname;
      return request.method() === "GET" && pathname.endsWith(endpoint);
    });
    await this.page.getByRole("button", { name: /重\s*置|Reset/u }).click();
    await response;
    await waitForTableReady(this.page, tableSelector);
  }
}
