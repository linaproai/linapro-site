// This file defines frontend marketplace API types used by plugin pages. The
// shapes mirror the plugin-owned backend DTOs and keep page code independent
// from host plugin-management DTOs.

export type MarketplaceArtifactType =
  | 'dynamic_tar_gz'
  | 'dynamic_zip'
  | 'plugin_wasm'
  | 'source_tar_gz'
  | 'source_zip';

export type MarketplaceDownloadSessionStatus =
  | 'active'
  | 'consumed'
  | 'expired'
  | 'revoked';

export type MarketplacePluginType = 'dynamic' | 'source';

export type MarketplaceReviewStatus =
  | 'approved'
  | 'draft'
  | 'rejected'
  | 'reviewing'
  | 'submitted';

export type MarketplaceRiskSeverity = 'high' | 'info' | 'warning';

export type MarketplaceRiskType =
  | 'data_table'
  | 'dependency'
  | 'docs'
  | 'dynamic_route'
  | 'external_network'
  | 'host_service'
  | 'install_sql'
  | 'menu_permission'
  | 'mock_sql'
  | 'multi_tenant'
  | 'uninstall_sql';

export type MarketplaceSourceDelivery =
  | 'dynamic_upload_required'
  | 'source_rebuild_required';

export type MarketplaceStatus =
  | 'delisted'
  | 'deprecated'
  | 'draft'
  | 'published';

export type MarketplaceProcessStatus =
  | 'completed'
  | 'failed'
  | 'pending_review'
  | 'pending_verify';

export type MarketplaceVisibility = 'private' | 'public' | 'reserved';

export interface MarketplacePageParams {
  pageNum?: number;
  pageSize?: number;
}

export interface MarketplacePageResult<T> {
  list: T[];
  total: number;
}

export interface MarketplaceGridResult<T> {
  items: T[];
  total: number;
}

export interface MarketplacePublisherItem {
  contactEmail?: string;
  homepage?: string;
  name: string;
  publisherKey: string;
  summary?: string;
  verified: boolean;
}

export interface MarketplacePublisherListParams extends MarketplacePageParams {
  keyword?: string;
}

export interface MarketplacePublisherCreatePayload {
  contactEmail?: string;
  homepage?: string;
  name: string;
  publisherKey: string;
  summary?: string;
}

export interface MarketplacePublisherUpdatePayload {
  contactEmail?: string;
  homepage?: string;
  name: string;
  publisherKey: string;
  summary?: string;
}

export interface MarketplaceTagItem {
  code: string;
  name: string;
  type: string;
}

export interface MarketplaceRiskCounts {
  high: number;
  info: number;
  warning: number;
}

export interface MarketplaceArtifactItem {
  artifactType: MarketplaceArtifactType;
  contentType: string;
  fileName: string;
  manifestSha256?: string;
  sha256: string;
  sizeBytes: number;
  wasmSha256?: string;
}

export interface MarketplaceReleaseItem {
  artifact?: MarketplaceArtifactItem;
  distribution?: MarketplaceDistributionItem;
  maxHostVersion?: string;
  minHostVersion?: string;
  pluginId: string;
  pluginType: MarketplacePluginType;
  processStatus?: MarketplaceProcessStatus;
  publishedAt?: null | number;
  releaseStatus: MarketplaceStatus;
  reviewMessage?: string;
  reviewStatus: MarketplaceReviewStatus;
  reviewedAt?: null | number;
  sourceCommit?: string;
  sourceRef?: string;
  submittedAt?: null | number;
  updatedAt?: null | number;
  version: string;
  visibility: MarketplaceVisibility;
}

export type MarketplaceSourceKind = "git" | "upload";
export type MarketplaceDistributionMode = "git" | "https";
export type MarketplaceRepoProvider = "github" | "gitee";

export interface MarketplaceDistributionItem {
  artifactType?: MarketplaceArtifactType;
  downloadSessionRequired?: boolean;
  mode: MarketplaceDistributionMode;
  path?: string;
  pluginId: string;
  pluginType: MarketplacePluginType;
  provider?: MarketplaceRepoProvider;
  ref?: string;
  repoUrl?: string;
  requiresAuth?: boolean;
  sha256?: string;
  sizeBytes?: number;
  version: string;
}

export interface MarketplacePluginListItem {
  downloadCount: number;
  lastSyncAt?: null | number;
  lastSyncMessage?: string;
  lastSyncStatus?: string;
  latestReviewStatus?: MarketplaceReviewStatus;
  latestVersion: string;
  marketStatus: MarketplaceStatus;
  maxHostVersion?: string;
  minHostVersion?: string;
  pluginId: string;
  pluginType: MarketplacePluginType;
  primaryTag?: string;
  processStatus?: MarketplaceProcessStatus;
  publishedAt?: null | number;
  publisher?: MarketplacePublisherItem;
  repoPath?: string;
  repoProvider?: MarketplaceRepoProvider;
  repoUrl?: string;
  riskCounts: MarketplaceRiskCounts;
  sourceKind?: MarketplaceSourceKind;
  summary: string;
  tagCodes: string[];
  updatedAt?: null | number;
  visibility: MarketplaceVisibility;
  name: string;
}

export type MarketplaceListStatusFilter =
  | MarketplaceProcessStatus
  | MarketplaceStatus;

export interface MarketplaceManagedPluginListParams extends MarketplacePageParams {
  keyword?: string;
  pluginType?: MarketplacePluginType;
  publisher?: string;
  status?: MarketplaceListStatusFilter;
}

