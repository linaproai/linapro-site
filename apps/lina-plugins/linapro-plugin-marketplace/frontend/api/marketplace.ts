// This file provides the plugin-owned marketplace frontend API adapter. It
// builds URLs through the host plugin API helper and keeps response projection
// thin so pages can use Vben grid conventions without duplicating DTO mapping.

import type { RequestClientConfig } from "@vben/request";

import type {
  MarketplaceDocumentBundle,
  MarketplaceDocumentParams,
  MarketplaceDownloadSessionCreateParams,
  MarketplaceDownloadSessionItem,
  MarketplaceGridResult,
  MarketplacePageResult,
  MarketplacePluginCreatePayload,
  MarketplacePluginDetailItem,
  MarketplaceManagedPluginListParams,
  MarketplacePluginListItem,
  MarketplacePluginListParams,
  MarketplacePluginStatusUpdatePayload,
  MarketplacePublisherCreatePayload,
  MarketplacePublisherItem,
  MarketplacePublisherListParams,
  MarketplacePublisherUpdatePayload,
  MarketplaceMyPluginListParams,
  MarketplaceReleaseItem,
  MarketplaceReleaseListParams,
  MarketplaceReleaseReviewPayload,
  MarketplaceReleaseSubmitPayload,
  MarketplaceReleaseUploadParams,
  MarketplaceReviewQueueItem,
  MarketplaceReviewQueueListParams,
  MarketplaceRiskItem,
  MarketplaceRiskListParams,
} from "../types/marketplace";

import { pluginApiPath, requestClient } from "#/api/request";

export const marketplacePluginId = "linapro-plugin-marketplace";

export type MarketplaceReadScope = "managed" | "mine" | "public";

const marketplaceUploadTimeout = 120_000;
const marketplaceDownloadTimeout = 120_000;
// Git monorepo registration can touch many remote roots; keep above browser default 10s.
const marketplaceGitTimeout = 120_000;

function marketplacePath(pathName: string) {
  return pluginApiPath(marketplacePluginId, pathName);
}

function encodePathSegment(value: string) {
  return encodeURIComponent(value.trim());
}

function releasePath(pluginId: string, version: string, suffix = "") {
  const base = `market/plugins/${encodePathSegment(pluginId)}/releases/${encodePathSegment(version)}`;
  return suffix ? `${base}/${suffix.replace(/^\/+/, "")}` : base;
}

function readPluginBase(scope: MarketplaceReadScope) {
  if (scope === "mine") {
    return "market/my-plugins";
  }
  if (scope === "managed") {
    return "market/managed-plugins";
  }
  return "market/plugins";
}

function readPluginPath(pluginId: string, scope: MarketplaceReadScope) {
  return `${readPluginBase(scope)}/${encodePathSegment(pluginId)}`;
}

function readReleasePath(
  pluginId: string,
  version: string,
  suffix: string,
  scope: MarketplaceReadScope,
) {
  const base = `${readPluginPath(pluginId, scope)}/releases/${encodePathSegment(version)}`;
  return suffix ? `${base}/${suffix.replace(/^\/+/, "")}` : base;
}

function appendString(formData: FormData, key: string, value?: string) {
  if (value) {
    formData.append(key, value);
  }
}

function appendBoolean(formData: FormData, key: string, value?: boolean) {
  if (typeof value === "boolean") {
    formData.append(key, value ? "1" : "0");
  }
}

function gridResult<T>(
  result: MarketplacePageResult<T>,
): MarketplaceGridResult<T> {
  return {
    items: result.list ?? [],
    total: result.total ?? 0,
  };
}

export async function marketplacePublisherList(
  params?: MarketplacePublisherListParams,
) {
  const res = await requestClient.get<
    MarketplacePageResult<MarketplacePublisherItem>
  >(marketplacePath("market/publishers"), { params });
  return gridResult(res);
}

export function marketplacePublisherCreate(
  data: MarketplacePublisherCreatePayload,
) {
  return requestClient.post<{ publisher: MarketplacePublisherItem }>(
    marketplacePath("market/publishers"),
    data,
  );
}

export function marketplacePublisherUpdate(
  publisherKey: string,
  data: MarketplacePublisherUpdatePayload,
) {
  return requestClient.put<{ publisher: MarketplacePublisherItem }>(
    marketplacePath(`market/publishers/${encodePathSegment(publisherKey)}`),
    data,
  );
}

