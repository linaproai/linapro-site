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
    assert.match(source, /mine\.columns\.pluginId/);
    assert.match(source, /mine\.columns\.name/);
    assert.match(source, /mine\.columns\.summary/);
    assert.match(source, /mine\.columns\.sourceKind/);
    assert.match(source, /field:\s*["']sourceKind["']/);
    assert.match(source, /mine\.actions\.registerPublisher/);
    // Status filter options must match publisher-visible process tags.
    assert.match(source, /value:\s*["']pending_verify["']/);
    assert.match(source, /value:\s*["']pending_review["']/);
    assert.match(source, /value:\s*["']published["']/);
    assert.match(source, /value:\s*["']failed["']/);
    assert.match(source, /field:\s*["']downloadCount["']/);
    // Source is a dedicated column; name no longer hosts a source-kind tag.
    assert.match(source, /slots:\s*\{\s*default:\s*["']sourceKind["']\s*\}/);
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
    // Drawer chrome keeps “添加插件”; body must not repeat pluginBasic heading.
    assert.doesNotMatch(mine, /mine\.sections\.pluginBasic/);
    // Process pipeline no longer exposes pending_fetch to the UI.
    assert.doesNotMatch(mine, /pending_fetch|pendingFetch/);
    const detail = readPluginFile("frontend/pages/detail/index.vue");
    const admin = readPluginFile("frontend/pages/admin-list/index.vue");
    const types = readPluginFile("frontend/types/marketplace.ts");
    assert.match(detail, /preferences\.app\.locale/);
    assert.match(detail, /marketplaceReleaseDocumentBundle/);
    assert.match(detail, /availableDocuments/);
    assert.match(detail, /availableDocumentLocaleOptions/);
    assert.match(detail, /handleSelectDocumentLocale/);
    assert.match(detail, /Segmented/);
    assert.match(detail, /chooseDocumentFromBundle/);
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
