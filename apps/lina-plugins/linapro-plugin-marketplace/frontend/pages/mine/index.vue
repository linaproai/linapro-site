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
  MarketplaceListStatusFilter,
  MarketplacePluginListItem,
  MarketplacePluginType,
  MarketplaceProcessStatus,
  MarketplacePublisherItem,
  MarketplaceStatus,
} from "../../types/marketplace";

import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";

import { Page, useVbenDrawer, useVbenModal } from "@vben/common-ui";
import { IconifyIcon } from "@vben/icons";
import { preferences } from "@vben/preferences";

import {
  Alert,
  Modal,
  Space,
  Tag,
  Tooltip,
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
  marketplacePublisherCreate,
  marketplacePublisherList,
  marketplacePublisherUpdate,
} from "../../api/marketplace";
import DetailModalContent from "../detail/detail-modal.vue";

type GridPageInfo = {
  currentPage: number;
  pageSize: number;
};

type MineFormValues = {
  keyword?: string;
  pluginType?: MarketplacePluginType;
  status?: MarketplaceListStatusFilter;
};

type MineGridOptions = NonNullable<
  VxeGridProps<MarketplacePluginListItem>["gridOptions"]
>;

type PublishSourceKind = "git" | "upload";
type PublishMode = "plugin" | "version";

const UploadDragger = Upload.Dragger;
const route = useRoute();
const workflowMessageKey = "plugin-marketplace-publish-workflow";
const mineRowClickIgnoredFields = new Set(["action"]);

const [DetailModal, detailModalApi] = useVbenModal({
  connectedComponent: DetailModalContent,
});
const publishMode = ref<PublishMode>("plugin");
const publishSourceKind = ref<PublishSourceKind>("git");
const publishers = ref<MarketplacePublisherItem[]>([]);
const boundPublisher = ref<MarketplacePublisherItem | null>(null);
/** Toolbar label only — not cleared when the publish drawer resets form state. */
const toolbarHasPublisher = ref(false);
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
const publisherToolbarLabel = computed(() =>
  toolbarHasPublisher.value
    ? t("plugin.linapro-plugin-marketplace.mine.actions.editPublisher")
    : t("plugin.linapro-plugin-marketplace.mine.actions.registerPublisher"),
);

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
    rowClassName: "cursor-pointer",
    showOverflow: "ellipsis",
    id: "plugin-marketplace-mine",
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
      if (shouldIgnoreMineRowClick(column?.field, $event)) {
        return;
      }
      handleOpenDetail(row);
    },
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
      publishSourceKind.value = "git";
      const result = await marketplacePublisherList({
        pageNum: 1,
        pageSize: 100,
      });
      publishers.value = result.items;
      pluginFormApi.setState({
        schema: buildPluginSchema(publishers.value),
      });
      await pluginFormApi.setValues({ sourceKind: "git" });
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

function trimOptional(value?: string) {
  const normalized = value?.trim();
  return normalized || undefined;
}

function buildFormOptions() {
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

function buildColumns(): MineGridOptions["columns"] {
  // Keep status early (ops-critical). pluginId/name minWidth floors stay readable
  // for typical plugin ids and localized names; do not tighten them for density.
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
      className: "mine-name-column",
      field: "name",
      headerAlign: "center",
      minWidth: 200,
      showOverflow: "ellipsis",
      slots: { default: "name" },
      title: t("plugin.linapro-plugin-marketplace.mine.columns.name"),
    },
    {
      align: "left",
      className: "mine-summary-column",
      field: "summary",
      headerAlign: "center",
      // Wider than name so short plugin descriptions stay readable before ellipsis.
      minWidth: 260,
      showOverflow: "ellipsis",
      slots: { default: "summary" },
      title: t("plugin.linapro-plugin-marketplace.mine.columns.summary"),
    },
    {
      field: "marketStatus",
      showOverflow: "ellipsis",
      slots: { default: "marketStatus" },
      title: t("plugin.linapro-plugin-marketplace.mine.fields.status"),
      // Wide enough for en-US "Pending Review" / zh-CN "待审核" tags.
      width: 124,
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
      field: "sourceKind",
      slots: { default: "sourceKind" },
      title: t("plugin.linapro-plugin-marketplace.mine.columns.sourceKind"),
      width: 108,
    },
    {
      field: "pluginType",
      slots: { default: "pluginType" },
      title: t("plugin.linapro-plugin-marketplace.catalog.columns.type"),
      width: 96,
    },
    {
      // Match review submittedAt: full "YYYY-MM-DD HH:mm:ss" without clipping.
      className: "mine-updated-at-column",
      field: "updatedAt",
      formatter: ({ cellValue }: { cellValue?: null | number | string }) =>
        formatTimestamp(cellValue),
      title: t("plugin.linapro-plugin-marketplace.catalog.columns.updatedAt"),
      width: 184,
    },
    {
      field: "action",
      fixed: "right",
      slots: { default: "action" },
      title: t("plugin.linapro-plugin-marketplace.catalog.columns.actions"),
      // Three ghost buttons: Details / New Version / Delist (en needs ~228px).
      width: 232,
    },
  ];
}

