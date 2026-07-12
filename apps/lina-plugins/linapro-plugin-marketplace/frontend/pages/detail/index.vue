<script lang="ts">
import type { PluginPageMeta } from "#/plugins/page-registry";

export const pluginPageMeta = {
  pluginId: "linapro-plugin-marketplace",
  routePath: "plugin-marketplace-detail",
  title: "Plugin Marketplace Detail",
} satisfies PluginPageMeta;
</script>

<script setup lang="ts">
import type { VxeGridProps } from "#/adapter/vxe-table";

import type {
  MarketplaceArtifactType,
  MarketplaceDocumentItem,
  MarketplacePluginDetailItem,
  MarketplaceReleaseItem,
  MarketplaceReviewStatus,
  MarketplaceRiskCounts,
  MarketplaceRiskItem,
  MarketplaceRiskSeverity,
  MarketplaceRiskType,
  MarketplaceStatus,
  MarketplaceVisibility,
} from "../../types/marketplace";
import type { MarketplaceReadScope } from "../../api/marketplace";

import { computed, h, nextTick, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import { useAccess } from "@vben/access";
import { Page } from "@vben/common-ui";
import { IconifyIcon } from "@vben/icons";
import { breakpointsTailwind, useBreakpoints } from "@vueuse/core";

import {
  Alert,
  Descriptions,
  DescriptionsItem,
  Empty,
  Modal,
  Space,
  Spin,
  TabPane,
  Tabs,
  Tag,
  Tooltip,
  message,
} from "ant-design-vue";

import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { pluginDynamicUpload } from "#/api/system/plugin";
import { $t } from "#/locales";
import { notifyPluginRegistryChanged } from "#/plugins/slot-registry";
import { downloadBlob } from "#/utils/download";
import { formatTimestamp } from "#/utils/time";

import {
  marketplaceDownloadSessionBlob,
  marketplaceDownloadSessionCreate,
  marketplacePluginDetail,
  marketplaceReleaseDocument,
  marketplaceReleaseList,
  marketplaceReleaseRisks,
} from "../../api/marketplace";
import { marketplaceBackPath } from "../../utils/routes";

type DetailGridOptions = NonNullable<
  VxeGridProps<MarketplaceReleaseItem>["gridOptions"]
>;

type GridPageInfo = {
  currentPage: number;
  pageSize: number;
};

type DetailTabKey = "docs" | "risks" | "versions";

type LoadState = "empty" | "error" | "idle" | "loading" | "ready";

type RuntimeErrorPayload = Record<string, unknown>;

const riskEmptyCounts: MarketplaceRiskCounts = {
  high: 0,
  info: 0,
  warning: 0,
};

const route = useRoute();
const router = useRouter();
const breakpoints = useBreakpoints(breakpointsTailwind);
const isMobile = breakpoints.smaller("md");
const readScope = computed<MarketplaceReadScope>(() => {
  if (route.query.from === "mine") {
    return "mine";
  }
  if (route.query.from === "admin-list" || route.query.from === "review") {
    return "managed";
  }
  return "public";
});
const { hasAccessByCodes } = useAccess();

const detail = ref<MarketplacePluginDetailItem | null>(null);
const currentDocument = ref<MarketplaceDocumentItem | null>(null);
const currentRisks = ref<MarketplaceRiskItem[]>([]);
const selectedRelease = ref<MarketplaceReleaseItem | null>(null);
const activeTab = ref<DetailTabKey>("versions");
const loading = ref(false);
const documentLoading = ref(false);
const riskLoading = ref(false);
const downloadingReleaseKeys = ref<Record<string, boolean>>({});
const detailLoadState = ref<LoadState>("idle");
const versionLoadState = ref<LoadState>("idle");
const documentLoadState = ref<LoadState>("idle");
const riskLoadState = ref<LoadState>("idle");
let pageRequestId = 0;
let releaseContextRequestId = 0;

const [VersionGrid, versionGridApi] = useVbenVxeGrid<MarketplaceReleaseItem>({
  gridOptions: {
    columns: [],
    height: "auto",
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      autoLoad: false,
      ajax: {
        query: async ({ page }: { page: GridPageInfo }) => {
          const pluginId = getRoutePluginId();
          if (!pluginId) {
            versionLoadState.value = "empty";
            return { items: [], total: 0 };
          }

          versionLoadState.value = "loading";
          try {
            const result = await marketplaceReleaseList(
              pluginId,
              {
                pageNum: page.currentPage,
                pageSize: page.pageSize,
              },
              readScope.value,
            );
            if (pluginId !== getRoutePluginId()) {
              return result;
            }
            if (!selectedRelease.value && result.items[0]) {
              selectedRelease.value = result.items[0];
            }
            versionLoadState.value = result.total > 0 ? "ready" : "empty";
            return result;
          } catch (error) {
            if (pluginId === getRoutePluginId()) {
              versionLoadState.value = "error";
            }
            throw error;
          }
        },
      },
    },
    rowConfig: {
      keyField: "version",
    },
    showOverflow: "tooltip",
    id: "plugin-marketplace-detail-releases",
  },
});

