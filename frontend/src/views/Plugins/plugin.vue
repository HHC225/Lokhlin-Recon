<script setup lang="tsx">
import { useI18n } from '@/hooks/web/useI18n'
import { ref, computed, provide, onMounted } from 'vue'
import { ElTabs, ElTabPane, ElMessage, ElMessageBox } from 'element-plus'
import ClientPlugin from './components/ClientPlugin.vue'
import ServerPlugin from './components/ServerPlugin.vue'
import PluginMarket from './components/PluginMarket.vue'
import {
  getLocalPluginListApi,
  getRemotePluginMarketApi,
  getPluginExportDataApi,
  importPluginApi
} from '@/api/plugins'
import type { RemotePluginData, pluginData } from '@/api/plugins/types'
import { catalogConfigured } from '@/integration/catalog'
import { isConsoleGateway } from '@/integration/console'

const { t } = useI18n()
const activeTab = ref('client')

const marketDrawerVisible = ref(false)
const remotePluginList = ref<RemotePluginData[]>([])
const marketLoading = ref(false)
const pluginKey = ref('')
const clientPluginRef = ref<InstanceType<typeof ClientPlugin>>()
const serverPluginRef = ref<InstanceType<typeof ServerPlugin>>()

const loadRemotePlugins = async () => {
  marketLoading.value = true
  try {
    const [localRes, remoteRes] = await Promise.all([
      getLocalPluginListApi(),
      getRemotePluginMarketApi()
    ])

    const localPluginsMap = new Map<string, pluginData>()
    if (localRes.code === 200 && localRes.data?.list) {
      localRes.data.list.forEach((plugin: pluginData) => {
        localPluginsMap.set(plugin.hash, plugin)
      })
    }

    let remotePlugins: RemotePluginData[] = []
    if (remoteRes.status === '200' && remoteRes.data?.data) {
      remotePlugins = remoteRes.data.data.map((plugin: any) => {
        const localPlugin = localPluginsMap.get(plugin.hash)
        const isInstalled = !!localPlugin
        let needUpdate = false

        if (isInstalled && localPlugin) {
          const localVersion = localPlugin.version || ''
          const remoteVersion = plugin.version || ''
          needUpdate =
            localVersion !== remoteVersion && remoteVersion !== null && remoteVersion !== ''
        }

        let isSystem = false
        if (plugin.tags) {
          const tags = plugin.tags.split(',').map((tag: string) => tag.trim())
          isSystem = tags.some((tag: string) =>
            ['built-in', 'builtin', '내장', '\u5185\u7f6e'].includes(tag.toLowerCase())
          )
        }

        return {
          id: plugin.id,
          name: plugin.name?.trim() || plugin.hash || `Plugin-${plugin.id}`,
          module: plugin.module || '',
          priceStatus: plugin.priceStatus,
          price: plugin.price,
          hash: plugin.hash || '',
          introduction: plugin.introduction || '',
          version: plugin.version || '',
          createTime: plugin.createTime || '',
          username: plugin.username || '',
          isInstalled,
          needUpdate,
          isSystem,
          type: plugin.type || 'scan'
        }
      })
    }

    remotePluginList.value = remotePlugins
  } catch (error) {
    console.error('Error loading plugins:', error)
    ElMessage.error(error instanceof Error ? error.message : t('console.loadFailed'))
  } finally {
    marketLoading.value = false
  }
}

const handleInstallPlugin = async (plugin: RemotePluginData, token?: string) => {
  try {
    const action =
      plugin.isInstalled && plugin.needUpdate ? t('plugin.update') : t('plugin.install')

    let pluginName = t('plugin.name')
    if (plugin && plugin.name) {
      const trimmedName = String(plugin.name).trim()
      if (trimmedName) {
        pluginName = trimmedName
      }
    }

    if (plugin.priceStatus === 0) {
      const confirmMessage = t('plugin.confirmInstall', {
        action,
        name: pluginName
      })

      await ElMessageBox.confirm(confirmMessage, t('common.reminder'), {
        confirmButtonText: t('common.ok'),
        cancelButtonText: t('common.cancel'),
        type: 'info'
      })
    }

    const loadingMessage = ElMessage({
      message: `${action}...`,
      type: 'info',
      duration: 0,
      showClose: false
    })

    try {
      const exportRes = await getPluginExportDataApi(plugin.hash, token)

      if (exportRes.status !== '200' || !exportRes.data) {
        loadingMessage.close()
        ElMessage.error(exportRes.message || t('console.loadFailed'))
        return
      }

      const { json, source } = exportRes.data

      const importRes = await importPluginApi(
        json || '',
        source || '',
        plugin.isSystem || false,
        pluginKey.value || ''
      )

      loadingMessage.close()

      if (importRes.code === 200) {
        const successMessage = t('plugin.installSuccess', { action })
        ElMessage.success(successMessage)
        ;(clientPluginRef.value as any)?.refreshList?.()
        ;(serverPluginRef.value as any)?.refreshList?.()

        loadRemotePlugins()
      } else {
        ElMessage.error(importRes.message || t('plugin.installFailed'))
      }
    } catch (error: any) {
      loadingMessage.close()
      console.error('Error installing plugin:', error)
      ElMessage.error(error?.response?.data?.message || error?.message || t('plugin.installFailed'))
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('Error:', error)
    }
  }
}

const openMarketDialog = () => {
  if (!catalogConfigured) {
    ElMessage.info(t('console.catalogUnavailable'))
    return
  }
  marketDrawerVisible.value = true
  if (remotePluginList.value.length === 0) {
    loadRemotePlugins()
  }
}

const pendingPluginsCount = computed(() => {
  return remotePluginList.value.filter((plugin) => !plugin.isInstalled || plugin.needUpdate).length
})

provide('openMarketDialog', openMarketDialog)
provide('pendingPluginsCount', pendingPluginsCount)

onMounted(() => {
  if (!isConsoleGateway) {
    const key = localStorage.getItem('plugin_key') as string
    pluginKey.value = key || ''
  }

  if (catalogConfigured) loadRemotePlugins()
})
</script>

<template>
  <ElTabs type="border-card" v-model="activeTab">
    <ElTabPane :label="t('plugin.clientPlugin')" name="client">
      <keep-alive>
        <ClientPlugin v-if="activeTab === 'client'" ref="clientPluginRef" />
      </keep-alive>
    </ElTabPane>
    <ElTabPane :label="t('plugin.serverPlugin')" name="server">
      <keep-alive>
        <ServerPlugin v-if="activeTab === 'server'" ref="serverPluginRef" />
      </keep-alive>
    </ElTabPane>
  </ElTabs>

  <PluginMarket
    :visible="marketDrawerVisible"
    :remote-plugin-list="remotePluginList"
    :market-loading="marketLoading"
    @close="marketDrawerVisible = false"
    @refresh="loadRemotePlugins"
    @install="handleInstallPlugin"
  />
</template>

<style scoped lang="less"></style>
