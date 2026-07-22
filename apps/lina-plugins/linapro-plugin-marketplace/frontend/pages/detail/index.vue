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
  MarketplaceProcessStatus,
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
import { preferences } from "@vben/preferences";
import { breakpointsTailwind, useBreakpoints } from "@vueuse/core";

import {
  Alert,
  Descriptions,
  DescriptionsItem,
  Empty,
  Modal,
  Segmented,
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
  marketplaceReleaseDocumentBundle,
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

const props = withDefaults(
  defineProps<{
    /** When true, render as modal/drawer content without Page chrome. */
    embedded?: boolean;
    /** Preferred source list context for API scope and back navigation. */
    from?: string;
    /** Explicit plugin id when opened from a list modal instead of the route. */
    pluginId?: string;
  }>(),
  {
    embedded: false,
    from: "",
    pluginId: "",
  },
);

const emit = defineEmits<{
  close: [];
}>();

const riskEmptyCounts: MarketplaceRiskCounts = {
  high: 0,
  info: 0,
  warning: 0,
};

const route = useRoute();
const router = useRouter();
const breakpoints = useBreakpoints(breakpointsTailwind);
const isMobile = breakpoints.smaller("md");
const activeFrom = computed(() => {
  if (props.from) {
    return props.from;
  }
  return typeof route.query.from === "string" ? route.query.from : "";
});
const readScope = computed<MarketplaceReadScope>(() => {
  if (activeFrom.value === "mine") {
    return "mine";
  }
  if (
    activeFrom.value === "admin-list" ||
    activeFrom.value === "review"
  ) {
    return "managed";
  }
  return "public";
});
const { hasAccessByCodes } = useAccess();

const detail = ref<MarketplacePluginDetailItem | null>(null);
const currentDocument = ref<MarketplaceDocumentItem | null>(null);
const availableDocuments = ref<MarketplaceDocumentItem[]>([]);
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

const activeDocumentLocale = computed(
  () =>
    currentDocument.value?.resolvedLocale ||
    currentDocument.value?.locale ||
    "",
);
const availableDocumentLocaleOptions = computed(() => {
  const seen = new Set<string>();
  return availableDocuments.value
    .map((document) => document.resolvedLocale || document.locale)
    .filter((locale) => {
      if (!locale || seen.has(locale)) {
        return false;
      }
      seen.add(locale);
      return true;
    })
    .map((locale) => ({ label: locale, value: locale }));
});

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
          const pluginId = getActivePluginId();
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
            if (pluginId !== getActivePluginId()) {
              return result;
            }
            if (!selectedRelease.value && result.items[0]) {
              selectedRelease.value = result.items[0];
            }
            versionLoadState.value = result.total > 0 ? "ready" : "empty";
            return result;
          } catch (error) {
            if (pluginId === getActivePluginId()) {
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

function getActivePluginId() {
  if (props.pluginId.trim()) {
    return props.pluginId.trim();
  }
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
  () => [props.pluginId, route.query.pluginId, props.from, route.query.from],
  () => {
    void initializePage();
  },
);

watch(isMobile, () => {
  versionGridApi.setGridOptions({ columns: buildVersionColumns() });
});

async function initializePage() {
  const pluginId = getActivePluginId();
  const requestId = ++pageRequestId;
  releaseContextRequestId += 1;
  detail.value = null;
  currentDocument.value = null;
  availableDocuments.value = [];
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
  if (props.embedded) {
    emit("close");
    return;
  }
  router.push(marketplaceBackPath(activeFrom.value || undefined));
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
  availableDocuments.value = [];
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
    // Pass the active UI locale so marketplace docs can prefer a matching
    // plugin i18n document; the server still applies single-locale / English
    // fallback when the preferred locale is missing.
    const result = await marketplaceReleaseDocumentBundle(
      row.pluginId,
      row.version,
      { locale: preferences.app.locale },
      readScope.value,
    );
    if (!isCurrentReleaseRequest(row, requestId)) {
      return;
    }
    availableDocuments.value =
      result.documents.length > 0
        ? result.documents
        : result.document
          ? [result.document]
          : [];
    currentDocument.value =
      result.document ?? availableDocuments.value[0] ?? null;
    documentLoadState.value = currentDocument.value ? "ready" : "empty";
  } catch (error) {
    if (!isCurrentReleaseRequest(row, requestId)) {
      return;
    }
    currentDocument.value = null;
    availableDocuments.value = [];
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

watch(
  () => preferences.app.locale,
  (locale) => {
    applyPreferredDocumentLocale(locale);
  },
);

function applyPreferredDocumentLocale(locale: string) {
  const document = chooseDocumentFromBundle(locale);
  if (document) {
    currentDocument.value = document;
  }
}

function chooseDocumentFromBundle(locale: string) {
  const documents = availableDocuments.value;
  if (documents.length === 0) {
    return null;
  }
  if (documents.length === 1) {
    return documents[0];
  }
  const preferred = locale.trim().toLowerCase();
  if (preferred) {
    const exact = documents.find(
      (document) =>
        (document.resolvedLocale || document.locale).toLowerCase() === preferred,
    );
    if (exact) {
      return exact;
    }
  }
  return (
    documents.find((document) =>
      ["en-us", "en"].includes(
        (document.resolvedLocale || document.locale).toLowerCase(),
      ),
    ) ?? documents[0]
  );
}

function handleSelectDocumentLocale(value: string | number) {
  const locale = String(value);
  const document = availableDocuments.value.find(
    (item) => (item.resolvedLocale || item.locale) === locale,
  );
  if (document) {
    currentDocument.value = document;
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

/** Show pinned git coordinates so historical installs stay inspectable. */
function formatReleaseSourcePin(row?: MarketplaceReleaseItem | null) {
  if (!row) {
    return "";
  }
  const commit = (row.sourceCommit || "").trim();
  const ref = (row.sourceRef || "").trim();
  if (!commit && !ref) {
    return "";
  }
  const shortCommit =
    commit.length > 12 ? `${commit.slice(0, 12)}…` : commit;
  if (commit && ref) {
    return t("plugin.linapro-plugin-marketplace.detail.sourcePin.refAndCommit", {
      ref,
      commit: shortCommit,
    });
  }
  if (commit) {
    return t("plugin.linapro-plugin-marketplace.detail.sourcePin.commitOnly", {
      commit: shortCommit,
    });
  }
  return t("plugin.linapro-plugin-marketplace.detail.sourcePin.refOnly", {
    ref,
  });
}

function formatStatus(
  status: MarketplaceStatus,
  processStatus?: MarketplaceProcessStatus | string,
) {
  if (status === "published") {
    return t("plugin.linapro-plugin-marketplace.detail.status.published");
  }
  if (status === "delisted") {
    return t("plugin.linapro-plugin-marketplace.detail.status.delisted");
  }
  if (status === "deprecated") {
    return t("plugin.linapro-plugin-marketplace.detail.status.deprecated");
  }
  switch (processStatus) {
    case "pending_verify": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.processStatus.pendingVerify",
      );
    }
    case "pending_review": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.processStatus.pendingReview",
      );
    }
    case "failed": {
      return t("plugin.linapro-plugin-marketplace.detail.processStatus.failed");
    }
    case "completed": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.processStatus.completed",
      );
    }
    default: {
      return t("plugin.linapro-plugin-marketplace.detail.status.draft");
    }
  }
}

