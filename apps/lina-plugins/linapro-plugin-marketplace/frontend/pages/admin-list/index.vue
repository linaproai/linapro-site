<script lang="ts">
import type { PluginPageMeta } from "#/plugins/page-registry";

export const pluginPageMeta = {
  pluginId: "linapro-plugin-marketplace",
  routePath: "plugin-marketplace-admin-list",
  title: "Marketplace Plugin List",
} satisfies PluginPageMeta;
</script>

<script setup lang="ts">
import type { VbenFormProps } from "@vben/common-ui";

import type { VxeGridProps } from "#/adapter/vxe-table";

import type {
  MarketplaceListStatusFilter,
  MarketplacePluginListItem,
  MarketplacePluginType,
  MarketplaceStatus,
} from "../../types/marketplace";

import { onMounted, watch } from "vue";
import { useRoute } from "vue-router";

import { Page, useVbenModal } from "@vben/common-ui";
import { preferences } from "@vben/preferences";

import { Tag, Tooltip } from "ant-design-vue";

import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { $t } from "#/locales";
import { formatTimestamp } from "#/utils/time";

import { marketplaceManagedPluginList } from "../../api/marketplace";
import DetailModalContent from "../detail/detail-modal.vue";

type GridPageInfo = {
  currentPage: number;
  pageSize: number;
};

type AdminFormValues = {
  keyword?: string;
  pluginType?: MarketplacePluginType;
  publisher?: string;
  status?: MarketplaceListStatusFilter;
};

type AdminGridOptions = NonNullable<
  VxeGridProps<MarketplacePluginListItem>["gridOptions"]
>;

const route = useRoute();
const adminRowClickIgnoredFields = new Set(["action"]);

const [DetailModal, detailModalApi] = useVbenModal({
  connectedComponent: DetailModalContent,
});

const [Grid, gridApi] = useVbenVxeGrid<MarketplacePluginListItem>({
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
          formValues: AdminFormValues = {},
        ) => {
          return await marketplaceManagedPluginList({
            keyword: trimOptional(formValues.keyword),
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            pluginType: formValues.pluginType,
            publisher: trimOptional(formValues.publisher),
            status: formValues.status,
          });
        },
      },
    },
    rowConfig: { keyField: "pluginId" },
    rowClassName: "cursor-pointer",
    showOverflow: "ellipsis",
    id: "plugin-marketplace-admin-list",
  },
  gridEvents: {
    cellClick: ({
      $event,
      column,
      row,
    }: {
      $event?: Event;
      column?: { field?: string };
      row: MarketplacePluginListItem;
    }) => {
      if (shouldIgnoreAdminRowClick(column?.field, $event)) {
        return;
      }
      handleOpenDetail(row);
    },
  },
});

function t(key: string) {
  return $t(key);
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
        component: "Select",
        componentProps: {
          options: [
            {
              label: t(
                "plugin.linapro-plugin-marketplace.catalog.pluginType.source",
              ),
              value: "source",
            },
            {
              label: t(
                "plugin.linapro-plugin-marketplace.catalog.pluginType.dynamic",
              ),
              value: "dynamic",
            },
          ],
        },
        fieldName: "pluginType",
        label: t("plugin.linapro-plugin-marketplace.catalog.fields.pluginType"),
      },
      {
        component: "Input",
        fieldName: "publisher",
        label: t("plugin.linapro-plugin-marketplace.catalog.fields.publisher"),
      },
      {
        component: "Select",
        componentProps: {
          // Options match formatMarketStatus() tags: process pipeline first,
          // then terminal marketplace lifecycle states operators filter by.
          options: [
            {
              label: t(
                "plugin.linapro-plugin-marketplace.detail.processStatus.pendingVerify",
              ),
              value: "pending_verify",
            },
            {
              label: t(
                "plugin.linapro-plugin-marketplace.detail.processStatus.pendingReview",
              ),
              value: "pending_review",
            },
            {
              label: t(
                "plugin.linapro-plugin-marketplace.detail.status.published",
              ),
              value: "published",
            },
            {
              label: t(
                "plugin.linapro-plugin-marketplace.detail.processStatus.failed",
              ),
              value: "failed",
            },
            {
              label: t(
                "plugin.linapro-plugin-marketplace.detail.status.delisted",
              ),
              value: "delisted",
            },
            {
              label: t(
                "plugin.linapro-plugin-marketplace.detail.status.deprecated",
              ),
              value: "deprecated",
            },
          ],
        },
        fieldName: "status",
        label: t("plugin.linapro-plugin-marketplace.mine.fields.status"),
      },
    ],
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4",
  };
}

