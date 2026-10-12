<script setup lang="ts">
import { ContentWrap } from '@/components/ContentWrap'
import { useI18n } from '@/hooks/web/useI18n'
import { displayName } from '@/integration/console'
const { t } = useI18n()
</script>

<template>
  <ContentWrap :title="displayName"
    ><p>{{ t('console.about') }}</p></ContentWrap
  >
</template>
