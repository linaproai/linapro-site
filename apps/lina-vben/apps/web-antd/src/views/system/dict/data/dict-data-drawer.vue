<script setup lang="ts">
import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { message } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import {
  dictDataAdd,
  dictDataInfo,
  dictDataUpdate,
} from '#/api/system/dict/dict-data';

import { drawerSchema } from './data';

const emit = defineEmits<{ reload: [dictType: string] }>();

interface DrawerProps {
  dictType: string;
  id?: number;
}

const isEdit = ref(false);
const editId = ref<number>(0);
const title = computed(() =>
  isEdit.value
    ? $t('pages.system.dict.data.drawer.editTitle')
    : $t('pages.system.dict.data.drawer.createTitle'),
);

const [Form, formApi] = useVbenForm({
  commonConfig: {
    componentProps: {
      class: 'w-full',
    },
    formItemClass: 'col-span-2',
    labelWidth: 112,
  },
  schema: drawerSchema,
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2',
});

const [Drawer, drawerApi] = useVbenDrawer({
  async onOpenChange(open) {
    if (!open) {
      return;
    }
    drawerApi.setState({ loading: true });

    const { dictType, id } = drawerApi.getData() as DrawerProps;
    isEdit.value = !!id;
    editId.value = id ?? 0;
    await formApi.setFieldValue('dictType', dictType);

    if (id && isEdit.value) {
      const record = await dictDataInfo(id);
      await formApi.setValues(record);
    }

    drawerApi.setState({ loading: false });
  },
  async onConfirm() {
    try {
      drawerApi.lock(true);
      const { valid } = await formApi.validate();
      if (!valid) {
        return;
      }
      const data = await formApi.getValues();

      if (isEdit.value) {
        await dictDataUpdate(editId.value, data);
        message.success($t('pages.common.updateSuccess'));
      } else {
        await dictDataAdd(data);
        message.success($t('pages.common.createSuccess'));
      }

      emit('reload', data.dictType);
      drawerApi.close();
    } catch (error) {
      console.error(error);
    } finally {
      drawerApi.lock(false);
    }
  },
  onClosed() {
    formApi.resetForm();
  },
});
</script>

<template>
  <Drawer :title="title" class="w-[700px] max-w-[calc(100vw-32px)]">
    <Form />
  </Drawer>
</template>
