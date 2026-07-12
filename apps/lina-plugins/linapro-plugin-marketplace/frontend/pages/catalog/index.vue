<script lang="ts">
import type { PluginPageMeta } from '#/plugins/page-registry';

// Legacy route kept for compatibility; redirects to the publisher workbench.
export const pluginPageMeta = {
  pluginId: 'linapro-plugin-marketplace',
  routePath: 'plugin-marketplace-catalog',
  title: 'Plugin Marketplace',
} satisfies PluginPageMeta;
</script>

<script setup lang="ts">
import { onMounted } from 'vue';
import { useRouter } from 'vue-router';

import { marketplaceMinePath } from '../../utils/routes';

const router = useRouter();

onMounted(() => {
  void router.replace(marketplaceMinePath());
});
</script>

<template>
  <div />
</template>