function getStatusColor(
  status: MarketplaceStatus,
  processStatus?: MarketplaceProcessStatus | string,
) {
  if (status === "published") {
    return "success";
  }
  if (status === "deprecated") {
    return "warning";
  }
  if (status === "delisted") {
    return "default";
  }
  switch (processStatus) {
    case "pending_verify": {
      return "processing";
    }
    case "pending_review": {
      return "gold";
    }
    case "failed": {
      return "error";
    }
    default: {
      return "processing";
    }
  }
}

function formatSourceKind(sourceKind?: string) {
  if (sourceKind === "git") {
    return t("plugin.linapro-plugin-marketplace.mine.sourceKind.git");
  }
  if (sourceKind === "upload") {
    return t("plugin.linapro-plugin-marketplace.mine.sourceKind.upload");
  }
  // Avoid treating missing projections as upload packages.
  return sourceKind?.trim() || "-";
}

function getSourceKindColor(sourceKind?: string) {
  return sourceKind === "git" ? "geekblue" : "default";
}

function processPipelineMessage() {
  if (!detail.value || detail.value.marketStatus === "published") {
    return "";
  }
  switch (detail.value.processStatus) {
    case "pending_verify": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.pipeline.pendingVerifyHint",
      );
    }
    case "pending_review": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.pipeline.pendingReviewHint",
      );
    }
    case "failed": {
      return (
        detail.value.lastSyncMessage ||
        t("plugin.linapro-plugin-marketplace.detail.pipeline.failedHint")
      );
    }
    default: {
      return "";
    }
  }
}

