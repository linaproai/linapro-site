<script lang="ts">
import type { PluginPageMeta } from "#/plugins/page-registry";

export const pluginPageMeta = {
  pluginId: "linapro-plugin-marketplace",
  routePath: "plugin-marketplace-mine",
  title: "My Plugins",
} satisfies PluginPageMeta;
</script>

<script setup lang="ts">
import type { VbenFormProps } from "@vben/common-ui";

import type { VbenFormSchema } from "#/adapter/form";
import type { VxeGridProps } from "#/adapter/vxe-table";
import type { UploadFile } from "ant-design-vue/es/upload/interface";

import type {
  MarketplacePluginListItem,
  MarketplacePluginType,
  MarketplacePublisherItem,
  MarketplaceReleaseItem,
  MarketplaceStatus,
  MarketplaceVisibility,
} from "../../types/marketplace";

import { computed, nextTick, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import { Page, useVbenDrawer } from "@vben/common-ui";
import { IconifyIcon } from "@vben/icons";
import { breakpointsTailwind, useBreakpoints } from "@vueuse/core";

import {
  Dropdown,
  Menu,
  MenuItem,
  Space,
  Tag,
  Upload,
  message,
} from "ant-design-vue";

import { useVbenForm } from "#/adapter/form";
import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { $t } from "#/locales";
import { formatTimestamp } from "#/utils/time";

import {
  marketplaceGitSourceRegister,
  marketplaceGitSourceSync,
  marketplaceMyPluginList,
  marketplacePluginCreate,
  marketplacePublisherCreate,
  marketplacePublisherList,
  marketplaceReleaseSubmitReview,
  marketplaceReleaseUpload,
} from "../../api/marketplace";
import { marketplaceDetailPath } from "../../utils/routes";
import MarketplaceDetail from "../detail/index.vue";

type GridPageInfo = {
  currentPage: number;
  pageSize: number;
};

type MineFormValues = {
  keyword?: string;
  pluginType?: MarketplacePluginType;
  status?: MarketplaceStatus;
};

type MineGridOptions = NonNullable<
  VxeGridProps<MarketplacePluginListItem>["gridOptions"]
>;

const UploadDragger = Upload.Dragger;
const route = useRoute();
const router = useRouter();
const breakpoints = useBreakpoints(breakpointsTailwind);
const isCompactTable = breakpoints.smaller("xl");
const workflowMessageKey = "plugin-marketplace-publish-workflow";
const showEmbeddedDetail = computed(
  () =>
    route.query.view === "detail" && typeof route.query.pluginId === "string",
);
const publishMode = ref<"plugin" | "version" | "git">("plugin");
const publishers = ref<MarketplacePublisherItem[]>([]);
const pluginDraftReady = ref(false);
const uploadFileList = ref<UploadFile[]>([]);
const latestDraftRelease = ref<MarketplaceReleaseItem | null>(null);
const gitRepoUrl = ref("");
const gitAccessToken = ref("");
const drawerTitle = computed(() => {
  if (publishMode.value === "version") {
    return t("plugin.linapro-plugin-marketplace.mine.actions.newVersion");
  }
  if (publishMode.value === "git") {
    return t("plugin.linapro-plugin-marketplace.mine.actions.registerGit");
  }
  return t("plugin.linapro-plugin-marketplace.mine.actions.publish");
});
const hasPublishers = computed(() => publishers.value.length > 0);
const canSubmitLatestDraft = computed(
  () => latestDraftRelease.value?.reviewStatus === "draft",
);

function buildCellConfig() {
  return { height: isCompactTable.value ? 84 : 48 };
}

const [Grid, gridApi] = useVbenVxeGrid<MarketplacePluginListItem>({
  formOptions: {
    commonConfig: {
      componentProps: { allowClear: true },
      labelWidth: 100,
    },
    schema: [],
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3",
  },
  gridOptions: {
    cellConfig: buildCellConfig(),
    columns: [],
    height: "auto",
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      autoLoad: false,
      ajax: {
        query: async (
          { page }: { page: GridPageInfo },
          formValues: MineFormValues = {},
        ) => {
          return await marketplaceMyPluginList({
            keyword: trimOptional(formValues.keyword),
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            pluginType: formValues.pluginType,
            status: formValues.status,
          });
        },
      },
    },
    rowConfig: { keyField: "pluginId" },
    showOverflow: "tooltip",
    id: "plugin-marketplace-mine",
  },
});

