import { Buffer } from "node:buffer";
import { existsSync, mkdirSync, readFileSync } from "node:fs";
import { join } from "node:path";

import type { Page, Route } from "@host-tests/fixtures/auth";

import { workspacePath } from "@host-tests/fixtures/config";
import { MainLayout } from "@host-tests/pages/MainLayout";
import { waitForRouteReady } from "@host-tests/support/ui";

type ArtifactType = "dynamic_zip" | "plugin_wasm" | "source_zip";
type MarketplaceLocale = "en-US" | "zh-CN";
type PluginType = "dynamic" | "source";
type ReviewStatus =
  | "approved"
  | "draft"
  | "rejected"
  | "reviewing"
  | "submitted";
type RiskSeverity = "high" | "info" | "warning";
type RiskType =
  | "docs"
  | "dynamic_route"
  | "host_service"
  | "install_sql"
  | "menu_permission";
type Visibility = "private" | "public" | "reserved";

type PublisherItem = {
  contactEmail?: string;
  homepage?: string;
  name: string;
  publisherKey: string;
  summary?: string;
  verified: boolean;
};

type ArtifactItem = {
  artifactType: ArtifactType;
  contentType: string;
  fileName: string;
  manifestSha256?: string;
  sha256: string;
  sizeBytes: number;
  wasmSha256?: string;
};

type ReleaseItem = {
  artifact?: ArtifactItem;
  maxHostVersion?: string;
  minHostVersion?: string;
  pluginId: string;
  pluginType: PluginType;
  publishedAt?: number;
  releaseStatus: "draft" | "published";
  reviewMessage?: string;
  reviewStatus: ReviewStatus;
  reviewedAt?: number;
  submittedAt?: number;
  updatedAt?: number;
  version: string;
  visibility: Visibility;
};

type PluginItem = {
  description?: string;
  downloadCount: number;
  homepage?: string;
  latestRelease?: ReleaseItem;
  latestReviewStatus?: ReviewStatus;
  latestVersion: string;
  license?: string;
  marketStatus: "draft" | "published";
  maxHostVersion?: string;
  minHostVersion?: string;
  name: string;
  pluginId: string;
  pluginType: PluginType;
  primaryTag?: string;
  publishedAt?: number;
  publisher?: PublisherItem;
  repository?: string;
  riskCounts: {
    high: number;
    info: number;
    warning: number;
  };
  sourceDelivery: "dynamic_upload_required" | "source_rebuild_required";
  summary: string;
  tagCodes: string[];
  tags: Array<{
    code: string;
    name: string;
    type: string;
  }>;
  updatedAt?: number;
  visibility: Visibility;
};

type DocumentItem = {
  content: string;
  contentHash: string;
  fallbackUsed: boolean;
  locale: string;
  path: string;
  pluginId: string;
  resolvedLocale: string;
  sourceKind: string;
  summary: string;
  title: string;
  updatedAt?: number;
  version: string;
};

type RiskItem = {
  createdAt?: number;
  payload: Record<string, unknown>;
  severity: RiskSeverity;
  source: string;
  summary: string;
  type: RiskType;
};

type MockOptions = {
  failDownloadsFor?: string[];
  includePrivatePlugins?: boolean;
  inspectionDelayMsByRelease?: Record<string, number>;
  menuRole?: "both" | "publish-only" | "review-only";
};

type MarketplaceMockData = {
  pluginsById: Map<string, PluginItem>;
  releasesByPlugin: Map<string, ReleaseItem[]>;
};

export type MarketplaceMockState = {
  deniedRequests: Array<{ method: string; path: string }>;
  downloadRequests: Array<{ pluginId: string; version: string }>;
  dynamicUploadRequests: number;
  inspectionResponses: Array<{
    kind: "docs" | "risks";
    pluginId: string;
    version: string;
  }>;
  listRequests: URLSearchParams[];
  releaseSnapshot: (
    pluginId: string,
    version: string,
  ) => {
    releaseStatus: "draft" | "published";
    reviewMessage?: string;
    reviewStatus: ReviewStatus;
  };
  reviewQueueRequests: URLSearchParams[];
  reviewRequests: Array<{
    message: string;
    pluginId: string;
    releaseStatus: "draft" | "published";
    status: string;
    version: string;
  }>;
  uploadRequests: Array<{
    pluginId: string;
    pluginType: PluginType;
    version: string;
  }>;
};

const mockNow = Date.UTC(2026, 6, 10, 8, 0, 0);
const pluginApiPrefix = "/x/linapro-plugin-marketplace/api/v1/";
const marketplacePluginId = "linapro-plugin-marketplace";
const sourcePluginId = "linapro-demo-source";
const dynamicPluginId = "linapro-demo-dynamic";
const privatePluginId = "linapro-private-reports";
const externalPluginId = "acme-observability";
const runtimeMessagesByLocale: Record<
  MarketplaceLocale,
  Record<string, string>
> = {
  "en-US": readRuntimeMessages("en-US"),
  "zh-CN": readRuntimeMessages("zh-CN"),
};

const linaPublisher: PublisherItem = {
  contactEmail: "plugins@linapro.ai",
  homepage: "https://linapro.ai",
  name: "LinaPro",
  publisherKey: "linapro",
  summary: "Official LinaPro publisher",
  verified: true,
};

const externalPublisher: PublisherItem = {
  contactEmail: "plugins@example.com",
  homepage: "https://example.com/acme",
  name: "Acme Labs",
  publisherKey: "acme-labs",
  summary: "Independent marketplace publisher",
  verified: true,
};

const sourceRelease: ReleaseItem = {
  artifact: {
    artifactType: "source_zip",
    contentType: "application/zip",
    fileName: "linapro-demo-source-v1.0.0.zip",
    manifestSha256: "source-manifest-sha256",
    sha256: "source-package-sha256",
    sizeBytes: 49152,
  },
  maxHostVersion: "v1.4.0",
  minHostVersion: "v1.2.0",
  pluginId: sourcePluginId,
  pluginType: "source",
  publishedAt: mockNow,
  releaseStatus: "published",
  reviewMessage: "Approved by marketplace review.",
  reviewStatus: "approved",
  reviewedAt: mockNow,
  submittedAt: mockNow - 300000,
  updatedAt: mockNow,
  version: "v1.0.0",
  visibility: "public",
};

const dynamicRelease: ReleaseItem = {
  artifact: {
    artifactType: "dynamic_zip",
    contentType: "application/zip",
    fileName: "linapro-demo-dynamic-v2.0.0.zip",
    manifestSha256: "dynamic-manifest-sha256",
    sha256: "dynamic-package-sha256",
    sizeBytes: 32768,
    wasmSha256: "dynamic-wasm-sha256",
  },
  minHostVersion: "v1.2.0",
  pluginId: dynamicPluginId,
  pluginType: "dynamic",
  publishedAt: mockNow,
  releaseStatus: "published",
  reviewMessage: "Approved with host service review.",
  reviewStatus: "approved",
  reviewedAt: mockNow,
  submittedAt: mockNow - 240000,
  updatedAt: mockNow,
  version: "v2.0.0",
  visibility: "public",
};