function versionsEmptyDescription() {
  switch (detail.value?.processStatus) {
    case "pending_verify": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.empty.versionsPendingVerify",
      );
    }
    case "pending_review": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.empty.versionsPendingReview",
      );
    }
    case "failed": {
      return t(
        "plugin.linapro-plugin-marketplace.detail.empty.versionsFailed",
      );
    }
    default: {
      return t("plugin.linapro-plugin-marketplace.detail.empty.versions");
    }
  }
}

function processPipelineAlertType() {
  if (detail.value?.processStatus === "failed") {
    return "error";
  }
  if (detail.value?.processStatus === "pending_review") {
    return "warning";
  }
  return "info";
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
  <component
    :is="embedded ? 'div' : Page"
    v-bind="
      embedded
        ? {
            class: 'marketplace-detail-modal-root',
            'data-testid': 'marketplace-detail-modal',
          }
        : { autoContentHeight: true }
    "
  >
    <div
      class="marketplace-detail-shell"
      :class="{ 'marketplace-detail-shell--embedded': embedded }"
    >
      <!-- Page mode keeps the full chrome; modal mode relies on dialog title/close. -->
      <div v-if="!embedded" class="marketplace-detail-header">
        <a-button @click="handleBack">
          <template #icon>
            <IconifyIcon icon="ant-design:arrow-left-outlined" />
          </template>
          {{ $t("plugin.linapro-plugin-marketplace.detail.actions.back") }}
        </a-button>

        <div class="marketplace-detail-heading">
          <h2>{{ detail?.name || getActivePluginId() || "-" }}</h2>
          <p>{{ detail?.summary || "-" }}</p>
        </div>

        <Space v-if="detail" wrap :size="[6, 6]">
          <Tag :color="detail.pluginType === 'source' ? 'blue' : 'green'">
            {{ formatPluginType(detail.pluginType) }}
          </Tag>
          <Tag
            :color="
              getStatusColor(detail.marketStatus, detail.processStatus)
            "
          >
            {{ formatStatus(detail.marketStatus, detail.processStatus) }}
          </Tag>
          <Tag>{{ formatVisibility(detail.visibility) }}</Tag>
        </Space>
      </div>

      <Spin :spinning="loading">
        <template v-if="detail">
          <div class="marketplace-detail-body">
            <div v-if="embedded" class="marketplace-detail-embedded-heading">
              <div class="marketplace-detail-embedded-title-row">
                <h3 class="marketplace-detail-embedded-title">
                  {{ detail.name || getActivePluginId() || "-" }}
                </h3>
                <Space wrap :size="[6, 6]">
                  <Tag :color="detail.pluginType === 'source' ? 'blue' : 'green'">
                    {{ formatPluginType(detail.pluginType) }}
                  </Tag>
                  <Tag
                    :color="
                      getStatusColor(detail.marketStatus, detail.processStatus)
                    "
                  >
                    {{ formatStatus(detail.marketStatus, detail.processStatus) }}
                  </Tag>
                  <Tag
                    :color="detail.visibility === 'public' ? 'green' : 'default'"
                  >
                    {{ formatVisibility(detail.visibility) }}
                  </Tag>
                  <Tag
                    v-if="detail.sourceKind"
                    :color="getSourceKindColor(detail.sourceKind)"
                  >
                    {{ formatSourceKind(detail.sourceKind) }}
                  </Tag>
                </Space>
              </div>
              <p v-if="detail.summary" class="marketplace-detail-embedded-summary">
                {{ detail.summary }}
              </p>
            </div>

            <Alert
              v-if="embedded && processPipelineMessage()"
              show-icon
              class="marketplace-detail-pipeline-alert"
              :type="processPipelineAlertType()"
              :message="processPipelineMessage()"
            />

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
                  detail.publisher?.name ||
                  detail.publisher?.publisherKey ||
                  "-"
                }}
              </DescriptionsItem>
              <DescriptionsItem
                :label="
                  $t('plugin.linapro-plugin-marketplace.detail.fields.status')
                "
              >
                <Tag
                  :color="
                    getStatusColor(detail.marketStatus, detail.processStatus)
                  "
                >
                  {{ formatStatus(detail.marketStatus, detail.processStatus) }}
                </Tag>
              </DescriptionsItem>
              <DescriptionsItem
                :label="
                  $t(
                    'plugin.linapro-plugin-marketplace.detail.fields.sourceKind',
                  )
                "
              >
                <Tag :color="getSourceKindColor(detail.sourceKind)">
                  {{ formatSourceKind(detail.sourceKind) }}
                </Tag>
              </DescriptionsItem>
              <DescriptionsItem
                :label="
                  $t(
                    'plugin.linapro-plugin-marketplace.detail.fields.latestVersion',
                  )
                "
              >
                <span class="marketplace-mono">
                  {{ detail.latestVersion || "-" }}
                </span>
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
                v-if="detail.sourceKind === 'git' && detail.repoUrl"
                :label="
                  $t('plugin.linapro-plugin-marketplace.detail.fields.repoUrl')
                "
                :span="2"
              >
                <a
                  class="marketplace-link"
                  :href="detail.repoUrl"
                  rel="noopener noreferrer"
                  target="_blank"
                >
                  {{ detail.repoUrl }}
                </a>
              </DescriptionsItem>
              <DescriptionsItem
                v-if="detail.lastSyncMessage"
                :label="
                  $t(
                    'plugin.linapro-plugin-marketplace.detail.fields.lastSyncMessage',
                  )
                "
                :span="2"
              >
                <span class="marketplace-muted">{{ detail.lastSyncMessage }}</span>
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

            <Tabs
              v-model:active-key="activeTab"
              class="marketplace-detail-tabs"
              :class="{ 'marketplace-detail-tabs--embedded': embedded }"
            >
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
                :description="versionsEmptyDescription()"
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
                    <span
                      v-if="formatReleaseSourcePin(row)"
                      class="marketplace-muted marketplace-source-pin"
                      :title="formatReleaseSourcePin(row)"
                    >
                      {{ formatReleaseSourcePin(row) }}
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
                      <Segmented
                        v-if="availableDocumentLocaleOptions.length > 1"
                        :aria-label="
                          $t(
                            'plugin.linapro-plugin-marketplace.detail.docs.locale',
                          )
                        "
                        :options="availableDocumentLocaleOptions"
                        size="small"
                        :value="activeDocumentLocale"
                        @change="handleSelectDocumentLocale"
                      />
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
          </div>
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
  </component>
