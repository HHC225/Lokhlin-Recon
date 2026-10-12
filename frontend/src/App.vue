<script setup lang="ts">
import { computed, watch } from 'vue'
import { useAppStore } from '@/store/modules/app'
import { ConfigGlobal } from '@/components/ConfigGlobal'
import { useDesign } from '@/hooks/web/useDesign'
import { useStorage } from '@/hooks/web/useStorage'
import { getCssVar, setCssVar } from './utils'
import { isDark } from '@/utils/is'
import { isConsoleGateway, isEmbedded, consoleSession, consoleTheme } from '@/integration/console'
import ConsoleSession from '@/views/ConsoleSession.vue'

const { getPrefixCls } = useDesign()

const prefixCls = getPrefixCls('app')

const appStore = useAppStore()
if (isConsoleGateway) {
  watch(
    () => consoleTheme.dark,
    (dark) => {
      appStore.isDark = dark
    },
    { immediate: true }
  )
}

const currentSize = computed(() => appStore.getCurrentSize)

const greyMode = computed(() => appStore.getGreyMode)

const { getStorage } = useStorage('localStorage')

const setDefaultTheme = () => {
  if (isConsoleGateway) {
    appStore.setFooter(false)
    if (isEmbedded) {
      appStore.setTagsView(false)
      setCssVar('--tags-view-height', '0px')
    }
    return
  }
  if (getStorage('isDark') !== null) {
    const isDarkValue = getStorage('isDark')
    appStore.setIsDark(isDarkValue)

    const color = getCssVar('--el-bg-color')
    appStore.setMenuTheme(color)
    appStore.setHeaderTheme(color)
  } else {
    const isDarkTheme = isDark()
    appStore.setIsDark(isDarkTheme)

    const color = getCssVar('--el-bg-color')
    appStore.setMenuTheme(color)
    appStore.setHeaderTheme(color)
  }

  const tagsViewEnabled = appStore.getTagsView
  setCssVar('--tags-view-height', tagsViewEnabled ? '35px' : '0px')
}

setDefaultTheme()
</script>

<template>
  <ConfigGlobal :size="currentSize">
    <ConsoleSession v-if="isConsoleGateway && consoleSession.status !== 'ready'" />
    <RouterView v-else :class="greyMode ? `${prefixCls}-grey-mode` : ''" />
  </ConfigGlobal>
</template>

<style lang="less">
@prefix-cls: ~'@{namespace}-app';

.size {
  width: 100%;
  height: 100%;
}

html,
body {
  padding: 0 !important;
  margin: 0;
  overflow: hidden;
  .size;

  #app {
    .size;
  }
}

.@{prefix-cls}-grey-mode {
  filter: grayscale(100%);
}
</style>
