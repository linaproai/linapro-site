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

import { defineAsyncComponent, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";

import { Page, useVbenDrawer, useVbenModal } from "@vben/common-ui";

import {
  Alert,
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
  connectedComponent: defineAsyncComponent(
    () => import("../detail/detail-modal.vue"),
  ),
});

const selectedRelease = ref<MarketplaceReviewQueueItem | null>(null);
const reviewRisks = ref<MarketplaceRiskItem[]>([]);
const reviewDocument = ref<MarketplaceDocumentItem | null>(null);
const reviewLoading = ref(false);
const reviewRisksReady = ref(false);
const reviewRiskLoadFailed = ref(false);
let inspectionRequestId = 0;

const [Grid, gridApi] = useVbenVxeGrid<ReviewQueueRow>({
  formOptions: {
    commonConfig: {
      componentProps: { allowClear: true },
      labelWidth: 100,
    },
    schema: [],
    wrapperClass: "grid-cols-1 md:grid-cols-2 xl:grid-cols-3",
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
    showOverflow: "tooltip",
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

function trimOptional(value?: string) {
  const normalized = value?.trim();
  return normalized || undefined;
}

function buildFormOptions(): VbenFormProps {
  return {
    commonConfig: {
      componentProps: { allowClear: true },
      labelWidth: 100,
    },
    schema: [
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
      {
        component: "Input",
        fieldName: "keyword",
        label: t("plugin.linapro-plugin-marketplace.catalog.fields.keyword"),
      },
    ],
    wrapperClass: "grid-cols-1 md:grid-cols-2 xl:grid-cols-3",
  };
}

function buildColumns(): ReviewGridOptions["columns"] {
  return [
    {
      field: "pluginId",
      minWidth: 200,
      slots: { default: "plugin" },
      title: t("plugin.linapro-plugin-marketplace.catalog.columns.plugin"),
    },
    {
      field: "version",
      minWidth: 120,
      title: t("plugin.linapro-plugin-marketplace.detail.columns.version"),
    },
    {
      field: "pluginType",
      slots: { default: "pluginType" },
      title: t("plugin.linapro-plugin-marketplace.catalog.fields.pluginType"),
      width: 120,
    },
    {
      field: "reviewStatus",
      slots: { default: "reviewStatus" },
      title: t("plugin.linapro-plugin-marketplace.detail.columns.reviewStatus"),
      width: 120,
    },
    {
      field: "submittedAt",
      formatter: ({ cellValue }: { cellValue?: null | number | string }) =>
        formatTimestamp(cellValue),
      title: t("plugin.linapro-plugin-marketplace.review.columns.submittedAt"),
      width: 180,
    },
    {
      field: "action",
      fixed: "right",
      slots: { default: "action" },
      title: t("plugin.linapro-plugin-marketplace.detail.columns.actions"),
      width: 120,
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
    marketplaceReleaseDocument(row.pluginId, row.version, undefined, "managed"),
  ]);

  if (requestId !== inspectionRequestId) {
    return;
  }

  if (riskResult.status === "fulfilled") {
    reviewRisks.value = riskResult.value.items;
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
                  :key="`${risk.type}:${risk.source}:${risk.summary}`"
                  class="marketplace-review-risk-item"
                >
                  <div class="marketplace-review-risk-meta">
                    <Tag :color="getRiskSeverityColor(risk.severity)">
                      {{ formatRiskSeverity(risk.severity) }}
                    </Tag>
                    <Tag>{{ formatRiskType(risk.type) }}</Tag>
                    <span>{{ risk.source }}</span>
                  </div>
                  <p>{{ risk.summary }}</p>
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
  max-height: 260px;
  overflow: auto;
  gap: 8px;
}

.marketplace-review-risk-item {
  padding: 10px;
  border: 1px solid var(--ant-color-border-secondary);
  border-radius: 6px;
  background: var(--ant-color-bg-container);
}

.marketplace-review-risk-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  color: var(--ant-color-text-secondary);
  overflow-wrap: anywhere;
}

.marketplace-review-risk-item p {
  margin: 8px 0 0;
  color: var(--ant-color-text);
}

@media (max-width: 640px) {
  .marketplace-audit-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