const dynamicPendingRelease: ReleaseItem = {
  ...dynamicRelease,
  artifact: {
    ...dynamicRelease.artifact!,
    fileName: "linapro-demo-dynamic-v2.1.0.zip",
    sha256: "dynamic-pending-package-sha256",
  },
  publishedAt: undefined,
  releaseStatus: "draft",
  reviewMessage: "Ready for marketplace review.",
  reviewStatus: "submitted",
  reviewedAt: undefined,
  submittedAt: mockNow + 60_000,
  updatedAt: mockNow + 60_000,
  version: "v2.1.0",
};

const privateRelease: ReleaseItem = {
  artifact: {
    artifactType: "source_zip",
    contentType: "application/zip",
    fileName: "linapro-private-reports-v1.0.0.zip",
    sha256: "private-package-sha256",
    sizeBytes: 24576,
  },
  pluginId: privatePluginId,
  pluginType: "source",
  publishedAt: mockNow,
  releaseStatus: "published",
  reviewStatus: "approved",
  updatedAt: mockNow,
  version: "v1.0.0",
  visibility: "private",
};

const externalPendingRelease: ReleaseItem = {
  artifact: {
    artifactType: "source_zip",
    contentType: "application/zip",
    fileName: "acme-observability-v0.9.0.zip",
    sha256: "acme-observability-package-sha256",
    sizeBytes: 28672,
  },
  pluginId: externalPluginId,
  pluginType: "source",
  releaseStatus: "draft",
  reviewMessage: "Awaiting marketplace policy review.",
  reviewStatus: "reviewing",
  submittedAt: mockNow + 30_000,
  updatedAt: mockNow + 30_000,
  version: "v0.9.0",
  visibility: "public",
};

const plugins: PluginItem[] = [
  {
    description: "Source package delivery demo with docs and SQL summaries.",
    downloadCount: 128,
    homepage: "https://linapro.ai/plugins/source-demo",
    latestRelease: sourceRelease,
    latestReviewStatus: sourceRelease.reviewStatus,
    latestVersion: sourceRelease.version,
    license: "Apache-2.0",
    marketStatus: "published",
    maxHostVersion: sourceRelease.maxHostVersion,
    minHostVersion: sourceRelease.minHostVersion,
    name: "LinaPro Source Demo",
    pluginId: sourcePluginId,
    pluginType: "source",
    primaryTag: "official",
    publishedAt: mockNow,
    publisher: linaPublisher,
    repository: "https://github.com/linaproai/linapro",
    riskCounts: {
      high: 1,
      info: 3,
      warning: 2,
    },
    sourceDelivery: "source_rebuild_required",
    summary: "Source plugin package for rebuild-based delivery.",
    tagCodes: ["official", "source"],
    tags: [
      { code: "official", name: "Official", type: "category" },
      { code: "source", name: "Source", type: "plugin-type" },
    ],
    updatedAt: mockNow,
    visibility: "public",
  },
  {
    description:
      "Dynamic runtime package that must pass local host governance.",
    downloadCount: 64,
    latestRelease: dynamicRelease,
    latestReviewStatus: dynamicRelease.reviewStatus,
    latestVersion: dynamicRelease.version,
    license: "Apache-2.0",
    marketStatus: "published",
    minHostVersion: dynamicRelease.minHostVersion,
    name: "LinaPro Dynamic Demo",
    pluginId: dynamicPluginId,
    pluginType: "dynamic",
    primaryTag: "runtime",
    publishedAt: mockNow,
    publisher: linaPublisher,
    riskCounts: {
      high: 0,
      info: 1,
      warning: 1,
    },
    sourceDelivery: "dynamic_upload_required",
    summary: "Dynamic plugin package for runtime upload governance.",
    tagCodes: ["official", "dynamic"],
    tags: [
      { code: "official", name: "Official", type: "category" },
      { code: "dynamic", name: "Dynamic", type: "plugin-type" },
    ],
    updatedAt: mockNow,
    visibility: "public",
  },
  {
    description: "Private report automation plugin hidden from public catalog.",
    downloadCount: 5,
    latestRelease: privateRelease,
    latestReviewStatus: privateRelease.reviewStatus,
    latestVersion: privateRelease.version,
    marketStatus: "published",
    name: "Private Report Automation",
    pluginId: privatePluginId,
    pluginType: "source",
    publisher: linaPublisher,
    riskCounts: {
      high: 0,
      info: 0,
      warning: 1,
    },
    sourceDelivery: "source_rebuild_required",
    summary: "Private tenant-scoped reporting plugin.",
    tagCodes: ["private"],
    tags: [{ code: "private", name: "Private", type: "visibility" }],
    updatedAt: mockNow,
    visibility: "private",
  },
  {
    description: "Independent operations plugin awaiting marketplace review.",
    downloadCount: 0,
    homepage: "https://example.com/acme/observability",
    latestRelease: externalPendingRelease,
    latestReviewStatus: externalPendingRelease.reviewStatus,
    latestVersion: externalPendingRelease.version,
    license: "MIT",
    marketStatus: "draft",
    name: "Acme Observability",
    pluginId: externalPluginId,
    pluginType: "source",
    publisher: externalPublisher,
    riskCounts: {
      high: 1,
      info: 0,
      warning: 0,
    },
    sourceDelivery: "source_rebuild_required",
    summary: "Operations telemetry plugin from an independent publisher.",
    tagCodes: ["observability"],
    tags: [{ code: "observability", name: "Observability", type: "category" }],
    updatedAt: mockNow + 30_000,
    visibility: "public",
  },
];

const baseReleasesByPlugin = new Map<string, ReleaseItem[]>([
  [sourcePluginId, [sourceRelease]],
  [dynamicPluginId, [dynamicPendingRelease, dynamicRelease]],
  [privatePluginId, [privateRelease]],
  [externalPluginId, [externalPendingRelease]],
]);