export async function marketplacePluginList(
  params?: MarketplacePluginListParams,
) {
  const res = await requestClient.get<
    MarketplacePageResult<MarketplacePluginListItem>
  >(marketplacePath("market/plugins"), { params });
  return gridResult(res);
}

export async function marketplaceMyPluginList(
  params?: MarketplaceMyPluginListParams,
) {
  const res = await requestClient.get<
    MarketplacePageResult<MarketplacePluginListItem>
  >(marketplacePath("market/my-plugins"), { params });
  return gridResult(res);
}

export async function marketplaceManagedPluginList(
  params?: MarketplaceManagedPluginListParams,
) {
  const res = await requestClient.get<
    MarketplacePageResult<MarketplacePluginListItem>
  >(marketplacePath("market/managed-plugins"), { params });
  return gridResult(res);
}

export async function marketplaceReviewQueueList(
  params?: MarketplaceReviewQueueListParams,
) {
  const res = await requestClient.get<
    MarketplacePageResult<MarketplaceReviewQueueItem>
  >(marketplacePath("market/review-queue"), { params });
  return gridResult(res);
}

export async function marketplacePluginDetail(
  pluginId: string,
  scope: MarketplaceReadScope = "public",
) {
  const res = await requestClient.get<{
    plugin: MarketplacePluginDetailItem;
  }>(marketplacePath(readPluginPath(pluginId, scope)));
  return res.plugin;
}

export function marketplacePluginCreate(data: MarketplacePluginCreatePayload) {
  return requestClient.post<{ plugin: MarketplacePluginDetailItem }>(
    marketplacePath("market/plugins"),
    data,
  );
}

export type MarketplaceGitSourceRegisterPayload = {
  accessToken?: string;
  homepage?: string;
  license?: string;
  publisherKey: string;
  repoUrl: string;
};

export function marketplaceGitSourceRegister(
  data: MarketplaceGitSourceRegisterPayload,
) {
  return requestClient.post<{ plugin: MarketplacePluginDetailItem }>(
    marketplacePath("market/plugins/git-sources"),
    data,
    { timeout: marketplaceGitTimeout },
  );
}

export type MarketplacePackageAddParams = {
  file: File;
  publisherKey?: string;
  replaceDraft?: boolean;
};

export function marketplacePackageAdd(params: MarketplacePackageAddParams) {
  const formData = new FormData();
  formData.append("file", params.file, params.file.name);
  appendString(formData, "publisherKey", params.publisherKey);
  appendBoolean(formData, "replaceDraft", params.replaceDraft ?? true);

  return requestClient.post<{
    plugin: MarketplacePluginDetailItem;
    release: MarketplaceReleaseItem;
  }>(marketplacePath("market/my-plugins/packages"), formData, {
    headers: {
      "Content-Type": "multipart/form-data",
    },
    timeout: marketplaceUploadTimeout,
  });
}

export function marketplacePluginPublish(
  pluginId: string,
  data?: { message?: string; version?: string },
) {
  return requestClient.post<{ release: MarketplaceReleaseItem }>(
    marketplacePath(
      `market/my-plugins/${encodePathSegment(pluginId)}/publish`,
    ),
    data ?? {},
  );
}

export function marketplacePluginDelist(
  pluginId: string,
  data?: { message?: string },
) {
  return requestClient.post<{ plugin: MarketplacePluginDetailItem }>(
    marketplacePath(`market/my-plugins/${encodePathSegment(pluginId)}/delist`),
    data ?? {},
  );
}

export function marketplaceGitSourceSync(pluginId: string) {
  return requestClient.post<{
    plugin: MarketplacePluginDetailItem;
    synced: number;
  }>(
    marketplacePath(`market/plugins/${encodePathSegment(pluginId)}/git-sync`),
    {},
    { timeout: marketplaceGitTimeout },
  );
}

export async function marketplaceReleaseDistribution(
  pluginId: string,
  version: string,
  scope: MarketplaceReadScope = "public",
) {
  const path =
    scope === "mine"
      ? `market/my-plugins/${encodePathSegment(pluginId)}/releases/${encodePathSegment(version)}/distribution`
      : `market/plugins/${encodePathSegment(pluginId)}/releases/${encodePathSegment(version)}/distribution`;
  const res = await requestClient.get<{
    distribution: import("../types/marketplace").MarketplaceDistributionItem;
  }>(marketplacePath(path));
  return res.distribution;
}

