<script setup lang="ts">
import { ElDrawer, ElDivider, ElMessage } from 'element-plus'
import { ref, unref, computed, watch } from 'vue'
import { useI18n } from '@/hooks/web/useI18n'
import { ThemeSwitch } from '@/components/ThemeSwitch'
import { colorIsDark, lighten, hexToRGB } from '@/utils/color'
import { useCssVar } from '@vueuse/core'
import { useAppStore } from '@/store/modules/app'
import { trim, setCssVar } from '@/utils'
import ColorRadioPicker from './components/ColorRadioPicker.vue'
import InterfaceDisplay from './components/InterfaceDisplay.vue'
import LayoutRadioPicker from './components/LayoutRadioPicker.vue'
import { useStorage } from '@/hooks/web/useStorage'
import { useClipboard } from '@vueuse/core'
import { useDesign } from '@/hooks/web/useDesign'

const { clear: storageClear } = useStorage('localStorage')

const { getPrefixCls } = useDesign()

const prefixCls = getPrefixCls('setting')

const appStore = useAppStore()

const { t } = useI18n()

const layout = computed(() => appStore.getLayout)

const drawer = ref(false)

const systemTheme = ref(appStore.getTheme.elColorPrimary)

const setSystemTheme = (color: string) => {
  setCssVar('--el-color-primary', color)
  appStore.setTheme({ elColorPrimary: color })
  const leftMenuBgColor = useCssVar('--left-menu-bg-color', document.documentElement)
  setMenuTheme(trim(unref(leftMenuBgColor)))
}

const headerTheme = ref(appStore.getTheme.topHeaderBgColor || '')

const setHeaderTheme = (color: string) => {
  const isDarkColor = colorIsDark(color)
  const textColor = isDarkColor ? '#fff' : 'inherit'
  const textHoverColor = isDarkColor ? lighten(color!, 6) : '#f6f6f6'
  const topToolBorderColor = isDarkColor ? color : '#eee'
  setCssVar('--top-header-bg-color', color)
  setCssVar('--top-header-text-color', textColor)
  setCssVar('--top-header-hover-color', textHoverColor)
  appStore.setTheme({
    topHeaderBgColor: color,
    topHeaderTextColor: textColor,
    topHeaderHoverColor: textHoverColor,
    topToolBorderColor
  })
  if (unref(layout) === 'top') {
    setMenuTheme(color)
  }
}

const menuTheme = ref(appStore.getTheme.leftMenuBgColor || '')

const setMenuTheme = (color: string) => {
  const primaryColor = useCssVar('--el-color-primary', document.documentElement)
  const isDarkColor = colorIsDark(color)
  const theme: Recordable = {
    leftMenuBorderColor: isDarkColor ? 'inherit' : '#eee',

    leftMenuBgColor: color,

    leftMenuBgLightColor: isDarkColor ? lighten(color!, 6) : color,

    leftMenuBgActiveColor: isDarkColor
      ? 'var(--el-color-primary)'
      : hexToRGB(unref(primaryColor), 0.1),

    leftMenuCollapseBgActiveColor: isDarkColor
      ? 'var(--el-color-primary)'
      : hexToRGB(unref(primaryColor), 0.1),

    leftMenuTextColor: isDarkColor ? '#bfcbd9' : '#333',

    leftMenuTextActiveColor: isDarkColor ? '#fff' : 'var(--el-color-primary)',

    logoTitleTextColor: isDarkColor ? '#fff' : 'inherit',

    logoBorderColor: isDarkColor ? color : '#eee'
  }
  appStore.setTheme(theme)
  appStore.setCssVarTheme()
}
setMenuTheme(unref(menuTheme))

watch(
  () => layout.value,
  (n) => {
    if (n === 'top' && !appStore.getIsDark) {
      headerTheme.value = '#fff'
      setHeaderTheme('#fff')
    } else {
      setMenuTheme(unref(menuTheme))
    }
  }
)