const documents = new Map<string, DocumentItem>([
  [
    releaseKey(sourcePluginId, sourceRelease.version),
    {
      content:
        "<h2>Source Demo Guide</h2><p>Place the source under apps/lina-plugins and rebuild the host before enabling it.</p>",
      contentHash: "source-doc-hash",
      fallbackUsed: false,
      locale: "en-US",
      path: "index.md",
      pluginId: sourcePluginId,
      resolvedLocale: "en-US",
      sourceKind: "manifest_docs",
      summary: "Source package delivery documentation.",
      title: "Source Demo Guide",
      updatedAt: mockNow,
      version: sourceRelease.version,
    },
  ],
  [
    releaseKey(dynamicPluginId, dynamicRelease.version),
    {
      content:
        "<h2>Dynamic Runtime Guide</h2><p>Download the runtime package and let the local host validate plugin.wasm.</p>",
      contentHash: "dynamic-doc-hash",
      fallbackUsed: false,
      locale: "en-US",
      path: "index.md",
      pluginId: dynamicPluginId,
      resolvedLocale: "en-US",
      sourceKind: "manifest_docs",
      summary: "Dynamic runtime import documentation.",
      title: "Dynamic Runtime Guide",
      updatedAt: mockNow,
      version: dynamicRelease.version,
    },
  ],
  [
    releaseKey(dynamicPluginId, dynamicPendingRelease.version),
    {
      content:
        "<h2>Dynamic Runtime Review Guide</h2><p>Review host service and route findings before approval.</p>",
      contentHash: "dynamic-pending-doc-hash",
      fallbackUsed: false,
      locale: "en-US",
      path: "index.md",
      pluginId: dynamicPluginId,
      resolvedLocale: "en-US",
      sourceKind: "manifest_docs",
      summary: "Dynamic runtime review documentation.",
      title: "Dynamic Runtime Review Guide",
      updatedAt: mockNow + 60_000,
      version: dynamicPendingRelease.version,
    },
  ],
  [
    releaseKey(externalPluginId, externalPendingRelease.version),
    {
      content:
        "<h2>Acme Observability Review Guide</h2><p>Confirm menu permissions before publishing this package.</p>",
      contentHash: "acme-observability-doc-hash",
      fallbackUsed: false,
      locale: "en-US",
      path: "index.md",
      pluginId: externalPluginId,
      resolvedLocale: "en-US",
      sourceKind: "manifest_docs",
      summary: "Independent publisher review documentation.",
      title: "Acme Observability Review Guide",
      updatedAt: mockNow + 30_000,
      version: externalPendingRelease.version,
    },
  ],
]);

const risks = new Map<string, RiskItem[]>([
  [
    releaseKey(sourcePluginId, sourceRelease.version),
    [
      {
        createdAt: mockNow,
        payload: { table: "plugin_marketplace_demo" },
        severity: "high",
        source: "manifest/sql/001-demo.sql",
        summary: "Install SQL creates plugin-owned demo tables.",
        type: "install_sql",
      },
      {
        createdAt: mockNow,
        payload: { docs: "index.md" },
        severity: "info",
        source: "manifest/docs/en-US/index.md",
        summary: "Documentation entry was indexed for marketplace display.",
        type: "docs",
      },
    ],
  ],
  [
    releaseKey(dynamicPluginId, dynamicRelease.version),
    [
      {
        createdAt: mockNow,
        payload: { services: ["host:storage"] },
        severity: "warning",
        source: "plugin.wasm",
        summary: "Requests host service access for local runtime governance.",
        type: "host_service",
      },
      {
        createdAt: mockNow,
        payload: { route: "/dynamic-demo" },
        severity: "info",
        source: "plugin.yaml",
        summary: "Registers a dynamic plugin route.",
        type: "dynamic_route",
      },
    ],
  ],
  [
    releaseKey(dynamicPluginId, dynamicPendingRelease.version),
    [
      {
        createdAt: mockNow + 60_000,
        payload: { services: ["host:storage"] },
        severity: "warning",
        source: "plugin.wasm",
        summary: "Requests host service access for reviewer approval.",
        type: "host_service",
      },
      {
        createdAt: mockNow + 60_000,
        payload: { route: "/dynamic-demo" },
        severity: "info",
        source: "plugin.yaml",
        summary: "Registers a dynamic route for reviewer inspection.",
        type: "dynamic_route",
      },
    ],
  ],
  [
    releaseKey(externalPluginId, externalPendingRelease.version),
    [
      {
        createdAt: mockNow + 30_000,
        payload: { permission: "acme-observability:dashboard:view" },
        severity: "high",
        source: "plugin.yaml",
        summary: "Requests a menu permission that requires reviewer approval.",
        type: "menu_permission",
      },
    ],
  ],
]);

export async function installMarketplaceApiMocks(
  page: Page,
  options: MockOptions = {},
): Promise<MarketplaceMockState> {
  const data = createMarketplaceMockData();
  const state: MarketplaceMockState = {
    deniedRequests: [],
    downloadRequests: [],
    dynamicUploadRequests: 0,
    inspectionResponses: [],
    listRequests: [],
    releaseSnapshot(pluginId, version) {
      const release = findRelease(data, pluginId, version);
      return {
        releaseStatus: release.releaseStatus,
        reviewMessage: release.reviewMessage,
        reviewStatus: release.reviewStatus,
      };
    },
    reviewQueueRequests: [],
    reviewRequests: [],
    uploadRequests: [],
  };
  const failedDownloads = new Set(options.failDownloadsFor ?? []);

  await page.route("**/api/v1/i18n/runtime/locales**", async (route) => {
    await fulfillData(route, runtimeLocaleOptions(resolveRequestLocale(route)));
  });

  await page.route("**/api/v1/i18n/runtime/messages**", async (route) => {
    await fulfillData(route, {
      messages: runtimeMessagesByLocale[resolveRequestLocale(route)],
    });
  });

  await page.route("**/api/v1/config/public/frontend", async (route) => {
    await fulfillData(route, publicFrontendSettings());
  });

  await page.route("**/api/v1/menus/all", async (route) => {
    await fulfillData(route, {
      list: marketplaceMenuRoutes(
        resolveRequestLocale(route),
        options.menuRole ?? "both",
      ),
    });
  });

  await page.route("**/api/v1/plugins/dynamic", async (route) => {
    await fulfillData(route, { list: [] });
  });

  await page.route("**/x/linapro-plugin-marketplace/**", async (route) => {
    await handleMarketplaceRoute(route, state, data, {
      failedDownloads,
      includePrivatePlugins: options.includePrivatePlugins === true,
      inspectionDelayMsByRelease: options.inspectionDelayMsByRelease ?? {},
      menuRole: options.menuRole ?? "both",
    });
  });

  await page.route("**/api/v1/plugins/dynamic/package", async (route) => {
    state.dynamicUploadRequests += 1;
    await fulfillData(route, {
      enabled: 0,
      installed: 0,
      name: "LinaPro Dynamic Demo",
      pluginId: dynamicPluginId,
      runtimeState: "uploaded",
      version: dynamicRelease.version,
    });
  });

  return state;
}

export async function openMarketplaceWorkbench(
  page: Page,
  language?: "English" | "简体中文",
) {
  // Prefer a host shell that exists even when dashboard analytics is disabled
  // or returns 404 in local environments.
  await page.goto(workspacePath("/"), {
    waitUntil: "domcontentloaded",
  });
  await waitForRouteReady(page, 15_000);

  const mainLayout = new MainLayout(page);
  if (language) {
    const languageToggle = page.getByTestId("language-toggle-trigger").first();
    if (await languageToggle.isVisible().catch(() => false)) {
      await mainLayout.switchLanguage(language);
      await page.reload({ waitUntil: "domcontentloaded" });
      await waitForRouteReady(page, 15_000);
    }
  }
  return mainLayout;
}