function buildColumns(): AdminGridOptions["columns"] {
  // Match My Plugins table chrome: separate pluginId/name (single-line cells),
  // status early, same minWidth floors, then admin-only publisher/review columns.
  return [
    {
      align: "left",
      field: "pluginId",
      headerAlign: "center",
      minWidth: 200,
      title: t("plugin.linapro-plugin-marketplace.mine.columns.pluginId"),
    },
    {
      align: "left",
      className: "admin-name-column",
      field: "name",
      headerAlign: "center",
      minWidth: 200,
      showOverflow: "ellipsis",
      slots: { default: "name" },
      title: t("plugin.linapro-plugin-marketplace.mine.columns.name"),
    },
    {
      field: "marketStatus",
      showOverflow: "ellipsis",
      slots: { default: "marketStatus" },
      // Same status header label as My Plugins ("状态" / "Status").
      title: t("plugin.linapro-plugin-marketplace.mine.fields.status"),
      // Wide enough for en-US "Pending Review" / zh-CN "待审核" tags.
      width: 124,
    },
    {
      field: "latestReviewStatus",
      showOverflow: "ellipsis",
      slots: { default: "reviewStatus" },
      title: t("plugin.linapro-plugin-marketplace.detail.columns.reviewStatus"),
      width: 112,
    },
    {
      field: "latestVersion",
      showOverflow: "ellipsis",
      slots: { default: "latestVersion" },
      title: t("plugin.linapro-plugin-marketplace.detail.fields.latestVersion"),
      width: 132,
    },
    {
      field: "downloadCount",
      formatter: ({ cellValue }: { cellValue?: null | number }) =>
        String(cellValue ?? 0),
      title: t("plugin.linapro-plugin-marketplace.catalog.columns.downloads"),
      width: 112,
    },
    {
      field: "pluginType",
      slots: { default: "pluginType" },
      title: t("plugin.linapro-plugin-marketplace.catalog.columns.type"),
      width: 96,
    },
    {
      field: "publisher",
      minWidth: 112,
      showOverflow: "ellipsis",
      slots: { default: "publisher" },
      title: t("plugin.linapro-plugin-marketplace.catalog.columns.publisher"),
    },
    {
      field: "updatedAt",
      formatter: ({ cellValue }: { cellValue?: null | number | string }) =>
        formatTimestamp(cellValue),
      title: t("plugin.linapro-plugin-marketplace.catalog.columns.updatedAt"),
      width: 160,
    },
    {
      field: "action",
      fixed: "right",
      slots: { default: "action" },
      title: t("plugin.linapro-plugin-marketplace.catalog.columns.actions"),
      // Detail ghost button (en-US "Details" / zh-CN "详情").
      width: 100,
    },
  ];
}

function shouldIgnoreAdminRowClick(
  columnField?: string,
  event?: Event,
): boolean {
  if (columnField && adminRowClickIgnoredFields.has(columnField)) {
    return true;
  }
  const target = event?.target;
  if (!(target instanceof Element)) {
    return false;
  }
  return Boolean(
    target.closest(
      'button, a, input, textarea, select, .ant-dropdown, .ant-dropdown-menu, [role="button"], [role="menuitem"]',
    ),
  );
}

onMounted(async () => {
  gridApi.setState({ formOptions: buildFormOptions() });
  gridApi.setGridOptions({ columns: buildColumns() });
  await gridApi.reload();
  openDetailFromRouteQuery();
});

watch(
  () => [route.query.view, route.query.pluginId] as const,
  () => {
    openDetailFromRouteQuery();
  },
);

