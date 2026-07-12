<script lang="ts">
import type { PluginPageMeta } from '#/plugins/page-registry';

// Legacy route kept for compatibility; redirects to the review workbench.
export const pluginPageMeta = {
  pluginId: 'linapro-plugin-marketplace',
  routePath: 'plugin-marketplace-console',
  title: 'Plugin Marketplace Console',
} satisfies PluginPageMeta;
</script>

<script setup lang="ts">
import { onMounted } from 'vue';
import { useRouter } from 'vue-router';

import { marketplaceReviewPath } from '../../utils/routes';

const router = useRouter();

onMounted(() => {
  void router.replace(marketplaceReviewPath());
});
</script>

<template>
  <div />
</template>