export function marketplaceSourcePluginId() {
  return sourcePluginId;
}

export function marketplaceDynamicPluginId() {
  return dynamicPluginId;
}

export function marketplacePrivatePluginId() {
  return privatePluginId;
}

export function marketplaceExternalPluginId() {
  return externalPluginId;
}

export async function callMarketplaceApiFromPage(
  page: Page,
  path: string,
  options: {
    body?: Record<string, unknown>;
    method?: "GET" | "PUT";
  } = {},
) {
  return await page.evaluate(
    async ({ body, method, url }) => {
      const response = await fetch(url, {
        body: body ? JSON.stringify(body) : undefined,
        headers: body ? { "content-type": "application/json" } : undefined,
        method,
      });
      return (await response.json()) as {
        code: number;
        data: unknown;
        messageKey?: string;
      };
    },
    {
      body: options.body,
      method: options.method ?? "GET",
      url: `${pluginApiPrefix}${path}`,
    },
  );
}

export function sourceMarketplaceZipUpload() {
  return {
    buffer: Buffer.from("source marketplace zip placeholder"),
    mimeType: "application/zip",
    name: "linapro-demo-source-v1.2.0.zip",
  };
}

export function dynamicMarketplaceZipUpload() {
  return {
    buffer: Buffer.from("dynamic marketplace zip placeholder"),
    mimeType: "application/zip",
    name: "linapro-demo-dynamic-v2.1.0.zip",
  };
}

export async function captureMarketplaceScreenshot(page: Page, label: string) {
  await page
    .locator(".ant-message-loading:visible")
    .first()
    .waitFor({ state: "hidden", timeout: 5000 })
    .catch(() => {});
  const now = new Date();
  const date = formatDatePart(now);
  const time = [now.getHours(), now.getMinutes(), now.getSeconds()]
    .map((value) => String(value).padStart(2, "0"))
    .join("");
  const screenshotDirectory = join(findRepositoryRoot(), "temp", date);
  mkdirSync(screenshotDirectory, { recursive: true });
  await page.screenshot({
    animations: "disabled",
    fullPage: false,
    path: join(screenshotDirectory, `${time}-${label}.png`),
  });
}

