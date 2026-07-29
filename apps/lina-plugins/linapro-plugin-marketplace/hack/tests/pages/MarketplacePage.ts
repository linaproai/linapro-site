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
const reviewTableSelector = ".plugin-marketplace-review .vxe-table";
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
      .filter({
        hasText:
          /添加插件|Add Plugin|发布插件|Publish Plugin|新版本|New Version/u,
      })
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

  async filterAdminByKeyword(keyword: string) {
    await this.fillField(this.page, /关键词|Keyword/u, keyword);
    await this.searchGrid(adminTableSelector, "/market/managed-plugins");
  }

  async filterReviewByPluginId(pluginId: string) {
    await this.fillField(this.page, /插件 ID|Plugin ID/u, pluginId);
    await this.searchGrid(reviewTableSelector, "/market/review-queue");
  }

  async expectColumn(
    table: "admin" | "mine" | "review",
    title: RegExp | string,
  ) {
    const selector = this.tableSelector(table);
    await expect(
      this.page
        .locator(`${selector} .vxe-header--column`)
        .filter({ hasText: title })
        .first(),
    ).toBeVisible();
  }

  async expectColumnHeaderFits(
    table: "admin" | "mine" | "review",
    title: string,
  ) {
    const header = this.page
      .locator(`${this.tableSelector(table)} .vxe-header--column`)
      .filter({ hasText: title })
      .first();
    await expect(header).toBeVisible();
    const label = header.locator(".vxe-cell--title").first();
    await expect(label).toBeVisible();
    await expect
      .poll(async () =>
        label.evaluate(
          (element) => element.scrollWidth <= element.clientWidth + 1,
        ),
      )
      .toBe(true);
  }

  async expectVisibleCellContentFits(
    table: "admin" | "mine" | "review",
    columnClass: string,
  ) {
    const cells = this.page.locator(
      `${this.tableSelector(table)} .vxe-body--column.${columnClass} .vxe-cell`,
    );
    await expect(cells.first()).toBeVisible();
    await expect
      .poll(async () =>
        cells.evaluateAll((elements) =>
          elements.every(
            (element) => element.scrollWidth <= element.clientWidth + 1,
          ),
        ),
      )
      .toBe(true);
  }

  async expectColumnBeforeFixedActions(
    table: "admin" | "mine" | "review",
    title: string,
  ) {
    const selector = this.tableSelector(table);
    const column = this.page
      .locator(`${selector} .vxe-table--main-wrapper .vxe-header--column`)
      .filter({ hasText: title })
      .first();
    const actionColumn = this.page
      .locator(
        `${selector} .vxe-table--fixed-right-wrapper .vxe-header--column`,
      )
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

  private tableSelector(table: "admin" | "mine" | "review") {
    return table === "admin"
      ? adminTableSelector
      : table === "mine"
        ? mineTableSelector
        : reviewTableSelector;
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
    const detailDialog = this.page
      .locator('[role="dialog"]:visible')
      .filter({ has: this.detailShell() })
      .last();
    const isModal = (await detailDialog.count()) > 0;
    if (isModal) {
      const backButton = this.detailShell().getByRole("button", {
        name: /返\s*回|Back/u,
      });
      if ((await backButton.count()) > 0) {
        await backButton.click();
      } else {
        await detailDialog.locator("button:has(svg.lucide-x)").last().click();
      }
      await expect(detailDialog).toBeHidden({ timeout: 15_000 });
    } else {
      await this.detailShell()
        .getByRole("button", { name: /返\s*回|Back/u })
        .click();
      await waitForRouteReady(this.page, 15_000);
    }
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
      .getByRole("button", { name: /添加插件|Add Plugin/u })
      .click();
    await expect(this.publishDrawer()).toBeVisible();
    await expect(
      this.publishDrawer()
        .locator("h2")
        .filter({
          hasText: /添加插件|Add Plugin/u,
        }),
    ).toBeVisible();
    await expect(
      this.publishDrawer().getByRole("heading", {
        name: /发布者资料|Publisher Profile/u,
      }),
    ).toHaveCount(0);
  }

  async openPublisherDrawer() {
    await this.page
      .getByRole("button", {
        name: /登记发布者|编辑发布者|Register Publisher|Edit Publisher/u,
      })
      .click();
    await expect(this.publisherDrawer()).toBeVisible();
    await expect(
      this.publisherDrawer().getByRole("heading", {
        name: /发布者资料|Publisher Profile/u,
      }),
    ).toBeVisible();
  }

  publisherDrawer() {
    return this.page
      .locator('[role="dialog"]:visible, .ant-drawer-content:visible')
      .filter({
        hasText: /登记发布者|编辑发布者|Register Publisher|Edit Publisher/u,
      })
      .last();
  }

  async expectPublisherFormValues(values: {
    contactEmail?: string;
    homepage?: string;
    name: string;
    summary?: string;
  }) {
    const drawer = this.publisherDrawer();
    // Publisher Key is an internal identifier and must not appear in the profile form.
    await expect(drawer.getByText(/发布者 Key|Publisher Key/u)).toHaveCount(0);
    await expect(
      this.formItem(drawer, /发布者名称|Publisher Name/u)
        .locator("input")
        .first(),
    ).toHaveValue(values.name);
    if (values.homepage !== undefined) {
      await expect(
        this.formItem(drawer, /主页|Homepage/u)
          .locator("input")
          .first(),
      ).toHaveValue(values.homepage);
    }
    if (values.contactEmail !== undefined) {
      await expect(
        this.formItem(drawer, /联系邮箱|Contact Email/u)
          .locator("input")
          .first(),
      ).toHaveValue(values.contactEmail);
    }
    if (values.summary !== undefined) {
      await expect(
        this.formItem(drawer, /摘要|Summary/u)
          .locator("textarea")
          .first(),
      ).toHaveValue(values.summary);
    }
  }

  async expectPublisherKeyHidden() {
    await expect(
      this.publisherDrawer().getByText(/发布者 Key|Publisher Key/u),
    ).toHaveCount(0);
  }

  async openNewVersionDrawer(pluginId: string) {
    // Row actions are fixed: Detail / New Version / Delist (no overflow menu).
    const newVersionButton = await this.rowActionButton(
      mineTableSelector,
      pluginId,
      /新版本|New Version/u,
    );
    await newVersionButton.click();
    await expect(this.publishDrawer()).toBeVisible();
    await expect(
      this.publishDrawer().getByText(/新版本|New Version/u, { exact: true }),
    ).toBeVisible();
  }

  async expectMineRowActions(pluginId: string) {
    const row = await this.rowActionScope(mineTableSelector, pluginId);
    await expect(
      row.getByRole("button", { name: /详\s*情|Details/u }),
    ).toBeVisible();
    await expect(
      row.getByRole("button", { name: /新\s*版\s*本|New\s+Version/u }),
    ).toBeVisible();
    await expect(
      row.getByRole("button", { name: /下\s*架|Delist/u }),
    ).toBeVisible();
    await expect(
      row.getByRole("button", { name: /发\s*布|Publish/u }),
    ).toHaveCount(0);
    await expect(
      row.getByRole("button", { name: /更\s*多|More/u }),
    ).toHaveCount(0);
  }

  async expectMineSourceColumn() {
    await this.expectColumn("mine", /来源|Source/u);
  }

  async closePublishDrawer() {
    await this.publishDrawer().getByRole("button").first().click();
    await expect(this.publishDrawer()).toBeHidden();
  }

  async closePublisherDrawer() {
    await this.publisherDrawer().getByRole("button").first().click();
    await expect(this.publisherDrawer()).toBeHidden();
  }

  async expectMineSearchFormStatusOptions() {
    const form = this.page.locator(".plugin-marketplace-mine").first();
    await expect(
      form.getByText(/关键词|Keyword/u, { exact: true }).first(),
    ).toBeVisible();
    // Prefer accessible label association; form-item full text also includes the
    // placeholder ("请选择"), so avoid exact full-item text matching.
    const statusSelect = form.getByLabel(/状态|Status/u).first();
    await expect(statusSelect).toBeVisible();
    await statusSelect.click();
    const dropdown = this.page.locator(".ant-select-dropdown:visible").last();
    await expect(dropdown.getByText("待验证", { exact: true })).toBeVisible();
    await expect(dropdown.getByText("待审核", { exact: true })).toBeVisible();
    await expect(dropdown.getByText("已发布", { exact: true })).toBeVisible();
    await expect(dropdown.getByText("处理失败", { exact: true })).toBeVisible();
    await expect(dropdown.getByText("已下架", { exact: true })).toBeVisible();
    await expect(dropdown.getByText("已废弃", { exact: true })).toBeVisible();
    // Market-only "草稿" must not appear; list tags show process states instead.
    await expect(dropdown.getByText("草稿", { exact: true })).toHaveCount(0);
    await this.page.keyboard.press("Escape");
  }

  async expectNoRegisterGitToolbarAction() {
    await expect(
      this.page.getByRole("button", {
        name: /登记 Git 源|Register Git Source/u,
      }),
    ).toHaveCount(0);
  }

  async expectPluginDraftReset() {
    // Package-add form only keeps distribution mode; upload list must reset.
    // Default distribution is Git repository (left option).
    await expect(
      this.publishDrawer().locator(".ant-upload-list-item"),
    ).toHaveCount(0);
    await expect(
      this.publishDrawer()
        .locator("label.ant-radio-button-wrapper-checked")
        .filter({ hasText: /Git 仓库|Git Repository/u }),
    ).toBeVisible();
  }

  async selectPublishSourceKind(kind: "git" | "upload") {
    const label =
      kind === "git" ? /Git 仓库|Git Repository/u : /上传包|Upload Package/u;
    const radio = this.publishDrawer().getByRole("radio", { name: label });
    await expect(radio).toBeAttached();
    // Ant Design button-style radios keep the native input hidden; click the
    // visible button label wrapper instead of the input itself.
    const button = this.publishDrawer()
      .locator("label.ant-radio-button-wrapper")
      .filter({ hasText: label })
      .first();
    if (await button.isVisible().catch(() => false)) {
      await button.click();
      return;
    }
    await radio.evaluate((element: HTMLInputElement) => {
      element.click();
    });
  }

  async expectAddPluginDrawerLayout() {
    const drawer = this.publishDrawer();
    // Only the drawer chrome title; body must not repeat the same heading.
    await expect(
      drawer.locator("h2").filter({ hasText: /添加插件|Add Plugin/u }),
    ).toBeVisible();
    await expect(
      drawer
        .locator("h3")
        .filter({ hasText: /添加插件|Add Plugin|插件基本信息|Plugin Basics/u }),
    ).toHaveCount(0);
    await expect(
      drawer.getByText(/分发方式|Distribution/u, { exact: true }),
    ).toBeVisible();
    await expect(
      drawer
        .locator(".mine-drawer-actions")
        .getByRole("button", { name: /添加插件|Add Plugin/u }),
    ).toBeVisible();
    await expect(
      drawer.locator(".mine-section-header").getByRole("button", {
        name: /添加插件|Add Plugin|保存并发现版本|Save and Discover|上传草稿|Upload Draft/u,
      }),
    ).toHaveCount(0);

    // Upload mode: package field is a form item aligned with sourceKind, not a
    // right-side aside panel. Dependencies use v-show, so assert visibility.
    const uploadChecked = drawer
      .locator("label.ant-radio-button-wrapper-checked")
      .filter({ hasText: /上传包|Upload Package/u });
    const packageField = drawer.getByTestId("mine-package-field");
    if (await uploadChecked.isVisible().catch(() => false)) {
      await expect(
        drawer
          .locator("label")
          .filter({ hasText: /上传压缩包|Upload Package/u }),
      ).toBeVisible();
      await expect(packageField).toBeVisible();
      await expect(packageField.locator('input[type="file"]')).toBeAttached();
      await expect(drawer.locator(".mine-add-aside")).toHaveCount(0);
    } else {
      await expect(packageField).toBeHidden();
    }
  }

  async expectReleaseTarget(_pluginId: string, _pluginType: PluginTypeLabel) {
    await expect(
      this.publishDrawer().getByRole("heading", {
        name: /上传压缩包|Upload Package/u,
      }),
    ).toBeVisible();
  }

  async submitEmptyPluginDraft() {
    await this.publishDrawer()
      .locator(".mine-drawer-actions")
      .getByRole("button", { name: /添加插件|Add Plugin/u })
      .click();
  }

  validationErrors() {
    return this.publishDrawer().getByRole("alert");
  }

  async fillPublisherProfile(values: {
    contactEmail?: string;
    homepage?: string;
    name: string;
    summary?: string;
  }) {
    const drawer = this.publisherDrawer();
    await this.fillField(drawer, /发布者名称|Publisher Name/u, values.name);
    if (values.homepage) {
      await this.fillField(drawer, /主页|Homepage/u, values.homepage);
    }
    if (values.contactEmail) {
      await this.fillField(
        drawer,
        /联系邮箱|Contact Email/u,
        values.contactEmail,
      );
    }
    if (values.summary) {
      await this.fillField(drawer, /摘要|Summary/u, values.summary);
    }
  }

  async savePublisher() {
    await this.publisherDrawer()
      .getByRole("button", { name: /保存发布者|Save Publisher/u })
      .click();
  }

  async fillPluginDraft(_values: {
    name: string;
    pluginId: string;
    pluginType: PluginTypeLabel;
    summary: string;
    visibility?: VisibilityLabel;
  }) {
    // Package-add flow no longer asks for hand-filled plugin basics; keep
    // compatibility by selecting the upload distribution mode only.
    await this.selectPublishSourceKind("upload");
  }

  async fillGitSource(values: {
    accessToken?: string;
    repoUrl: string;
    visibility?: VisibilityLabel;
  }) {
    // Add-plugin mode no longer has an inner section heading; scope to drawer.
    const drawer = this.publishDrawer();
    await this.selectPublishSourceKind("git");
    await this.fillField(drawer, /仓库地址|Repository URL/u, values.repoUrl);
    if (values.accessToken) {
      await this.fillField(
        drawer,
        /访问令牌|Access Token/u,
        values.accessToken,
      );
    }
    // Visibility is no longer selectable on Git add.
    void values.visibility;
    await expect(
      drawer
        .locator(".ant-form-item:visible")
        .filter({ hasText: /可见性|Visibility/u }),
    ).toHaveCount(0);
  }

  async saveGitSource() {
    await this.publishDrawer()
      .locator(".mine-drawer-actions")
      .getByRole("button", { name: /添加插件|Add Plugin/u })
      .click();
  }

  async savePluginDraft() {
    // Compatibility no-op: package add no longer has a separate save-draft step.
  }

  async fillReleaseUpload(_values: { version: string }) {
    // Version is parsed from the uploaded package; keep as compatibility no-op.
  }

  async setUploadFile(file: Parameters<Locator["setInputFiles"]>[0]) {
    const drawer = this.publishDrawer();
    // Prefer the form-aligned package field (add-plugin), then section heading
    // (new-version), then any file input in the drawer.
    // Add-plugin defaults to Git; switch to upload only when source-kind radios exist.
    const packageField = drawer.getByTestId("mine-package-field");
    if ((await packageField.count()) > 0) {
      if (await packageField.isVisible().catch(() => false)) {
        await packageField
          .locator('input[type="file"]')
          .first()
          .setInputFiles(file);
        return;
      }
      const uploadRadio = drawer
        .locator("label.ant-radio-button-wrapper")
        .filter({ hasText: /上传包|Upload Package/u });
      if (await uploadRadio.isVisible().catch(() => false)) {
        await this.selectPublishSourceKind("upload");
        await packageField
          .locator('input[type="file"]')
          .first()
          .setInputFiles(file);
        return;
      }
      // Hidden package field without radios = new-version mode; fall through.
    }
    const sectionInput = this.publishSection(
      /上传压缩包|Upload Package|版本包|Release Package/u,
    )
      .locator('input[type="file"]')
      .first();
    if ((await sectionInput.count()) > 0) {
      await sectionInput.setInputFiles(file);
      return;
    }
    await drawer.locator('input[type="file"]').first().setInputFiles(file);
  }

  async uploadDraft() {
    await this.publishDrawer()
      .locator(".mine-drawer-actions")
      .getByRole("button", {
        name: /添加插件|Add Plugin|上传草稿|Upload Draft/u,
      })
      .click();
  }

  async submitLatestDraft() {
    // Publish is now a row action; compatibility no-op for legacy call sites.
  }

  async publishOwnedPlugin(pluginId: string) {
    // Manual publish is no longer a row action; async pipeline owns listing.
    void pluginId;
  }

  async delistOwnedPlugin(pluginId: string) {
    const row = this.mineRow(pluginId);
    await row.getByRole("button", { name: /下架|Delist/u }).click();
    const confirm = this.page.locator(".ant-modal:visible").last();
    await expect(confirm).toBeVisible();
    await confirm.getByRole("button", { name: /确 定|OK|确定/u }).click();
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

  detailReleaseRow(version: string) {
    return this.tableRow(detailReleaseTableSelector, version);
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

  async expectDocumentLocaleOptions(locales: string[]) {
    const segmented = this.detailShell().locator(".ant-segmented").first();
    await expect(segmented).toBeVisible();
    for (const locale of locales) {
      await expect(segmented.getByText(locale, { exact: true })).toBeVisible();
    }
  }

  async switchDocumentLocale(locale: string) {
    await this.detailShell()
      .locator(".ant-segmented")
      .first()
      .getByText(locale, { exact: true })
      .click();
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
    // Layout uses separate pluginId / name columns (no combined plugin cell).
    const firstCell = firstRow.locator(".vxe-body--column").first();
    await expect(secondRow).toBeVisible();
    await expect(firstCell).toBeVisible();
    await expect
      .poll(async () => {
        const [firstBox, secondBox, cellBox] = await Promise.all([
          firstRow.boundingBox(),
          secondRow.boundingBox(),
          firstCell.boundingBox(),
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
    const row = await this.rowActionScope(tableSelector, rowText);
    return row.getByRole("button", { name: buttonName }).first();
  }

  private async rowActionScope(tableSelector: string, rowText: string) {
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
        return fixedRow;
      }
      return row;
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