const [PublisherForm, publisherFormApi] = useVbenForm({
  commonConfig: { componentProps: { class: "w-full" }, labelWidth: 132 },
  schema: [],
  showDefaultActions: false,
  wrapperClass: "grid-cols-1",
});

const [PluginForm, pluginFormApi] = useVbenForm({
  commonConfig: { componentProps: { class: "w-full" }, labelWidth: 132 },
  schema: [],
  showDefaultActions: false,
  wrapperClass: "grid-cols-1 md:grid-cols-2",
});

const [UploadForm, uploadFormApi] = useVbenForm({
  commonConfig: { componentProps: { class: "w-full" }, labelWidth: 132 },
  schema: [],
  showDefaultActions: false,
  wrapperClass: "grid-cols-1 md:grid-cols-2",
});

const [PublishDrawer, publishDrawerApi] = useVbenDrawer({
  destroyOnClose: true,
  footer: false,
  async onClosed() {
    await resetPublishWorkflow();
  },
  async onOpenChange(open) {
    if (!open) {
      return;
    }

    publishDrawerApi.setState({ loading: true });
    try {
      const requestedMode = publishMode.value;
      await resetPublishWorkflow();
      const { row } = publishDrawerApi.getData() as {
        row?: MarketplacePluginListItem;
      };

      if (row) {
        publishMode.value = "version";
        pluginDraftReady.value = true;
        uploadFormApi.setState({ schema: buildUploadSchema(true) });
        await uploadFormApi.setValues({
          pluginId: row.pluginId,
          pluginType: row.pluginType,
        });
        return;
      }

      publishMode.value = requestedMode === "git" ? "git" : "plugin";
      const result = await marketplacePublisherList({
        pageNum: 1,
        pageSize: 100,
      });
      publishers.value = result.items;
      publisherFormApi.setState({
        schema: buildPublisherSchema(publishers.value),
      });
      if (publishers.value[0]) {
        await publisherFormApi.setFieldValue(
          "publisherKey",
          publishers.value[0].publisherKey,
        );
      }
    } finally {
      publishDrawerApi.setState({ loading: false });
    }
  },
});

function t(key: string, params?: Record<string, number | string>) {
  return $t(key, params);
}

function trimOptional(value?: string) {
  const normalized = value?.trim();
  return normalized || undefined;
}

function requiredString(value: unknown) {
  const result = typeof value === "string" ? value.trim() : "";
  if (!result) {
    message.warning(
      t("plugin.linapro-plugin-marketplace.console.messages.fieldRequired"),
    );
    throw new Error("validation");
  }
  return result;
}

function stringValue(value: unknown) {
  return typeof value === "string" ? value.trim() : "";
}

function pluginTypeValue(value: unknown): MarketplacePluginType {
  return value === "dynamic" ? "dynamic" : "source";
}

function visibilityValue(value: unknown): MarketplaceVisibility {
  if (value === "private" || value === "reserved") {
    return value;
  }
  return "public";
}

function booleanValue(value: unknown) {
  return value === true || value === 1 || value === "1";
}

function tagCodesValue(value: unknown) {
  return stringValue(value)
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
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
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3",
  };
}