function shouldIgnoreMineRowClick(
  columnField?: string,
  event?: Event,
): boolean {
  if (columnField && mineRowClickIgnoredFields.has(columnField)) {
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
      label: t("plugin.linapro-plugin-marketplace.mine.sourceKind.git"),
      value: "git",
    },
    {
      label: t("plugin.linapro-plugin-marketplace.mine.sourceKind.upload"),
      value: "upload",
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
      defaultValue: "git",
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
      label: t("plugin.linapro-plugin-marketplace.mine.sections.packageUpload"),
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

async function loadPublishersForToolbar() {
  try {
    const result = await marketplacePublisherList({
      pageNum: 1,
      pageSize: 1,
    });
    toolbarHasPublisher.value = result.items.length > 0;
  } catch {
    // Toolbar label falls back to register; list reload still works independently.
  }
}

onMounted(async () => {
  gridApi.setState({ formOptions: buildFormOptions() });
  gridApi.setGridOptions({ columns: buildColumns() });
  publisherFormApi.setState({ schema: buildPublisherSchema() });
  pluginFormApi.setState({ schema: buildPluginSchema([]) });
  await Promise.all([gridApi.reload(), loadPublishersForToolbar()]);
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
    publisherFormApi.setState({ schema: buildPublisherSchema() });
    pluginFormApi.setState({ schema: buildPluginSchema(publishers.value) });
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

function formatMarketStatus(
  status: MarketplaceStatus,
  processStatus?: MarketplaceProcessStatus,
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

function getMarketStatusColor(
  status: MarketplaceStatus,
  processStatus?: MarketplaceProcessStatus,
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

function buildStatusTooltip(row: MarketplacePluginListItem) {
  const parts = [
    formatMarketStatus(row.marketStatus, row.processStatus),
    row.latestReviewStatus
      ? `${t("plugin.linapro-plugin-marketplace.detail.columns.reviewStatus")}: ${formatReviewStatus(row.latestReviewStatus)}`
      : "",
    row.lastSyncMessage || "",
  ].filter(Boolean);
  return parts.join(" · ");
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
  detailModalApi.setData({ from: "mine", pluginId: row.pluginId });
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
  detailModalApi.setData({ from: "mine", pluginId });
  detailModalApi.open();
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
  publishSourceKind.value = "git";
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
    toolbarHasPublisher.value = true;
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

async function handleDelistPlugin(row: MarketplacePluginListItem) {
  if (!canDelistRow(row)) {
    message.warning(
      t("plugin.linapro-plugin-marketplace.mine.messages.delistOnlyPublished"),
    );
    return;
  }
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

/**
 * "新版本" asks the server to refresh version metadata:
 * - git: trigger metadata discovery/sync
 * - upload: open package upload so the server can parse a new package version
 */
async function handleNewVersion(row: MarketplacePluginListItem) {
  if (row.sourceKind === "git") {
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
    return;
  }
  openPublishDrawer(row);
}
</script>

<template>
  <Page :auto-content-height="true">
    <DetailModal />
    <Grid
      class="plugin-marketplace-mine"
      :table-title="$t('plugin.linapro-plugin-marketplace.mine.tableTitle')"
    >
      <template #toolbar-tools>
        <Space>
          <a-button @click="openPublisherDrawer()">
            {{ publisherToolbarLabel }}
          </a-button>
          <a-button type="primary" @click="openPublishDrawer()">
            {{ $t("plugin.linapro-plugin-marketplace.mine.actions.add") }}
          </a-button>
        </Space>
      </template>

      <template #name="{ row }">
        <div class="max-w-full truncate font-medium" :title="row.name">
          {{ row.name }}
        </div>
      </template>

      <template #sourceKind="{ row }">
        <Tag :color="getSourceKindColor(row.sourceKind)">
          {{ formatSourceKind(row.sourceKind) }}
        </Tag>
      </template>

      <template #summary="{ row }">
        <div class="max-w-full truncate" :title="row.summary || '-'">
          {{ row.summary || "-" }}
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

      <template #marketStatus="{ row }">
        <Tooltip :title="buildStatusTooltip(row)">
          <Tag
            :color="getMarketStatusColor(row.marketStatus, row.processStatus)"
          >
            {{ formatMarketStatus(row.marketStatus, row.processStatus) }}
          </Tag>
        </Tooltip>
      </template>

      <template #action="{ row }">
        <Space :size="4" :wrap="false">
          <ghost-button @click.stop="handleOpenDetail(row)">
            {{ $t("plugin.linapro-plugin-marketplace.catalog.actions.detail") }}
          </ghost-button>
          <ghost-button @click.stop="handleNewVersion(row)">
            {{
              $t("plugin.linapro-plugin-marketplace.mine.actions.newVersion")
            }}
          </ghost-button>
          <ghost-button
            :disabled="!canDelistRow(row)"
            @click.stop="handleDelistPlugin(row)"
          >
            {{ $t("plugin.linapro-plugin-marketplace.mine.actions.delist") }}
          </ghost-button>
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
:deep(.mine-name-column .vxe-cell),
:deep(.mine-summary-column .vxe-cell) {
  overflow: hidden;
}

:deep(.mine-updated-at-column .vxe-cell) {
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
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

/* Keep dense action buttons on one line inside the fixed action column. */
:deep(.plugin-marketplace-mine .vxe-body--column.col--fixed .ant-space) {
  flex-wrap: nowrap;
}
</style>