function t(key: string, params?: Record<string, number | string>) {
  return params ? $t(key, params) : $t(key);
}

function getRoutePluginId() {
  const value = route.query.pluginId;
  if (Array.isArray(value)) {
    return value[0] || "";
  }
  return typeof value === "string" ? value : "";
}

function buildVersionColumns(): DetailGridOptions["columns"] {
  return [
    {
      align: "left",
      field: "version",
      headerAlign: "center",
      minWidth: 160,
      showOverflow: false,
      slots: { default: "version" },
      title: t("plugin.linapro-plugin-marketplace.detail.columns.version"),
    },
    {
      field: "releaseStatus",
      slots: { default: "releaseStatus" },
      title: t("plugin.linapro-plugin-marketplace.detail.columns.status"),
      width: 120,
    },
    {
      field: "reviewStatus",
      slots: { default: "reviewStatus" },
      title: t("plugin.linapro-plugin-marketplace.detail.columns.reviewStatus"),
      width: 130,
    },
    {
      align: "left",
      field: "artifact",
      headerAlign: "center",
      minWidth: 190,
      showOverflow: false,
      slots: { default: "artifact" },
      title: t("plugin.linapro-plugin-marketplace.detail.columns.artifact"),
    },
    {
      align: "left",
      field: "compatibility",
      headerAlign: "center",
      minWidth: 180,
      slots: { default: "compatibility" },
      title: t(
        "plugin.linapro-plugin-marketplace.detail.columns.compatibility",
      ),
    },
    {
      field: "publishedAt",
      formatter: ({ cellValue }: { cellValue?: null | number | string }) =>
        formatTimestamp(cellValue),
      title: t("plugin.linapro-plugin-marketplace.detail.columns.publishedAt"),
      width: 180,
    },
    {
      field: "action",
      fixed: "right",
      showOverflow: false,
      slots: { default: "action" },
      title: t("plugin.linapro-plugin-marketplace.detail.columns.actions"),
      width: isMobile.value ? 165 : 230,
    },
  ];
}

onMounted(async () => {
  versionGridApi.setGridOptions({
    columns: buildVersionColumns(),
  });
  await nextTick();
  await initializePage();
});

watch(
  () => route.query.pluginId,
  () => {
    void initializePage();
  },
);

watch(isMobile, () => {
  versionGridApi.setGridOptions({ columns: buildVersionColumns() });
});

async function initializePage() {
  const pluginId = getRoutePluginId();
  const requestId = ++pageRequestId;
  releaseContextRequestId += 1;
  detail.value = null;
  currentDocument.value = null;
  currentRisks.value = [];
  selectedRelease.value = null;
  activeTab.value = "versions";
  detailLoadState.value = "idle";
  versionLoadState.value = "idle";
  documentLoadState.value = "idle";
  riskLoadState.value = "idle";

  if (!pluginId) {
    return;
  }

  loading.value = true;
  detailLoadState.value = "loading";
  try {
    const result = await marketplacePluginDetail(pluginId, readScope.value);
    if (requestId !== pageRequestId) {
      return;
    }
    detail.value = result;
    detailLoadState.value = "ready";
    selectedRelease.value = detail.value.latestRelease ?? null;
    await nextTick();
    await versionGridApi.reload();
    if (requestId === pageRequestId && selectedRelease.value) {
      await loadReleaseContext(selectedRelease.value);
    }
  } catch {
    if (requestId === pageRequestId) {
      detailLoadState.value = "error";
    }
  } finally {
    if (requestId === pageRequestId) {
      loading.value = false;
    }
  }
}

function handleBack() {
  const from =
    typeof route.query.from === "string" ? route.query.from : undefined;
  router.push(marketplaceBackPath(from));
}

async function handleSelectRelease(
  row: MarketplaceReleaseItem,
  tab: DetailTabKey,
) {
  selectedRelease.value = row;
  activeTab.value = tab;
  await loadReleaseContext(row);
}

async function loadReleaseContext(row: MarketplaceReleaseItem) {
  const requestId = ++releaseContextRequestId;
  currentDocument.value = null;
  currentRisks.value = [];
  documentLoading.value = true;
  riskLoading.value = true;
  documentLoadState.value = "loading";
  riskLoadState.value = "loading";
  await Promise.all([
    loadReleaseDocument(row, requestId),
    loadReleaseRisks(row, requestId),
  ]);
}