function buildColumns(): MineGridOptions["columns"] {
  const compact = isCompactTable.value;
  return [
    {
      align: "left",
      field: "pluginId",
      headerAlign: "center",
      minWidth: compact ? 175 : 240,
      slots: { default: "plugin" },
      title: t("plugin.linapro-plugin-marketplace.catalog.columns.plugin"),
    },
    ...(compact
      ? []
      : [
          {
            field: "pluginType",
            slots: { default: "pluginType" },
            title: t("plugin.linapro-plugin-marketplace.catalog.columns.type"),
            width: 100,
          },
          {
            field: "marketStatus",
            slots: { default: "marketStatus" },
            title: t(
              "plugin.linapro-plugin-marketplace.mine.columns.marketStatus",
            ),
            width: 105,
          },
        ]),
    {
      field: "sourceKind",
      formatter: ({ cellValue }: { cellValue?: null | string }) =>
        cellValue === "git"
          ? t("plugin.linapro-plugin-marketplace.mine.sourceKind.git")
          : t("plugin.linapro-plugin-marketplace.mine.sourceKind.upload"),
      title: t("plugin.linapro-plugin-marketplace.mine.columns.sourceKind"),
      width: compact ? 80 : 90,
    },
    {
      field: "visibility",
      slots: { default: "visibility" },
      title: t("plugin.linapro-plugin-marketplace.console.fields.visibility"),
      width: compact ? 75 : 90,
    },
    {
      field: "latestVersion",
      minWidth: compact ? 110 : 105,
      slots: { default: "latestVersion" },
      title: t("plugin.linapro-plugin-marketplace.detail.fields.latestVersion"),
    },
    ...(compact
      ? []
      : [
          {
            field: "latestReviewStatus",
            slots: { default: "reviewStatus" },
            title: t(
              "plugin.linapro-plugin-marketplace.detail.columns.reviewStatus",
            ),
            width: 105,
          },
        ]),
    {
      field: "updatedAt",
      formatter: ({ cellValue }: { cellValue?: null | number | string }) =>
        formatTimestamp(cellValue),
      title: t("plugin.linapro-plugin-marketplace.catalog.columns.updatedAt"),
      width: compact ? 170 : 180,
    },
    {
      field: "action",
      fixed: "right",
      slots: { default: "action" },
      title: t("plugin.linapro-plugin-marketplace.catalog.columns.actions"),
      width: compact ? 120 : 130,
    },
  ];
}

function buildPublisherSchema(
  availablePublishers: MarketplacePublisherItem[],
): VbenFormSchema[] {
  if (availablePublishers.length > 0) {
    return [
      {
        component: "Select",
        componentProps: {
          optionFilterProp: "label",
          options: availablePublishers.map((publisher) => ({
            label: `${publisher.name} (${publisher.publisherKey})`,
            value: publisher.publisherKey,
          })),
          showSearch: true,
        },
        fieldName: "publisherKey",
        label: t("plugin.linapro-plugin-marketplace.catalog.fields.publisher"),
        rules: "required",
      },
    ];
  }

  return [
    {
      component: "Input",
      fieldName: "publisherKey",
      label: t("plugin.linapro-plugin-marketplace.console.fields.publisherKey"),
      rules: "required",
    },
    {
      component: "Input",
      fieldName: "name",
      label: t(
        "plugin.linapro-plugin-marketplace.console.fields.publisherName",
      ),
      rules: "required",
    },
    {
      component: "Input",
      fieldName: "homepage",
      label: t("plugin.linapro-plugin-marketplace.console.fields.homepage"),
    },
    {
      component: "Input",
      fieldName: "contactEmail",
      label: t("plugin.linapro-plugin-marketplace.console.fields.contactEmail"),
    },
    {
      component: "Textarea",
      fieldName: "summary",
      label: t("plugin.linapro-plugin-marketplace.console.fields.summary"),
    },
  ];
}

