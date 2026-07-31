<script lang="ts">
import type { PluginPageMeta } from "#/plugins/page-registry";

export const pluginPageMeta = {
  pluginId: "linapro-plugin-marketplace",
  routePath: "plugin-marketplace-review",
  title: "Marketplace Review",
} satisfies PluginPageMeta;
</script>

<script setup lang="ts">
import type { VbenFormProps } from "@vben/common-ui";

import type { VbenFormSchema } from "#/adapter/form";
import type { VxeGridProps } from "#/adapter/vxe-table";

import type {
  MarketplaceDocumentItem,
  MarketplaceReleaseReviewPayload,
  MarketplaceReviewQueueItem,
  MarketplaceReviewStatus,
  MarketplaceRiskItem,
  MarketplaceRiskSeverity,
  MarketplaceRiskType,
} from "../../types/marketplace";

import { nextTick, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";

import { Page, useVbenDrawer, useVbenModal } from "@vben/common-ui";
import { preferences } from "@vben/preferences";

import {
  Alert,
  Button,
  Descriptions,
  DescriptionsItem,
  Spin,
  Tag,
  message,
} from "ant-design-vue";

import { useVbenForm } from "#/adapter/form";
import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { $t } from "#/locales";
import { formatTimestamp } from "#/utils/time";

import {
  marketplaceReleaseDocument,
  marketplaceReleaseReview,
  marketplaceReleaseRisks,
  marketplaceReviewQueueList,
} from "../../api/marketplace";
import {
  formatMarketplaceRiskDisposition,
  filterMarketplaceRiskFindingsActionable,
  formatMarketplaceRiskFindingGuidance,
  formatMarketplaceRiskFindingSummary,
  marketplaceRiskBlocking,
  marketplaceRiskDisposition,
  marketplaceRiskEvidence,
  marketplaceRiskHasEvidence,
} from "../../utils/risk";
import DetailModalContent from "../detail/detail-modal.vue";

type GridPageInfo = {
  currentPage: number;
  pageSize: number;
};

type ReviewFormValues = {
  keyword?: string;
  pluginId?: string;
  reviewStatus?: MarketplaceReviewStatus;
};

type ReviewGridOptions = NonNullable<
  VxeGridProps<ReviewQueueRow>["gridOptions"]
>;

type ReviewQueueRow = MarketplaceReviewQueueItem & {
  reviewKey: string;
};

const route = useRoute();

const [DetailModal, detailModalApi] = useVbenModal({
  connectedComponent: DetailModalContent,
});

const selectedRelease = ref<MarketplaceReviewQueueItem | null>(null);
const reviewRisks = ref<MarketplaceRiskItem[]>([]);
const expandedReviewRiskKeys = ref<Record<string, boolean>>({});
const reviewDocument = ref<MarketplaceDocumentItem | null>(null);
const reviewLoading = ref(false);
const reviewRisksReady = ref(false);
const reviewRiskLoadFailed = ref(false);
let inspectionRequestId = 0;

const [Grid, gridApi] = useVbenVxeGrid<ReviewQueueRow>({
  formOptions: {
    commonConfig: {
      componentProps: { allowClear: true },
      labelWidth: 80,
    },
    schema: [],
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4",
  },
  gridOptions: {
    columns: [],
    height: "auto",
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      autoLoad: false,
      ajax: {
        query: async (
          { page }: { page: GridPageInfo },
          formValues: ReviewFormValues = {},
        ) => {
          const result = await marketplaceReviewQueueList({
            keyword: trimOptional(formValues.keyword),
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            pluginId: trimOptional(formValues.pluginId),
            reviewStatus: formValues.reviewStatus,
          });
          return {
            ...result,
            items: result.items.map((item) => ({
              ...item,
              reviewKey: getReleaseKey(item),
            })),
          };
        },
      },
    },
    rowConfig: { keyField: "reviewKey" },
    showOverflow: "ellipsis",
    id: "plugin-marketplace-review-queue",
  },
});