async function loadReleaseDocument(
  row: MarketplaceReleaseItem,
  requestId: number,
) {
  try {
    const result = await marketplaceReleaseDocument(
      row.pluginId,
      row.version,
      undefined,
      readScope.value,
    );
    if (!isCurrentReleaseRequest(row, requestId)) {
      return;
    }
    currentDocument.value = result;
    documentLoadState.value = "ready";
  } catch (error) {
    if (!isCurrentReleaseRequest(row, requestId)) {
      return;
    }
    currentDocument.value = null;
    documentLoadState.value = isMarketplaceErrorKey(
      error,
      "error.plugin.marketplace.document.not.found",
    )
      ? "empty"
      : "error";
  } finally {
    if (isCurrentReleaseRequest(row, requestId)) {
      documentLoading.value = false;
    }
  }
}

async function loadReleaseRisks(
  row: MarketplaceReleaseItem,
  requestId: number,
) {
  try {
    const result = await marketplaceReleaseRisks(
      row.pluginId,
      row.version,
      {
        pageNum: 1,
        pageSize: 50,
      },
      readScope.value,
    );
    if (!isCurrentReleaseRequest(row, requestId)) {
      return;
    }
    currentRisks.value = result.items;
    riskLoadState.value = result.items.length > 0 ? "ready" : "empty";
  } catch {
    if (!isCurrentReleaseRequest(row, requestId)) {
      return;
    }
    currentRisks.value = [];
    riskLoadState.value = "error";
  } finally {
    if (isCurrentReleaseRequest(row, requestId)) {
      riskLoading.value = false;
    }
  }
}

function isCurrentReleaseRequest(
  row: MarketplaceReleaseItem,
  requestId: number,
) {
  return (
    requestId === releaseContextRequestId &&
    selectedRelease.value !== null &&
    getReleaseKey(selectedRelease.value) === getReleaseKey(row)
  );
}

function isMarketplaceErrorKey(error: unknown, messageKey: string) {
  return stringValue(extractErrorPayload(error).messageKey) === messageKey;
}

function handleConfirmDownload(row: MarketplaceReleaseItem) {
  const shouldImportDynamic = shouldImportDynamicPackage(row);
  Modal.confirm({
    cancelText: t("plugin.linapro-plugin-marketplace.detail.actions.cancel"),
    content: buildDownloadConfirmContent(row),
    okText: shouldImportDynamic
      ? t("plugin.linapro-plugin-marketplace.detail.actions.downloadAndImport")
      : t("plugin.linapro-plugin-marketplace.detail.actions.download"),
    title: t("plugin.linapro-plugin-marketplace.detail.download.confirmTitle"),
    onOk: async () => {
      const releaseKey = getReleaseKey(row);
      setDownloading(releaseKey, true);
      try {
        const packageBlob = await downloadReleaseArtifact(row);
        downloadBlob(packageBlob, buildDownloadFileName(row));
        if (shouldImportDynamic) {
          const imported = await importDynamicPluginWasm(row);
          if (!imported) {
            return;
          }
          message.success(
            t(
              "plugin.linapro-plugin-marketplace.detail.download.importSuccess",
            ),
          );
        } else {
          message.success(
            t("plugin.linapro-plugin-marketplace.detail.download.success"),
          );
        }
      } catch (error) {
        message.error(resolveDownloadErrorMessage(error));
      } finally {
        setDownloading(releaseKey, false);
      }
    },
  });
}

async function downloadReleaseArtifact(
  row: MarketplaceReleaseItem,
  artifactType?: MarketplaceArtifactType,
) {
  const response = await marketplaceDownloadSessionCreate(
    {
      artifactType: artifactType || row.artifact?.artifactType,
      pluginId: row.pluginId,
      version: row.version,
    },
    {
      silentErrorMessage: true,
    },
  );
  return marketplaceDownloadSessionBlob(response.session);
}

function resolveDownloadErrorMessage(error: unknown) {
  const data = extractErrorPayload(error);
  const messageKey = stringValue(data.messageKey);
  if (messageKey) {
    const localized = t(messageKey, normalizeMessageParams(data.messageParams));
    if (localized && localized !== messageKey) {
      return localized;
    }
  }
  return (
    stringValue(data.error) ||
    stringValue(data.message) ||
    stringValue(isRecord(error) ? error.message : undefined) ||
    t("error.plugin.marketplace.download.session.unavailable")
  );
}

function extractErrorPayload(error: unknown): RuntimeErrorPayload {
  if (!isRecord(error)) {
    return {};
  }
  const response = error.response;
  if (isRecord(response) && isRecord(response.data)) {
    return response.data;
  }
  return hasRuntimeErrorFields(error) ? error : {};
}