</template>

<style scoped>
.marketplace-detail-modal-root {
  min-height: 0;
}

.marketplace-detail-shell {
  display: flex;
  min-height: 100%;
  flex-direction: column;
  gap: 16px;
}

.marketplace-detail-shell--embedded {
  min-height: 0;
  max-height: min(72vh, 820px);
  overflow: auto;
  padding-right: 2px;
}

.marketplace-detail-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
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

.marketplace-detail-embedded-heading {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.marketplace-detail-embedded-title-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px 12px;
}

.marketplace-detail-embedded-title {
  margin: 0;
  min-width: 0;
  color: var(--ant-color-text);
  font-size: 18px;
  font-weight: 650;
  line-height: 1.3;
}

.marketplace-detail-embedded-summary {
  margin: 0;
  color: var(--ant-color-text-secondary);
  font-size: 13px;
  line-height: 1.5;
}

.marketplace-detail-pipeline-alert {
  margin: 0;
}

.marketplace-detail-descriptions {
  background: var(--ant-color-bg-container);
}

.marketplace-detail-tabs {
  min-height: 0;
}

.marketplace-detail-tabs--embedded :deep(.ant-tabs-content-holder) {
  min-height: 180px;
}

.marketplace-link {
  color: var(--ant-color-primary);
  word-break: break-all;
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

.marketplace-source-pin {
  display: block;
  max-width: 100%;
  overflow: hidden;
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
  text-overflow: ellipsis;
  white-space: nowrap;
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
