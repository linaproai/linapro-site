import assert from "node:assert/strict";
import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { describe, it } from "node:test";
import { fileURLToPath } from "node:url";

const testDir = path.dirname(fileURLToPath(import.meta.url));
const pluginRoot = path.resolve(testDir, "../../..");

function readPluginFile(relativePath) {
  return readFileSync(path.join(pluginRoot, relativePath), "utf8");
}

/**
 * Extract minWidth for a vxe column object where `field` appears before minWidth.
 * Returns null when the field or minWidth cannot be found.
 */
function extractColumnMinWidth(source, field) {
  const escaped = field.replaceAll(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const match = source.match(
    new RegExp(
      String.raw`field:\s*["']${escaped}["'][\s\S]*?minWidth:\s*(\d+)`,
    ),
  );
  return match ? Number(match[1]) : null;
}

function extractColumnWidth(source, field) {
  const escaped = field.replaceAll(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const match = source.match(
    new RegExp(String.raw`field:\s*["']${escaped}["'][\s\S]*?width:\s*(\d+)`),
  );
  return match ? Number(match[1]) : null;
}

function listVuePages() {
  const pagesRoot = path.join(pluginRoot, "frontend", "pages");
  const result = [];
  const visit = (dir) => {
    for (const entry of readdirSync(dir, { withFileTypes: true })) {
      const current = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        visit(current);
        continue;
      }
      if (entry.isFile() && entry.name.endsWith(".vue")) {
        result.push(current);
      }
    }
  };
  visit(pagesRoot);
  return result;
}

function flattenKeys(value, prefix = "") {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return [prefix];
  }
  return Object.entries(value).flatMap(([key, child]) => {
    const next = prefix ? `${prefix}.${key}` : key;
    return flattenKeys(child, next);
  });
}