function normalizeMessageParams(value: unknown) {
  if (!isRecord(value)) {
    return {};
  }
  return Object.fromEntries(
    Object.entries(value).filter(
      (entry): entry is [string, number | string] =>
        typeof entry[1] === "number" || typeof entry[1] === "string",
    ),
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function hasRuntimeErrorFields(value: RuntimeErrorPayload) {
  return (
    typeof value.error === "string" ||
    typeof value.message === "string" ||
    typeof value.messageKey === "string"
  );
}

function stringValue(value: unknown) {
  return typeof value === "string" ? value.trim() : "";
}

function buildDownloadConfirmContent(row: MarketplaceReleaseItem) {
  const boundaryKey =
    row.pluginType === "source"
      ? "plugin.linapro-plugin-marketplace.detail.download.sourceBoundary"
      : shouldImportDynamicPackage(row)
        ? "plugin.linapro-plugin-marketplace.detail.download.dynamicImportBoundary"
        : "plugin.linapro-plugin-marketplace.detail.download.dynamicBoundary";
  const alerts = [
    h(Alert, {
      message: t(boundaryKey),
      showIcon: true,
      type: row.pluginType === "source" ? "warning" : "info",
    }),
  ];

  if (row.pluginType === "dynamic" && !canImportDynamicPlugin()) {
    alerts.push(
      h(Alert, {
        message: t(
          "plugin.linapro-plugin-marketplace.detail.download.dynamicImportNoPermission",
        ),
        showIcon: true,
        type: "warning",
      }),
    );
  }

  return h("div", { class: "marketplace-download-confirm" }, [
    h("p", t("plugin.linapro-plugin-marketplace.detail.download.confirmBody")),
    h(
      "dl",
      { class: "marketplace-download-meta" },
      [
        [
          t("plugin.linapro-plugin-marketplace.detail.fields.artifact"),
          formatArtifactType(row.artifact?.artifactType),
        ],
        [
          t("plugin.linapro-plugin-marketplace.detail.fields.size"),
          formatBytes(row.artifact?.sizeBytes),
        ],
        [
          t("plugin.linapro-plugin-marketplace.detail.fields.sha256"),
          row.artifact?.sha256 || "-",
        ],
      ].flatMap(([label, value]) => [h("dt", label), h("dd", value)]),
    ),
    ...alerts,
  ]);
}

function canImportDynamicPlugin() {
  return hasAccessByCodes(["plugin:install", "*:*:*"]);
}

function canDownloadMarketplacePlugin() {
  return hasAccessByCodes(["market:plugin:download", "*:*:*"]);
}

function shouldImportDynamicPackage(row: MarketplaceReleaseItem) {
  return row.pluginType === "dynamic" && canImportDynamicPlugin();
}

async function importDynamicPluginWasm(row: MarketplaceReleaseItem) {
  const wasmBlob = await downloadReleaseArtifact(row, "plugin_wasm");
  const bytes = new Uint8Array(await wasmBlob.arrayBuffer());
  if (!isWasmBinary(bytes)) {
    message.error(
      t("plugin.linapro-plugin-marketplace.detail.download.wasmInvalid"),
    );
    return false;
  }
  const file = new File([bytes], `${row.pluginId}-${row.version}.wasm`, {
    type: "application/wasm",
  });
  await pluginDynamicUpload(file, false);
  await notifyPluginRegistryChanged();
  return true;
}

function isWasmBinary(bytes: Uint8Array) {
  return (
    bytes.length >= 4 &&
    bytes[0] === 0x00 &&
    bytes[1] === 0x61 &&
    bytes[2] === 0x73 &&
    bytes[3] === 0x6d
  );
}

function setDownloading(releaseKey: string, downloading: boolean) {
  const next = { ...downloadingReleaseKeys.value };
  if (downloading) {
    next[releaseKey] = true;
  } else {
    delete next[releaseKey];
  }
  downloadingReleaseKeys.value = next;
}

function isDownloading(row: MarketplaceReleaseItem) {
  return downloadingReleaseKeys.value[getReleaseKey(row)] === true;
}

function getReleaseKey(row: MarketplaceReleaseItem) {
  return `${row.pluginId}:${row.version}`;
}

function buildDownloadFileName(row: MarketplaceReleaseItem) {
  if (row.artifact?.fileName) {
    return row.artifact.fileName;
  }
  const extension =
    row.artifact?.artifactType === "plugin_wasm" ? "wasm" : "zip";
  return `${row.pluginId}-${row.version}.${extension}`;
}

function getRiskCounts() {
  return detail.value?.riskCounts || riskEmptyCounts;
}

function hasRiskCounts() {
  const counts = getRiskCounts();
  return counts.high > 0 || counts.warning > 0 || counts.info > 0;
}

function hasRiskAssessment() {
  return Boolean(detail.value?.latestRelease?.artifact);
}

function getCompatibilityLabel(row: MarketplaceReleaseItem) {
  if (row.minHostVersion && row.maxHostVersion) {
    return t("plugin.linapro-plugin-marketplace.detail.compatibility.range", {
      max: row.maxHostVersion,
      min: row.minHostVersion,
    });
  }
  if (row.minHostVersion) {
    return t("plugin.linapro-plugin-marketplace.detail.compatibility.min", {
      min: row.minHostVersion,
    });
  }
  if (row.maxHostVersion) {
    return t("plugin.linapro-plugin-marketplace.detail.compatibility.max", {
      max: row.maxHostVersion,
    });
  }
  return t("plugin.linapro-plugin-marketplace.detail.compatibility.any");
}

function formatPluginType(type?: string) {
  if (type === "source") {
    return t("plugin.linapro-plugin-marketplace.catalog.pluginType.source");
  }
  if (type === "dynamic") {
    return t("plugin.linapro-plugin-marketplace.catalog.pluginType.dynamic");
  }
  return type || "-";
}

function formatStatus(status: MarketplaceStatus) {
  switch (status) {
    case "delisted": {
      return t("plugin.linapro-plugin-marketplace.detail.status.delisted");
    }
    case "deprecated": {
      return t("plugin.linapro-plugin-marketplace.detail.status.deprecated");
    }
    case "draft": {
      return t("plugin.linapro-plugin-marketplace.detail.status.draft");
    }
    case "published": {
      return t("plugin.linapro-plugin-marketplace.detail.status.published");
    }
  }
}

function getStatusColor(status: MarketplaceStatus) {
  switch (status) {
    case "published": {
      return "success";
    }
    case "deprecated": {
      return "warning";
    }
    case "delisted": {
      return "default";
    }
    default: {
      return "processing";
    }
  }
}

function formatReviewStatus(status: MarketplaceReviewStatus) {
  switch (status) {
    case "approved": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.reviewStatus.approved",
      );
    }
    case "draft": {
      return t("plugin.linapro-plugin-marketplace.detail.reviewStatus.draft");
    }
    case "rejected": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.reviewStatus.rejected",
      );
    }
    case "reviewing": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.reviewStatus.reviewing",
      );
    }
    case "submitted": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.reviewStatus.submitted",
      );
    }
  }
}

