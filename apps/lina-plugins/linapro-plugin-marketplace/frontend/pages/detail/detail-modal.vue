<script setup lang="ts">
import type { MarketplaceDetailSource } from "../../utils/routes";

import { computed, ref } from "vue";

import { useVbenModal } from "@vben/common-ui";

import { $t } from "#/locales";

import MarketplaceDetail from "./index.vue";

type DetailModalData = {
  from?: MarketplaceDetailSource;
  pluginId: string;
};

const activePluginId = ref("");
const activeFrom = ref<MarketplaceDetailSource | "">("");

const [BasicModal, modalApi] = useVbenModal({
  destroyOnClose: true,
  onClosed() {
    activePluginId.value = "";
    activeFrom.value = "";
  },
  onOpenChange(isOpen) {
    if (!isOpen) {
      return;
    }
    const data = modalApi.getData<DetailModalData>();
    activePluginId.value = data?.pluginId?.trim() || "";
    activeFrom.value = data?.from || "";
  },
});

const modalTitle = computed(() => {
  return $t("plugin.linapro-plugin-marketplace.detail.modalTitle");
});

function handleClose() {
  modalApi.close();
}
</script>

<template>
  <BasicModal
    :footer="false"
    :title="modalTitle"
    class="marketplace-detail-modal w-[920px] max-w-[calc(100vw-32px)]"
    content-class="marketplace-detail-modal-content"
  >
    <MarketplaceDetail
      v-if="activePluginId"
      :key="`${activeFrom}:${activePluginId}`"
      embedded
      :from="activeFrom"
      :plugin-id="activePluginId"
      @close="handleClose"
    />
  </BasicModal>
</template>

<style scoped>
/*
 * content-class lands on Vben's flex-1 overflow pane. Keep it a flex column so
 * MarketplaceDetail can stretch when the dialog is maximized (size-full).
 */
:deep(.marketplace-detail-modal-content) {
  display: flex;
  min-height: min(72vh, 820px);
  flex-direction: column;
  padding-top: 8px;
}
</style>