function buildPluginSchema(): VbenFormSchema[] {
  return [
    {
      component: "Input",
      fieldName: "pluginId",
      label: t("plugin.linapro-plugin-marketplace.detail.fields.pluginId"),
      rules: "required",
    },
    {
      component: "Input",
      fieldName: "name",
      label: t("plugin.linapro-plugin-marketplace.console.fields.pluginName"),
      rules: "required",
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
      rules: "required",
    },
    {
      component: "Select",
      componentProps: {
        options: [
          {
            label: t(
              "plugin.linapro-plugin-marketplace.detail.visibility.public",
            ),
            value: "public",
          },
          {
            label: t(
              "plugin.linapro-plugin-marketplace.detail.visibility.private",
            ),
            value: "private",
          },
          {
            label: t(
              "plugin.linapro-plugin-marketplace.detail.visibility.reserved",
            ),
            value: "reserved",
          },
        ],
      },
      defaultValue: "public",
      fieldName: "visibility",
      label: t("plugin.linapro-plugin-marketplace.console.fields.visibility"),
    },
    {
      component: "Textarea",
      fieldName: "summary",
      formItemClass: "md:col-span-2",
      label: t("plugin.linapro-plugin-marketplace.console.fields.summary"),
      rules: "required",
    },
  ];
}