function getReviewStatusColor(status: MarketplaceReviewStatus) {
  switch (status) {
    case "approved": {
      return "success";
    }
    case "rejected": {
      return "error";
    }
    case "reviewing": {
      return "processing";
    }
    case "submitted": {
      return "warning";
    }
    default: {
      return "default";
    }
  }
}

function formatVisibility(visibility: MarketplaceVisibility) {
  switch (visibility) {
    case "private": {
      return t("plugin.linapro-plugin-marketplace.detail.visibility.private");
    }
    case "public": {
      return t("plugin.linapro-plugin-marketplace.detail.visibility.public");
    }
    case "reserved": {
      return t("plugin.linapro-plugin-marketplace.detail.visibility.reserved");
    }
  }
}

function formatArtifactType(type?: MarketplaceArtifactType) {
  switch (type) {
    case "dynamic_zip": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.artifactType.dynamic_zip",
      );
    }
    case "plugin_wasm": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.artifactType.plugin_wasm",
      );
    }
    case "source_zip": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.artifactType.source_zip",
      );
    }
    default: {
      return "-";
    }
  }
}

function formatRiskSeverity(severity: MarketplaceRiskSeverity) {
  switch (severity) {
    case "high": {
      return t("plugin.linapro-plugin-marketplace.detail.riskSeverity.high");
    }
    case "info": {
      return t("plugin.linapro-plugin-marketplace.detail.riskSeverity.info");
    }
    case "warning": {
      return t("plugin.linapro-plugin-marketplace.detail.riskSeverity.warning");
    }
  }
}

function getRiskSeverityColor(severity: MarketplaceRiskSeverity) {
  switch (severity) {
    case "high": {
      return "error";
    }
    case "warning": {
      return "warning";
    }
    default: {
      return "processing";
    }
  }
}