async function handleMarketplaceRoute(
  route: Route,
  state: MarketplaceMockState,
  data: MarketplaceMockData,
  options: {
    failedDownloads: Set<string>;
    includePrivatePlugins: boolean;
    inspectionDelayMsByRelease: Record<string, number>;
    menuRole: "both" | "publish-only" | "review-only";
  },
) {
  const request = route.request();
  const requestURL = new URL(request.url());
  const pathName = requestURL.pathname;

  if (!pathName.includes(pluginApiPrefix)) {
    if (pathName.includes("/market/download-sessions/")) {
      await fulfillBinary(route, dynamicZipWithPluginWasm());
      return;
    }
    await route.continue();
    return;
  }

  const apiPath = pathName.slice(
    pathName.indexOf(pluginApiPrefix) + pluginApiPrefix.length,
  );
  const segments = apiPath.split("/").filter(Boolean).map(decodeURIComponent);
  const method = request.method();

  if (
    options.menuRole === "publish-only" &&
    isManagedOrReviewRequest(method, apiPath, segments)
  ) {
    state.deniedRequests.push({ method, path: apiPath });
    await fulfillError(route, "error.plugin.marketplace.permission.denied");
    return;
  }

  if (
    method === "GET" &&
    (apiPath === "market/plugins" ||
      apiPath === "market/my-plugins" ||
      apiPath === "market/managed-plugins")
  ) {
    state.listRequests.push(new URLSearchParams(requestURL.searchParams));
    await fulfillData(
      route,
      listPlugins(data, apiPath, requestURL.searchParams, options),
    );
    return;
  }

  if (method === "GET" && apiPath === "market/review-queue") {
    state.reviewQueueRequests.push(
      new URLSearchParams(requestURL.searchParams),
    );
    const pluginId = requestURL.searchParams.get("pluginId")?.trim() || "";
    const keyword = requestURL.searchParams
      .get("keyword")
      ?.trim()
      .toLowerCase();
    const requestedStatus = requestURL.searchParams.get("reviewStatus")?.trim();
    const releases = [...data.releasesByPlugin.values()]
      .flat()
      .filter((item) => {
        if (pluginId && item.pluginId !== pluginId) {
          return false;
        }
        if (requestedStatus) {
          return item.reviewStatus === requestedStatus;
        }
        return ["reviewing", "submitted"].includes(item.reviewStatus);
      })
      .map((item) => ({
        ...item,
        pluginName: data.pluginsById.get(item.pluginId)?.name || item.pluginId,
      }))
      .filter((item) => {
        if (!keyword) {
          return true;
        }
        return `${item.pluginId} ${item.pluginName} ${item.version}`
          .toLowerCase()
          .includes(keyword);
      });
    await fulfillData(route, pageResult(releases));
    return;
  }

  if (method === "GET" && apiPath === "market/publishers") {
    await fulfillData(route, pageResult([linaPublisher]));
    return;
  }

  if (method === "POST" && apiPath === "market/publishers") {
    await fulfillData(route, {
      publisher: linaPublisher,
    });
    return;
  }

  if (
    method === "PUT" &&
    segments.length === 3 &&
    segments[0] === "market" &&
    segments[1] === "publishers"
  ) {
    const payload = postDataObject(request.postDataJSON());
    await fulfillData(route, {
      publisher: {
        ...linaPublisher,
        contactEmail:
          typeof payload.contactEmail === "string"
            ? payload.contactEmail
            : linaPublisher.contactEmail,
        homepage:
          typeof payload.homepage === "string"
            ? payload.homepage
            : linaPublisher.homepage,
        name:
          typeof payload.name === "string" ? payload.name : linaPublisher.name,
        publisherKey:
          typeof payload.publisherKey === "string"
            ? payload.publisherKey
            : (segments[2] ?? linaPublisher.publisherKey),
        summary:
          typeof payload.summary === "string"
            ? payload.summary
            : linaPublisher.summary,
      },
    });
    return;
  }

  if (method === "POST" && apiPath === "market/plugins") {
    const payload = postDataObject(request.postDataJSON());
    const plugin = pluginFromPayload(payload);
    data.pluginsById.set(plugin.pluginId, plugin);
    await fulfillData(route, { plugin });
    return;
  }

  if (method === "POST" && apiPath === "market/my-plugins/packages") {
    const bodyText = request.postData() ?? "";
    const fileName = multipartFilename(bodyText, "file");
    const inferredPluginId = pluginIdFromPackageFileName(fileName);
    const release = releaseFromUpload(inferredPluginId, bodyText);
    if (!multipartField(bodyText, "version")) {
      // First add of the E2E private plugin uses v1.0.0; later packages keep
      // the version encoded in the mock file name when present.
      release.version = versionFromPackageFileName(fileName) || "v1.0.0";
    }
    release.pluginId = inferredPluginId;
    const existing = data.pluginsById.get(inferredPluginId);
    const plugin: PluginItem = {
      description: existing?.description ?? "",
      downloadCount: existing?.downloadCount ?? 0,
      latestRelease: cloneRelease(release),
      latestReviewStatus: release.reviewStatus,
      latestVersion: release.version,
      license: existing?.license ?? "",
      marketStatus: existing?.marketStatus ?? "draft",
      name: existing?.name ?? inferredPluginId,
      pluginId: inferredPluginId,
      pluginType: release.pluginType,
      publisher: existing?.publisher ?? linaPublisher,
      riskCounts: existing?.riskCounts ?? { high: 0, info: 0, warning: 0 },
      sourceDelivery:
        release.pluginType === "dynamic"
          ? "dynamic_upload_required"
          : "source_rebuild_required",
      summary: existing?.summary ?? "Auto-parsed marketplace package draft.",
      tagCodes: existing?.tagCodes ?? [],
      tags: existing?.tags ?? [],
      updatedAt: mockNow + 60000,
      visibility: existing?.visibility ?? "private",
    };
    data.pluginsById.set(plugin.pluginId, plugin);
    state.uploadRequests.push({
      pluginId: release.pluginId,
      pluginType: release.pluginType,
      version: release.version,
    });
    addRelease(data, release);
    await fulfillData(route, { plugin, release });
    return;
  }

  if (
    segments.length === 4 &&
    method === "POST" &&
    segments[0] === "market" &&
    segments[1] === "my-plugins" &&
    segments[3] === "publish"
  ) {
    const pluginId = segments[2] ?? "";
    const plugin = data.pluginsById.get(pluginId);
    const version = plugin?.latestVersion || "v1.0.0";
    const release = updateRelease(data, pluginId, version, {
      reviewStatus: "submitted",
      submittedAt: mockNow + 60000,
    });
    updatePluginLatestRelease(data, release);
    await fulfillData(route, { release });
    return;
  }

  if (
    segments.length === 4 &&
    method === "POST" &&
    segments[0] === "market" &&
    segments[1] === "my-plugins" &&
    segments[3] === "delist"
  ) {
    const pluginId = segments[2] ?? "";
    const plugin = data.pluginsById.get(pluginId);
    if (plugin) {
      const version = plugin.latestVersion || "v1.0.0";
      const release = updateRelease(data, pluginId, version, {
        releaseStatus: "delisted",
      });
      data.pluginsById.set(pluginId, {
        ...plugin,
        latestRelease: cloneRelease(release),
        latestReviewStatus: release.reviewStatus,
        latestVersion: release.version,
        marketStatus: "delisted",
        updatedAt: mockNow + 120000,
        visibility: "private",
      });
      await fulfillData(route, {
        plugin: data.pluginsById.get(pluginId),
      });
      return;
    }
  }

  if (
    segments.length === 3 &&
    method === "GET" &&
    segments[0] === "market" &&
    isMarketplaceReadCollection(segments[1])
  ) {
    await fulfillData(route, {
      plugin: pluginDetail(data, segments[2], {
        includePrivatePlugins:
          options.includePrivatePlugins || segments[1] !== "plugins",
      }),
    });
    return;
  }

  if (
    segments.length === 4 &&
    method === "GET" &&
    segments[0] === "market" &&
    isMarketplaceReadCollection(segments[1]) &&
    segments[3] === "releases"
  ) {
    await fulfillData(route, pageResult(releaseList(data, segments[2])));
    return;
  }

  if (
    segments.length === 4 &&
    method === "POST" &&
    segments[0] === "market" &&
    segments[1] === "plugins" &&
    segments[3] === "releases"
  ) {
    const release = releaseFromUpload(segments[2], request.postData() ?? "");
    state.uploadRequests.push({
      pluginId: release.pluginId,
      pluginType: release.pluginType,
      version: release.version,
    });
    addRelease(data, release);
    updatePluginLatestRelease(data, release);
    await fulfillData(route, { release });
    return;
  }

  if (
    segments.length === 6 &&
    method === "POST" &&
    segments[0] === "market" &&
    segments[1] === "plugins" &&
    segments[3] === "releases" &&
    segments[5] === "submit-review"
  ) {
    const release = updateRelease(data, segments[2], segments[4], {
      reviewStatus: "submitted",
      submittedAt: mockNow + 60000,
    });
    updatePluginLatestRelease(data, release);
    await fulfillData(route, { release });
    return;
  }

  if (
    segments.length === 6 &&
    method === "PUT" &&
    segments[0] === "market" &&
    segments[1] === "plugins" &&
    segments[3] === "releases" &&
    segments[5] === "review"
  ) {
    const payload = postDataObject(request.postDataJSON());
    const status = stringValue(payload.reviewStatus) || "approved";
    const release = updateRelease(data, segments[2], segments[4], {
      releaseStatus: status === "approved" ? "published" : "draft",
      reviewMessage: stringValue(payload.message) || "Reviewed by E2E mock.",
      reviewStatus: status === "rejected" ? "rejected" : "approved",
      reviewedAt: mockNow + 120000,
    });
    state.reviewRequests.push({
      message: release.reviewMessage ?? "",
      pluginId: release.pluginId,
      releaseStatus: release.releaseStatus,
      status: release.reviewStatus,
      version: release.version,
    });
    updatePluginLatestRelease(data, release);
    await fulfillData(route, { release });
    return;
  }

  if (
    segments.length === 6 &&
    method === "GET" &&
    segments[0] === "market" &&
    isMarketplaceReadCollection(segments[1]) &&
    segments[3] === "releases" &&
    segments[5] === "docs"
  ) {
    await waitForInspectionDelay(
      options.inspectionDelayMsByRelease,
      segments[2],
      segments[4],
    );
    await fulfillData(route, {
      document: documents.get(releaseKey(segments[2], segments[4])) ?? null,
    });
    state.inspectionResponses.push({
      kind: "docs",
      pluginId: segments[2],
      version: segments[4],
    });
    return;
  }

  if (
    segments.length === 6 &&
    method === "GET" &&
    segments[0] === "market" &&
    isMarketplaceReadCollection(segments[1]) &&
    segments[3] === "releases" &&
    segments[5] === "risks"
  ) {
    await waitForInspectionDelay(
      options.inspectionDelayMsByRelease,
      segments[2],
      segments[4],
    );
    await fulfillData(
      route,
      pageResult(risks.get(releaseKey(segments[2], segments[4])) ?? []),
    );
    state.inspectionResponses.push({
      kind: "risks",
      pluginId: segments[2],
      version: segments[4],
    });
    return;
  }

  if (
    segments.length === 6 &&
    method === "POST" &&
    segments[0] === "market" &&
    segments[1] === "plugins" &&
    segments[3] === "releases" &&
    segments[5] === "downloads"
  ) {
    const pluginId = segments[2];
    const version = segments[4];
    state.downloadRequests.push({ pluginId, version });
    if (options.failedDownloads.has(pluginId)) {
      await fulfillError(
        route,
        "error.plugin.marketplace.download.session.not.found",
      );
      return;
    }
    const release = findRelease(data, pluginId, version);
    await fulfillData(route, {
      session: {
        artifactType: release.artifact?.artifactType ?? "source_zip",
        createdAt: mockNow,
        downloadUrl: `/x/linapro-plugin-marketplace/api/v1/market/download-sessions/e2e-${pluginId}/content`,
        expiresAt: mockNow + 900000,
        pluginId,
        sessionId: `e2e-${pluginId}`,
        sha256: release.artifact?.sha256 ?? "sha256",
        sizeBytes: release.artifact?.sizeBytes ?? 0,
        status: "active",
        version,
      },
    });
    return;
  }

  await route.continue();
}