const copyConfig = async () => {
  const { copy, copied, isSupported } = useClipboard({
    source: `

      breadcrumb: ${appStore.getBreadcrumb},

      breadcrumbIcon: ${appStore.getBreadcrumbIcon},

      hamburger: ${appStore.getHamburger},

      screenfull: ${appStore.getScreenfull},

      size: ${appStore.getSize},

      locale: ${appStore.getLocale},

      tagsView: ${appStore.getTagsView},

      getTagsViewIcon: ${appStore.getTagsViewIcon},
      // logo
      logo: ${appStore.getLogo},

      uniqueOpened: ${appStore.getUniqueOpened},

      fixedHeader: ${appStore.getFixedHeader},

      footer: ${appStore.getFooter},

      greyMode: ${appStore.getGreyMode},

      layout: '${appStore.getLayout}',

      isDark: ${appStore.getIsDark},

      currentSize: '${appStore.getCurrentSize}',

      theme: {

        elColorPrimary: '${appStore.getTheme.elColorPrimary}',

        leftMenuBorderColor: '${appStore.getTheme.leftMenuBorderColor}',

        leftMenuBgColor: '${appStore.getTheme.leftMenuBgColor}',

        leftMenuBgLightColor: '${appStore.getTheme.leftMenuBgLightColor}',

        leftMenuBgActiveColor: '${appStore.getTheme.leftMenuBgActiveColor}',

        leftMenuCollapseBgActiveColor: '${appStore.getTheme.leftMenuCollapseBgActiveColor}',

        leftMenuTextColor: '${appStore.getTheme.leftMenuTextColor}',

        leftMenuTextActiveColor: '${appStore.getTheme.leftMenuTextActiveColor}',

        logoTitleTextColor: '${appStore.getTheme.logoTitleTextColor}',

        logoBorderColor: '${appStore.getTheme.logoBorderColor}',

        topHeaderBgColor: '${appStore.getTheme.topHeaderBgColor}',

        topHeaderTextColor: '${appStore.getTheme.topHeaderTextColor}',

        topHeaderHoverColor: '${appStore.getTheme.topHeaderHoverColor}',

        topToolBorderColor: '${appStore.getTheme.topToolBorderColor}'
      }
    `,
    legacy: true
  })
  if (!isSupported) {
    ElMessage.error(t('setting.copyFailed'))
  } else {
    await copy()
    if (unref(copied)) {
      ElMessage.success(t('setting.copySuccess'))
    }
  }
}

const clear = () => {
  storageClear()
  window.location.reload()
}
</script>

<template>
  <div
    :class="prefixCls"
    class="fixed top-[90%] right-0 w-40px h-40px flex items-center justify-center bg-[var(--el-color-primary)] cursor-pointer z-10"
    @click="drawer = true"
  >
    <Icon icon="ant-design:setting-outlined" color="#fff" />
  </div>

  <ElDrawer v-model="drawer" direction="rtl" size="350px" :z-index="4000">
    <template #header>
      <span class="text-16px font-700">{{ t('setting.projectSetting') }}</span>
    </template>

    <div class="text-center">
      <ElDivider>{{ t('setting.theme') }}</ElDivider>
      <ThemeSwitch />

      <ElDivider>{{ t('setting.layout') }}</ElDivider>
      <LayoutRadioPicker />

      <ElDivider>{{ t('setting.systemTheme') }}</ElDivider>
      <ColorRadioPicker
        v-model="systemTheme"
        :schema="[
          '#409eff',
          '#009688',
          '#536dfe',
          '#ff5c93',
          '#ee4f12',
          '#0096c7',
          '#9c27b0',
          '#ff9800'
        ]"
        @change="setSystemTheme"
      />

      <ElDivider>{{ t('setting.headerTheme') }}</ElDivider>
      <ColorRadioPicker
        v-model="headerTheme"
        :schema="[
          '#fff',
          '#151515',
          '#5172dc',
          '#e74c3c',
          '#24292e',
          '#394664',
          '#009688',
          '#383f45'
        ]"
        @change="setHeaderTheme"
      />

      <template v-if="layout !== 'top'">
        <ElDivider>{{ t('setting.menuTheme') }}</ElDivider>
        <ColorRadioPicker
          v-model="menuTheme"
          :schema="[
            '#fff',
            '#001529',
            '#212121',
            '#273352',
            '#191b24',
            '#383f45',
            '#001628',
            '#344058'
          ]"
          @change="setMenuTheme"
        />
      </template>
    </div>

    <ElDivider>{{ t('setting.interfaceDisplay') }}</ElDivider>
    <InterfaceDisplay />

    <ElDivider />
    <div>
      <BaseButton type="primary" class="w-full" @click="copyConfig">{{
        t('setting.copy')
      }}</BaseButton>
    </div>
    <div class="mt-5px">
      <BaseButton type="danger" class="w-full" @click="clear">
        {{ t('setting.clearAndReset') }}
      </BaseButton>
    </div>
  </ElDrawer>
</template>

<style lang="less" scoped>
@prefix-cls: ~'@{namespace}-setting';

.@{prefix-cls} {
  border-radius: 6px 0 0 6px;
}
</style>