function formatRiskType(type: MarketplaceRiskType) {
  switch (type) {
    case "data_table": {
      return t("plugin.linapro-plugin-marketplace.detail.riskType.data_table");
    }
    case "dependency": {
      return t("plugin.linapro-plugin-marketplace.detail.riskType.dependency");
    }
    case "docs": {
      return t("plugin.linapro-plugin-marketplace.detail.riskType.docs");
    }
    case "dynamic_route": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.riskType.dynamic_route",
      );
    }
    case "external_network": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.riskType.external_network",
      );
    }
    case "host_service": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.riskType.host_service",
      );
    }
    case "install_sql": {
      return t("plugin.linapro-plugin-marketplace.detail.riskType.install_sql");
    }
    case "menu_permission": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.riskType.menu_permission",
      );
    }
    case "mock_sql": {
      return t("plugin.linapro-plugin-marketplace.detail.riskType.mock_sql");
    }
    case "multi_tenant": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.riskType.multi_tenant",
      );
    }
    case "uninstall_sql": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.riskType.uninstall_sql",
      );
    }
  }
}

function formatBytes(value?: number) {
  if (!value || value <= 0) {
    return "-";
  }
  const units = ["B", "KB", "MB", "GB"];
  let size = value;
  let unitIndex = 0;
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex += 1;
  }
  return `${size.toFixed(unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`;
}
</script>