function listPlugins(
  data: MarketplaceMockData,
  apiPath: string,
  searchParams: URLSearchParams,
  options: { includePrivatePlugins: boolean },
) {
  const keyword = searchParams.get("keyword")?.trim().toLowerCase() ?? "";
  const pluginType = searchParams.get("pluginType")?.trim() ?? "";
  const publisher = searchParams.get("publisher")?.trim().toLowerCase() ?? "";
  const status = searchParams.get("status")?.trim() ?? "";
  const tagCode = searchParams.get("tagCode")?.trim().toLowerCase() ?? "";
  const availablePlugins = [...data.pluginsById.values()].filter((item) => {
    if (apiPath === "market/my-plugins") {
      return item.publisher?.publisherKey === linaPublisher.publisherKey;
    }
    if (apiPath === "market/managed-plugins") {
      return true;
    }
    return options.includePrivatePlugins || item.visibility !== "private";
  });
  const filtered = availablePlugins.filter((item) => {
    if (keyword) {
      const haystack =
        `${item.pluginId} ${item.name} ${item.summary}`.toLowerCase();
      if (!haystack.includes(keyword)) {
        return false;
      }
    }
    if (pluginType && item.pluginType !== pluginType) {
      return false;
    }
    if (
      publisher &&
      !`${item.publisher?.publisherKey ?? ""} ${item.publisher?.name ?? ""}`
        .toLowerCase()
        .includes(publisher)
    ) {
      return false;
    }
    if (status && item.marketStatus !== status) {
      return false;
    }
    if (
      tagCode &&
      !item.tagCodes.some((code) => code.toLowerCase() === tagCode)
    ) {
      return false;
    }
    return true;
  });
  return pageResult(filtered);
}

function isMarketplaceReadCollection(value: string) {
  return ["managed-plugins", "my-plugins", "plugins"].includes(value);
}

function isManagedOrReviewRequest(
  method: string,
  apiPath: string,
  segments: string[],
) {
  if (apiPath === "market/review-queue") {
    return true;
  }
  if (segments[0] === "market" && segments[1] === "managed-plugins") {
    return true;
  }
  return (
    method === "PUT" &&
    segments[0] === "market" &&
    segments[1] === "plugins" &&
    segments[3] === "releases" &&
    segments[5] === "review"
  );
}

async function waitForInspectionDelay(
  delays: Record<string, number>,
  pluginId: string,
  version: string,
) {
  const delayMs = delays[releaseKey(pluginId, version)] ?? 0;
  if (delayMs > 0) {
    await new Promise((resolve) => setTimeout(resolve, delayMs));
  }
}

function publicFrontendSettings() {
  return {
    app: {
      logo: "",
      logoDark: "",
      name: "LinaPro",
    },
    auth: {
      loginSubtitle: "",
      panelLayout: "panel-right",
      pageDesc: "",
      pageTitle: "",
    },
    cron: {
      logRetention: {
        mode: "days",
        value: 30,
      },
      shell: {
        disabledReason: "",
        enabled: false,
        supported: true,
      },
      timezone: {
        current: "Asia/Shanghai",
      },
    },
    ui: {
      layout: "sidebar-mixed-nav",
      themeMode: "light",
      watermarkContent: "",
      watermarkEnabled: false,
    },
    user: {
      defaultAvatar: "",
    },
    workspace: {
      basePath: "/admin",
    },
  };
}

function findRepositoryRoot() {
  for (const root of [
    process.cwd(),
    join(process.cwd(), ".."),
    join(process.cwd(), "..", ".."),
  ]) {
    if (existsSync(join(root, "AGENTS.md"))) {
      return root;
    }
  }
  throw new Error("LinaPro repository root not found");
}

function formatDatePart(value: Date) {
  return [value.getFullYear(), value.getMonth() + 1, value.getDate()]
    .map((part) => String(part).padStart(2, "0"))
    .join("");
}

function readRuntimeMessages(locale: MarketplaceLocale) {
  const filePath = marketplaceRuntimeMessagesPath(locale);
  return JSON.parse(readFileSync(filePath, "utf8")) as Record<string, string>;
}

function marketplaceRuntimeMessagesPath(locale: MarketplaceLocale) {
  for (const root of [
    process.cwd(),
    join(process.cwd(), ".."),
    join(process.cwd(), "..", ".."),
  ]) {
    const filePath = join(
      root,
      "apps",
      "lina-plugins",
      "linapro-plugin-marketplace",
      "manifest",
      "i18n",
      locale,
      "plugin.json",
    );
    if (existsSync(filePath)) {
      return filePath;
    }
  }
  throw new Error(`Marketplace runtime messages not found for ${locale}`);
}

function resolveRequestLocale(route: Route): MarketplaceLocale {
  const request = route.request();
  const requestURL = new URL(request.url());
  const requested =
    requestURL.searchParams.get("lang") ||
    request.headers()["accept-language"] ||
    "";
  return requested.toLowerCase().includes("en") ? "en-US" : "zh-CN";
}

function runtimeLocaleOptions(locale: MarketplaceLocale) {
  return {
    enabled: true,
    items: [
      {
        isDefault: true,
        locale: "en-US",
        name: "English",
        nativeName: "English",
      },
      {
        isDefault: false,
        locale: "zh-CN",
        name: "Chinese",
        nativeName: "简体中文",
      },
    ],
    locale,
  };
}