const [DecisionForm, decisionFormApi] = useVbenForm({
  commonConfig: {
    componentProps: { class: "w-full" },
    labelWidth: 112,
  },
  schema: [],
  showDefaultActions: false,
  wrapperClass: "grid-cols-1",
});

const [ReviewDrawer, reviewDrawerApi] = useVbenDrawer({
  onClosed: resetInspection,
  onConfirm: handleSubmitDecision,
});

function t(key: string, params?: Record<string, number | string>) {
  return params ? $t(key, params) : $t(key);
}

/** Localize scanner finding body text via payload.code; English summary is fallback. */
function formatRiskFindingSummary(risk: MarketplaceRiskItem) {
  return formatMarketplaceRiskFindingSummary(t, risk);
}

function reviewRiskKey(risk: MarketplaceRiskItem) {
  const code =
    typeof risk.payload?.code === "string" ? risk.payload.code : risk.summary;
  return `${risk.type}:${risk.source}:${code}:${risk.severity}`;
}

function isReviewRiskExpanded(risk: MarketplaceRiskItem) {
  return !!expandedReviewRiskKeys.value[reviewRiskKey(risk)];
}

function toggleReviewRiskExpanded(risk: MarketplaceRiskItem) {
  const key = reviewRiskKey(risk);
  expandedReviewRiskKeys.value = {
    ...expandedReviewRiskKeys.value,
    [key]: !expandedReviewRiskKeys.value[key],
  };
}

function formatReviewRiskDisposition(risk: MarketplaceRiskItem) {
  return formatMarketplaceRiskDisposition(t, marketplaceRiskDisposition(risk));
}

function getReviewDispositionColor(risk: MarketplaceRiskItem) {
  switch (marketplaceRiskDisposition(risk)) {
    case "need_fix": {
      return "error";
    }
    case "need_attention": {
      return "warning";
    }
    default: {
      return "default";
    }
  }
}

function reviewRiskReason(risk: MarketplaceRiskItem) {
  return formatMarketplaceRiskFindingGuidance(t, risk, "reason");
}

function reviewRiskRemediation(risk: MarketplaceRiskItem) {
  return formatMarketplaceRiskFindingGuidance(t, risk, "remediation");
}

function reviewRiskAcceptance(risk: MarketplaceRiskItem) {
  return formatMarketplaceRiskFindingGuidance(t, risk, "acceptance");
}

function reviewRiskHasEvidence(risk: MarketplaceRiskItem) {
  return marketplaceRiskHasEvidence(marketplaceRiskEvidence(risk));
}

function reviewRiskEvidence(risk: MarketplaceRiskItem) {
  return marketplaceRiskEvidence(risk);
}

function isReviewRiskBlocking(risk: MarketplaceRiskItem) {
  return marketplaceRiskBlocking(risk);
}

function trimOptional(value?: string) {
  const normalized = value?.trim();
  return normalized || undefined;
}

function buildFormOptions(): VbenFormProps {
  return {
    commonConfig: {
      componentProps: { allowClear: true },
      labelWidth: 80,
    },
    schema: [
      {
        component: "Input",
        fieldName: "keyword",
        label: t("plugin.linapro-plugin-marketplace.catalog.fields.keyword"),
      },
      {
        component: "Input",
        fieldName: "pluginId",
        label: t("plugin.linapro-plugin-marketplace.detail.fields.pluginId"),
      },
      {
        component: "Select",
        componentProps: {
          options: [
            {
              label: t(
                "plugin.linapro-plugin-marketplace.detail.reviewStatus.submitted",
              ),
              value: "submitted",
            },
            {
              label: t(
                "plugin.linapro-plugin-marketplace.detail.reviewStatus.reviewing",
              ),
              value: "reviewing",
            },
            {
              label: t(
                "plugin.linapro-plugin-marketplace.detail.reviewStatus.approved",
              ),
              value: "approved",
            },
            {
              label: t(
                "plugin.linapro-plugin-marketplace.detail.reviewStatus.rejected",
              ),
              value: "rejected",
            },
          ],
        },
        fieldName: "reviewStatus",
        label: t(
          "plugin.linapro-plugin-marketplace.detail.columns.reviewStatus",
        ),
      },
    ],
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4",
  };
}