// Name/summary are backend-localized per request locale. Rebuild chrome and
// re-query when the workbench language changes so headers/filters/cells match.
watch(
  () => preferences.app.locale,
  async (locale, previousLocale) => {
    if (!previousLocale || locale === previousLocale) {
      return;
    }
    gridApi.setState({ formOptions: buildFormOptions() });
    gridApi.setGridOptions({ columns: buildColumns() });
    await gridApi.query();
  },
);

function formatPluginType(type: MarketplacePluginType) {
  return type === "source"
    ? t("plugin.linapro-plugin-marketplace.catalog.pluginType.source")
    : t("plugin.linapro-plugin-marketplace.catalog.pluginType.dynamic");
}

function getPluginTypeColor(type: MarketplacePluginType) {
  return type === "source" ? "blue" : "green";
}

function formatMarketStatus(status: MarketplaceStatus, processStatus?: string) {
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

function formatReviewStatus(status?: string) {
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
      return status || "-";
    }
  }
}

function getMarketStatusColor(
  status: MarketplaceStatus,
  processStatus?: string,
) {
  if (status === "published") {
    return "success";
  }
  if (status === "delisted") {
    return "default";
  }
  if (status === "deprecated") {
    return "warning";
  }
  // Match My Plugins tag colors for process pipeline states.
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

function getReviewStatusColor(status?: string) {
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

function buildStatusTooltip(row: MarketplacePluginListItem) {
  const parts = [
    formatMarketStatus(row.marketStatus, row.processStatus),
    row.latestReviewStatus
      ? `${t("plugin.linapro-plugin-marketplace.detail.columns.reviewStatus")}: ${formatReviewStatus(row.latestReviewStatus)}`
      : "",
  ].filter(Boolean);
  return parts.join(" · ");
}

function handleOpenDetail(row: MarketplacePluginListItem) {
  detailModalApi.setData({ from: "admin-list", pluginId: row.pluginId });
  detailModalApi.open();
}

function openDetailFromRouteQuery() {
  if (route.query.view !== "detail") {
    return;
  }
  const pluginId =
    typeof route.query.pluginId === "string" ? route.query.pluginId.trim() : "";
  if (!pluginId) {
    return;
  }
  detailModalApi.setData({ from: "admin-list", pluginId });
  detailModalApi.open();
}
</script>

<template>
  <Page :auto-content-height="true">
    <DetailModal />
    <Grid
      class="plugin-marketplace-admin-list"
      :table-title="
        $t('plugin.linapro-plugin-marketplace.adminList.tableTitle')
      "
    >
      <template #name="{ row }">
        <div class="max-w-full truncate font-medium" :title="row.name">
          {{ row.name }}
        </div>
      </template>

      <template #latestVersion="{ row }">
        <span
          class="inline-flex max-w-full items-center gap-1 whitespace-nowrap font-mono text-xs tabular-nums"
          :title="row.latestVersion || '-'"
        >
          <span class="shrink-0">{{ row.latestVersion || "-" }}</span>
        </span>
      </template>

      <template #pluginType="{ row }">
        <Tag :color="getPluginTypeColor(row.pluginType)">
          {{ formatPluginType(row.pluginType) }}
        </Tag>
      </template>

      <template #publisher="{ row }">
        <div
          class="max-w-full truncate"
          :title="row.publisher?.name || row.publisher?.publisherKey || '-'"
        >
          {{ row.publisher?.name || row.publisher?.publisherKey || "-" }}
        </div>
      </template>

      <template #marketStatus="{ row }">
        <Tooltip :title="buildStatusTooltip(row)">
          <Tag
            :color="getMarketStatusColor(row.marketStatus, row.processStatus)"
          >
            {{ formatMarketStatus(row.marketStatus, row.processStatus) }}
          </Tag>
        </Tooltip>
      </template>

      <template #reviewStatus="{ row }">
        <Tag
          v-if="row.latestReviewStatus"
          :color="getReviewStatusColor(row.latestReviewStatus)"
        >
          {{ formatReviewStatus(row.latestReviewStatus) }}
        </Tag>
        <span v-else>-</span>
      </template>

      <template #action="{ row }">
        <ghost-button @click.stop="handleOpenDetail(row)">
          {{ $t("plugin.linapro-plugin-marketplace.catalog.actions.detail") }}
        </ghost-button>
      </template>
    </Grid>
  </Page>
</template>