<template>
  <Page :auto-content-height="true">
    <div class="marketplace-detail-shell">
      <div class="marketplace-detail-header">
        <a-button @click="handleBack">
          <template #icon>
            <IconifyIcon icon="ant-design:arrow-left-outlined" />
          </template>
          {{ $t("plugin.linapro-plugin-marketplace.detail.actions.back") }}
        </a-button>

        <div class="marketplace-detail-heading">
          <h2>{{ detail?.name || getRoutePluginId() || "-" }}</h2>
          <p>{{ detail?.summary || "-" }}</p>
        </div>

        <Space v-if="detail" wrap :size="[6, 6]">
          <Tag :color="detail.pluginType === 'source' ? 'blue' : 'green'">
            {{ formatPluginType(detail.pluginType) }}
          </Tag>
          <Tag :color="getStatusColor(detail.marketStatus)">
            {{ formatStatus(detail.marketStatus) }}
          </Tag>
          <Tag>{{ formatVisibility(detail.visibility) }}</Tag>
        </Space>
      </div>

      <Spin :spinning="loading">
        <template v-if="detail">
          <Descriptions
            :column="{ xs: 1, sm: 2 }"
            bordered
            class="marketplace-detail-descriptions"
            size="small"
          >
            <DescriptionsItem
              :label="
                $t('plugin.linapro-plugin-marketplace.detail.fields.pluginId')
              "
            >
              <span class="marketplace-mono">{{ detail.pluginId }}</span>
            </DescriptionsItem>
            <DescriptionsItem
              :label="
                $t('plugin.linapro-plugin-marketplace.detail.fields.publisher')
              "
            >
              {{
                detail.publisher?.name || detail.publisher?.publisherKey || "-"
              }}
            </DescriptionsItem>
            <DescriptionsItem
              :label="
                $t(
                  'plugin.linapro-plugin-marketplace.detail.fields.latestVersion',
                )
              "
            >
              {{ detail.latestVersion || "-" }}
            </DescriptionsItem>
            <DescriptionsItem
              :label="
                $t('plugin.linapro-plugin-marketplace.detail.fields.downloads')
              "
            >
              {{ detail.downloadCount }}
            </DescriptionsItem>
            <DescriptionsItem
              :label="
                $t('plugin.linapro-plugin-marketplace.detail.fields.license')
              "
            >
              {{ detail.license || "-" }}
            </DescriptionsItem>
            <DescriptionsItem
              :label="
                $t('plugin.linapro-plugin-marketplace.detail.fields.updatedAt')
              "
            >
              {{ formatTimestamp(detail.updatedAt) }}
            </DescriptionsItem>
            <DescriptionsItem
              :label="
                $t('plugin.linapro-plugin-marketplace.detail.fields.risk')
              "
              :span="2"
            >
              <Space v-if="hasRiskCounts()" wrap :size="[4, 4]">
                <Tag v-if="getRiskCounts().high > 0" color="error">
                  {{
                    $t("plugin.linapro-plugin-marketplace.catalog.risk.high", {
                      count: getRiskCounts().high,
                    })
                  }}
                </Tag>
                <Tag v-if="getRiskCounts().warning > 0" color="warning">
                  {{
                    $t(
                      "plugin.linapro-plugin-marketplace.catalog.risk.warning",
                      { count: getRiskCounts().warning },
                    )
                  }}
                </Tag>
                <Tag v-if="getRiskCounts().info > 0" color="processing">
                  {{
                    $t("plugin.linapro-plugin-marketplace.catalog.risk.info", {
                      count: getRiskCounts().info,
                    })
                  }}
                </Tag>
              </Space>
              <Tag v-else-if="hasRiskAssessment()" color="success">
                {{ $t("plugin.linapro-plugin-marketplace.catalog.risk.none") }}
              </Tag>
              <Tag v-else>
                {{
                  $t(
                    "plugin.linapro-plugin-marketplace.catalog.risk.unassessed",
                  )
                }}
              </Tag>
            </DescriptionsItem>
          </Descriptions>

          <Tabs v-model:active-key="activeTab" class="marketplace-detail-tabs">
            <TabPane
              key="versions"
              :tab="
                $t('plugin.linapro-plugin-marketplace.detail.tabs.versions')
              "
            >
              <Alert
                v-if="versionLoadState === 'error'"
                show-icon
                type="error"
                :message="
                  $t(
                    'plugin.linapro-plugin-marketplace.detail.errors.versionsLoad',
                  )
                "
              />
              <Empty
                v-else-if="versionLoadState === 'empty'"
                :description="
                  $t('plugin.linapro-plugin-marketplace.detail.empty.versions')
                "
              />
              <VersionGrid
                v-else
                :table-title="
                  $t(
                    'plugin.linapro-plugin-marketplace.detail.versionTableTitle',
                  )
                "
              >
                <template #version="{ row }">
                  <Space direction="vertical" :size="2">
                    <span class="marketplace-version">{{ row.version }}</span>
                    <span class="marketplace-muted">
                      {{ formatPluginType(row.pluginType) }}
                    </span>
                  </Space>
                </template>

                <template #releaseStatus="{ row }">
                  <Tag :color="getStatusColor(row.releaseStatus)">
                    {{ formatStatus(row.releaseStatus) }}
                  </Tag>
                </template>

                <template #reviewStatus="{ row }">
                  <Tag :color="getReviewStatusColor(row.reviewStatus)">
                    {{ formatReviewStatus(row.reviewStatus) }}
                  </Tag>
                </template>

                <template #artifact="{ row }">
                  <Space direction="vertical" :size="2">
                    <Tag>{{
                      formatArtifactType(row.artifact?.artifactType)
                    }}</Tag>
                    <Tooltip :title="row.artifact?.sha256">
                      <span class="marketplace-muted">
                        {{ formatBytes(row.artifact?.sizeBytes) }}
                      </span>
                    </Tooltip>
                  </Space>
                </template>

                <template #compatibility="{ row }">
                  <span class="marketplace-muted">
                    {{ getCompatibilityLabel(row) }}
                  </span>
                </template>

                <template #action="{ row }">
                  <Space :size="[4, 4]" :wrap="true">
                    <ghost-button
                      @click.stop="handleSelectRelease(row, 'docs')"
                    >
                      {{
                        $t(
                          "plugin.linapro-plugin-marketplace.detail.actions.viewDocs",
                        )
                      }}
                    </ghost-button>
                    <ghost-button
                      @click.stop="handleSelectRelease(row, 'risks')"
                    >
                      {{
                        $t(
                          "plugin.linapro-plugin-marketplace.detail.actions.viewRisks",
                        )
                      }}
                    </ghost-button>
                    <ghost-button
                      v-if="canDownloadMarketplacePlugin()"
                      :loading="isDownloading(row)"
                      @click.stop="handleConfirmDownload(row)"
                    >
                      {{
                        $t(
                          "plugin.linapro-plugin-marketplace.detail.actions.download",
                        )
                      }}
                    </ghost-button>
                  </Space>
                </template>
              </VersionGrid>
            </TabPane>

            <TabPane
              key="docs"
              :tab="$t('plugin.linapro-plugin-marketplace.detail.tabs.docs')"
            >
              <Spin :spinning="documentLoading">
                <Alert
                  v-if="documentLoadState === 'error'"
                  show-icon
                  type="error"
                  :message="
                    $t(
                      'plugin.linapro-plugin-marketplace.detail.errors.documentLoad',
                    )
                  "
                />
                <div v-else-if="currentDocument" class="marketplace-doc-panel">
                  <div class="marketplace-doc-title">
                    <h3>
                      {{ currentDocument.title || selectedRelease?.version }}
                    </h3>
                    <Space wrap :size="[6, 6]">
                      <Tag>{{ currentDocument.resolvedLocale }}</Tag>
                      <Tag>{{ currentDocument.path }}</Tag>
                    </Space>
                  </div>
                  <p class="marketplace-muted">{{ currentDocument.summary }}</p>
                  <div
                    class="marketplace-doc-body"
                    v-html="currentDocument.content"
                  ></div>
                </div>
                <Empty
                  v-else
                  :description="
                    $t(
                      documentLoadState === 'idle'
                        ? 'plugin.linapro-plugin-marketplace.detail.empty.selectVersion'
                        : 'plugin.linapro-plugin-marketplace.detail.empty.document',
                    )
                  "
                />
              </Spin>
            </TabPane>

            <TabPane
              key="risks"
              :tab="$t('plugin.linapro-plugin-marketplace.detail.tabs.risks')"
            >
              <Spin :spinning="riskLoading">
                <Alert
                  v-if="riskLoadState === 'error'"
                  show-icon
                  type="error"
                  :message="
                    $t(
                      'plugin.linapro-plugin-marketplace.detail.errors.risksLoad',
                    )
                  "
                />
                <div
                  v-else-if="currentRisks.length > 0"
                  class="marketplace-risk-list"
                >
                  <div
                    v-for="risk in currentRisks"
                    :key="`${risk.type}:${risk.source}:${risk.summary}`"
                    class="marketplace-risk-item"
                  >
                    <Space wrap :size="[6, 6]">
                      <Tag :color="getRiskSeverityColor(risk.severity)">
                        {{ formatRiskSeverity(risk.severity) }}
                      </Tag>
                      <Tag>{{ formatRiskType(risk.type) }}</Tag>
                      <span class="marketplace-muted">{{ risk.source }}</span>
                    </Space>
                    <p>{{ risk.summary }}</p>
                  </div>
                </div>
                <Empty
                  v-else
                  :description="
                    $t(
                      riskLoadState === 'idle'
                        ? 'plugin.linapro-plugin-marketplace.detail.empty.selectVersion'
                        : 'plugin.linapro-plugin-marketplace.detail.empty.risks',
                    )
                  "
                />
              </Spin>
            </TabPane>
          </Tabs>
        </template>

        <Alert
          v-else-if="detailLoadState === 'error'"
          show-icon
          type="error"
          :message="
            $t('plugin.linapro-plugin-marketplace.detail.errors.pluginLoad')
          "
        />
        <Empty
          v-else-if="detailLoadState !== 'loading'"
          :description="
            $t('plugin.linapro-plugin-marketplace.detail.empty.plugin')
          "
        />
      </Spin>
    </div>
  </Page>