function marketplaceMenuRoutes(
  locale: MarketplaceLocale,
  role: "both" | "publish-only" | "review-only",
) {
  const title = marketplaceMenuTitle(locale);
  const marketplaceChildren = [
    role !== "review-only"
      ? marketplaceMenuRoute({
          authority: "market:plugin:publish",
          icon: "ant-design:appstore-outlined",
          name: "PluginMarketplaceMine",
          order: 1,
          path: "plugin-marketplace-mine",
          title: title.mine,
        })
      : null,
    role !== "publish-only"
      ? marketplaceMenuRoute({
          authority: "market:plugin:review",
          icon: "ant-design:unordered-list-outlined",
          name: "PluginMarketplaceAdminList",
          order: 2,
          path: "plugin-marketplace-admin-list",
          title: title.adminList,
        })
      : null,
    role !== "publish-only"
      ? marketplaceMenuRoute({
          authority: "market:plugin:review",
          icon: "ant-design:audit-outlined",
          name: "PluginMarketplaceReview",
          order: 3,
          path: "plugin-marketplace-review",
          title: title.review,
        })
      : null,
    marketplaceMenuRoute({
      authority: "market:plugin:view",
      hidden: true,
      icon: "ant-design:profile-outlined",
      name: "PluginMarketplaceDetail",
      order: 4,
      path: "plugin-marketplace-detail",
      title: title.detail,
    }),
  ].filter((item) => item !== null);
  return [
    {
      children: [
        {
          component: "dashboard/analytics/index",
          meta: {
            icon: "lucide:area-chart",
            title: "page.dashboard.analytics",
          },
          name: "Analytics",
          path: "/dashboard/analytics",
        },
      ],
      component: "BasicLayout",
      meta: {
        icon: "lucide:layout-dashboard",
        order: -1,
        title: "page.dashboard.title",
      },
      name: "Dashboard",
      path: "/dashboard",
    },
    {
      children: [
        marketplaceMenuRoute({
          authority: "plugin:list",
          icon: "lucide:plug",
          name: "SystemPlugin",
          order: 1,
          path: "/system/plugin",
          title: "Plugins",
        }),
      ],
      component: "BasicLayout",
      meta: {
        icon: "lucide:puzzle",
        order: 9,
        title: "page.routes.system.extensionCenter",
      },
      name: "Extension",
      path: "/extension",
    },
    {
      children: marketplaceChildren,
      component: "BasicLayout",
      meta: {
        icon: "ant-design:shop-outlined",
        order: 8,
        title: title.directory,
      },
      name: "PluginMarketplaceDirectory",
      path: "/plugin-marketplace",
    },
  ];
}

function marketplaceMenuTitle(locale: MarketplaceLocale) {
  if (locale === "en-US") {
    return {
      adminList: "Plugin List",
      detail: "Plugin Marketplace Detail",
      directory: "Plugin Marketplace",
      mine: "My Plugins",
      review: "Plugin Review",
    };
  }
  return {
    adminList: "插件列表",
    detail: "插件市场详情",
    directory: "插件市场",
    mine: "我的插件",
    review: "插件审核",
  };
}

function marketplaceMenuRoute(options: {
  authority: string;
  hidden?: boolean;
  icon: string;
  name: string;
  order: number;
  path: string;
  title: string;
}) {
  return {
    component: "system/plugin/dynamic-page",
    meta: {
      authority: [options.authority, marketplacePluginId],
      hideInMenu: options.hidden === true,
      icon: options.icon,
      order: options.order,
      title: options.title,
    },
    name: options.name,
    path: options.path,
  };
}

function pluginDetail(
  data: MarketplaceMockData,
  pluginId: string,
  options: { includePrivatePlugins: boolean },
) {
  const plugin = data.pluginsById.get(pluginId);
  if (
    !plugin ||
    (!options.includePrivatePlugins && plugin.visibility === "private")
  ) {
    return null;
  }
  return plugin;
}

function pageResult<T>(items: T[]) {
  return {
    list: items,
    total: items.length,
  };
}

function createMarketplaceMockData(): MarketplaceMockData {
  return {
    pluginsById: new Map(
      plugins.map((plugin) => [plugin.pluginId, clonePlugin(plugin)]),
    ),
    releasesByPlugin: new Map(
      [...baseReleasesByPlugin.entries()].map(([pluginId, releases]) => [
        pluginId,
        releases.map(cloneRelease),
      ]),
    ),
  };
}

function clonePlugin(plugin: PluginItem): PluginItem {
  return {
    ...plugin,
    latestRelease: plugin.latestRelease
      ? cloneRelease(plugin.latestRelease)
      : undefined,
    publisher: plugin.publisher ? { ...plugin.publisher } : undefined,
    riskCounts: { ...plugin.riskCounts },
    tagCodes: [...plugin.tagCodes],
    tags: plugin.tags.map((tag) => ({ ...tag })),
  };
}

function cloneRelease(release: ReleaseItem): ReleaseItem {
  return {
    ...release,
    artifact: release.artifact ? { ...release.artifact } : undefined,
  };
}

function releaseList(data: MarketplaceMockData, pluginId: string) {
  return [...(data.releasesByPlugin.get(pluginId) ?? [])];
}

function addRelease(data: MarketplaceMockData, release: ReleaseItem) {
  const releases = data.releasesByPlugin.get(release.pluginId) ?? [];
  const next = [
    release,
    ...releases.filter((item) => item.version !== release.version),
  ];
  data.releasesByPlugin.set(release.pluginId, next);
}

function findRelease(
  data: MarketplaceMockData,
  pluginId: string,
  version: string,
) {
  const release = data.releasesByPlugin
    .get(pluginId)
    ?.find((item) => item.version === version);
  if (!release) {
    throw new Error(
      `Marketplace E2E mock release not found: ${pluginId}@${version}`,
    );
  }
  return release;
}

function updateRelease(
  data: MarketplaceMockData,
  pluginId: string,
  version: string,
  updates: Partial<ReleaseItem>,
) {
  const release = {
    ...findRelease(data, pluginId, version),
    ...updates,
    updatedAt: mockNow + 120000,
  };
  addRelease(data, release);
  return release;
}

function updatePluginLatestRelease(
  data: MarketplaceMockData,
  release: ReleaseItem,
) {
  const plugin = data.pluginsById.get(release.pluginId);
  if (!plugin) {
    return;
  }
  data.pluginsById.set(release.pluginId, {
    ...plugin,
    latestRelease: cloneRelease(release),
    latestReviewStatus: release.reviewStatus,
    latestVersion: release.version,
    marketStatus:
      release.releaseStatus === "published" ? "published" : plugin.marketStatus,
    updatedAt: release.updatedAt,
  });
}

function pluginFromPayload(payload: Record<string, unknown>): PluginItem {
  const pluginId = stringValue(payload.pluginId) || sourcePluginId;
  const pluginType =
    stringValue(payload.pluginType) === "dynamic" ? "dynamic" : "source";
  return {
    description: stringValue(payload.description),
    downloadCount: 0,
    latestVersion: "",
    license: stringValue(payload.license),
    marketStatus: "draft",
    name: stringValue(payload.name) || pluginId,
    pluginId,
    pluginType,
    publisher: linaPublisher,
    riskCounts: { high: 0, info: 0, warning: 0 },
    sourceDelivery:
      pluginType === "dynamic"
        ? "dynamic_upload_required"
        : "source_rebuild_required",
    summary: stringValue(payload.summary) || "Draft marketplace plugin.",
    tagCodes: stringArray(payload.tagCodes),
    tags: [],
    updatedAt: mockNow,
    visibility: visibilityValue(payload.visibility),
  };
}

