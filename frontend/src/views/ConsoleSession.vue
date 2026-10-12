<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '@/hooks/web/useI18n'
import { consoleSession } from '@/integration/console'
import { bootstrapConsoleSession } from '@/integration/console'
import { ElButton } from 'element-plus'

const { t } = useI18n()
const noticeStatus = computed(() =>
  consoleSession.status === 'ready' ? 'loading' : consoleSession.status
)
const title = computed(() => t(`console.${noticeStatus.value}Title`))
const description = computed(() => t(`console.${noticeStatus.value}Description`))
const retry = async () => {
  if (await bootstrapConsoleSession(true)) {
    window.location.reload()
  }
}
</script>

<template>
  <main class="recon-session" role="status" aria-live="polite">
    <div>
      <h1>{{ title }}</h1>
      <p>{{ description }}</p>
      <div
        v-if="!['loading', 'ready'].includes(consoleSession.status)"
        class="recon-session-actions"
      >
        <ElButton v-if="consoleSession.status !== 'denied'" @click="retry">{{
          t('console.retry')
        }}</ElButton>
        <a href="/" target="_top">{{ t('console.returnToConsole') }}</a>
      </div>
    </div>
  </main>
</template>