</template>

<style scoped>
.marketplace-detail-shell {
  display: flex;
  min-height: 100%;
  flex-direction: column;
  gap: 16px;
}

.marketplace-detail-header {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  gap: 16px;
  align-items: center;
}

.marketplace-detail-heading {
  min-width: 0;
}

.marketplace-detail-heading h2 {
  margin: 0;
  overflow: hidden;
  color: var(--ant-color-text);
  font-size: 20px;
  font-weight: 650;
  line-height: 1.25;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.marketplace-detail-heading p {
  margin: 4px 0 0;
  overflow: hidden;
  color: var(--ant-color-text-secondary);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.marketplace-detail-descriptions {
  background: var(--ant-color-bg-container);
}

.marketplace-detail-tabs {
  min-height: 0;
}

.marketplace-version {
  color: var(--ant-color-text);
  font-weight: 600;
}

.marketplace-muted {
  color: var(--ant-color-text-secondary);
  font-size: 12px;
  line-height: 1.45;
}

.marketplace-mono {
  font-family: var(
    --font-family-mono,
    ui-monospace,
    SFMono-Regular,
    Menlo,
    Monaco,
    Consolas,
    "Liberation Mono",
    "Courier New",
    monospace
  );
}

.marketplace-doc-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.marketplace-doc-title {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  justify-content: space-between;
}

.marketplace-doc-title h3 {
  margin: 0;
  color: var(--ant-color-text);
  font-size: 16px;
  font-weight: 600;
}

.marketplace-doc-body {
  overflow: auto;
  max-height: 520px;
  padding: 14px;
  border: 1px solid var(--ant-color-border-secondary);
  border-radius: 6px;
  background: var(--ant-color-bg-container);
  color: var(--ant-color-text);
  line-height: 1.7;
}

.marketplace-risk-list {
  display: grid;
  gap: 10px;
}

.marketplace-risk-item {
  padding: 12px;
  border: 1px solid var(--ant-color-border-secondary);
  border-radius: 6px;
  background: var(--ant-color-bg-container);
}

.marketplace-risk-item p {
  margin: 8px 0 0;
  color: var(--ant-color-text);
}

:global(.marketplace-download-confirm) {
  display: grid;
  gap: 10px;
}

:global(.marketplace-download-confirm p) {
  margin: 0;
}

:global(.marketplace-download-meta) {
  display: grid;
  grid-template-columns: max-content minmax(0, 1fr);
  gap: 6px 12px;
  margin: 0;
}

:global(.marketplace-download-meta dt) {
  color: var(--ant-color-text-secondary);
}

:global(.marketplace-download-meta dd) {
  min-width: 0;
  margin: 0;
  overflow-wrap: anywhere;
  color: var(--ant-color-text);
}

@media (max-width: 768px) {
  .marketplace-detail-shell {
    padding-right: 32px;
  }

  .marketplace-detail-header {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