function buildColumns(): ReviewGridOptions["columns"] {
  // Dense layout aligned with My Plugins: keep review status early (ops-critical),
  // leave plugin name flexible, and keep Inspect fixed on the right.
  return [
    {
      align: "left",
      field: "pluginId",
      headerAlign: "center",
      minWidth: 176,
      showOverflow: "ellipsis",
      slots: { default: "plugin" },
      title: t("plugin.linapro-plugin-marketplace.catalog.columns.plugin"),
    },
    {
      field: "reviewStatus",
      showOverflow: "ellipsis",
      slots: { default: "reviewStatus" },
      title: t("plugin.linapro-plugin-marketplace.detail.columns.reviewStatus"),
      // Wide enough for en-US "Submitted" / zh-CN "已提交" tags.
      width: 116,
    },
    {
      className: "review-submitted-at-column",
      field: "submittedAt",
      formatter: ({ cellValue }: { cellValue?: null | number | string }) =>
        formatTimestamp(cellValue),
      title: t("plugin.linapro-plugin-marketplace.review.columns.submittedAt"),
      width: 184,
    },
    {
      field: "version",
      showOverflow: "ellipsis",
      title: t("plugin.linapro-plugin-marketplace.detail.columns.version"),
      width: 92,
    },
    {
      field: "pluginType",
      slots: { default: "pluginType" },
      title: t("plugin.linapro-plugin-marketplace.catalog.columns.type"),
      width: 96,
    },
    {
      field: "action",
      fixed: "right",
      slots: { default: "action" },
      title: t("plugin.linapro-plugin-marketplace.detail.columns.actions"),
      // Inspect ghost button (en-US "Inspect" / zh-CN "检查").
      width: 88,
    },
  ];
}

function buildDecisionSchema(): VbenFormSchema[] {
  return [
    {
      component: "Select",
      componentProps: {
        options: [
          {
            label: t(
              "plugin.linapro-plugin-marketplace.detail.reviewStatus.approved",
            ),
            value: "approved",
          },
          {
            label: t(
              "plugin.linapro-plugin-marketplace.detail.reviewStatus.rejected",
            ),
            value: "rejected",
          },
        ],
      },
      fieldName: "reviewStatus",
      label: t("plugin.linapro-plugin-marketplace.detail.columns.reviewStatus"),
      rules: "required",
    },
    {
      component: "Textarea",
      fieldName: "message",
      label: t(
        "plugin.linapro-plugin-marketplace.console.fields.reviewMessage",
      ),
    },
  ];
}

onMounted(async () => {
  gridApi.setState({ formOptions: buildFormOptions() });
  gridApi.setGridOptions({
    columns: buildColumns(),
    emptyText: t("plugin.linapro-plugin-marketplace.console.empty.queue"),
  });
  decisionFormApi.setState({ schema: buildDecisionSchema() });
  await gridApi.reload();
  openDetailFromRouteQuery();
});

watch(
  () => [route.query.view, route.query.pluginId] as const,
  () => {
    openDetailFromRouteQuery();
  },
);

// Rebuild chrome and re-query when the workbench language changes so
// headers/filters/decision form match the active locale.
watch(
  () => preferences.app.locale,
  async (locale, previousLocale) => {
    if (!previousLocale || locale === previousLocale) {
      return;
    }
    gridApi.setState({ formOptions: buildFormOptions() });
    gridApi.setGridOptions({
      columns: buildColumns(),
      emptyText: t("plugin.linapro-plugin-marketplace.console.empty.queue"),
    });
    decisionFormApi.setState({ schema: buildDecisionSchema() });
    await gridApi.query();
  },
);

function openDetailFromRouteQuery() {
  if (route.query.view !== "detail") {
    return;
  }
  const pluginId =
    typeof route.query.pluginId === "string" ? route.query.pluginId.trim() : "";
  if (!pluginId) {
    return;
  }
  detailModalApi.setData({ from: "review", pluginId });
  detailModalApi.open();
}

