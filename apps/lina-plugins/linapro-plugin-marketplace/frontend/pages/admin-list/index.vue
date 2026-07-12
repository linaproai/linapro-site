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
  MarketplacePluginListItem,
  MarketplacePluginType,
  MarketplaceStatus,
} from "../../types/marketplace";

import { computed, nextTick, onMounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import { Page } from "@vben/common-ui";

import { Tag } from "ant-design-vue";

import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { $t } from "#/locales";
import { formatTimestamp } from "#/utils/time";

import { marketplaceManagedPluginList } from "../../api/marketplace";
import { marketplaceDetailPath } from "../../utils/routes";
import MarketplaceDetail from "../detail/index.vue";

type GridPageInfo = {
  currentPage: number;
  pageSize: number;
};

type AdminFormValues = {
  keyword?: string;
  pluginType?: MarketplacePluginType;
  publisher?: string;
  status?: MarketplaceStatus;
};

type AdminGridOptions = NonNullable<
  VxeGridProps<MarketplacePluginListItem>["gridOptions"]
>;

const route = useRoute();
const router = useRouter();
const showEmbeddedDetail = computed(
  () =>
    route.query.view === "detail" && typeof route.query.pluginId === "string",
);

const [Grid, gridApi] = useVbenVxeGrid<MarketplacePluginListItem>({
  formOptions: {
    commonConfig: {
      componentProps: { allowClear: true },
      labelWidth: 100,
    },
    schema: [],
    wrapperClass: "grid-cols-1 md:grid-cols-2 xl:grid-cols-4",
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
    showOverflow: "tooltip",
    id: "plugin-marketplace-admin-list",
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
      labelWidth: 100,
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
          options: [
            {
              label: t("plugin.linapro-plugin-marketplace.detail.status.draft"),
              value: "draft",
            },
            {
              label: t(
                "plugin.linapro-plugin-marketplace.detail.status.published",
              ),
              value: "published",
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
    wrapperClass: "grid-cols-1 md:grid-cols-2 xl:grid-cols-4",
  };
}

function buildColumns(): AdminGridOptions["columns"] {
  return [
    {
      align: "left",
      field: "pluginId",
      headerAlign: "center",
      minWidth: 220,
      slots: { default: "plugin" },
      title: t("plugin.linapro-plugin-marketplace.catalog.columns.plugin"),
    },
    {
      field: "pluginType",
      slots: { default: "pluginType" },
      title: t("plugin.linapro-plugin-marketplace.catalog.columns.type"),
      width: 100,
    },
    {
      field: "publisher",
      minWidth: 130,
      slots: { default: "publisher" },
      title: t("plugin.linapro-plugin-marketplace.catalog.columns.publisher"),
    },
    {
      field: "marketStatus",
      slots: { default: "marketStatus" },
      title: t("plugin.linapro-plugin-marketplace.mine.columns.marketStatus"),
      width: 105,
    },
    {
      field: "latestVersion",
      minWidth: 100,
      title: t("plugin.linapro-plugin-marketplace.detail.fields.latestVersion"),
    },
    {
      field: "latestReviewStatus",
      slots: { default: "reviewStatus" },
      title: t("plugin.linapro-plugin-marketplace.detail.columns.reviewStatus"),
      width: 105,
    },
    {
      field: "updatedAt",
      formatter: ({ cellValue }: { cellValue?: null | number | string }) =>
        formatTimestamp(cellValue),
      title: t("plugin.linapro-plugin-marketplace.catalog.columns.updatedAt"),
      width: 180,
    },
    {
      field: "action",
      fixed: "right",
      slots: { default: "action" },
      title: t("plugin.linapro-plugin-marketplace.catalog.columns.actions"),
      width: 100,
    },
  ];
}

onMounted(async () => {
  gridApi.setState({ formOptions: buildFormOptions() });
  gridApi.setGridOptions({ columns: buildColumns() });
  if (!showEmbeddedDetail.value) {
    await gridApi.reload();
  }
});

watch(showEmbeddedDetail, async (detailMode) => {
  if (!detailMode) {
    await nextTick();
    await gridApi.reload();
  }
});

function formatPluginType(type: MarketplacePluginType) {
  return type === "source"
    ? t("plugin.linapro-plugin-marketplace.catalog.pluginType.source")
    : t("plugin.linapro-plugin-marketplace.catalog.pluginType.dynamic");
}

function formatMarketStatus(status: MarketplaceStatus) {
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
    default: {
      return status || "-";
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

function getMarketStatusColor(status: MarketplaceStatus) {
  switch (status) {
    case "published": {
      return "success";
    }
    case "deprecated": {
      return "warning";
    }
    case "draft": {
      return "processing";
    }
    default: {
      return "default";
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

function handleOpenDetail(row: MarketplacePluginListItem) {
  router.push(marketplaceDetailPath(row.pluginId, "admin-list"));
}
</script>

<template>
  <MarketplaceDetail v-if="showEmbeddedDetail" />
  <Page v-else :auto-content-height="true">
    <Grid
      class="plugin-marketplace-admin-list"
      :table-title="
        $t('plugin.linapro-plugin-marketplace.adminList.tableTitle')
      "
    >
      <template #plugin="{ row }">
        <div class="admin-plugin-cell">
          <span class="admin-plugin-name">{{ row.name }}</span>
          <span class="admin-plugin-id">{{ row.pluginId }}</span>
        </div>
      </template>

      <template #pluginType="{ row }">
        <Tag :color="row.pluginType === 'source' ? 'blue' : 'green'">
          {{ formatPluginType(row.pluginType) }}
        </Tag>
      </template>

      <template #publisher="{ row }">
        {{ row.publisher?.name || row.publisher?.publisherKey || "-" }}
      </template>

      <template #marketStatus="{ row }">
        <Tag :color="getMarketStatusColor(row.marketStatus)">
          {{ formatMarketStatus(row.marketStatus) }}
        </Tag>
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

<style scoped>
.admin-plugin-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
  text-align: left;
}

.admin-plugin-name {
  font-weight: 600;
}

.admin-plugin-id {
  color: var(--ant-color-text-tertiary);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
}
</style>
