import assert from "node:assert/strict";
import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { describe, it } from "node:test";
import { fileURLToPath } from "node:url";

const testDir = path.dirname(fileURLToPath(import.meta.url));
const pluginRoot = path.resolve(testDir, "../../..");
const componentPrefix = "plugins.linapro_plugin_marketplace.api.market.v1";

const exactlyTranslatedDtosByFile = {
  "backend/api/market/v1/market_catalog.go": [
    "MyPluginDetailReq",
    "MyPluginDetailRes",
    "ManagedPluginDetailReq",
    "ManagedPluginDetailRes",
    "MyReleaseListReq",
    "MyReleaseListRes",
    "ManagedReleaseListReq",
    "ManagedReleaseListRes",
  ],
  "backend/api/market/v1/market_docs.go": [
    "MyReleaseDocsReq",
    "MyReleaseDocsRes",
    "ManagedReleaseDocsReq",
    "ManagedReleaseDocsRes",
    "MyReleaseRisksReq",
    "MyReleaseRisksRes",
    "ManagedReleaseRisksReq",
    "ManagedReleaseRisksRes",
  ],
};

const commonFallbackSuffixes = [
  "Res.schema.dc",
  ".fields.pageNum.dc",
  ".fields.pageSize.dc",
  ".fields.total.dc",
  ".fields.createdAt.dc",
  ".fields.updatedAt.dc",
  ".fields.deletedAt.dc",
];

function readPluginFile(relativePath) {
  return readFileSync(path.join(pluginRoot, relativePath), "utf8");
}

function structTagValue(tag, name) {
  return tag.match(new RegExp(`(?:^|\\s)${name}:"([^"]*)"`))?.[1] ?? "";
}

function collectStructTranslationKeys(typeName, structBody) {
  const componentKey = `${componentPrefix}.${typeName}`;
  const keys = [];
  for (const match of structBody.matchAll(/`([^`]+)`/g)) {
    const tag = match[1];
    if (structTagValue(tag, "tags")) {
      keys.push(`${componentKey}.meta.tags`);
    }
    if (structTagValue(tag, "summary")) {
      keys.push(`${componentKey}.meta.summary`);
    }
    if (structTagValue(tag, "path") && structTagValue(tag, "dc")) {
      keys.push(`${componentKey}.meta.dc`, `${componentKey}.schema.dc`);
    }

    const jsonName = structTagValue(tag, "json").split(",")[0];
    if (jsonName && jsonName !== "-" && structTagValue(tag, "dc")) {
      keys.push(`${componentKey}.fields.${jsonName}.dc`);
    }
  }
  return keys;
}

function collectFileTranslationKeys(source) {
  const keys = [];
  for (const match of source.matchAll(
    /type\s+(\w+)\s+struct\s*\{[ \t]*\n([\s\S]*?)\n\}/g,
  )) {
    keys.push(...collectStructTranslationKeys(match[1], match[2]));
  }
  return keys;
}

function collectDtoTranslationKeys(source, typeName) {
  const block = source.match(
    new RegExp(
      `type\\s+${typeName}\\s+struct\\s*\\{[ \\t]*\\n([\\s\\S]*?)\\n\\}`,
    ),
  );
  assert.ok(block, `missing governed DTO ${typeName}`);
  return collectStructTranslationKeys(typeName, block[1]);
}

function listApiSourceFiles() {
  const apiRoot = path.join(pluginRoot, "backend", "api");
  const files = [];
  const visit = (dir) => {
    for (const entry of readdirSync(dir, { withFileTypes: true })) {
      const current = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        visit(current);
      } else if (entry.isFile() && entry.name.endsWith(".go")) {
        files.push(current);
      }
    }
  };
  visit(apiRoot);
  return files.sort();
}

function hasCommonFallback(key) {
  return commonFallbackSuffixes.some((suffix) => key.endsWith(suffix));
}

describe("marketplace apidoc i18n", () => {
  it("keeps English API documentation sourced from Go DTO metadata", () => {
    const en = JSON.parse(
      readPluginFile("manifest/i18n/en-US/apidoc/plugin-marketplace.json"),
    );
    assert.deepEqual(en, {});
  });

  it("covers all marketplace API DTO metadata", () => {
    const zh = JSON.parse(
      readPluginFile("manifest/i18n/zh-CN/apidoc/plugin-marketplace.json"),
    );
    const requiredKeys = listApiSourceFiles().flatMap((filePath) =>
      collectFileTranslationKeys(readFileSync(filePath, "utf8")),
    );

    assert.ok(requiredKeys.length > 300, "expected marketplace API DTO keys");
    assert.equal(new Set(requiredKeys).size, requiredKeys.length);
    for (const key of requiredKeys) {
      if (hasCommonFallback(key)) {
        continue;
      }
      assert.equal(typeof zh[key], "string", `missing zh-CN key ${key}`);
      assert.match(
        zh[key],
        /[\u3400-\u9fff]/,
        `expected Chinese text for ${key}`,
      );
    }
  });

  it("keeps My and Managed detail, release, docs, and risks keys exact", () => {
    const zh = JSON.parse(
      readPluginFile("manifest/i18n/zh-CN/apidoc/plugin-marketplace.json"),
    );
    const requiredKeys = [];

    for (const [relativePath, typeNames] of Object.entries(
      exactlyTranslatedDtosByFile,
    )) {
      const source = readPluginFile(relativePath);
      for (const typeName of typeNames) {
        requiredKeys.push(...collectDtoTranslationKeys(source, typeName));
      }
    }

    assert.equal(new Set(requiredKeys).size, requiredKeys.length);
    assert.equal(requiredKeys.length, 76);
    for (const key of requiredKeys) {
      assert.equal(typeof zh[key], "string", `missing zh-CN key ${key}`);
      assert.match(
        zh[key],
        /[\u3400-\u9fff]/,
        `expected Chinese text for ${key}`,
      );
    }
  });
});