function formatPluginType(type: string) {
  return type === "source"
    ? t("plugin.linapro-plugin-marketplace.catalog.pluginType.source")
    : t("plugin.linapro-plugin-marketplace.catalog.pluginType.dynamic");
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
    default: {
      return status;
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
    case "draft": {
      return "default";
    }
    default: {
      return "default";
    }
  }
}

function hasRiskType(types: MarketplaceRiskType[]) {
  const set = new Set(reviewRisks.value.map((item) => item.type));
  return types.some((type) => set.has(type));
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

function getAuditStatusColor(passed: boolean, evaluated = true) {
  if (!evaluated) {
    return "default";
  }
  return passed ? "success" : "warning";
}

function getAuditStatusLabel(passed: boolean, evaluated = true) {
  if (!evaluated) {
    return t("plugin.linapro-plugin-marketplace.console.audit.unavailable");
  }
  return passed
    ? t("plugin.linapro-plugin-marketplace.console.audit.passed")
    : t("plugin.linapro-plugin-marketplace.console.audit.needsAttention");
}

async function handleInspect(row: MarketplaceReviewQueueItem) {
  const requestId = ++inspectionRequestId;
  selectedRelease.value = row;
  reviewRisks.value = [];
  expandedReviewRiskKeys.value = {};
  reviewDocument.value = null;
  reviewRisksReady.value = false;
  reviewRiskLoadFailed.value = false;
  reviewLoading.value = true;
  reviewDrawerApi.open();
  await nextTick();
  await decisionFormApi.resetForm();

  const [riskResult, documentResult] = await Promise.allSettled([
    marketplaceReleaseRisks(
      row.pluginId,
      row.version,
      {
        pageNum: 1,
        pageSize: 100,
      },
      "managed",
    ),
    marketplaceReleaseDocument(
      row.pluginId,
      row.version,
      { locale: preferences.app.locale },
      "managed",
    ),
  ]);

  if (requestId !== inspectionRequestId) {
    return;
  }

  if (riskResult.status === "fulfilled") {
    // Hide info_only tips; keep need_fix / need_attention for reviewers.
    reviewRisks.value = filterMarketplaceRiskFindingsActionable(
      riskResult.value.items,
    );
    expandedReviewRiskKeys.value = {};
    for (const risk of reviewRisks.value) {
      if (
        marketplaceRiskBlocking(risk) ||
        marketplaceRiskDisposition(risk) === "need_fix"
      ) {
        expandedReviewRiskKeys.value[reviewRiskKey(risk)] = true;
      }
    }
    reviewRisksReady.value = true;
  } else {
    reviewRiskLoadFailed.value = true;
  }
  reviewDocument.value =
    documentResult.status === "fulfilled" ? documentResult.value : null;
  reviewLoading.value = false;
}

async function handleSubmitDecision() {
  const release = selectedRelease.value;
  if (!release) {
    message.warning(
      t("plugin.linapro-plugin-marketplace.console.messages.releaseRequired"),
    );
    return;
  }
  try {
    reviewDrawerApi.lock();
    const { valid } = await decisionFormApi.validate();
    if (!valid) {
      return;
    }
    const values = await decisionFormApi.getValues();
    const result = await marketplaceReleaseReview(
      release.pluginId,
      release.version,
      {
        message:
          typeof values.message === "string" ? values.message.trim() : "",
        reviewStatus: (values.reviewStatus === "rejected"
          ? "rejected"
          : "approved") as MarketplaceReleaseReviewPayload["reviewStatus"],
      },
    );
    selectedRelease.value = {
      ...release,
      reviewMessage: result.release.reviewMessage,
      reviewStatus: result.release.reviewStatus,
    };
    message.success(
      t("plugin.linapro-plugin-marketplace.console.messages.reviewSaved"),
    );
    await gridApi.query();
    await reviewDrawerApi.close();
  } catch {
    // Form/API feedback is handled by adapters.
  } finally {
    reviewDrawerApi.unlock();
  }
}

function resetInspection() {
  inspectionRequestId += 1;
  selectedRelease.value = null;
  reviewRisks.value = [];
  expandedReviewRiskKeys.value = {};
  reviewDocument.value = null;
  reviewLoading.value = false;
  reviewRisksReady.value = false;
  reviewRiskLoadFailed.value = false;
  void decisionFormApi.resetForm();
}

function getReleaseKey(row: MarketplaceReviewQueueItem) {
  return `${row.pluginId}:${row.version}`;
}

function getReviewDrawerTitle() {
  const title = t("plugin.linapro-plugin-marketplace.console.sections.audit");
  if (!selectedRelease.value) {
    return title;
  }
  return `${title} - ${selectedRelease.value.pluginId} ${selectedRelease.value.version}`;
}
</script>

<template>
  <Page :auto-content-height="true">
    <DetailModal />
    <Grid
      class="plugin-marketplace-review"
      :table-title="$t('plugin.linapro-plugin-marketplace.review.tableTitle')"
    >
      <template #plugin="{ row }">
        <div class="review-plugin-cell">
          <span class="review-plugin-name">{{ row.pluginName || "-" }}</span>
          <span class="review-plugin-id">{{ row.pluginId }}</span>
        </div>
      </template>
      <template #pluginType="{ row }">
        <Tag :color="row.pluginType === 'source' ? 'blue' : 'green'">
          {{ formatPluginType(row.pluginType) }}
        </Tag>
      </template>
      <template #reviewStatus="{ row }">
        <Tag :color="getReviewStatusColor(row.reviewStatus)">
          {{ formatReviewStatus(row.reviewStatus) }}
        </Tag>
      </template>
      <template #action="{ row }">
        <ghost-button @click.stop="handleInspect(row)">
          {{ $t("plugin.linapro-plugin-marketplace.console.actions.inspect") }}
        </ghost-button>
      </template>
    </Grid>

    <ReviewDrawer
      class="w-[720px] max-w-[calc(100vw-16px)] sm:max-w-[calc(100vw-32px)]"
      :confirm-text="
        $t('plugin.linapro-plugin-marketplace.console.actions.submitDecision')
      "
      :show-confirm-button="
        !!selectedRelease && !reviewLoading && !reviewRiskLoadFailed
      "
      :title="getReviewDrawerTitle()"
    >
      <div class="marketplace-review-drawer-content">
        <Spin :spinning="reviewLoading">
          <template v-if="selectedRelease">
            <div class="marketplace-review-inspection">
              <Descriptions :column="1" bordered size="small">
                <DescriptionsItem
                  :label="
                    $t(
                      'plugin.linapro-plugin-marketplace.detail.fields.pluginId',
                    )
                  "
                >
                  {{ selectedRelease.pluginId }}
                </DescriptionsItem>
                <DescriptionsItem
                  :label="
                    $t(
                      'plugin.linapro-plugin-marketplace.detail.columns.version',
                    )
                  "
                >
                  {{ selectedRelease.version }}
                </DescriptionsItem>
                <DescriptionsItem
                  :label="
                    $t(
                      'plugin.linapro-plugin-marketplace.detail.columns.artifact',
                    )
                  "
                >
                  {{ selectedRelease.artifact?.fileName || "-" }}
                </DescriptionsItem>
              </Descriptions>

              <div class="marketplace-audit-grid">
                <div class="marketplace-audit-item">
                  <span>
                    {{
                      $t(
                        "plugin.linapro-plugin-marketplace.console.audit.manifest",
                      )
                    }}
                  </span>
                  <Tag :color="getAuditStatusColor(!!selectedRelease.artifact)">
                    {{ getAuditStatusLabel(!!selectedRelease.artifact) }}
                  </Tag>
                </div>
                <div class="marketplace-audit-item">
                  <span>
                    {{
                      $t("plugin.linapro-plugin-marketplace.console.audit.sql")
                    }}
                  </span>
                  <Tag
                    :color="
                      getAuditStatusColor(
                        !hasRiskType([
                          'install_sql',
                          'uninstall_sql',
                          'mock_sql',
                        ]),
                        reviewRisksReady,
                      )
                    "
                  >
                    {{
                      getAuditStatusLabel(
                        !hasRiskType([
                          "install_sql",
                          "uninstall_sql",
                          "mock_sql",
                        ]),
                        reviewRisksReady,
                      )
                    }}
                  </Tag>
                </div>
                <div class="marketplace-audit-item">
                  <span>
                    {{
                      $t(
                        "plugin.linapro-plugin-marketplace.console.audit.hostServices",
                      )
                    }}
                  </span>
                  <Tag
                    :color="
                      getAuditStatusColor(
                        !hasRiskType(['host_service']),
                        reviewRisksReady,
                      )
                    "
                  >
                    {{
                      getAuditStatusLabel(
                        !hasRiskType(["host_service"]),
                        reviewRisksReady,
                      )
                    }}
                  </Tag>
                </div>
                <div class="marketplace-audit-item">
                  <span>
                    {{
                      $t("plugin.linapro-plugin-marketplace.console.audit.docs")
                    }}
                  </span>
                  <Tag :color="getAuditStatusColor(!!reviewDocument)">
                    {{ getAuditStatusLabel(!!reviewDocument) }}
                  </Tag>
                </div>
              </div>

              <Alert
                v-if="reviewRisks.length > 0"
                show-icon
                type="warning"
                :message="
                  $t(
                    'plugin.linapro-plugin-marketplace.console.audit.riskCount',
                    { count: reviewRisks.length },
                  )
                "
              />

              <div
                v-if="reviewRisks.length > 0"
                class="marketplace-review-risk-list"
              >
                <div
                  v-for="risk in reviewRisks"
                  :key="reviewRiskKey(risk)"
                  class="marketplace-review-risk-item"
                  :class="{
                    'marketplace-review-risk-item--blocking':
                      isReviewRiskBlocking(risk),
                  }"
                >
                  <div class="marketplace-review-risk-item-header">
                    <div class="marketplace-review-risk-meta">
                      <Tag v-if="isReviewRiskBlocking(risk)" color="error">
                        {{
                          $t(
                            'plugin.linapro-plugin-marketplace.detail.riskGuide.blockingTag',
                          )
                        }}
                      </Tag>
                      <Tag :color="getReviewDispositionColor(risk)">
                        {{ formatReviewRiskDisposition(risk) }}
                      </Tag>
                      <Tag :color="getRiskSeverityColor(risk.severity)">
                        {{ formatRiskSeverity(risk.severity) }}
                      </Tag>
                      <Tag>{{ formatRiskType(risk.type) }}</Tag>
                      <span>{{ risk.source }}</span>
                    </div>
                    <Button
                      type="link"
                      size="small"
                      class="marketplace-review-risk-toggle"
                      @click="toggleReviewRiskExpanded(risk)"
                    >
                      {{
                        isReviewRiskExpanded(risk)
                          ? $t(
                              'plugin.linapro-plugin-marketplace.detail.riskGuide.collapse',
                            )
                          : $t(
                              'plugin.linapro-plugin-marketplace.detail.riskGuide.expand',
                            )
                      }}
                    </Button>
                  </div>
                  <p>{{ formatRiskFindingSummary(risk) }}</p>
                  <div
                    v-if="isReviewRiskExpanded(risk)"
                    class="marketplace-review-risk-guidance"
                  >
                    <div
                      v-if="reviewRiskReason(risk)"
                      class="marketplace-review-risk-guidance-section"
                    >
                      <div class="marketplace-review-risk-guidance-label">
                        {{
                          $t(
                            'plugin.linapro-plugin-marketplace.detail.riskGuide.reason',
                          )
                        }}
                      </div>
                      <p>{{ reviewRiskReason(risk) }}</p>
                    </div>
                    <div
                      v-if="reviewRiskRemediation(risk)"
                      class="marketplace-review-risk-guidance-section"
                    >
                      <div class="marketplace-review-risk-guidance-label">
                        {{
                          $t(
                            'plugin.linapro-plugin-marketplace.detail.riskGuide.remediation',
                          )
                        }}
                      </div>
                      <p>{{ reviewRiskRemediation(risk) }}</p>
                    </div>
                    <div
                      v-if="reviewRiskAcceptance(risk)"
                      class="marketplace-review-risk-guidance-section"
                    >
                      <div class="marketplace-review-risk-guidance-label">
                        {{
                          $t(
                            'plugin.linapro-plugin-marketplace.detail.riskGuide.acceptance',
                          )
                        }}
                      </div>
                      <p>{{ reviewRiskAcceptance(risk) }}</p>
                    </div>
                    <div
                      v-if="reviewRiskHasEvidence(risk)"
                      class="marketplace-review-risk-guidance-section marketplace-review-risk-evidence"
                    >
                      <div class="marketplace-review-risk-guidance-label">
                        {{
                          $t(
                            'plugin.linapro-plugin-marketplace.detail.riskGuide.evidence',
                          )
                        }}
                      </div>
                      <ul v-if="reviewRiskEvidence(risk).files.length > 0">
                        <li
                          v-for="file in reviewRiskEvidence(risk).files"
                          :key="file"
                        >
                          <code>{{ file }}</code>
                        </li>
                      </ul>
                      <ul v-if="reviewRiskEvidence(risk).services.length > 0">
                        <li
                          v-for="(svc, idx) in reviewRiskEvidence(risk)
                            .services"
                          :key="`${svc.service}-${idx}`"
                        >
                          <code>{{ svc.service }}</code>
                        </li>
                      </ul>
                      <ul v-if="reviewRiskEvidence(risk).routes.length > 0">
                        <li
                          v-for="(routeItem, idx) in reviewRiskEvidence(risk)
                            .routes"
                          :key="`${routeItem.method}-${routeItem.path}-${idx}`"
                        >
                          <code
                            >{{ routeItem.method }} {{ routeItem.path }}</code
                          >
                        </li>
                      </ul>
                      <p
                        v-if="
                          reviewRiskEvidence(risk).expectedPath ||
                          reviewRiskEvidence(risk).expectedField
                        "
                      >
                        <code>
                          {{
                            [
                              reviewRiskEvidence(risk).expectedPath,
                              reviewRiskEvidence(risk).expectedField,
                            ]
                              .filter(Boolean)
                              .join(" · ")
                          }}
                        </code>
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              <Alert
                v-if="reviewRiskLoadFailed"
                show-icon
                type="error"
                :message="
                  $t(
                    'plugin.linapro-plugin-marketplace.console.messages.inspectFailed',
                  )
                "
              />

              <DecisionForm />
            </div>
          </template>
        </Spin>
      </div>
    </ReviewDrawer>
  </Page>