function buildUploadSchema(lockPluginIdentity = false): VbenFormSchema[] {
  return [
    {
      component: "Input",
      componentProps: { disabled: lockPluginIdentity },
      fieldName: "pluginId",
      label: t("plugin.linapro-plugin-marketplace.detail.fields.pluginId"),
      rules: "required",
    },
    {
      component: "Input",
      fieldName: "version",
      label: t("plugin.linapro-plugin-marketplace.detail.columns.version"),
      rules: "required",
    },
    {
      component: "Select",
      componentProps: {
        disabled: lockPluginIdentity,
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
      rules: "required",
    },
    {
      component: "Switch",
      defaultValue: false,
      fieldName: "replaceDraft",
      label: t("plugin.linapro-plugin-marketplace.console.fields.replaceDraft"),
    },
  ];
}

onMounted(async () => {
  gridApi.setState({ formOptions: buildFormOptions() });
  gridApi.setGridOptions({
    cellConfig: buildCellConfig(),
    columns: buildColumns(),
  });
  publisherFormApi.setState({ schema: buildPublisherSchema([]) });
  pluginFormApi.setState({ schema: buildPluginSchema() });
  uploadFormApi.setState({ schema: buildUploadSchema() });
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

watch(isCompactTable, () => {
  gridApi.setGridOptions({
    cellConfig: buildCellConfig(),
    columns: buildColumns(),
  });
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

function formatVisibility(visibility: MarketplaceVisibility) {
  switch (visibility) {
    case "private": {
      return t("plugin.linapro-plugin-marketplace.detail.visibility.private");
    }
    case "reserved": {
      return t("plugin.linapro-plugin-marketplace.detail.visibility.reserved");
    }
    default: {
      return t("plugin.linapro-plugin-marketplace.detail.visibility.public");
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

function handleOpenDetail(row: MarketplacePluginListItem) {
  router.push(marketplaceDetailPath(row.pluginId, "mine"));
}

function openPublishDrawer(row?: MarketplacePluginListItem) {
  if (!row && publishMode.value !== "git") {
    publishMode.value = "plugin";
  }
  publishDrawerApi.setData(row ? { row } : {});
  publishDrawerApi.open();
}

async function resetPublishWorkflow() {
  await Promise.all([
    publisherFormApi.resetForm(),
    pluginFormApi.resetForm(),
    uploadFormApi.resetForm(),
  ]);
  publishMode.value = "plugin";
  publishers.value = [];
  pluginDraftReady.value = false;
  latestDraftRelease.value = null;
  uploadFileList.value = [];
  publisherFormApi.setState({ schema: buildPublisherSchema([]) });
  uploadFormApi.setState({ schema: buildUploadSchema() });
}

async function handleCreatePublisher() {
  publishDrawerApi.lock(true);
  try {
    const { valid } = await publisherFormApi.validate();
    if (!valid) {
      return;
    }
    const values = await publisherFormApi.getValues();
    const result = await marketplacePublisherCreate({
      contactEmail: stringValue(values.contactEmail),
      homepage: stringValue(values.homepage),
      name: requiredString(values.name),
      publisherKey: requiredString(values.publisherKey),
      summary: stringValue(values.summary),
    });
    message.success({
      content: t(
        "plugin.linapro-plugin-marketplace.console.messages.publisherSaved",
      ),
      key: workflowMessageKey,
    });
    publishers.value = [result.publisher];
    publisherFormApi.setState({
      schema: buildPublisherSchema(publishers.value),
    });
    await publisherFormApi.setFieldValue(
      "publisherKey",
      result.publisher.publisherKey,
    );
  } catch {
    // Validation or API errors are shown by form/message handlers.
  } finally {
    publishDrawerApi.lock(false);
  }
}

async function handleCreatePluginDraft() {
  publishDrawerApi.lock(true);
  try {
    const [{ valid: publisherValid }, { valid: pluginValid }] =
      await Promise.all([
        publisherFormApi.validate(),
        pluginFormApi.validate(),
      ]);
    if (!publisherValid || !pluginValid) {
      return;
    }
    const [publisherValues, values] = await Promise.all([
      publisherFormApi.getValues(),
      pluginFormApi.getValues(),
    ]);
    const plugin = await marketplacePluginCreate({
      description: stringValue(values.description),
      homepage: stringValue(values.homepage),
      icon: "",
      license: stringValue(values.license),
      name: requiredString(values.name),
      pluginId: requiredString(values.pluginId),
      pluginType: pluginTypeValue(values.pluginType),
      publisherKey: requiredString(publisherValues.publisherKey),
      repository: stringValue(values.repository),
      summary: requiredString(values.summary),
      tagCodes: tagCodesValue(values.tagCodes),
      visibility: visibilityValue(values.visibility),
    });
    message.success({
      content: t(
        "plugin.linapro-plugin-marketplace.console.messages.pluginSaved",
      ),
      key: workflowMessageKey,
    });
    uploadFormApi.setState({ schema: buildUploadSchema(true) });
    await uploadFormApi.setValues({
      pluginId: plugin.plugin.pluginId,
      pluginType: plugin.plugin.pluginType,
    });
    pluginDraftReady.value = true;
    await gridApi.reload();
  } catch {
    // Validation or API errors are shown by form/message handlers.
  } finally {
    publishDrawerApi.lock(false);
  }
}

async function handleUploadDraft() {
  publishDrawerApi.lock(true);
  try {
    const uploadFile = uploadFileList.value[0]?.originFileObj as
      | File
      | undefined;
    if (!uploadFile) {
      message.warning(
        t("plugin.linapro-plugin-marketplace.console.messages.fileRequired"),
      );
      return;
    }
    const { valid } = await uploadFormApi.validate();
    if (!valid) {
      return;
    }
    const values = await uploadFormApi.getValues();
    const result = await marketplaceReleaseUpload({
      file: uploadFile,
      pluginId: requiredString(values.pluginId),
      pluginType: pluginTypeValue(values.pluginType),
      replaceDraft: booleanValue(values.replaceDraft),
      version: requiredString(values.version),
      visibility: "public",
    });
    latestDraftRelease.value = result.release;
    message.success({
      content: t(
        "plugin.linapro-plugin-marketplace.console.messages.releaseUploaded",
      ),
      key: workflowMessageKey,
    });
  } catch {
    // Validation or API errors are shown by form/message handlers.
  } finally {
    publishDrawerApi.lock(false);
  }
}

async function handleSubmitLatestDraft() {
  publishDrawerApi.lock(true);
  try {
    const release = latestDraftRelease.value;
    const values = await uploadFormApi.getValues();
    const pluginId = release?.pluginId || requiredString(values.pluginId);
    const version = release?.version || requiredString(values.version);
    const result = await marketplaceReleaseSubmitReview(pluginId, version, {});
    latestDraftRelease.value = result.release;
    message.success({
      content: t(
        "plugin.linapro-plugin-marketplace.console.messages.reviewSubmitted",
      ),
      key: workflowMessageKey,
    });
    await gridApi.reload();
  } catch {
    // Validation or API errors are shown by form/message handlers.
  } finally {
    publishDrawerApi.lock(false);
  }
}

async function handleRegisterGitSource() {
  publishDrawerApi.lock(true);
  try {
    const { valid } = await publisherFormApi.validate();
    if (!valid) {
      return;
    }
    if (!gitRepoUrl.value.trim()) {
      message.warning(
        t("plugin.linapro-plugin-marketplace.console.messages.fieldRequired"),
      );
      return;
    }
    const publisherValues = await publisherFormApi.getValues();
    await marketplaceGitSourceRegister({
      accessToken: gitAccessToken.value.trim() || undefined,
      publisherKey: requiredString(publisherValues.publisherKey),
      repoUrl: gitRepoUrl.value.trim(),
      visibility: "public",
    });
    message.success({
      content: t(
        "plugin.linapro-plugin-marketplace.mine.messages.gitRegistered",
      ),
      key: workflowMessageKey,
    });
    gitRepoUrl.value = "";
    gitAccessToken.value = "";
    await gridApi.reload();
    publishDrawerApi.close();
  } catch {
    // Validation or API errors are shown by form/message handlers.
  } finally {
    publishDrawerApi.lock(false);
  }
}

async function handleSyncGitSource(row: MarketplacePluginListItem) {
  try {
    const result = await marketplaceGitSourceSync(row.pluginId);
    message.success(
      t("plugin.linapro-plugin-marketplace.mine.messages.gitSynced", {
        count: result.synced,
      }),
    );
    await gridApi.reload();
  } catch {
    // API errors are shown by request client handlers.
  }
}

function openGitPublishDrawer() {
  publishMode.value = "git";
  openPublishDrawer();
}
</script>

<template>
  <MarketplaceDetail v-if="showEmbeddedDetail" />
  <Page v-else :auto-content-height="true">
    <Grid
      class="plugin-marketplace-mine"
      :table-title="$t('plugin.linapro-plugin-marketplace.mine.tableTitle')"
    >
      <template #toolbar-tools>
        <Space>
          <a-button @click="openGitPublishDrawer()">
            {{ $t("plugin.linapro-plugin-marketplace.mine.actions.registerGit") }}
          </a-button>
          <a-button type="primary" @click="openPublishDrawer()">
            {{ $t("plugin.linapro-plugin-marketplace.mine.actions.publish") }}
          </a-button>
        </Space>
      </template>

      <template #plugin="{ row }">
        <div class="mine-plugin-cell">
          <span class="mine-plugin-name" :title="row.name">{{ row.name }}</span>
          <span class="mine-plugin-id" :title="row.pluginId">
            {{ row.pluginId }}
          </span>
          <div v-if="isCompactTable" class="mine-plugin-meta">
            <Tag :color="row.pluginType === 'source' ? 'blue' : 'green'">
              {{ formatPluginType(row.pluginType) }}
            </Tag>
            <Tag>{{ formatMarketStatus(row.marketStatus) }}</Tag>
          </div>
        </div>
      </template>

      <template #pluginType="{ row }">
        <Tag :color="row.pluginType === 'source' ? 'blue' : 'green'">
          {{ formatPluginType(row.pluginType) }}
        </Tag>
      </template>

      <template #marketStatus="{ row }">
        <Tag>{{ formatMarketStatus(row.marketStatus) }}</Tag>
      </template>

      <template #visibility="{ row }">
        <Tag :color="row.visibility === 'public' ? 'green' : undefined">
          {{ formatVisibility(row.visibility) }}
        </Tag>
      </template>

      <template #latestVersion="{ row }">
        <Space v-if="isCompactTable" direction="vertical" :size="2">
          <span>{{ row.latestVersion || "-" }}</span>
          <Tag v-if="row.latestReviewStatus">
            {{ formatReviewStatus(row.latestReviewStatus) }}
          </Tag>
        </Space>
        <span v-else>{{ row.latestVersion || "-" }}</span>
      </template>

      <template #reviewStatus="{ row }">
        <Tag v-if="row.latestReviewStatus">
          {{ formatReviewStatus(row.latestReviewStatus) }}
        </Tag>
        <span v-else class="mine-muted">-</span>
      </template>

      <template #action="{ row }">
        <Space>
          <ghost-button @click.stop="handleOpenDetail(row)">
            {{ $t("plugin.linapro-plugin-marketplace.catalog.actions.detail") }}
          </ghost-button>
          <Dropdown placement="bottomRight">
            <template #overlay>
              <Menu>
                <MenuItem key="new-version" @click="openPublishDrawer(row)">
                  {{
                    $t(
                      "plugin.linapro-plugin-marketplace.mine.actions.newVersion",
                    )
                  }}
                </MenuItem>
                <MenuItem
                  v-if="row.sourceKind === 'git'"
                  key="sync-git"
                  @click="handleSyncGitSource(row)"
                >
                  {{
                    $t(
                      "plugin.linapro-plugin-marketplace.mine.actions.syncGit",
                    )
                  }}
                </MenuItem>
              </Menu>
            </template>
            <a-button size="small" type="link">
              {{ $t("pages.common.more") }}
            </a-button>
          </Dropdown>
        </Space>
      </template>
    </Grid>

    <PublishDrawer
      :title="drawerTitle"
      class="w-[760px] max-w-[calc(100vw-32px)]"
    >
      <div class="mine-publish-sections">
        <section v-show="publishMode === 'plugin'">
          <div class="mine-section-header">
            <h3>
              {{
                $t(
                  "plugin.linapro-plugin-marketplace.console.sections.publisher",
                )
              }}
            </h3>
            <a-button
              v-if="!hasPublishers"
              type="primary"
              @click="handleCreatePublisher"
            >
              {{
                $t(
                  "plugin.linapro-plugin-marketplace.console.actions.savePublisher",
                )
              }}
            </a-button>
          </div>
          <PublisherForm />
        </section>

        <section
          v-show="publishMode === 'plugin' && hasPublishers"
          class="mine-section-divider"
        >
          <div class="mine-section-header">
            <h3>
              {{
                $t(
                  "plugin.linapro-plugin-marketplace.console.sections.pluginDraft",
                )
              }}
            </h3>
            <a-button type="primary" @click="handleCreatePluginDraft">
              {{
                $t(
                  "plugin.linapro-plugin-marketplace.console.actions.savePlugin",
                )
              }}
            </a-button>
          </div>
          <PluginForm />
        </section>

        <section
          v-show="publishMode === 'version' || pluginDraftReady"
          :class="{ 'mine-section-divider': publishMode === 'plugin' }"
        >
          <div class="mine-section-header">
            <h3>
              {{
                $t(
                  "plugin.linapro-plugin-marketplace.console.sections.releaseUpload",
                )
              }}
            </h3>
            <Space>
              <a-button
                :disabled="!canSubmitLatestDraft"
                @click="handleSubmitLatestDraft"
              >
                {{
                  $t(
                    "plugin.linapro-plugin-marketplace.console.actions.submitReview",
                  )
                }}
              </a-button>
              <a-button type="primary" @click="handleUploadDraft">
                {{
                  $t(
                    "plugin.linapro-plugin-marketplace.console.actions.uploadDraft",
                  )
                }}
              </a-button>
            </Space>
          </div>
          <UploadForm />
          <UploadDragger
            v-model:file-list="uploadFileList"
            accept=".zip,.tar.gz,.tgz"
            :before-upload="() => false"
            :max-count="1"
            class="mine-upload"
          >
            <p class="ant-upload-drag-icon flex items-center justify-center">
              <IconifyIcon
                class="text-primary text-5xl"
                icon="ant-design:inbox-outlined"
              />
            </p>
            <p class="ant-upload-text">
              {{
                $t("plugin.linapro-plugin-marketplace.console.upload.dragText")
              }}
            </p>
            <p class="ant-upload-hint">
              {{ $t("plugin.linapro-plugin-marketplace.console.upload.hint") }}
            </p>
          </UploadDragger>
          <div v-if="latestDraftRelease" class="mine-draft-result">
            <Tag color="processing">
              {{ latestDraftRelease.version }} /
              {{ formatReviewStatus(latestDraftRelease.reviewStatus) }}
            </Tag>
          </div>
        </section>

        <section
          v-show="publishMode === 'git'"
          class="mine-section-divider"
        >
          <div class="mine-section-header">
            <h3>
              {{
                $t(
                  "plugin.linapro-plugin-marketplace.console.sections.publisher",
                )
              }}
            </h3>
            <a-button
              v-if="!hasPublishers"
              type="primary"
              @click="handleCreatePublisher"
            >
              {{
                $t(
                  "plugin.linapro-plugin-marketplace.console.actions.savePublisher",
                )
              }}
            </a-button>
          </div>
          <PublisherForm />
          <div v-if="hasPublishers" class="mine-section-header">
            <h3>
              {{
                $t("plugin.linapro-plugin-marketplace.mine.sections.gitSource")
              }}
            </h3>
            <a-button type="primary" @click="handleRegisterGitSource">
              {{
                $t(
                  "plugin.linapro-plugin-marketplace.mine.actions.saveGitSource",
                )
              }}
            </a-button>
          </div>
          <div v-if="hasPublishers" class="mine-git-fields">
            <label>
              {{ $t("plugin.linapro-plugin-marketplace.mine.fields.repoUrl") }}
              <input
                v-model="gitRepoUrl"
                class="mine-git-input"
                type="url"
                :placeholder="
                  $t(
                    'plugin.linapro-plugin-marketplace.mine.placeholders.repoUrl',
                  )
                "
              />
            </label>
            <label>
              {{
                $t("plugin.linapro-plugin-marketplace.mine.fields.accessToken")
              }}
              <input
                v-model="gitAccessToken"
                class="mine-git-input"
                type="password"
                autocomplete="off"
                :placeholder="
                  $t(
                    'plugin.linapro-plugin-marketplace.mine.placeholders.accessToken',
                  )
                "
              />
            </label>
          </div>
        </section>
      </div>
    </PublishDrawer>
  </Page>
</template>

<style scoped>
.mine-git-fields {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.mine-git-input {
  display: block;
  width: 100%;
  margin-top: 4px;
  padding: 6px 10px;
  border: 1px solid hsl(var(--border));
  border-radius: 6px;
  background: transparent;
}

.mine-plugin-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
  text-align: left;
}

.mine-plugin-name {
  max-width: 100%;
  overflow: hidden;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mine-plugin-id {
  max-width: 100%;
  overflow: hidden;
  color: var(--ant-color-text-tertiary);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mine-plugin-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 2px;
}

.mine-muted {
  color: var(--ant-color-text-secondary);
}

.mine-publish-sections {
  display: grid;
  gap: 20px;
}

.mine-section-divider {
  padding-top: 20px;
  border-top: 1px solid var(--ant-color-border-secondary);
}

.mine-section-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.mine-section-header h3 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
}

.mine-upload {
  margin-top: 12px;
}

.mine-draft-result {
  margin-top: 12px;
}
</style>