export interface MarketplaceMyPluginListParams extends MarketplacePageParams {
  keyword?: string;
  orderBy?: "downloadCount" | "marketStatus" | "pluginId" | "updatedAt" | string;
  orderDirection?: "asc" | "desc" | string;
  pluginType?: MarketplacePluginType;
  status?: MarketplaceListStatusFilter;
}

export interface MarketplaceReviewQueueItem {
  artifact?: MarketplaceArtifactItem;
  pluginId: string;
  pluginName: string;
  pluginType: MarketplacePluginType;
  publisher?: MarketplacePublisherItem;
  releaseStatus: MarketplaceStatus;
  reviewMessage?: string;
  reviewStatus: MarketplaceReviewStatus;
  submittedAt?: null | number;
  updatedAt?: null | number;
  version: string;
  visibility: MarketplaceVisibility;
}

export interface MarketplaceReviewQueueListParams extends MarketplacePageParams {
  keyword?: string;
  pluginId?: string;
  reviewStatus?: MarketplaceReviewStatus;
}

export interface MarketplacePluginDetailItem {
  description?: string;
  distribution?: MarketplaceDistributionItem;
  downloadCount: number;
  homepage?: string;
  icon?: string;
  lastSyncAt?: null | number;
  lastSyncMessage?: string;
  lastSyncStatus?: string;
  latestRelease?: MarketplaceReleaseItem;
  latestVersion: string;
  license?: string;
  marketStatus: MarketplaceStatus;
  name: string;
  pluginId: string;
  pluginType: MarketplacePluginType;
  processStatus?: MarketplaceProcessStatus;
  publishedAt?: null | number;
  publisher?: MarketplacePublisherItem;
  repoPath?: string;
  repoProvider?: MarketplaceRepoProvider;
  repoUrl?: string;
  repository?: string;
  requiresAuth?: boolean;
  riskCounts: MarketplaceRiskCounts;
  sourceDelivery: MarketplaceSourceDelivery;
  sourceKind?: MarketplaceSourceKind;
  summary: string;
  tags: MarketplaceTagItem[];
  updatedAt?: null | number;
  visibility: MarketplaceVisibility;
}

export interface MarketplacePluginListParams extends MarketplacePageParams {
  hostVersion?: string;
  keyword?: string;
  pluginType?: MarketplacePluginType;
  publisher?: string;
  tagCode?: string;
}

export interface MarketplaceReleaseListParams extends MarketplacePageParams {
  reviewStatus?: MarketplaceReviewStatus;
  status?: MarketplaceStatus;
}

export interface MarketplaceDocumentParams {
  locale?: string;
  path?: string;
}

export interface MarketplaceDocumentItem {
  content: string;
  contentHash: string;
  fallbackUsed: boolean;
  locale: string;
  /** Raw Markdown source preferred for client-side rendering. */
  markdown?: string;
  path: string;
  pluginId: string;
  resolvedLocale: string;
  sourceKind: string;
  summary: string;
  title: string;
  updatedAt?: null | number;
  version: string;
}

export interface MarketplaceDocumentCatalogItem {
  locales: string[];
  path: string;
  sourceKind: string;
  title: string;
}

export interface MarketplaceDocumentBundle {
  catalog: MarketplaceDocumentCatalogItem[];
  document?: MarketplaceDocumentItem | null;
  documents: MarketplaceDocumentItem[];
}

export interface MarketplaceRiskListParams extends MarketplacePageParams {
  severity?: MarketplaceRiskSeverity;
  type?: MarketplaceRiskType;
}

export interface MarketplaceRiskItem {
  createdAt?: null | number;
  payload: Record<string, unknown>;
  severity: MarketplaceRiskSeverity;
  source: string;
  summary: string;
  type: MarketplaceRiskType;
}

export interface MarketplacePluginCreatePayload {
  description?: string;
  homepage?: string;
  icon?: string;
  license?: string;
  name: string;
  pluginId: string;
  pluginType: MarketplacePluginType;
  publisherKey: string;
  repository?: string;
  summary: string;
  tagCodes?: string[];
  visibility?: MarketplaceVisibility;
}

export interface MarketplaceReleaseUploadParams {
  file: File;
  maxHostVersion?: string;
  minHostVersion?: string;
  pluginId: string;
  pluginType: MarketplacePluginType;
  replaceDraft?: boolean;
  version: string;
  visibility?: MarketplaceVisibility;
}

export interface MarketplaceReleaseSubmitPayload {
  message?: string;
}

export interface MarketplaceReleaseReviewPayload {
  message?: string;
  reviewStatus: Extract<MarketplaceReviewStatus, 'approved' | 'rejected'>;
}

export interface MarketplacePluginStatusUpdatePayload {
  message?: string;
  status: Exclude<MarketplaceStatus, 'draft'>;
}

export interface MarketplaceDownloadSessionCreateParams {
  artifactType?: MarketplaceArtifactType;
  pluginId: string;
  version: string;
}

export interface MarketplaceDownloadSessionItem {
  artifactType: MarketplaceArtifactType;
  consumedAt?: null | number;
  createdAt?: null | number;
  downloadUrl: string;
  expiresAt?: null | number;
  pluginId: string;
  sessionId: string;
  sha256: string;
  sizeBytes: number;
  status: MarketplaceDownloadSessionStatus;
  version: string;
}