</template>

<style scoped>
.marketplace-review-drawer-content {
  display: grid;
  gap: 12px;
}

.marketplace-review-inspection {
  display: grid;
  gap: 12px;
}

.review-plugin-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
  text-align: left;
}

.review-plugin-name {
  font-weight: 600;
}

.review-plugin-id {
  color: var(--ant-color-text-tertiary);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
}

.marketplace-audit-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.marketplace-audit-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 10px;
  border: 1px solid var(--ant-color-border-secondary);
  border-radius: 6px;
  background: var(--ant-color-bg-layout);
}

.marketplace-review-risk-list {
  display: grid;
  max-height: 320px;
  overflow: auto;
  gap: 12px;
  padding: 12px;
  border: 1px solid var(--ant-color-border-secondary);
  border-radius: 10px;
  background: var(--ant-color-bg-layout);
}

/* Align with detail risk cards for consistent multi-item separation. */
.marketplace-review-risk-item {
  position: relative;
  padding: 12px 14px 12px 16px;
  border: 1px solid var(--ant-color-border);
  border-radius: 8px;
  background: var(--ant-color-bg-container);
  box-shadow:
    0 1px 2px rgb(0 0 0 / 4%),
    0 1px 6px -1px rgb(0 0 0 / 4%);
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    background-color 0.2s ease;
}