function releaseFromUpload(pluginId: string, bodyText: string): ReleaseItem {
  const pluginType = bodyText.includes("\r\n\r\ndynamic\r\n")
    ? "dynamic"
    : "source";
  const version =
    multipartField(bodyText, "version") ||
    (pluginType === "dynamic" ? "v2.1.0" : "v1.2.0");
  return {
    artifact: {
      artifactType: pluginType === "dynamic" ? "dynamic_zip" : "source_zip",
      contentType: "application/zip",
      fileName:
        pluginType === "dynamic"
          ? "linapro-demo-dynamic-v2.1.0.zip"
          : "linapro-demo-source-v1.2.0.zip",
      sha256:
        pluginType === "dynamic"
          ? "dynamic-uploaded-sha256"
          : "source-uploaded-sha256",
      sizeBytes: pluginType === "dynamic" ? 32768 : 49152,
      wasmSha256:
        pluginType === "dynamic" ? "dynamic-uploaded-wasm-sha256" : undefined,
    },
    minHostVersion: multipartField(bodyText, "minHostVersion") || "v1.2.0",
    pluginId,
    pluginType,
    releaseStatus: "draft",
    reviewMessage: "Draft uploaded from E2E marketplace package.",
    reviewStatus: "draft",
    updatedAt: mockNow + 60000,
    version,
    visibility: visibilityValue(multipartField(bodyText, "visibility")),
  };
}

function multipartField(bodyText: string, fieldName: string) {
  const pattern = new RegExp(
    `name="${escapeRegExp(fieldName)}"\\r\\n\\r\\n([^\\r]+)`,
    "u",
  );
  return bodyText.match(pattern)?.[1]?.trim() ?? "";
}

function multipartFilename(bodyText: string, fieldName: string) {
  const pattern = new RegExp(
    `name="${escapeRegExp(fieldName)}";\\s*filename="([^"]+)"`,
    "u",
  );
  return bodyText.match(pattern)?.[1]?.trim() ?? "";
}

function pluginIdFromPackageFileName(fileName: string) {
  const normalized = fileName.trim().toLowerCase();
  if (normalized.includes("demo-source")) {
    return sourcePluginId;
  }
  if (normalized.includes("demo-dynamic")) {
    return dynamicPluginId;
  }
  if (normalized.includes("e2e-private") || normalized.includes("private")) {
    return "linapro-e2e-private";
  }
  // Default first-add package identity for TC-1b when the buffer name is generic.
  if (normalized.includes("linapro-demo-source")) {
    return sourcePluginId;
  }
  return "linapro-e2e-private";
}

function versionFromPackageFileName(fileName: string) {
  const match = fileName.match(/v\d+\.\d+\.\d+/u);
  return match?.[0] ?? "";
}

function postDataObject(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {};
}

function stringValue(value: unknown) {
  return typeof value === "string" ? value.trim() : "";
}

function stringArray(value: unknown) {
  return Array.isArray(value)
    ? value.filter((item): item is string => typeof item === "string")
    : [];
}

function visibilityValue(value: unknown): Visibility {
  return value === "private" || value === "reserved" ? value : "public";
}

async function fulfillData(route: Route, data: unknown) {
  await route.fulfill({
    body: JSON.stringify({
      code: 0,
      data,
      message: "success",
    }),
    contentType: "application/json",
    status: 200,
  });
}

async function fulfillError(route: Route, messageKey: string) {
  await route.fulfill({
    body: JSON.stringify({
      code: 403,
      data: null,
      error: messageKey,
      messageKey,
    }),
    contentType: "application/json",
    status: 200,
  });
}

async function fulfillBinary(route: Route, body: Buffer) {
  await route.fulfill({
    body,
    contentType: "application/zip",
    status: 200,
  });
}

function releaseKey(pluginId: string, version: string) {
  return `${pluginId}@${version}`;
}

function dynamicZipWithPluginWasm() {
  const wasm = Buffer.from([0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00]);
  return storedZip("plugin.wasm", wasm);
}

function storedZip(fileName: string, content: Buffer) {
  const name = Buffer.from(fileName);
  const localHeader = Buffer.alloc(30);
  localHeader.writeUInt32LE(0x04034b50, 0);
  localHeader.writeUInt16LE(20, 4);
  localHeader.writeUInt16LE(0, 6);
  localHeader.writeUInt16LE(0, 8);
  localHeader.writeUInt32LE(0, 10);
  localHeader.writeUInt32LE(0, 14);
  localHeader.writeUInt32LE(content.length, 18);
  localHeader.writeUInt32LE(content.length, 22);
  localHeader.writeUInt16LE(name.length, 26);
  localHeader.writeUInt16LE(0, 28);

  const centralDirectory = Buffer.alloc(46);
  centralDirectory.writeUInt32LE(0x02014b50, 0);
  centralDirectory.writeUInt16LE(20, 4);
  centralDirectory.writeUInt16LE(20, 6);
  centralDirectory.writeUInt16LE(0, 8);
  centralDirectory.writeUInt16LE(0, 10);
  centralDirectory.writeUInt32LE(0, 12);
  centralDirectory.writeUInt32LE(0, 16);
  centralDirectory.writeUInt32LE(content.length, 20);
  centralDirectory.writeUInt32LE(content.length, 24);
  centralDirectory.writeUInt16LE(name.length, 28);
  centralDirectory.writeUInt16LE(0, 30);
  centralDirectory.writeUInt16LE(0, 32);
  centralDirectory.writeUInt16LE(0, 34);
  centralDirectory.writeUInt16LE(0, 36);
  centralDirectory.writeUInt32LE(0, 38);
  centralDirectory.writeUInt32LE(0, 42);

  const centralDirectoryOffset =
    localHeader.length + name.length + content.length;
  const centralDirectorySize = centralDirectory.length + name.length;
  const endRecord = Buffer.alloc(22);
  endRecord.writeUInt32LE(0x06054b50, 0);
  endRecord.writeUInt16LE(0, 4);
  endRecord.writeUInt16LE(0, 6);
  endRecord.writeUInt16LE(1, 8);
  endRecord.writeUInt16LE(1, 10);
  endRecord.writeUInt32LE(centralDirectorySize, 12);
  endRecord.writeUInt32LE(centralDirectoryOffset, 16);
  endRecord.writeUInt16LE(0, 20);

  return Buffer.concat([
    localHeader,
    name,
    content,
    centralDirectory,
    name,
    endRecord,
  ]);
}

function escapeRegExp(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/gu, "\\$&");
}