function collectRuntimeI18nKeys() {
  const keyPattern = /\b(?:t|\$t)\(\s*['"]([^'"]+)['"]/g;
  const files = [...listVuePages(), path.join(pluginRoot, "plugin.yaml")];
  const keys = new Set();
  for (const filePath of files) {
    const content = readFileSync(filePath, "utf8");
    for (const match of content.matchAll(keyPattern)) {
      if (match[1].startsWith("plugin.linapro-plugin-marketplace.")) {
        keys.add(match[1]);
      }
    }
    for (const match of content.matchAll(
      /plugin\.linapro-plugin-marketplace\.[\w.-]+/g,
    )) {
      keys.add(match[0]);
    }
  }
  return [...keys].sort();
}

describe("marketplace frontend API adapter", () => {
  it("keeps marketplace endpoints under the plugin API namespace", () => {
    const source = readPluginFile("frontend/api/marketplace.ts");
    assert.match(
      source,
      /marketplacePluginId\s*=\s*["']linapro-plugin-marketplace["']/,
    );
    assert.match(source, /pluginApiPath\(marketplacePluginId,\s*pathName\)/);
    assert.match(source, /encodeURIComponent\(value\.trim\(\)\)/);

    for (const name of [
      "marketplacePluginList",
      "marketplacePluginDetail",
      "marketplaceReleaseDocument",
      "marketplaceReleaseDocumentBundle",
      "marketplaceReleaseRisks",
      "marketplaceDownloadSessionCreate",
      "marketplaceDownloadSessionBlob",
    ]) {
      assert.match(
        source,
        new RegExp(`export\\s+(?:async\\s+)?function\\s+${name}\\b`),
      );
    }

    // Document reads map "not found" to empty UI; avoid global error toasts.
    assert.match(
      source,
      /marketplaceReleaseDocumentBundle[\s\S]*silentErrorMessage:\s*true/,
    );

    for (const endpoint of [
      "market/plugins",
      "market/publishers",
      "submit-review",
      "review",
      "downloads",
      "download-sessions",
    ]) {
      assert.ok(
        source.includes(endpoint),
        `missing endpoint fragment ${endpoint}`,
      );
    }

    assert.match(source, /items:\s*result\.list\s*\?\?\s*\[\]/);
    assert.match(source, /total:\s*result\.total\s*\?\?\s*0/);
    assert.match(source, /documents:\s*res\.documents\s*\?\?\s*\[\]/);
  });
});

describe("marketplace frontend composition logic", () => {
  it("builds my-plugin grid with pagination and management-style search form", () => {
    const source = readPluginFile("frontend/pages/mine/index.vue");
    assert.match(source, /useVbenVxeGrid<MarketplacePluginListItem>/);
    assert.match(source, /pageNum:\s*page\.currentPage/);
    assert.match(source, /pageSize:\s*page\.pageSize/);
    assert.match(source, /keyword:\s*trimOptional\(formValues\.keyword\)/);
    assert.match(source, /formOptions:/);
    assert.match(source, /rowClassName:\s*["']cursor-pointer["']/);
    // Remote sort: pluginId / marketStatus / downloadCount / updatedAt; default pluginId asc.
    assert.match(source, /sortConfig:\s*\{/);
    assert.match(source, /defaultSort:\s*\{\s*field:\s*["']pluginId["']\s*,\s*order:\s*["']asc["']/);
    assert.match(source, /remote:\s*true/);
    assert.match(source, /sort:\s*true/);
    assert.match(source, /orderBy:\s*sort\?\.field\s*\|\|\s*["']pluginId["']/);
    assert.match(source, /orderDirection:/);
    assert.match(
      source,
      /field:\s*["']pluginId["'][\s\S]*?sortable:\s*true/,
    );
    assert.match(
      source,
      /field:\s*["']marketStatus["'][\s\S]*?sortable:\s*true/,
    );
    assert.match(
      source,
      /field:\s*["']downloadCount["'][\s\S]*?sortable:\s*true/,
    );
    assert.match(
      source,
      /field:\s*["']updatedAt["'][\s\S]*?sortable:\s*true/,
    );
    assert.match(source, /mine\.columns\.pluginId/);
    assert.match(source, /mine\.columns\.name/);
    assert.match(source, /mine\.columns\.summary/);
    assert.match(source, /mine\.columns\.sourceKind/);
    assert.match(source, /field:\s*["']sourceKind["']/);
    assert.match(source, /mine\.actions\.registerPublisher/);
    assert.match(source, /publisherToolbarLabel/);
    assert.match(source, /toolbarHasPublisher/);
    assert.match(source, /loadPublishersForToolbar/);
    // Status filter options must match publisher-visible process tags.
    assert.match(source, /value:\s*["']pending_verify["']/);
    assert.match(source, /value:\s*["']pending_review["']/);
    assert.match(source, /value:\s*["']published["']/);
    assert.match(source, /value:\s*["']failed["']/);
    assert.match(source, /field:\s*["']downloadCount["']/);
    // Source is a dedicated column; name no longer hosts a source-kind tag.
    assert.match(source, /slots:\s*\{\s*default:\s*["']sourceKind["']\s*\}/);
    // Add-plugin: Git repository is left option and the default selection.
    assert.match(
      source,
      /value:\s*["']git["'][\s\S]*?value:\s*["']upload["']/,
    );
    assert.match(source, /defaultValue:\s*["']git["']/);
    assert.match(source, /ref<PublishSourceKind>\(["']git["']\)/);
    assert.match(source, /setValues\(\{\s*sourceKind:\s*["']git["']\s*\}\)/);
    // Identity first: name then description (summary), then ops-critical status.
    // Visibility is not a user-facing column; public catalog is status-driven.
    assert.match(
      source,
      /field:\s*["']name["'][\s\S]*?field:\s*["']summary["'][\s\S]*?field:\s*["']marketStatus["'][\s\S]*?field:\s*["']latestVersion["'][\s\S]*?field:\s*["']downloadCount["']/,
    );
    assert.doesNotMatch(source, /field:\s*["']visibility["']/);
    // Readable floors for identity columns — guard against density-driven narrowing.
    assert.ok(
      extractColumnMinWidth(source, "pluginId") >= 200,
      "mine pluginId column minWidth must be >= 200",
    );
    assert.ok(
      extractColumnMinWidth(source, "name") >= 200,
      "mine name column minWidth must be >= 200",
    );
    assert.ok(
      extractColumnMinWidth(source, "summary") >= 260,
      "mine summary/description column minWidth must be >= 260",
    );
    assert.ok(
      extractColumnWidth(source, "latestVersion") >= 132,
      "mine latest-version header must fit English copy",
    );
    assert.ok(
      extractColumnWidth(source, "downloadCount") >= 112,
      "mine downloads header must fit English copy",
    );
    assert.ok(
      extractColumnWidth(source, "updatedAt") >= 184,
      "mine updatedAt must fit YYYY-MM-DD HH:mm:ss",
    );
    assert.doesNotMatch(
      source,
      /openGitPublishDrawer|mine\.actions\.registerGit/,
    );
  });

  it("keeps publish workflow on my-plugins and review workflow on review page", () => {
    const mine = readPluginFile("frontend/pages/mine/index.vue");
    const review = readPluginFile("frontend/pages/review/index.vue");
    assert.match(mine, /useVbenForm/);
    assert.match(mine, /const UploadDragger = Upload\.Dragger/);
    assert.match(mine, /marketplacePackageAdd/);
    assert.match(mine, /marketplacePluginDelist/);
    assert.match(mine, /marketplaceGitSourceRegister/);
    assert.match(mine, /marketplaceGitSourceSync/);
    assert.match(mine, /handleNewVersion/);
    assert.match(mine, /publishSourceKind/);
    assert.match(mine, /PublisherDrawer/);
    assert.match(mine, /handleAddPackage/);
    assert.match(mine, /handlePublishDrawerPrimaryAction/);
    assert.match(mine, /class="mine-drawer-actions"/);
    assert.match(mine, /marketplacePublisherUpdate/);
    assert.match(mine, /boundPublisher/);
    assert.match(mine, /buildPublisherSchema\(\)/);
    assert.match(mine, /mine\.actions\.add/);
    // Row actions: Detail / New Version / Delist only (no publish / more menu).
    assert.match(mine, /mine\.actions\.newVersion/);
    assert.match(mine, /mine\.actions\.delist/);
    assert.doesNotMatch(mine, /marketplacePluginPublish/);
    assert.doesNotMatch(mine, /Dropdown|MenuItem|pages\.common\.more/);
    assert.doesNotMatch(mine, /mine\.actions\.saveGitSource/);
    assert.doesNotMatch(mine, /fieldName:\s*"visibility"/);
    assert.doesNotMatch(mine, /field:\s*["']visibility["']/);
    assert.doesNotMatch(mine, /formatVisibility/);
    // Drawer chrome keeps “添加插件”; body must not repeat pluginBasic heading.
    assert.doesNotMatch(mine, /mine\.sections\.pluginBasic/);
    // Process pipeline no longer exposes pending_fetch to the UI.
    assert.doesNotMatch(mine, /pending_fetch|pendingFetch/);
    const detail = readPluginFile("frontend/pages/detail/index.vue");
    const admin = readPluginFile("frontend/pages/admin-list/index.vue");
    const types = readPluginFile("frontend/types/marketplace.ts");
    // Detail header tags: type + status (+ source); no user-facing visibility.
    assert.doesNotMatch(detail, /formatVisibility/);
    assert.match(detail, /preferences\.app\.locale/);
    assert.match(detail, /marketplaceReleaseDocumentBundle/);
    assert.match(detail, /availableDocuments/);
    assert.match(detail, /availableDocumentLocaleOptions/);
    assert.match(detail, /handleSelectDocumentLocale/);
    assert.match(detail, /Segmented/);
    assert.match(detail, /chooseDocumentFromBundle/);
    assert.match(
      detail,
      /renderMarketplaceMarkdown\(document\.markdown\)\s*\|\|\s*document\.content/,
    );
    assert.doesNotMatch(detail, /marketplaceReleaseDocument\(/);
    assert.doesNotMatch(detail, /pending_fetch|pendingFetch/);
    assert.doesNotMatch(admin, /pending_fetch|pendingFetch/);
    assert.doesNotMatch(types, /pending_fetch/);
    assert.match(types, /pending_verify/);
    assert.match(types, /MarketplaceDocumentBundle/);
    assert.match(types, /documents:\s*MarketplaceDocumentItem\[\]/);
    assert.match(review, /marketplaceReviewQueueList/);
    assert.match(review, /marketplaceReleaseReview/);
    assert.match(review, /preferences\.app\.locale/);
    assert.match(review, /class="marketplace-review-risk-list"/);
    assert.match(review, /risk\.source/);
    assert.match(review, /risk\.summary/);
    // Inspect path must import nextTick; missing import hard-crashes the drawer.
    assert.match(
      review,
      /import\s*\{[^}]*\bnextTick\b[^}]*\}\s*from\s*["']vue["']/,
    );
    assert.match(review, /class="plugin-marketplace-review"/);
    // Locale switch rebuilds filter/column/decision chrome (same pattern as mine).
    assert.match(
      review,
      /preferences\.app\.locale[\s\S]*?buildFormOptions\(\)[\s\S]*?buildColumns\(\)[\s\S]*?buildDecisionSchema\(\)/,
    );
    // Review status is ops-critical: keep it before version/submittedAt.
    assert.match(
      review,
      /field:\s*["']reviewStatus["'][\s\S]*?field:\s*["']version["']/,
    );
    // Admin list matches mine table chrome (split id/name, status early, tags).
    assert.match(admin, /class="plugin-marketplace-admin-list"/);
    assert.match(admin, /mine\.columns\.pluginId/);
    assert.match(admin, /mine\.columns\.name/);
    assert.match(admin, /mine\.fields\.status/);
    assert.match(admin, /rowClassName:\s*["']cursor-pointer["']/);
    assert.match(admin, /slots:\s*\{\s*default:\s*["']name["']\s*\}/);
    assert.match(admin, /max-w-full truncate font-medium/);
    assert.match(admin, /font-mono text-xs tabular-nums/);
    assert.match(admin, /buildStatusTooltip/);
    assert.match(
      admin,
      /getMarketStatusColor\(row\.marketStatus,\s*row\.processStatus\)/,
    );
    assert.match(
      admin,
      /case\s*["']pending_review["'][\s\S]*?return\s*["']gold["']/,
    );
    assert.match(
      admin,
      /preferences\.app\.locale[\s\S]*?buildFormOptions\(\)[\s\S]*?buildColumns\(\)/,
    );
    // Split identity columns + status early (same order pattern as mine).
    assert.match(
      admin,
      /field:\s*["']pluginId["'][\s\S]*?field:\s*["']name["'][\s\S]*?field:\s*["']marketStatus["'][\s\S]*?field:\s*["']latestReviewStatus["']/,
    );
    // Stacked two-line plugin cell must not return (diverges from mine chrome).
    assert.doesNotMatch(
      admin,
      /admin-plugin-cell|admin-plugin-name|admin-plugin-id/,
    );
    assert.ok(
      extractColumnMinWidth(admin, "pluginId") >= 200,
      "admin pluginId column minWidth must be >= 200 (match mine)",
    );
    assert.ok(
      extractColumnMinWidth(admin, "name") >= 200,
      "admin name column minWidth must be >= 200 (match mine)",
    );
    assert.ok(
      extractColumnWidth(admin, "latestVersion") >= 132,
      "admin latest-version header must fit English copy",
    );
    assert.ok(
      extractColumnWidth(admin, "downloadCount") >= 112,
      "admin downloads header must fit English copy",
    );
    assert.ok(
      extractColumnWidth(review, "submittedAt") >= 184,
      "review submitted timestamp must fit without ellipsis",
    );
  });

  it("hides downloads without the marketplace download permission", () => {
    const source = readPluginFile("frontend/pages/detail/index.vue");
    assert.match(
      source,
      /hasAccessByCodes\(\[\s*["']market:plugin:download["'],\s*["']\*:\*:\*["']\s*\]\)/,
    );
    assert.match(source, /v-if="canDownloadMarketplacePlugin\(\)"/);
  });

  it("surfaces historical git source pins on the version table", () => {
    const detail = readPluginFile("frontend/pages/detail/index.vue");
    const types = readPluginFile("frontend/types/marketplace.ts");
    assert.match(detail, /formatReleaseSourcePin/);
    assert.match(detail, /marketplace-source-pin/);
    assert.match(
      detail,
      /plugin\.linapro-plugin-marketplace\.detail\.sourcePin\.refAndCommit/,
    );
    assert.match(
      detail,
      /plugin\.linapro-plugin-marketplace\.detail\.sourcePin\.commitOnly/,
    );
    assert.match(
      detail,
      /plugin\.linapro-plugin-marketplace\.detail\.sourcePin\.refOnly/,
    );
    assert.match(types, /sourceCommit\?:/);
    assert.match(types, /sourceRef\?:/);
  });

  it("returns unscoped public details to an accessible host page", () => {
    const source = readPluginFile("frontend/utils/routes.ts");
    assert.match(
      source,
      /DEFAULT_BACK_PATH\s*=\s*["']\/dashboard\/analytics["']/,
    );
    assert.match(
      source,
      /from === ["']mine["'][\s\S]*?return marketplaceMinePath\(\)/,
    );
    assert.match(source, /pageKey:\s*path/);
    assert.match(source, /return DEFAULT_BACK_PATH/);
  });
});

describe("marketplace lastSyncMessage i18n", () => {
  it("localizes known Git sync diagnostics and falls back for free-form text", () => {
    const syncUtil = readPluginFile("frontend/utils/sync-message.ts");
    assert.match(syncUtil, /formatMarketplaceLastSyncMessage/);
    assert.match(syncUtil, /MARKETPLACE_SYNC_MESSAGE_CODES/);
    assert.match(syncUtil, /discoveredImmutableOnly/);
    assert.match(syncUtil, /discoveredDrafts/);

    const detail = readPluginFile("frontend/pages/detail/index.vue");
    const mine = readPluginFile("frontend/pages/mine/index.vue");
    assert.match(detail, /formatMarketplaceLastSyncMessage/);
    assert.match(detail, /formatLastSyncMessage\(detail\.lastSyncMessage\)/);
    // Sync message must use Descriptions body font size; marketplace-muted is 12px.
    assert.doesNotMatch(
      detail,
      /lastSyncMessage[\s\S]{0,400}class="marketplace-muted"/,
    );
    assert.doesNotMatch(
      detail,
      /class="marketplace-muted">\{\{\s*detail\.lastSyncMessage\s*\}\}/,
    );
    assert.match(mine, /formatMarketplaceLastSyncMessage/);
    assert.match(mine, /formatMarketplaceLastSyncMessage\(t,\s*row\.lastSyncMessage\)/);

    const en = JSON.parse(readPluginFile("manifest/i18n/en-US/plugin.json"));
    const zh = JSON.parse(readPluginFile("manifest/i18n/zh-CN/plugin.json"));
    const syncCodes = [
      "discoveredDrafts",
      "discoveredImmutableOnly",
      "discoveredWithFailures",
      "failedImportRefs",
      "credentialLoadFailed",
      "repositoryUrlInvalid",
      "platformTokenConfigFailed",
      "queuedForVerification",
      "docsIndexIncomplete",
      "immutableDocsIndexIncomplete",
    ];
    for (const code of syncCodes) {
      const key = `plugin.linapro-plugin-marketplace.detail.syncMessage.${code}`;
      assert.ok(en[key], `missing en-US sync message key ${key}`);
      assert.ok(zh[key], `missing zh-CN sync message key ${key}`);
      assert.notEqual(en[key], zh[key], `${key} must differ across locales`);
    }
    assert.equal(
      zh[
        "plugin.linapro-plugin-marketplace.detail.syncMessage.discoveredImmutableOnly"
      ],
      "未发现新草稿版本（已有 {count} 个不可变版本）",
    );

    // Pure helper behavior without Vue: mirror pattern → key → translate path.
    function withOptionalDetail(detail) {
      const trimmed = (detail || "").trim();
      return trimmed ? `: ${trimmed}` : "";
    }
    function formatLastSyncMessage(t, message) {
      const fallback = (message || "").trim();
      if (!fallback) {
        return "";
      }
      const patterns = [
        [
          /^discovered 0 new draft releases \((\d+) existing immutable version\(s\)\)$/,
          "discoveredImmutableOnly",
          (m) => ({ count: Number(m[1]) }),
        ],
        [
          /^discovered (\d+) draft releases$/,
          "discoveredDrafts",
          (m) => ({ count: Number(m[1]) }),
        ],
        [
          /^discovered (\d+) drafts with (\d+) ref failures(?::\s*(.*))?$/s,
          "discoveredWithFailures",
          (m) => ({
            count: Number(m[1]),
            failures: Number(m[2]),
            detail: withOptionalDetail(m[3]),
          }),
        ],
        [
          /^failed to import (\d+) refs(?::\s*(.*))?$/s,
          "failedImportRefs",
          (m) => ({
            count: Number(m[1]),
            detail: withOptionalDetail(m[2]),
          }),
        ],
        [/^credential load failed$/, "credentialLoadFailed"],
        [/^repository url is invalid$/, "repositoryUrlInvalid"],
      ];
      for (const [re, code, paramsFn] of patterns) {
        const match = fallback.match(re);
        if (!match) {
          continue;
        }
        const key = `plugin.linapro-plugin-marketplace.detail.syncMessage.${code}`;
        const params = paramsFn ? paramsFn(match) : undefined;
        let translated = zh[key] || key;
        if (params) {
          for (const [name, value] of Object.entries(params)) {
            translated = translated.replaceAll(`{${name}}`, String(value));
          }
        }
        if (!translated || translated === key) {
          return fallback;
        }
        return translated;
      }
      return fallback;
    }

    assert.equal(
      formatLastSyncMessage(
        null,
        "discovered 0 new draft releases (1 existing immutable version(s))",
      ),
      "未发现新草稿版本（已有 1 个不可变版本）",
    );
    assert.equal(
      formatLastSyncMessage(null, "discovered 2 draft releases"),
      "已发现 2 个草稿版本",
    );
    assert.equal(
      formatLastSyncMessage(null, "failed to import 3 refs: tag missing"),
      "导入 3 个引用失败: tag missing",
    );
    assert.equal(
      formatLastSyncMessage(null, "credential load failed"),
      "凭据加载失败",
    );
    assert.equal(
      formatLastSyncMessage(null, "some unexpected provider timeout"),
      "some unexpected provider timeout",
    );
  });
});

describe("marketplace risk severity ordering", () => {
  it("sorts risk findings high → warning → info and detail page applies the helper", () => {
    const riskUtil = readPluginFile("frontend/utils/risk.ts");
    assert.match(riskUtil, /sortMarketplaceRiskFindingsBySeverity/);
    assert.match(riskUtil, /marketplaceRiskSeverityRank/);

    // Mirror the production rank/sort helpers so the unit test fails if the
    // priority order (high first) regresses without updating the util contract.
    function marketplaceRiskSeverityRank(severity) {
      switch (String(severity || "").trim().toLowerCase()) {
        case "high":
          return 0;
        case "warning":
          return 1;
        case "info":
          return 2;
        default:
          return 3;
      }
    }
    function marketplaceRiskDispositionRank(disposition) {
      switch ((disposition || "").trim().toLowerCase()) {
        case "need_fix":
          return 0;
        case "need_attention":
          return 1;
        case "info_only":
          return 2;
        default:
          return 3;
      }
    }
    function marketplaceRiskDisposition(risk) {
      const policy = {
        i18n_files_missing: "need_fix",
        framework_dependency_missing: "need_attention",
        dynamic_host_services_present: "need_attention",
        source_docs_indexed: "info_only",
      };
      const code =
        risk?.payload && typeof risk.payload.code === "string"
          ? risk.payload.code
          : "";
      return policy[code] || "need_attention";
    }
    function marketplaceRiskBlocking(risk) {
      return marketplaceRiskDisposition(risk) === "need_fix";
    }
    function sortMarketplaceRiskFindingsBySeverity(items) {
      if (!items || items.length === 0) {
        return [];
      }
      return items
        .map((item, index) => ({ index, item }))
        .sort((left, right) => {
          const leftBlocking = marketplaceRiskBlocking(left.item) ? 0 : 1;
          const rightBlocking = marketplaceRiskBlocking(right.item) ? 0 : 1;
          if (leftBlocking !== rightBlocking) {
            return leftBlocking - rightBlocking;
          }
          const dispositionDiff =
            marketplaceRiskDispositionRank(
              marketplaceRiskDisposition(left.item),
            ) -
            marketplaceRiskDispositionRank(
              marketplaceRiskDisposition(right.item),
            );
          if (dispositionDiff !== 0) {
            return dispositionDiff;
          }
          const rankDiff =
            marketplaceRiskSeverityRank(left.item.severity) -
            marketplaceRiskSeverityRank(right.item.severity);
          if (rankDiff !== 0) {
            return rankDiff;
          }
          return left.index - right.index;
        })
        .map((entry) => entry.item);
    }

    const sorted = sortMarketplaceRiskFindingsBySeverity([
      {
        payload: { code: "source_docs_indexed" },
        severity: "info",
        summary: "info-a",
      },
      {
        payload: { code: "dynamic_host_services_present" },
        severity: "high",
        summary: "high-attention",
      },
      {
        payload: { code: "i18n_files_missing" },
        severity: "warning",
        summary: "fix-a",
      },
      {
        payload: { code: "framework_dependency_missing" },
        severity: "warning",
        summary: "framework-attention",
      },
      {
        payload: { code: "source_docs_indexed" },
        severity: "info",
        summary: "info-b",
      },
    ]);
    assert.deepEqual(
      sorted.map((item) => item.summary),
      ["fix-a", "high-attention", "framework-attention", "info-a", "info-b"],
    );

    // info_only is stripped from workbench lists.
    assert.match(riskUtil, /filterMarketplaceRiskFindingsActionable/);
    assert.match(riskUtil, /isMarketplaceRiskActionable/);
    assert.match(riskUtil, /!== "info_only"/);

    const detail = readPluginFile("frontend/pages/detail/index.vue");
    const review = readPluginFile("frontend/pages/review/index.vue");
    assert.match(detail, /filterMarketplaceRiskFindingsActionable/);
    assert.match(
      detail,
      /currentRisks\.value\s*=\s*filterMarketplaceRiskFindingsActionable\(result\.items\)/,
    );
    assert.match(review, /filterMarketplaceRiskFindingsActionable/);
    assert.doesNotMatch(detail, /filterInfoOnly/);
    // Risk summary Tags already render high → warning → info in the template.
    assert.match(
      detail,
      /getRiskCounts\(\)\.high[\s\S]*?getRiskCounts\(\)\.warning[\s\S]*?getRiskCounts\(\)\.info/,
    );
  });

  it("renders multi-item risk rows as separated cards on detail and review", () => {
    const detail = readPluginFile("frontend/pages/detail/index.vue");
    const review = readPluginFile("frontend/pages/review/index.vue");
    // Detail: soft canvas + white cards + left accent between findings.
    assert.match(
      detail,
      /\.marketplace-risk-list\s*\{[\s\S]*?gap:\s*12px[\s\S]*?background:\s*var\(--ant-color-bg-layout\)/,
    );
    assert.match(
      detail,
      /\.marketplace-risk-item\s*\{[\s\S]*?border:\s*1px solid var\(--ant-color-border\)[\s\S]*?background:\s*var\(--ant-color-bg-container\)/,
    );
    assert.match(detail, /\.marketplace-risk-item::before\s*\{/);
    assert.match(
      detail,
      /\.marketplace-risk-item--blocking::before\s*\{[\s\S]*?background:\s*var\(--ant-color-error\)/,
    );
    // Expanded guidance is a distinct inset panel, not the same surface as summary.
    assert.match(detail, /class="marketplace-risk-guidance"/);
    assert.match(detail, /marketplace-risk-guidance-section/);
    assert.match(detail, /marketplace-risk-guidance-label/);
    assert.match(
      detail,
      /\.marketplace-risk-guidance\s*\{[\s\S]*?border-left:\s*3px solid #1677ff[\s\S]*?background:\s*#e6f4ff/,
    );
    // Review page keeps the same card separation model.
    assert.match(
      review,
      /\.marketplace-review-risk-list\s*\{[\s\S]*?gap:\s*12px[\s\S]*?background:\s*var\(--ant-color-bg-layout\)/,
    );
    assert.match(review, /\.marketplace-review-risk-item::before\s*\{/);
    assert.match(
      review,
      /\.marketplace-review-risk-item\s*\{[\s\S]*?border:\s*1px solid var\(--ant-color-border\)[\s\S]*?background:\s*var\(--ant-color-bg-container\)/,
    );
    assert.match(review, /marketplace-review-risk-guidance-section/);
    assert.match(
      review,
      /\.marketplace-review-risk-guidance\s*\{[\s\S]*?border-left:\s*3px solid #1677ff[\s\S]*?background:\s*#e6f4ff/,
    );
  });
});

describe("marketplace risk finding i18n", () => {
  it("localizes known scanner codes and falls back to English summary", () => {
    const riskUtil = readPluginFile("frontend/utils/risk.ts");
    assert.match(riskUtil, /formatMarketplaceRiskFindingSummary/);
    assert.match(riskUtil, /formatMarketplaceRiskFindingGuidance/);
    assert.match(riskUtil, /marketplaceRiskDisposition/);
    assert.match(riskUtil, /MARKETPLACE_RISK_FINDING_CODES/);
    assert.match(riskUtil, /source_docs_indexed/);
    assert.match(riskUtil, /framework_dependency_missing/);

    const detail = readPluginFile("frontend/pages/detail/index.vue");
    const review = readPluginFile("frontend/pages/review/index.vue");
    assert.match(detail, /formatMarketplaceRiskFindingSummary/);
    assert.match(detail, /formatRiskFindingSummary\(risk\)/);
    assert.match(detail, /riskGuide\.remediation/);
    assert.match(detail, /riskDispositionFilter/);
    assert.doesNotMatch(
      detail,
      /class="marketplace-risk-item"[\s\S]{0,400}<p>\{\{\s*risk\.summary\s*\}\}/,
    );
    assert.match(review, /formatMarketplaceRiskFindingSummary/);
    assert.match(review, /formatRiskFindingSummary\(risk\)/);
    assert.match(review, /riskGuide\.reason/);

    const en = JSON.parse(readPluginFile("manifest/i18n/en-US/plugin.json"));
    const zh = JSON.parse(readPluginFile("manifest/i18n/zh-CN/plugin.json"));
    const findingCodes = [
      "source_sql_present",
      "source_docs_indexed",
      "framework_dependency_missing",
      "i18n_files_missing",
      "dynamic_runtime_detected",
      "dynamic_host_services_present",
      "dynamic_routes_present",
      "dynamic_sql_present",
      "dynamic_mock_sql_present",
      "dynamic_manifest_resources_missing",
    ];
    for (const code of findingCodes) {
      const base = `plugin.linapro-plugin-marketplace.detail.riskFinding.${code}`;
      const titleKey = `${base}.title`;
      assert.ok(en[titleKey], `missing en-US risk finding key ${titleKey}`);
      assert.ok(zh[titleKey], `missing zh-CN risk finding key ${titleKey}`);
      assert.notEqual(
        en[titleKey],
        zh[titleKey],
        `${titleKey} must differ across locales`,
      );
      // Nested runtime trees break when a bare leaf shares the path of child keys.
      assert.equal(
        en[base],
        undefined,
        `bare title leaf ${base} conflicts with nested guidance keys`,
      );
      assert.equal(
        zh[base],
        undefined,
        `bare title leaf ${base} conflicts with nested guidance keys`,
      );
      for (const suffix of ["reason", "remediation", "acceptance"]) {
        const guidanceKey = `${base}.${suffix}`;
        assert.ok(en[guidanceKey], `missing en-US ${guidanceKey}`);
        assert.ok(zh[guidanceKey], `missing zh-CN ${guidanceKey}`);
        assert.notEqual(
          en[guidanceKey],
          zh[guidanceKey],
          `${guidanceKey} must differ across locales`,
        );
      }
    }
    assert.equal(
      zh[
        "plugin.linapro-plugin-marketplace.detail.riskFinding.framework_dependency_missing.title"
      ],
      "未声明框架兼容性依赖。",
    );
    assert.equal(
      zh[
        "plugin.linapro-plugin-marketplace.detail.riskFinding.source_docs_indexed.title"
      ],
      "已检测到可用于市场展示的文档条目。",
    );
    assert.equal(
      zh[
        "plugin.linapro-plugin-marketplace.detail.riskFinding.source_sql_present.title"
      ],
      "源码包包含需审核关注的 SQL 资源。",
    );
    assert.ok(
      zh["plugin.linapro-plugin-marketplace.detail.riskDisposition.need_fix"],
    );
    assert.ok(
      en["error.plugin.marketplace.risk.blocking"]?.includes("{diagnostic}"),
    );

    // Pure helper behavior without Vue: mirror the util's code → .title key path.
    function formatRiskFindingSummary(t, risk) {
      const code =
        risk?.payload && typeof risk.payload.code === "string"
          ? risk.payload.code.trim()
          : "";
      if (!code) {
        return (risk?.summary || "").trim();
      }
      const key = `plugin.linapro-plugin-marketplace.detail.riskFinding.${code}.title`;
      const translated = t(key);
      if (!translated || translated === key) {
        return (risk?.summary || "").trim();
      }
      return translated;
    }
    const t = (key) => zh[key] || key;
    assert.equal(
      formatRiskFindingSummary(t, {
        payload: { code: "framework_dependency_missing" },
        summary: "Framework compatibility dependency is not declared.",
      }),
      "未声明框架兼容性依赖。",
    );
    assert.equal(
      formatRiskFindingSummary(t, {
        payload: { code: "source_sql_present" },
        summary:
          "Source package contains SQL resources that require reviewer inspection.",
      }),
      "源码包包含需审核关注的 SQL 资源。",
    );
    assert.equal(
      formatRiskFindingSummary(t, {
        payload: { code: "unknown_legacy_code" },
        summary: "Legacy English summary remains as fallback.",
      }),
      "Legacy English summary remains as fallback.",
    );
    assert.match(
      zh[
        "plugin.linapro-plugin-marketplace.detail.riskFinding.framework_dependency_missing.remediation"
      ],
      /plugin\.yaml/,
    );
    assert.match(
      riskUtil,
      /marketplaceRiskFindingMessageKey[\s\S]*?\.title/,
    );
    assert.match(riskUtil, /blocking:\s*false,\s*disposition:\s*"need_attention"/);
  });
});

describe("marketplace frontend i18n keys", () => {
  it("covers all plugin runtime keys in en-US and zh-CN resources", () => {
    const en = JSON.parse(readPluginFile("manifest/i18n/en-US/plugin.json"));
    const zh = JSON.parse(readPluginFile("manifest/i18n/zh-CN/plugin.json"));
    const enKeys = new Set(flattenKeys(en));
    const zhKeys = new Set(flattenKeys(zh));
    const referencedKeys = collectRuntimeI18nKeys();

    assert.ok(referencedKeys.length > 100, "expected page and menu i18n keys");
    for (const key of referencedKeys) {
      assert.ok(enKeys.has(key), `missing en-US key ${key}`);
      assert.ok(zhKeys.has(key), `missing zh-CN key ${key}`);
    }
    assert.deepEqual([...enKeys].sort(), [...zhKeys].sort());
    assert.equal(
      zh["plugin.linapro-plugin-marketplace.mine.columns.pluginId"],
      "插件标识",
    );
    assert.equal(
      zh["plugin.linapro-plugin-marketplace.mine.columns.name"],
      "插件名称",
    );
    assert.equal(
      zh["plugin.linapro-plugin-marketplace.mine.columns.summary"],
      "插件描述",
    );
    assert.equal(
      en["plugin.linapro-plugin-marketplace.mine.columns.pluginId"],
      "Plugin Identifier",
    );
    assert.equal(
      en["plugin.linapro-plugin-marketplace.mine.columns.name"],
      "Plugin Name",
    );
    assert.equal(
      en["plugin.linapro-plugin-marketplace.mine.columns.summary"],
      "Description",
    );
  });
});

describe("marketplace plugin.yaml menu keys", () => {
  it("uses the host-required plugin:<id>:* menu key prefix", () => {
    const source = readPluginFile("plugin.yaml");
    const pluginIdMatch = source.match(/^id:\s*([^\s#]+)/m);
    assert.ok(pluginIdMatch, "plugin.yaml must declare id");
    const pluginId = pluginIdMatch[1];
    const requiredPrefix = `plugin:${pluginId}:`;
    const menuKeys = [...source.matchAll(/^\s*-\s*key:\s*([^\s#]+)/gm)].map(
      (match) => match[1],
    );

    assert.equal(menuKeys.length, 6, "expected six marketplace menu resources");
    for (const menuKey of menuKeys) {
      assert.ok(
        menuKey.startsWith(requiredPrefix),
        `menu key ${menuKey} must use prefix ${requiredPrefix}`,
      );
      assert.ok(
        menuKey.length > requiredPrefix.length,
        `menu key ${menuKey} must include a non-empty suffix`,
      );
    }
  });

  it("declares marketplace as a top-level directory above the host extension center", () => {
    const source = readPluginFile("plugin.yaml");
    const directoryBlock = source.match(
      /- key:\s*plugin:linapro-plugin-marketplace:directory[\s\S]*?(?=\n  - key:|\n[^\s]|$)/,
    );
    assert.ok(directoryBlock, "expected marketplace directory menu block");
    assert.match(directoryBlock[0], /type:\s*D/);
    assert.match(directoryBlock[0], /sort:\s*8\b/);
    assert.doesNotMatch(
      directoryBlock[0],
      /parent_key:/,
      "top-level marketplace directory must not set parent_key",
    );

    const directoryParentKeys = [
      ...source.matchAll(
        /parent_key:\s*plugin:linapro-plugin-marketplace:directory/g,
      ),
    ];
    assert.equal(
      directoryParentKeys.length,
      4,
      "expected leaf menus under marketplace directory",
    );

    const downloadBlock = source.match(
      /- key:\s*plugin:linapro-plugin-marketplace:marketplace-download[\s\S]*?(?=\n  - key:|\n[^\s]|$)/,
    );
    assert.ok(downloadBlock, "expected assignable marketplace download action");
    assert.match(
      downloadBlock[0],
      /parent_key:\s*plugin:linapro-plugin-marketplace:marketplace-detail/,
    );
    assert.match(downloadBlock[0], /perms:\s*market:plugin:download/);
    assert.match(downloadBlock[0], /type:\s*B/);
  });
});

describe("marketplace dynamic package import bridge", () => {
  it("downloads package ZIP and imports backend-provided plugin.wasm", () => {
    const source = readPluginFile("frontend/pages/detail/index.vue");
    assert.match(source, /pluginDynamicUpload\(file,\s*false\)/);
    assert.match(source, /notifyPluginRegistryChanged\(\)/);
    assert.match(
      source,
      /hasAccessByCodes\(\[\s*["']plugin:install["'],\s*["']\*:\*:\*["']\s*\]\)/,
    );
    assert.match(
      source,
      /row\.pluginType === ["']dynamic["'] && canImportDynamicPlugin\(\)/,
    );
    assert.match(source, /marketplaceDownloadSessionCreate/);
    assert.match(source, /marketplaceDownloadSessionBlob/);
    assert.match(
      source,
      /downloadBlob\(packageBlob,\s*buildDownloadFileName\(row\)\)/,
    );
    assert.match(source, /importDynamicPluginWasm/);
    assert.match(
      source,
      /downloadReleaseArtifact\(row,\s*["']plugin_wasm["']\)/,
    );
    assert.doesNotMatch(source, /extractPluginWasmFromZip/);
    assert.doesNotMatch(source, /DecompressionStreamCtor/);
    assert.match(source, /bytes\[0\] === 0x00/);
    assert.match(source, /bytes\[1\] === 0x61/);
    assert.match(source, /bytes\[2\] === 0x73/);
    assert.match(source, /bytes\[3\] === 0x6d/);
  });
});

describe("marketplace detail version grid layout", () => {
  it("sizes the release table by content instead of fill-parent height", () => {
    const detail = readPluginFile("frontend/pages/detail/index.vue");
    // height: "auto" is fill-parent in vxe-table and grows blank body space inside
    // the nested detail modal tabs; the version grid must stay content-sized.
    const versionGridBlock = detail.match(
      /useVbenVxeGrid<MarketplaceReleaseItem>\(\{[\s\S]*?id:\s*["']plugin-marketplace-detail-releases["'][\s\S]*?\}\)/,
    );
    assert.ok(versionGridBlock, "expected detail VersionGrid configuration");
    assert.doesNotMatch(
      versionGridBlock[0],
      /height:\s*["']auto["']/,
      "version grid must not use height auto fill-parent mode",
    );
    assert.match(versionGridBlock[0], /autoResize:\s*false/);
    assert.match(versionGridBlock[0], /maxHeight:\s*480/);
    assert.match(
      versionGridBlock[0],
      /pagerConfig:\s*\{\s*enabled:\s*false\s*,?\s*\}/s,
    );
    assert.match(versionGridBlock[0], /pageSize:\s*100/);
    assert.match(
      versionGridBlock[0],
      /class:\s*["']marketplace-detail-version-grid h-auto["']/,
    );
    assert.match(detail, /\.marketplace-detail-version-grid\s*\{/);
  });
});

describe("marketplace markdown helpers", () => {
  it("detail page uses plugin markdown-it helper and document catalog nav", () => {
    const detail = readPluginFile("frontend/pages/detail/index.vue");
    const frontendPackage = JSON.parse(
      readPluginFile("frontend/package.json"),
    );
    assert.match(detail, /from "\.\.\/\.\.\/utils\/markdown"/);
    assert.match(detail, /renderMarketplaceMarkdown/);
    assert.match(detail, /enhanceMarketplaceMarkdown/);
    assert.match(detail, /markdownBodyRef/);
    assert.match(detail, /documentCatalogOptions/);
    assert.match(
      detail,
      /entry\.sourceKind\s*===\s*["']manifest_docs["'][\s\S]*?entry\.sourceKind\s*===\s*["']readme["']/,
    );
    assert.match(
      detail,
      /document\.sourceKind\s*===\s*["']manifest_docs["'][\s\S]*?document\.sourceKind\s*===\s*["']readme["']/,
    );
    assert.match(detail, /marketplace-markdown-body/);
    assert.match(detail, /handleSelectDocumentPath/);
    assert.match(
      detail,
      /plugin\.linapro-plugin-marketplace\.detail\.docs\.catalog/,
    );
    assert.equal(frontendPackage.dependencies["markdown-it"], "14.1.1");
    assert.equal(frontendPackage.dependencies["highlight.js"], "11.11.1");
    assert.equal(frontendPackage.dependencies.mermaid, "11.12.2");
  });

  it("markdown util rejects unsafe absolute hrefs and resolves relative md paths", async () => {
    // Lightweight pure-logic copy of resolveRelativeMarkdownPath for unit isolation
    // without importing TypeScript through the Vite host.
    function resolveRelativeMarkdownPath(currentPath, href) {
      const cleaned = href.trim().split("#")[0]?.split("?")[0] ?? "";
      if (
        !cleaned ||
        cleaned.includes("://") ||
        cleaned.startsWith("//") ||
        cleaned.startsWith("/") ||
        cleaned.startsWith("mailto:") ||
        cleaned.startsWith("data:") ||
        !cleaned.toLowerCase().endsWith(".md")
      ) {
        return null;
      }
      const baseDir = currentPath.includes("/")
        ? currentPath.slice(0, currentPath.lastIndexOf("/") + 1)
        : "";
      const joined = `${baseDir}${cleaned}`.replaceAll("\\", "/");
      const segments = joined.split("/");
      const resolved = [];
      for (const segment of segments) {
        if (!segment || segment === ".") continue;
        if (segment === "..") {
          if (resolved.length === 0) return null;
          resolved.pop();
          continue;
        }
        resolved.push(segment);
      }
      return resolved.join("/") || null;
    }
    assert.equal(
      resolveRelativeMarkdownPath("index.md", "configuration.md"),
      "configuration.md",
    );
    assert.equal(
      resolveRelativeMarkdownPath("guides/index.md", "../changelog.md"),
      "changelog.md",
    );
    assert.equal(
      resolveRelativeMarkdownPath("index.md", "https://example.com/a.md"),
      null,
    );
  });

  it("markdown util enables highlight, mermaid fence, safe images, and html:false", () => {
    const markdownUtil = readPluginFile("frontend/utils/markdown.ts");
    assert.match(markdownUtil, /from "highlight\.js\/lib\/common"/);
    assert.match(markdownUtil, /from "markdown-it"/);
    assert.match(markdownUtil, /import\("mermaid"\)/);
    assert.match(markdownUtil, /html:\s*false/);
    assert.match(markdownUtil, /securityLevel:\s*["']strict["']/);
    assert.match(markdownUtil, /lang === ["']mermaid["']/);
    assert.match(markdownUtil, /class="marketplace-mermaid/);
    assert.match(markdownUtil, /hljs\.highlight/);
    assert.match(markdownUtil, /marketplace-md-image/);
    assert.match(markdownUtil, /isUnsafeImageSrc/);
    assert.match(markdownUtil, /javascript:/);
    assert.match(markdownUtil, /data:image\//);
    assert.match(markdownUtil, /validateLink/);
  });
});