.marketplace-review-risk-item::before {
  content: "";
  position: absolute;
  top: 8px;
  bottom: 8px;
  left: 0;
  width: 3px;
  border-radius: 0 2px 2px 0;
  background: var(--ant-color-text-quaternary, var(--ant-color-border));
}

.marketplace-review-risk-item:hover {
  border-color: var(--ant-color-primary-border-hover, var(--ant-color-primary));
  box-shadow:
    0 2px 4px rgb(0 0 0 / 5%),
    0 4px 12px -2px rgb(0 0 0 / 8%);
}

.marketplace-review-risk-item--blocking {
  border-color: var(--ant-color-error-border);
  background: var(--ant-color-error-bg);
  box-shadow: none;
}

.marketplace-review-risk-item--blocking::before {
  background: var(--ant-color-error);
}

.marketplace-review-risk-item--blocking:hover {
  border-color: var(--ant-color-error-border);
  box-shadow: 0 1px 4px color-mix(in srgb, var(--ant-color-error) 16%, transparent);
}

.marketplace-review-risk-item-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.marketplace-review-risk-meta {
  display: flex;
  flex: 1 1 auto;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  min-width: 0;
  color: var(--ant-color-text-secondary);
  overflow-wrap: anywhere;
}

.marketplace-review-risk-item p {
  margin: 8px 0 0;
  color: var(--ant-color-text);
  line-height: 1.5;
}

