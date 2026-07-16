<script lang="ts">
import type { PluginPageMeta } from "#/plugins/page-registry";

export const pluginPageMeta = {
  pluginId: "linapro-plugin-marketplace",
  routePath: "plugin-marketplace-mine",
  title: "My Plugins",
} satisfies PluginPageMeta;
</script>

<script setup lang="ts">
import type { VbenFormSchema } from "#/adapter/form";
import type { VxeGridProps } from "#/adapter/vxe-table";
import type { UploadFile } from "ant-design-vue/es/upload/interface";

import type {
  MarketplacePluginListItem,
  MarketplacePluginType,
  MarketplacePublisherItem,
  MarketplaceStatus,
  MarketplaceVisibility,
} from "../../types/marketplace";

import { computed, nextTick, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import { Page, useVbenDrawer } from "@vben/common-ui";
import { IconifyIcon } from "@vben/icons";
import { breakpointsTailwind, useBreakpoints } from "@vueuse/core";

import {
  Alert,
  Dropdown,
  Menu,
  MenuItem,
  Modal,
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
  marketplacePackageAdd,
  marketplacePluginDelist,
  marketplacePluginPublish,
  marketplacePublisherCreate,
  marketplacePublisherList,
  marketplacePublisherUpdate,
} from "../../api/marketplace";
import { marketplaceDetailPath } from "../../utils/routes";
import MarketplaceDetail from "../detail/index.vue";

type GridPageInfo = {
  currentPage: number;
  pageSize: number;
};

type MineGridOptions = NonNullable<
  VxeGridProps<MarketplacePluginListItem>["gridOptions"]
>;

type PublishSourceKind = "git" | "upload";
type PublishMode = "plugin" | "version";

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
const publishMode = ref<PublishMode>("plugin");
const publishSourceKind = ref<PublishSourceKind>("upload");
const publishers = ref<MarketplacePublisherItem[]>([]);
const boundPublisher = ref<MarketplacePublisherItem | null>(null);
const uploadFileList = ref<UploadFile[]>([]);
const versionTarget = ref<MarketplacePluginListItem | null>(null);
const drawerTitle = computed(() => {
  if (publishMode.value === "version") {
    return t("plugin.linapro-plugin-marketplace.mine.actions.newVersion");
  }
  return t("plugin.linapro-plugin-marketplace.mine.actions.add");
});
const publishPrimaryActionLabel = computed(() =>
  publishMode.value === "version"
    ? t("plugin.linapro-plugin-marketplace.console.actions.uploadDraft")
    : t("plugin.linapro-plugin-marketplace.mine.actions.add"),
);
const publisherDrawerTitle = computed(() =>
  boundPublisher.value
    ? t("plugin.linapro-plugin-marketplace.mine.actions.editPublisher")
    : t("plugin.linapro-plugin-marketplace.mine.actions.registerPublisher"),
);
const hasPublishers = computed(() => publishers.value.length > 0);

function buildCellConfig() {
  return { height: isCompactTable.value ? 84 : 48 };
}

const [Grid, gridApi] = useVbenVxeGrid<MarketplacePluginListItem>({
  showSearchForm: false,
  gridOptions: {
    cellConfig: buildCellConfig(),
    columns: [],
    height: "auto",
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      autoLoad: false,
      ajax: {
        query: async ({ page }: { page: GridPageInfo }) => {
          return await marketplaceMyPluginList({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
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
  handleValuesChange(values) {
    const nextKind =
      values.sourceKind === "git" ? "git" : ("upload" as PublishSourceKind);
    if (publishSourceKind.value !== nextKind) {
      publishSourceKind.value = nextKind;
      uploadFileList.value = [];
    }
  },
  schema: [],
  showDefaultActions: false,
  wrapperClass: "grid-cols-1",
});

const [PublisherDrawer, publisherDrawerApi] = useVbenDrawer({
  destroyOnClose: true,
  footer: false,
  async onClosed() {
    boundPublisher.value = null;
    await publisherFormApi.resetForm();
  },
  async onOpenChange(open) {
    if (!open) {
      return;
    }

    publisherDrawerApi.setState({ loading: true });
    try {
      const result = await marketplacePublisherList({
        pageNum: 1,
        pageSize: 1,
      });
      const existing = result.items[0] ?? null;
      boundPublisher.value = existing;
      publisherFormApi.setState({
        schema: buildPublisherSchema(),
      });
      await publisherFormApi.resetForm();
      if (existing) {
        await publisherFormApi.setValues({
          contactEmail: existing.contactEmail || "",
          homepage: existing.homepage || "",
          name: existing.name,
          summary: existing.summary || "",
        });
      }
    } finally {
      publisherDrawerApi.setState({ loading: false });
    }
  },
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
      await resetPublishWorkflow();
      const { row } = publishDrawerApi.getData() as {
        row?: MarketplacePluginListItem;
      };

      if (row) {
        publishMode.value = "version";
        versionTarget.value = row;
        publishSourceKind.value = "upload";
        return;
      }

      publishMode.value = "plugin";
      publishSourceKind.value = "upload";
      const result = await marketplacePublisherList({
        pageNum: 1,
        pageSize: 100,
      });
      publishers.value = result.items;
      pluginFormApi.setState({
        schema: buildPluginSchema(publishers.value),
      });
      await pluginFormApi.setValues({ sourceKind: "upload" });
      if (publishers.value[0]) {
        await pluginFormApi.setFieldValue(
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
      width: compact ? 90 : 100,
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

function buildPublisherSchema(): VbenFormSchema[] {
  // publisherKey is an internal stable ID: generated on create, never edited here.
  return [
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

function generatePublisherKey() {
  const random =
    typeof crypto !== "undefined" && "randomUUID" in crypto
      ? crypto.randomUUID().replaceAll("-", "")
      : `${Date.now().toString(36)}${Math.random().toString(36).slice(2, 10)}`;
  return `pub-${random}`.slice(0, 64);
}

function buildPublisherSelectField(
  availablePublishers: MarketplacePublisherItem[],
): VbenFormSchema | null {
  if (availablePublishers.length <= 1) {
    return null;
  }
  return {
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
    formItemClass: "md:col-span-2",
    label: t("plugin.linapro-plugin-marketplace.catalog.fields.publisher"),
    rules: "required",
  };
}

function buildSourceKindOptions() {
  return [
    {
      label: t("plugin.linapro-plugin-marketplace.mine.sourceKind.upload"),
      value: "upload",
    },
    {
      label: t("plugin.linapro-plugin-marketplace.mine.sourceKind.git"),
      value: "git",
    },
  ];
}

function buildPluginSchema(
  availablePublishers: MarketplacePublisherItem[],
): VbenFormSchema[] {
  const publisherField = buildPublisherSelectField(availablePublishers);
  return [
    {
      component: "RadioGroup",
      componentProps: {
        buttonStyle: "solid",
        options: buildSourceKindOptions(),
        optionType: "button",
      },
      defaultValue: "upload",
      fieldName: "sourceKind",
      formItemClass: "md:col-span-2",
      label: t("plugin.linapro-plugin-marketplace.mine.fields.sourceKind"),
      rules: "required",
    },
    ...(publisherField ? [publisherField] : []),
    // Slot-rendered upload control; keep as form item so the label shares the
    // same left column as sourceKind instead of sitting in a right-side aside.
    {
      component: "Input",
      dependencies: {
        show: (values: Record<string, unknown>) =>
          values.sourceKind === "upload",
        triggerFields: ["sourceKind"],
      },
      fieldName: "packageFile",
      formItemClass: "mine-package-form-item md:col-span-2",
      // Tall control stack (hint + dragger): keep label top-aligned with sourceKind.
      labelClass: "self-start pt-1",
      label: t(
        "plugin.linapro-plugin-marketplace.mine.sections.packageUpload",
      ),
    },
    {
      component: "Input",
      componentProps: {
        placeholder: t(
          "plugin.linapro-plugin-marketplace.mine.placeholders.repoUrl",
        ),
      },
      dependencies: {
        show: (values: Record<string, unknown>) => values.sourceKind === "git",
        triggerFields: ["sourceKind"],
      },
      fieldName: "repoUrl",
      formItemClass: "md:col-span-2",
      label: t("plugin.linapro-plugin-marketplace.mine.fields.repoUrl"),
      rules: "required",
    },
    {
      component: "Input",
      componentProps: {
        autocomplete: "off",
        placeholder: t(
          "plugin.linapro-plugin-marketplace.mine.placeholders.accessToken",
        ),
        type: "password",
      },
      dependencies: {
        show: (values: Record<string, unknown>) => values.sourceKind === "git",
        triggerFields: ["sourceKind"],
      },
      fieldName: "accessToken",
      formItemClass: "md:col-span-2",
      help: t("plugin.linapro-plugin-marketplace.mine.hints.gitDiscover"),
      label: t("plugin.linapro-plugin-marketplace.mine.fields.accessToken"),
    },
  ];
}

function canPublishRow(row: MarketplacePluginListItem) {
  if (
    row.latestReviewStatus === "submitted" ||
    row.latestReviewStatus === "reviewing"
  ) {
    return false;
  }
  if (row.marketStatus === "delisted") {
    return true;
  }
  if (row.marketStatus === "draft") {
    return Boolean(row.latestVersion);
  }
  // Published plugins can publish again when a draft/rejected version is ready.
  return (
    row.latestReviewStatus === "draft" || row.latestReviewStatus === "rejected"
  );
}

function canDelistRow(row: MarketplacePluginListItem) {
  return row.marketStatus === "published";
}

function resolvePublisherKey(
  values: Record<string, unknown>,
  availablePublishers: MarketplacePublisherItem[],
) {
  const fromForm = stringValue(values.publisherKey);
  if (fromForm) {
    return fromForm;
  }
  const first = availablePublishers[0]?.publisherKey;
  if (first) {
    return first;
  }
  message.warning(
    t("plugin.linapro-plugin-marketplace.mine.messages.publisherRequired"),
  );
  throw new Error("validation");
}

onMounted(async () => {
  gridApi.setGridOptions({
    cellConfig: buildCellConfig(),
    columns: buildColumns(),
  });
  publisherFormApi.setState({ schema: buildPublisherSchema() });
  pluginFormApi.setState({ schema: buildPluginSchema([]) });
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

function openPublisherDrawer() {
  publisherDrawerApi.open();
}

function openPublishDrawer(row?: MarketplacePluginListItem) {
  publishMode.value = row ? "version" : "plugin";
  publishDrawerApi.setData(row ? { row } : {});
  publishDrawerApi.open();
}

async function resetPublishWorkflow() {
  await pluginFormApi.resetForm();
  publishMode.value = "plugin";
  publishSourceKind.value = "upload";
  publishers.value = [];
  versionTarget.value = null;
  uploadFileList.value = [];
  pluginFormApi.setState({ schema: buildPluginSchema([]) });
}

async function handleSavePublisher() {
  publisherDrawerApi.lock(true);
  try {
    const { valid } = await publisherFormApi.validate();
    if (!valid) {
      return;
    }
    const values = await publisherFormApi.getValues();
    const existing = boundPublisher.value;
    // Key is not user-editable: create allocates one, update keeps the bound key.
    const publisherKey = existing?.publisherKey || generatePublisherKey();
    const payload = {
      contactEmail: stringValue(values.contactEmail),
      homepage: stringValue(values.homepage),
      name: requiredString(values.name),
      publisherKey,
      summary: stringValue(values.summary),
    };
    const result = existing
      ? await marketplacePublisherUpdate(existing.publisherKey, payload)
      : await marketplacePublisherCreate(payload);
    // Keep the drawer open so the operator can continue editing; sync local
    // binding so the next save uses update and the title switches to edit.
    if (result?.publisher) {
      boundPublisher.value = result.publisher;
    } else if (!existing) {
      boundPublisher.value = {
        contactEmail: payload.contactEmail || "",
        homepage: payload.homepage || "",
        name: payload.name,
        publisherKey: payload.publisherKey,
        summary: payload.summary || "",
        verified: false,
      };
    }
    message.success({
      content: t(
        "plugin.linapro-plugin-marketplace.console.messages.publisherSaved",
      ),
      key: workflowMessageKey,
    });
  } catch {
    // Validation or API errors are shown by form/message handlers.
  } finally {
    publisherDrawerApi.lock(false);
  }
}

async function handleSubmitPluginBasics() {
  if (!hasPublishers.value) {
    message.warning(
      t("plugin.linapro-plugin-marketplace.mine.messages.publisherRequired"),
    );
    return;
  }

  const values = await pluginFormApi.getValues();
  if (values.sourceKind === "git") {
    await handleRegisterGitSource();
    return;
  }
  await handleAddPackage();
}

async function handlePublishDrawerPrimaryAction() {
  if (publishMode.value === "version") {
    await handleAddPackage();
    return;
  }
  await handleSubmitPluginBasics();
}

async function handleAddPackage() {
  publishDrawerApi.lock(true);
  try {
    if (publishMode.value === "plugin" && !hasPublishers.value) {
      message.warning(
        t("plugin.linapro-plugin-marketplace.mine.messages.publisherRequired"),
      );
      return;
    }
    const uploadFile = uploadFileList.value[0]?.originFileObj as
      | File
      | undefined;
    if (!uploadFile) {
      message.warning(
        t("plugin.linapro-plugin-marketplace.console.messages.fileRequired"),
      );
      return;
    }

    let publisherKey = "";
    if (publishMode.value === "plugin") {
      const values = await pluginFormApi.getValues();
      publisherKey = resolvePublisherKey(values, publishers.value);
    }

    await marketplacePackageAdd({
      file: uploadFile,
      publisherKey: publisherKey || undefined,
      replaceDraft: true,
    });
    message.success({
      content: t(
        "plugin.linapro-plugin-marketplace.mine.messages.packageAdded",
      ),
      key: workflowMessageKey,
    });
    await gridApi.reload();
    publishDrawerApi.close();
  } catch {
    // Validation or API errors are shown by form/message handlers.
  } finally {
    publishDrawerApi.lock(false);
  }
}

async function handleRegisterGitSource() {
  publishDrawerApi.lock(true);
  try {
    const { valid } = await pluginFormApi.validate();
    if (!valid) {
      return;
    }
    const values = await pluginFormApi.getValues();
    await marketplaceGitSourceRegister({
      accessToken: stringValue(values.accessToken) || undefined,
      publisherKey: resolvePublisherKey(values, publishers.value),
      repoUrl: requiredString(values.repoUrl),
    });
    message.success({
      content: t(
        "plugin.linapro-plugin-marketplace.mine.messages.gitRegistered",
      ),
      key: workflowMessageKey,
    });
    await gridApi.reload();
    publishDrawerApi.close();
  } catch {
    // Validation or API errors are shown by form/message handlers.
  } finally {
    publishDrawerApi.lock(false);
  }
}

async function handlePublishPlugin(row: MarketplacePluginListItem) {
  try {
    await marketplacePluginPublish(row.pluginId, {
      version: row.latestVersion || undefined,
    });
    message.success(
      t("plugin.linapro-plugin-marketplace.mine.messages.publishSubmitted"),
    );
    await gridApi.reload();
  } catch {
    // API errors are shown by request client handlers.
  }
}

async function handleDelistPlugin(row: MarketplacePluginListItem) {
  Modal.confirm({
    title: t("plugin.linapro-plugin-marketplace.mine.actions.delist"),
    content: t("plugin.linapro-plugin-marketplace.mine.confirm.delist"),
    async onOk() {
      await marketplacePluginDelist(row.pluginId);
      message.success(
        t("plugin.linapro-plugin-marketplace.mine.messages.delisted"),
      );
      await gridApi.reload();
    },
  });
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
          <a-button @click="openPublisherDrawer()">
            {{
              $t(
                "plugin.linapro-plugin-marketplace.mine.actions.registerPublisher",
              )
            }}
          </a-button>
          <a-button type="primary" @click="openPublishDrawer()">
            {{ $t("plugin.linapro-plugin-marketplace.mine.actions.add") }}
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
                <MenuItem
                  v-if="canPublishRow(row)"
                  key="publish"
                  @click="handlePublishPlugin(row)"
                >
                  {{
                    $t("plugin.linapro-plugin-marketplace.mine.actions.publish")
                  }}
                </MenuItem>
                <MenuItem
                  v-if="canDelistRow(row)"
                  key="delist"
                  @click="handleDelistPlugin(row)"
                >
                  {{
                    $t("plugin.linapro-plugin-marketplace.mine.actions.delist")
                  }}
                </MenuItem>
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
                    $t("plugin.linapro-plugin-marketplace.mine.actions.syncGit")
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

    <PublisherDrawer
      :title="publisherDrawerTitle"
      class="w-[560px] max-w-[calc(100vw-32px)]"
    >
      <div class="mine-publish-sections">
        <section>
          <div class="mine-section-header">
            <h3>
              {{
                $t(
                  "plugin.linapro-plugin-marketplace.console.sections.publisher",
                )
              }}
            </h3>
            <a-button type="primary" @click="handleSavePublisher">
              {{
                $t(
                  "plugin.linapro-plugin-marketplace.console.actions.savePublisher",
                )
              }}
            </a-button>
          </div>
          <PublisherForm />
        </section>
      </div>
    </PublisherDrawer>

    <PublishDrawer
      :title="drawerTitle"
      class="w-[760px] max-w-[calc(100vw-32px)]"
    >
      <div class="mine-publish-sections">
        <Alert
          v-if="publishMode === 'plugin' && !hasPublishers"
          type="warning"
          show-icon
          class="mine-publisher-alert"
          :message="
            $t(
              'plugin.linapro-plugin-marketplace.mine.messages.publisherRequired',
            )
          "
        >
          <template #action>
            <a-button size="small" type="primary" @click="openPublisherDrawer">
              {{
                $t(
                  "plugin.linapro-plugin-marketplace.mine.actions.registerPublisher",
                )
              }}
            </a-button>
          </template>
        </Alert>

        <section v-show="publishMode === 'plugin' && hasPublishers">
          <div class="mine-section-header">
            <h3>
              {{
                $t(
                  "plugin.linapro-plugin-marketplace.mine.sections.pluginBasic",
                )
              }}
            </h3>
          </div>
          <PluginForm>
            <template #packageFile>
              <div class="mine-package-field" data-testid="mine-package-field">
                <Alert
                  type="info"
                  show-icon
                  class="mine-package-hint"
                  :message="
                    $t(
                      'plugin.linapro-plugin-marketplace.mine.hints.packageAutoParse',
                    )
                  "
                />
                <UploadDragger
                  v-model:file-list="uploadFileList"
                  accept=".zip,.tar.gz,.tgz"
                  :before-upload="() => false"
                  :max-count="1"
                  class="mine-upload"
                >
                  <p
                    class="ant-upload-drag-icon flex items-center justify-center"
                  >
                    <IconifyIcon
                      class="text-primary text-5xl"
                      icon="ant-design:inbox-outlined"
                    />
                  </p>
                  <p class="ant-upload-text">
                    {{
                      $t(
                        "plugin.linapro-plugin-marketplace.console.upload.dragText",
                      )
                    }}
                  </p>
                  <p class="ant-upload-hint">
                    {{
                      $t(
                        "plugin.linapro-plugin-marketplace.console.upload.hint",
                      )
                    }}
                  </p>
                </UploadDragger>
              </div>
            </template>
          </PluginForm>
        </section>

        <section v-show="publishMode === 'version'">
          <div class="mine-section-header">
            <h3>
              {{
                $t(
                  "plugin.linapro-plugin-marketplace.mine.sections.packageUpload",
                )
              }}
            </h3>
          </div>
          <Alert
            type="info"
            show-icon
            class="mine-package-hint"
            :message="
              $t(
                'plugin.linapro-plugin-marketplace.mine.hints.packageAutoParse',
              )
            "
          />
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
        </section>

        <div
          v-if="
            publishMode === 'version' ||
            (publishMode === 'plugin' && hasPublishers)
          "
          class="mine-drawer-actions"
        >
          <a-button type="primary" @click="handlePublishDrawerPrimaryAction">
            {{ publishPrimaryActionLabel }}
          </a-button>
        </div>
      </div>
    </PublishDrawer>
  </Page>
</template>

<style scoped>
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

.mine-package-hint {
  margin-bottom: 0;
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

.mine-package-field {
  display: flex;
  width: 100%;
  min-width: 0;
  flex-direction: column;
  gap: 12px;
}

.mine-drawer-actions {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
  border-top: 1px solid var(--ant-color-border-secondary);
}

.mine-publisher-alert {
  margin-bottom: 0;
}

.mine-upload {
  width: 100%;
}

.mine-draft-result {
  margin-top: 12px;
}
</style>
