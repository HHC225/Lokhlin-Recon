<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElButton, ElInput, ElPagination, ElAlert, ElTabs, ElTabPane } from 'element-plus'
import { ContentWrap } from '@/components/ContentWrap'
import { Dialog } from '@/components/Dialog'
import { useI18n } from '@/hooks/web/useI18n'
import { resourcePath } from '@/integration/console'
import {
  getMapHostsApi,
  getMapAssetsApi,
  getMapSensitiveApi,
  getMapVulnerabilityApi,
  type MapAsset
} from '@/api/asset/map'

const { t } = useI18n()
const route = useRoute()
const rootDomain = ref(
  String(route.query.rootDomain || new URLSearchParams(location.search).get('rootDomain') || '')
)
const domain = ref('')
const hosts = ref<string[]>([])
const assets = ref<Record<string, { list: MapAsset[]; total: number }>>({})
const hostPage = ref(1)
const hostTotal = ref(0)
const assetPage = ref(1)
const selectedHost = ref('')
const loading = ref(false)
const error = ref('')
let generation = 0

const selectedAssets = computed(() => assets.value[selectedHost.value] || { list: [], total: 0 })
const errorText = (value: unknown) =>
  value instanceof Error ? value.message : t('console.loadFailed')

const loadHosts = async (page = 1) => {
  if (!rootDomain.value.trim()) {
    error.value = t('assetMap.domainRequired')
    return
  }
  const current = ++generation
  loading.value = true
  error.value = ''
  try {
    const response = await getMapHostsApi(rootDomain.value.trim(), page)
    if (response.code !== 200 || !Array.isArray(response.data?.list))
      throw new Error(t('console.loadFailed'))
    const pairs = await Promise.all(
      response.data.list.map(async (host) => {
        const rows = await getMapAssetsApi(host)
        if (rows.code !== 200 || !Array.isArray(rows.data?.list))
          throw new Error(t('console.loadFailed'))
        return [host, rows.data] as const
      })
    )
    if (current !== generation) return
    domain.value = rootDomain.value.trim()
    hosts.value = response.data.list
    hostTotal.value = response.data.total
    hostPage.value = page
    assets.value = Object.fromEntries(pairs)
    selectedHost.value = hosts.value[0] || ''
    assetPage.value = 1
  } catch (value) {
    if (current === generation) error.value = errorText(value)
  } finally {
    if (current === generation) loading.value = false
  }
}

const selectHost = (host: string) => {
  selectedHost.value = host
  assetPage.value = 1
  void loadAssetPage(1)
}
const loadAssetPage = async (page: number) => {
  loading.value = true
  error.value = ''
  try {
    const response = await getMapAssetsApi(selectedHost.value, page)
    if (response.code !== 200) throw new Error(t('console.loadFailed'))
    assets.value[selectedHost.value] = response.data
    assetPage.value = page
  } catch (value) {
    error.value = errorText(value)
  } finally {
    loading.value = false
  }
}

const detail = ref<MapAsset | null>(null)
const detailVisible = ref(false)
const sensitive = ref<unknown[]>([])
const vulnerabilities = ref<unknown[]>([])
const vulnerabilityPage = ref(1)
const vulnerabilityTotal = ref(0)
const detailLoading = ref(false)
const detailError = ref('')
const safeURL = computed(() =>
  /^https?:\/\//i.test(detail.value?.url || '') ? detail.value!.url : undefined
)
const loadVulnerabilities = async (page: number) => {
  if (!detail.value?.url) return
  const response = await getMapVulnerabilityApi(detail.value.url, page)
  if (response.code !== 200) throw new Error(t('console.loadFailed'))
  vulnerabilities.value = response.data.list || []
  vulnerabilityTotal.value = response.data.total || 0
  vulnerabilityPage.value = page
}
const changeVulnerabilityPage = async (page: number) => {
  try {
    await loadVulnerabilities(page)
  } catch (value) {
    detailError.value = errorText(value)
  }
}
const openAsset = async (asset: MapAsset) => {
  detail.value = asset
  detailVisible.value = true
  sensitive.value = []
  vulnerabilities.value = []
  vulnerabilityTotal.value = 0
  vulnerabilityPage.value = 1
  detailError.value = ''
  if (!asset.url) return
  detailLoading.value = true
  try {
    const [response] = await Promise.all([getMapSensitiveApi(asset.url), loadVulnerabilities(1)])
    if (response.code !== 200) throw new Error(t('console.loadFailed'))
    sensitive.value = response.data.list || []
  } catch (value) {
    detailError.value = errorText(value)
  } finally {
    detailLoading.value = false
  }
}