.marketplace-review-risk-toggle {
  flex: 0 0 auto;
  height: auto;
  padding: 0;
  margin: 0;
  line-height: 22px;
  white-space: nowrap;
}

.marketplace-review-risk-guidance {
  display: grid;
  gap: 0;
  margin-top: 12px;
  overflow: hidden;
  border: 1px solid #91caff;
  border: 1px solid var(--ant-color-primary-border, #91caff);
  border-left: 3px solid #1677ff;
  border-left-color: var(--ant-color-primary, #1677ff);
  border-radius: 8px;
  background: #e6f4ff;
  background: var(--ant-color-primary-bg, #e6f4ff);
  box-shadow: inset 0 1px 0 rgb(255 255 255 / 55%);
}

.marketplace-review-risk-guidance-section {
  padding: 12px 14px;
}

.marketplace-review-risk-guidance-section
  + .marketplace-review-risk-guidance-section {
  border-top: 1px solid #bae0ff;
  border-top-color: var(--ant-color-primary-border, #bae0ff);
}

.marketplace-review-risk-guidance-label {
  margin: 0 0 6px;
  color: rgba(0, 0, 0, 0.88);
  color: var(--ant-color-text, rgba(0, 0, 0, 0.88));
  font-size: 13px;
  font-weight: 600;
  line-height: 1.4;
}

.marketplace-review-risk-guidance p {
  margin: 0;
  color: rgba(0, 0, 0, 0.65);
  color: var(--ant-color-text-secondary, rgba(0, 0, 0, 0.65));
  font-size: 13px;
  line-height: 1.65;
}

.marketplace-review-risk-evidence {
  display: grid;
  gap: 10px;
}

.marketplace-review-risk-evidence ul {
  display: grid;
  gap: 6px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.marketplace-review-risk-evidence li {
  padding: 6px 10px;
  border: 1px solid #d9d9d9;
  border-color: var(--ant-color-border-secondary, #d9d9d9);
  border-radius: 6px;
  background: #fff;
  background: var(--ant-color-bg-container, #fff);
  line-height: 1.5;
  word-break: break-all;
}

.marketplace-review-risk-evidence li code {
  padding: 0;
  border: none;
  border-radius: 0;
  background: transparent;
  color: rgba(0, 0, 0, 0.88);
  color: var(--ant-color-text, rgba(0, 0, 0, 0.88));
  font-size: 12px;
  word-break: break-all;
}

.marketplace-review-risk-evidence > p code,
.marketplace-review-risk-evidence > code {
  display: block;
  max-width: 100%;
  padding: 6px 10px;
  border: 1px solid #d9d9d9;
  border-color: var(--ant-color-border-secondary, #d9d9d9);
  border-radius: 6px;
  background: #fff;
  background: var(--ant-color-bg-container, #fff);
  font-size: 12px;
  word-break: break-all;
}

@media (max-width: 640px) {
  .marketplace-audit-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