export async function marketplaceReleaseList(
  pluginId: string,
  params?: MarketplaceReleaseListParams,
  scope: MarketplaceReadScope = "public",
) {
  const res = await requestClient.get<
    MarketplacePageResult<MarketplaceReleaseItem>
  >(marketplacePath(`${readPluginPath(pluginId, scope)}/releases`), {
    params,
  });
  return gridResult(res);
}

export function marketplaceReleaseUpload(
  params: MarketplaceReleaseUploadParams,
) {
  const formData = new FormData();
  formData.append("file", params.file, params.file.name);
  formData.append("pluginId", params.pluginId);
  formData.append("version", params.version);
  formData.append("pluginType", params.pluginType);
  appendString(formData, "visibility", params.visibility);
  appendString(formData, "minHostVersion", params.minHostVersion);
  appendString(formData, "maxHostVersion", params.maxHostVersion);
  appendBoolean(formData, "replaceDraft", params.replaceDraft);

  return requestClient.post<{ release: MarketplaceReleaseItem }>(
    marketplacePath(
      `market/plugins/${encodePathSegment(params.pluginId)}/releases`,
    ),
    formData,
    {
      headers: {
        "Content-Type": "multipart/form-data",
      },
      timeout: marketplaceUploadTimeout,
    },
  );
}

export function marketplaceReleaseSubmitReview(
  pluginId: string,
  version: string,
  data?: MarketplaceReleaseSubmitPayload,
) {
  return requestClient.post<{ release: MarketplaceReleaseItem }>(
    marketplacePath(releasePath(pluginId, version, "submit-review")),
    data ?? {},
  );
}

export function marketplaceReleaseReview(
  pluginId: string,
  version: string,
  data: MarketplaceReleaseReviewPayload,
) {
  return requestClient.put<{ release: MarketplaceReleaseItem }>(
    marketplacePath(releasePath(pluginId, version, "review")),
    data,
  );
}

export function marketplacePluginStatusUpdate(
  pluginId: string,
  data: MarketplacePluginStatusUpdatePayload,
) {
  return requestClient.put<{ plugin: MarketplacePluginDetailItem }>(
    marketplacePath(`market/plugins/${encodePathSegment(pluginId)}/status`),
    data,
  );
}

export async function marketplaceReleaseDocument(
  pluginId: string,
  version: string,
  params?: MarketplaceDocumentParams,
  scope: MarketplaceReadScope = "public",
) {
  const res = await marketplaceReleaseDocumentBundle(
    pluginId,
    version,
    params,
    scope,
  );
  return res.document ?? null;
}

export async function marketplaceReleaseDocumentBundle(
  pluginId: string,
  version: string,
  params?: MarketplaceDocumentParams,
  scope: MarketplaceReadScope = "public",
) {
  const res = await requestClient.get<MarketplaceDocumentBundle>(
    marketplacePath(readReleasePath(pluginId, version, "docs", scope)),
    { params },
  );
  return {
    document: res.document ?? null,
    documents: res.documents ?? [],
  };
}

export async function marketplaceReleaseRisks(
  pluginId: string,
  version: string,
  params?: MarketplaceRiskListParams,
  scope: MarketplaceReadScope = "public",
) {
  const res = await requestClient.get<
    MarketplacePageResult<MarketplaceRiskItem>
  >(marketplacePath(readReleasePath(pluginId, version, "risks", scope)), {
    params,
  });
  return gridResult(res);
}

export function marketplaceDownloadSessionCreate(
  params: MarketplaceDownloadSessionCreateParams,
  options?: Pick<RequestClientConfig, "silentErrorMessage">,
) {
  return requestClient.post<{ session: MarketplaceDownloadSessionItem }>(
    marketplacePath(
      `market/plugins/${encodePathSegment(params.pluginId)}/releases/${encodePathSegment(params.version)}/downloads`,
    ),
    {
      artifactType: params.artifactType,
      pluginId: params.pluginId,
      version: params.version,
    },
    options,
  );
}

export async function marketplaceDownloadSessionGet(sessionId: string) {
  const res = await requestClient.get<{
    session: MarketplaceDownloadSessionItem;
  }>(
    marketplacePath(`market/download-sessions/${encodePathSegment(sessionId)}`),
  );
  return res.session;
}

export function marketplaceDownloadSessionBlob(
  session: MarketplaceDownloadSessionItem,
  options?: Pick<RequestClientConfig, "timeout">,
) {
  return requestClient.download<Blob>(session.downloadUrl, {
    timeout: options?.timeout ?? marketplaceDownloadTimeout,
  });
}