onMounted(() => {
  if (rootDomain.value) void loadHosts()
})
</script>

<template>
  <ContentWrap :title="t('assetMap.title')" :message="t('assetMap.description')">
    <form class="map-search" @submit.prevent="loadHosts()">
      <label for="map-domain">{{ t('assetMap.rootDomain') }}</label>
      <ElInput id="map-domain" v-model="rootDomain" placeholder="example.com" :disabled="loading" />
      <ElButton native-type="submit" type="primary" :loading="loading">{{
        t('assetMap.load')
      }}</ElButton>
    </form>
    <ElAlert v-if="error" :title="error" type="error" :closable="false" show-icon />
    <p v-else-if="!domain" class="map-notice" role="status">{{ t('assetMap.enterDomain') }}</p>
    <p v-else-if="!hosts.length" class="map-notice" role="status">{{ t('assetMap.noHosts') }}</p>
    <template v-if="hosts.length">
      <div class="topology-scroll" :aria-label="t('assetMap.title')" tabindex="0">
        <div class="topology">
          <div class="map-domain-node"
            ><strong>{{ domain }}</strong
            ><span>{{ hostTotal }} {{ t('assetMap.hosts') }}</span></div
          >
          <div class="map-hosts">
            <button
              v-for="host in hosts"
              :key="host"
              class="map-node"
              :disabled="loading || detailLoading"
              :class="{ selected: host === selectedHost }"
              :aria-pressed="host === selectedHost"
              @click="selectHost(host)"
            >
              <strong>{{ host }}</strong
              ><span>{{ assets[host]?.total || 0 }} {{ t('assetMap.ports') }}</span>
            </button>
          </div>
          <div class="map-ports">
            <p>{{ selectedHost }}</p>
            <button
              v-for="(asset, index) in selectedAssets.list"
              :key="`${asset.ip}:${asset.port}:${index}`"
              class="map-node"
              :disabled="loading || detailLoading"
              @click="openAsset(asset)"
            >
              <strong>{{ asset.ip || selectedHost }}:{{ asset.port }}</strong>
              <span
                >{{ asset.service || '-'
                }}<template v-if="asset.title"> · {{ asset.title }}</template></span
              >
            </button>
            <p v-if="!selectedAssets.list.length" class="map-notice">{{ t('assetMap.noPorts') }}</p>
          </div>
        </div>
      </div>
      <div class="map-pagination">
        <div
          ><span>{{ t('assetMap.hosts') }}</span
          ><ElPagination
            :disabled="loading"
            :current-page="hostPage"
            :page-size="10"
            :total="hostTotal"
            layout="prev, pager, next"
            @current-change="loadHosts"
        /></div>
        <div
          ><span>{{ t('assetMap.ports') }}</span
          ><ElPagination
            :disabled="loading"
            :current-page="assetPage"
            :page-size="10"
            :total="selectedAssets.total"
            layout="prev, pager, next"
            @current-change="loadAssetPage"
        /></div>
      </div>
    </template>
  </ContentWrap>
  <Dialog v-model="detailVisible" :title="t('asset.assetDetail')" width="900px">
    <template v-if="detail">
      <a v-if="safeURL" :href="safeURL" target="_blank" rel="noopener noreferrer">{{
        detail.url
      }}</a>
      <ElAlert v-if="detailError" :title="detailError" type="error" :closable="false" />
      <ElTabs v-loading="detailLoading">
        <ElTabPane :label="t('console.rawData')">
          <pre class="map-json">{{ JSON.stringify(detail, null, 2) }}</pre>
        </ElTabPane>
        <ElTabPane :label="t('sensitiveInformation.sensitiveInformationName')">
          <pre class="map-json">{{
            sensitive.length ? JSON.stringify(sensitive, null, 2) : t('common.noData')
          }}</pre>
        </ElTabPane>
        <ElTabPane :label="t('vulnerability.vulnerabilityName')">
          <pre class="map-json">{{
            vulnerabilities.length ? JSON.stringify(vulnerabilities, null, 2) : t('common.noData')
          }}</pre>
          <ElPagination
            :current-page="vulnerabilityPage"
            :page-size="20"
            :total="vulnerabilityTotal"
            layout="prev, pager, next"
            @current-change="changeVulnerabilityPage"
          />
        </ElTabPane>
        <ElTabPane v-if="detail.ResponseBodyHash" :label="t('asset.screenshot')"
          ><img
            class="map-screenshot"
            :src="resourcePath(`/images/screenshots/${detail.ResponseBodyHash}.png`)"
            :alt="t('asset.screenshot')"
        /></ElTabPane>
      </ElTabs>
    </template>
  </Dialog>
</template>

<style scoped>
.map-search {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
}
.map-search .el-input {
  flex: 1 1 180px;
  max-width: 360px;
}
.map-notice {
  color: var(--recon-muted-foreground);
  padding: 24px 0;
  font-size: 14px;
}
.topology-scroll {
  overflow: auto;
  padding: 24px 0;
}
.topology {
  display: grid;
  grid-template-columns: 180px minmax(220px, 1fr) minmax(240px, 1fr);
  gap: 40px;
  min-width: 720px;
}
.map-domain-node {
  align-self: start;
  margin-top: 32px;
  padding: 16px;
  background: var(--recon-muted);
  border: 1px solid var(--recon-border);
  border-radius: 4px;
}
.map-domain-node strong,
.map-node strong {
  font-weight: 500;
  overflow-wrap: anywhere;
}
.map-domain-node span,
.map-node span {
  display: block;
  font-size: 12px;
  color: var(--recon-muted-foreground);
  margin-top: 6px;
  overflow-wrap: anywhere;
}
.map-hosts,
.map-ports {
  position: relative;
  padding-left: 20px;
  border-left: 1px solid var(--recon-border);
}
.map-node {
  position: relative;
  display: block;
  width: 100%;
  padding: 12px;
  margin-bottom: 12px;
  text-align: left;
  background: var(--recon-card);
  color: var(--recon-foreground);
  border: 1px solid var(--recon-border);
  border-radius: 4px;
  cursor: pointer;
  font: inherit;
}
.map-node::before {
  content: '';
  position: absolute;
  width: 20px;
  height: 1px;
  left: -21px;
  top: 24px;
  background: var(--recon-border);
}
.map-node:hover,
.map-node.selected {
  background: var(--recon-muted);
  border-color: var(--recon-ring);
}
.map-ports > p {
  font-size: 12px;
  color: var(--recon-muted-foreground);
  margin-top: 0;
  overflow-wrap: anywhere;
}
.map-pagination {
  display: flex;
  flex-wrap: wrap;
  gap: 24px;
  justify-content: space-between;
  border-top: 1px solid var(--recon-border);
  padding-top: 16px;
}
.map-pagination > div {
  display: flex;
  align-items: center;
  gap: 8px;
}
.map-pagination span {
  font-size: 12px;
  color: var(--recon-muted-foreground);
}
.map-json {
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  font-size: 13px;
  padding: 16px;
  background: var(--recon-muted);
  border-radius: 4px;
  max-height: 480px;
  overflow: auto;
}
.map-screenshot {
  max-width: 100%;
  height: auto;
}
</style>
