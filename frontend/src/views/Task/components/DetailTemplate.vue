<script setup lang="ts">
import { ref, reactive, watch, nextTick } from 'vue'
import {
  ElMessage,
  ElTooltip,
  ElTag,
  ElCard,
  ElInput,
  ElSwitch,
  ElButton,
  ElForm,
  ElFormItem,
  ElRow,
  ElCol,
  ElSelectV2,
  ElTreeV2,
  ElSelect,
  ElOption,
  ElDrawer
} from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import { Dialog } from '@/components/Dialog'
import { useI18n } from '@/hooks/web/useI18n'
import { getPluginDataByModuleApi } from '@/api/plugins'
import { getPocDataAllApi } from '@/api/poc'
import { getTemplateDetailApi, saveTemplateDetailApi } from '@/api/task'
import { pocData } from '@/api/poc/types'
import { getManagetListApi, getPortDictDataApi } from '@/api/DictionaryManagement'
import type { fileData, portDictData } from '@/api/DictionaryManagement/types'

const { t } = useI18n()

const props = defineProps<{
  closeDialog: () => void
  getList: () => void
  id: string
}>()

const modules = [
  'TargetHandler',
  'SubdomainScan',
  'SubdomainSecurity',
  'PortScanPreparation',
  'PortScan',
  'PortFingerprint',
  'AssetMapping',
  'AssetHandle',
  'URLScan',
  'WebCrawler',
  'URLSecurity',
  'DirScan',
  'VulnerabilityScan',
  'PassiveScan'
]

interface ParameterItem {
  name: string
  type: 'string' | 'bool' | 'dict' | 'port'
  defaultValue?: string
  dictCategory?: string
  dictName?: string
  portName?: string
}

const plugins = reactive<
  Record<
    string,
    {
      name: string
      hash: string
      parameter: string
      parameterList?: string
      help: string
      introduction: string
      enabled: boolean
    }[]
  >
>({})

const parameterLists = reactive<Record<string, Record<string, ParameterItem[]>>>({})

const parameters = reactive<Record<string, Record<string, string>>>({})
const selectPlugin = reactive<Record<string, string[]>>({})

const dictList = ref<fileData[]>([])
const portList = ref<portDictData[]>([])

const loadDictList = async () => {
  try {
    const res = await getManagetListApi()
    if (res.code === 200) {
      dictList.value = res.data.list
    }
  } catch (error) {
    console.error('Failed to load dictionaries:', error)
  }
}

const loadPortList = async () => {
  try {
    const res = await getPortDictDataApi('', 1, 1000)
    if (res.code === 200) {
      portList.value = res.data.list
    }
  } catch (error) {
    console.error('Failed to load ports:', error)
  }
}

const formatParameters = (params: ParameterItem[]): string => {
  if (!params || params.length === 0) {
    return ''
  }
  return params
    .filter((param) => param && param.name && param.type)
    .map((param) => {
      if (param.type === 'dict') {
        if (param.dictCategory && param.dictName) {
          return `-${param.name} {dict.${param.dictCategory}.${param.dictName}}`
        }
        return null
      }
      if (param.type === 'port') {
        if (param.portName) {
          return `-${param.name} {port.${param.portName}}`
        }
        return null
      }

      if (
        param.defaultValue !== undefined &&
        param.defaultValue !== null &&
        param.defaultValue !== ''
      ) {
        return `-${param.name} ${param.defaultValue}`
      }
      return null
    })
    .filter((item) => item !== null)
    .join(' ')
}

const parseParameters = (paramStr: string): ParameterItem[] => {
  if (!paramStr || paramStr.trim() === '') {
    return []
  }

  const parts = paramStr.match(/-(\w+)\s+(\S+)/g)
  if (!parts) {
    return []
  }
  return parts
    .map((part) => {
      const match = part.match(/-(\w+)\s+(.+)/)
      if (match) {
        const name = match[1]
        const value = match[2]

        if (value === 'true' || value === 'false') {
          return { name, type: 'bool', defaultValue: value }
        }

        if (value.startsWith('{dict.')) {
          const dictMatch = value.match(/{dict\.(.+)\.(.+)}/)
          if (dictMatch) {
            return { name, type: 'dict', dictCategory: dictMatch[1], dictName: dictMatch[2] }
          }
        }

        if (value.startsWith('{port.')) {
          const portMatch = value.match(/{port\.(.+)}/)
          if (portMatch) {
            return { name, type: 'port', portName: portMatch[1] }
          }
        }

        return { name, type: 'string', defaultValue: value }
      }
      return null
    })
    .filter((item) => item !== null) as ParameterItem[]
}

const handleParameterListChange = (module: string, hash: string) => {
  setTimeout(() => {
    const params = parameterLists[module]?.[hash] || []
    const formattedParams = formatParameters(params)
    if (!parameters[module]) {
      parameters[module] = {}
    }
    parameters[module][hash] = formattedParams
  }, 0)
}

const getFilteredDictNames = (category: string | undefined) => {
  if (!category) {
    return []
  }
  return dictList.value.filter((dict) => (dict.category || dict.name) === category)
}

const handleDictCategoryChange = (module: string, hash: string, index: number) => {
  const param = parameterLists[module][hash][index]
  param.dictName = undefined
  handleParameterListChange(module, hash)
}

const addParameter = (module: string, hash: string) => {
  if (!parameterLists[module]) {
    parameterLists[module] = {}
  }
  if (!parameterLists[module][hash]) {
    parameterLists[module][hash] = []
  }
  parameterLists[module][hash].push({
    name: '',
    type: 'string',
    defaultValue: ''
  })
  handleParameterListChange(module, hash)
}

const removeParameter = (module: string, hash: string, index: number) => {
  parameterLists[module][hash].splice(index, 1)
  handleParameterListChange(module, hash)
}

const initPlugins = async () => {
  for (const module of modules) {
    const modulePlugins = await getPluginDataByModuleApi(module)
    parameters[module] = {}
    parameterLists[module] = {}
    selectPlugin[module] = []

    plugins[module] = modulePlugins.data.list.map((plugin) => {
      let paramList: ParameterItem[] = []
      if (plugin.parameterList) {
        try {
          paramList = JSON.parse(plugin.parameterList) as ParameterItem[]
        } catch (error) {
          console.error('Failed to parse parameterList:', error)

          paramList = parseParameters(plugin.parameter || '')
        }
      } else {
        paramList = parseParameters(plugin.parameter || '')
      }

      if (!parameterLists[module]) {
        parameterLists[module] = {}
      }
      parameterLists[module][plugin.hash] = paramList

      const formattedParams = formatParameters(paramList)
      parameters[module][plugin.hash] = formattedParams

      return {
        ...plugin,
        enabled: false
      }
    })
  }
}

const vulList = ref<string[]>([])

const loadTemplate = async (id: string) => {
  const template = await getTemplateDetailApi(id)
  console.log(template)
  templateName.value = template.data.name
  vulList.value = template.data.vullist
  for (const module of modules) {
    parameters[module] = {}
    parameterLists[module] = {}

    const moduleData = template.data?.[module] || []
    const modulePlugins = await getPluginDataByModuleApi(module)

    plugins[module] = modulePlugins.data.list.map((plugin) => {
      const templateData = template.data as any
      const savedParameterList = templateData.ParameterLists?.[module]?.[plugin.hash]
      const savedParameterString = templateData.Parameters?.[module]?.[plugin.hash]
      let paramList: ParameterItem[] = []

      if (savedParameterList) {
        if (typeof savedParameterList === 'string') {
          try {
            paramList = JSON.parse(savedParameterList) as ParameterItem[]
          } catch (error) {
            console.error('Failed to parse saved parameterList:', error)

            if (savedParameterString && typeof savedParameterString === 'string') {
              paramList = parseParameters(savedParameterString)
            } else {
              paramList = plugin.parameterList
                ? (JSON.parse(plugin.parameterList) as ParameterItem[])
                : []
            }
          }
        } else if (Array.isArray(savedParameterList)) {
          paramList = savedParameterList as ParameterItem[]
        } else {
          if (savedParameterString && typeof savedParameterString === 'string') {
            paramList = parseParameters(savedParameterString)
          } else {
            paramList = plugin.parameterList
              ? (JSON.parse(plugin.parameterList) as ParameterItem[])
              : []
          }
        }
      } else if (savedParameterString && typeof savedParameterString === 'string') {
        const parsedParams = parseParameters(savedParameterString)

        if (plugin.parameterList) {
          try {
            const pluginParamList = JSON.parse(plugin.parameterList) as ParameterItem[]

            paramList = pluginParamList.map((pluginParam) => {
              const savedParam = parsedParams.find((p) => p.name === pluginParam.name)
              if (savedParam) {
                return {
                  ...pluginParam,
                  defaultValue: savedParam.defaultValue,
                  dictCategory: savedParam.dictCategory,
                  dictName: savedParam.dictName,
                  portName: savedParam.portName
                }
              }
              return pluginParam
            })
          } catch (error) {
            console.error('Failed to parse plugin parameterList:', error)

            paramList = parsedParams
          }
        } else {
          paramList = parsedParams
        }
      } else if (plugin.parameterList) {
        try {
          paramList = JSON.parse(plugin.parameterList) as ParameterItem[]
        } catch (error) {
          console.error('Failed to parse plugin parameterList:', error)
          paramList = []
        }
      } else {
        paramList = parseParameters(plugin.parameter || '')
      }

      parameterLists[module][plugin.hash] = paramList

      const formattedParams = formatParameters(paramList)
      parameters[module][plugin.hash] = formattedParams

      return {
        ...plugin,
        enabled: moduleData.includes(plugin.hash) || false
      }
    })
  }
}

watch(
  () => props.id,
  async (newId) => {
    if (newId) {
      await loadTemplate(newId)
    } else {
      await initPlugins()
    }
  },
  { immediate: true }
)
const saveLoading = ref(false)

const onSubmit = async () => {
  saveLoading.value = true
  const result: Record<string, any> = {}
  if (templateName.value == '') {
    ElMessage.error(t('console.nameRequired'))
    saveLoading.value = false
    return
  }

  for (const module of modules) {
    const enabledPlugins = plugins[module].filter((plugin) => plugin.enabled)
    result[module] = enabledPlugins.map((plugin) => plugin.hash)

    result.Parameters = result.Parameters || {}
    result.Parameters[module] = {}

    result.ParameterLists = result.ParameterLists || {}
    result.ParameterLists[module] = {}

    for (const plugin of enabledPlugins) {
      if (parameters[module]?.[plugin.hash]) {
        result.Parameters[module][plugin.hash] = parameters[module][plugin.hash]
      }

      if (parameterLists[module]?.[plugin.hash]) {
        result.ParameterLists[module][plugin.hash] = JSON.stringify(
          parameterLists[module][plugin.hash]
        )
      }
    }
  }
  result['name'] = templateName.value
  result['vullist'] = vulList.value
  try {
    const res = await saveTemplateDetailApi(result, props.id)
    console.log(result)

    if (res.code === 200) {
      ElMessage.success('success')

      props.closeDialog()
      props.getList()
    } else {
      ElMessage.error(res.message || 'error')
    }
  } catch (error) {
    ElMessage.error('error')
    console.error(error)
  } finally {
    saveLoading.value = false
  }
}
const moduleColorMap = {
  TargetHandler: '#409EFF',
  SubdomainScan: '#E6A23C',
  SubdomainSecurity: '#F56C6C',
  PortScanPreparation: '#67C23A',
  PortScan: '#00CED1',
  AssetMapping: '#8A2BE2',
  URLScan: '#C71585',
  WebCrawler: '#FF4500',
  DirScan: '#20B2AA',
  VulnerabilityScan: '#DC143C',
  AssetHandle: '#4682B4',
  PortFingerprint: '#DAA520',
  URLSecurity: '#9370DB',
  PassiveScan: '#5F9EA0'
}
const templateName = ref('')
interface TreeNode {
  value: string
  label: string
  children: TreeNode[]
}

const vulOptions = reactive<TreeNode[]>([])

const buildTree = (data: pocData[]): TreeNode[] => {
  const tree: TreeNode[] = []

  data.forEach((item) => {
    let currentLevel = tree
    item.tags.forEach((tag) => {
      const existingNode = currentLevel.find((node) => node.label === tag)
      if (!existingNode) {
        const randomString = Math.random().toString(36).substring(2, 8)
        const newNode: TreeNode = { value: randomString, label: tag, children: [] }
        currentLevel.push(newNode)
        currentLevel = newNode.children
      } else {
        currentLevel = existingNode.children
      }
    })

    currentLevel.push({ value: item.id, label: item.name, children: [] })
  })

  return tree
}

const vulSelectOptions = reactive<{ value: string; label: string }[]>([])

const getPocList = async () => {
  const res = await getPocDataAllApi()
  if (res.data.list.length > 0) {
    vulSelectOptions.push({ value: 'All Poc', label: 'All Poc' })
    res.data.list.forEach((item) => {
      vulSelectOptions.push({ value: item.id, label: item.name })
    })
  }

  vulOptions.push({ value: 'All Poc', label: 'All Poc', children: [] })
  const tree = buildTree(res.data.list)
  vulOptions.push(...tree)
}

getPocList()

loadDictList()
loadPortList()

const dialogVisible = ref(false)
const openPocList = async () => {
  dialogVisible.value = true
}

const parameterDrawerVisible = ref(false)
const currentPluginModule = ref('')
const currentPluginHash = ref('')
const currentPluginName = ref('')
const currentPluginHelp = ref('')
const drawerSize = ref('50%')

const calculateDrawerSize = () => {
  const screenWidth = window.innerWidth
  if (screenWidth < 768) {
    drawerSize.value = '90%'
  } else if (screenWidth < 1024) {
    drawerSize.value = '60%'
  } else {
    drawerSize.value = '50%'
  }
}

const openParameterDialog = (module: string, hash: string, name: string) => {
  currentPluginModule.value = module
  currentPluginHash.value = hash
  currentPluginName.value = name

  const plugin = plugins[module]?.find((p) => p.hash === hash)
  currentPluginHelp.value = plugin?.help || ''
  calculateDrawerSize()
  parameterDrawerVisible.value = true
}

const closeParameterDialog = () => {
  parameterDrawerVisible.value = false
  currentPluginModule.value = ''
  currentPluginHash.value = ''
  currentPluginName.value = ''
  currentPluginHelp.value = ''
}
watch(dialogVisible, (newVal) => {
  if (newVal) {
    nextTick(() => {
      const tree = treeRef.value
      if (tree) {
        console.log('treeRef instance ready:', tree)
        setTreeCheckedNodes(vulList.value)
      } else {
        console.log('treeRef instance unavailable')
      }
    })
  }
})
const treeRef = ref<InstanceType<typeof ElTreeV2> | null>(null)
const setTreeCheckedNodes = (checkedKeys) => {
  nextTick(() => {
    const tree = treeRef.value
    if (tree) {
      tree.setCheckedKeys(checkedKeys)
    } else {
      console.log('treeRef instance unavailable')
    }
  })
}
const propss = {
  value: 'value',
  label: 'label',
  children: 'children'
}

const handleCheckChange = (data, checked) => {
  const nodeValue = data.value

  const isLeafNode = !data.children || data.children.length === 0

  if (isLeafNode) {
    if (checked && !vulList.value.includes(nodeValue)) {
      vulList.value.push(nodeValue)
    } else if (!checked) {
      const index = vulList.value.indexOf(nodeValue)
      if (index > -1) {
        vulList.value.splice(index, 1)
      }
    }
  } else {
    const addLeafNodes = (node) => {
      if (node.children && node.children.length > 0) {
        node.children.forEach((child) => {
          const isChildLeafNode = !child.children || child.children.length === 0
          if (isChildLeafNode) {
            if (checked && !vulList.value.includes(child.value)) {
              vulList.value.push(child.value)
            } else if (!checked) {
              const index = vulList.value.indexOf(child.value)
              if (index > -1) {
                vulList.value.splice(index, 1)
              }
            }
          } else {
            addLeafNodes(child)
          }
        })
      }
    }

    addLeafNodes(data)
  }

  console.log('Selected leaf node values:', vulList.value)
}
</script>

<template>
  <ElForm @submit.prevent="onSubmit" label-width="auto">
    <ElFormItem :label="t('task.templateName')">
      <ElInput v-model="templateName" />
    </ElFormItem>
    <div v-for="module in modules" :key="module">
      <ElCard class="module-card" :body-style="{ padding: '20px' }" shadow="always">
        <div
          style="
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
            border-bottom: 1px solid #f0f2f5;
            padding-bottom: 15px;
          "
        >
          <div style="display: flex; align-items: center; gap: 10px">
            <div
              :style="{
                width: '4px',
                height: '16px',
                backgroundColor: moduleColorMap[module],
                borderRadius: '2px'
              }"
            ></div>
            <span style="font-weight: 600; font-size: 16px; color: #303133">
              {{ t(`scanTemplate.${module}`) }}
            </span>
          </div>
        </div>

        <div class="plugins-container">
          <ElCard
            v-for="plugin in plugins[module]"
            :key="plugin.hash"
            :class="['plugin-card', { 'plugin-card-enabled': plugin.enabled }]"
            :style="{ '--card-accent-color': moduleColorMap[module] }"
            :body-style="{ padding: '0', height: '100%', display: 'flex', flexDirection: 'column' }"
            shadow="hover"
          >
            <div class="plugin-card-header">
              <div class="plugin-title">
                <div class="plugin-name-wrapper">
                  <span class="plugin-name" :title="plugin.name">{{ plugin.name }}</span>
                </div>
                <ElTooltip
                  placement="top"
                  effect="dark"
                  :content="plugin.enabled ? t('common.enabled') : t('common.disabled')"
                >
                  <ElSwitch
                    v-model="plugin.enabled"
                    class="plugin-switch"
                    style="--el-switch-on-color: var(--card-accent-color)"
                  />
                </ElTooltip>
              </div>
              <div class="plugin-desc" v-if="plugin.introduction">
                <ElTooltip placement="top" effect="light" :content="plugin.introduction">
                  <span class="text-truncate">{{ plugin.introduction }}</span>
                </ElTooltip>
              </div>
            </div>
            <div class="plugin-card-body">
              <ElFormItem
                :label="t('task.vulList')"
                prop="type"
                v-if="plugin.enabled && plugin.hash === 'ed93b8af6b72fe54a60efdb932cf6fbc'"
              >
                <ElSelectV2
                  v-model="vulList"
                  filterable
                  :options="vulSelectOptions"
                  placeholder="Please select vul"
                  style="width: 80%; margin-right: 10px"
                  multiple
                  collapse-tags
                  collapse-tags-tooltip
                  tag-type="info"
                  :max-collapse-tags="3"
                />
                <ElButton type="primary" @click="openPocList" :loading="saveLoading">
                  {{ t('common.selectCategory') }}
                </ElButton>
              </ElFormItem>

              <ElFormItem v-if="plugin.enabled" :label="t('plugin.parameter')">
                <div style="display: flex; gap: 8px; align-items: flex-start">
                  <ElInput
                    :model-value="parameters[module]?.[plugin.hash] || ''"
                    readonly
                    style="background-color: #f5f5f5; flex: 1"
                    type="textarea"
                    :rows="2"
                  />
                  <ElButton
                    type="primary"
                    size="small"
                    @click="openParameterDialog(module, plugin.hash, plugin.name)"
                  >
                    {{ t('plugin.parameterConfig') }}
                  </ElButton>
                </div>
              </ElFormItem>
            </div>
          </ElCard>
        </div>
      </ElCard>
    </div>
    <ElRow>
      <ElCol :span="12" style="text-align: right">
        <ElButton type="primary" @click="onSubmit" :loading="saveLoading">
          {{ t('task.save') }}
        </ElButton>
      </ElCol>
    </ElRow>
  </ElForm>
  <Dialog
    v-model="dialogVisible"
    title="POC"
    center
    fullscreen
    style="border-radius: 15px; box-shadow: 5px 5px 10px rgba(0, 0, 0, 0.3)"
  >
    <ElTreeV2
      ref="treeRef"
      style="max-width: 100%"
      :data="vulOptions"
      :props="propss"
      show-checkbox
      :height="600"
      @check-change="handleCheckChange"
    />
  </Dialog>

  <ElDrawer
    v-model="parameterDrawerVisible"
    :title="
      currentPluginName
        ? `${currentPluginName} - ${t('plugin.parameterConfig')}`
        : t('plugin.parameterConfig')
    "
    :size="drawerSize"
    direction="rtl"
    :close-on-click-modal="false"
  >
    <ElForm label-width="100px">
      <ElRow v-if="currentPluginHelp" style="margin-bottom: 20px">
        <ElCol :span="24">
          <ElCard shadow="never" style="background-color: #f0f9ff; border: 1px solid #b3d8ff">
            <div style="display: flex; align-items: flex-start; gap: 8px">
              <span style="font-weight: 600; color: #409eff; flex-shrink: 0">
                {{ t('plugin.help') }}:
              </span>
              <span style="color: #606266; line-height: 1.6">{{ currentPluginHelp }}</span>
            </div>
          </ElCard>
        </ElCol>
      </ElRow>
      <ElRow :gutter="20" v-if="parameterLists[currentPluginModule]?.[currentPluginHash]?.length">
        <template
          v-for="(param, index) in parameterLists[currentPluginModule]?.[currentPluginHash] || []"
          :key="index"
        >
          <ElCol :span="24" style="margin-bottom: 16px">
            <div style="padding: 12px; border: 1px solid #dcdfe6; border-radius: 4px">
              <ElRow :gutter="10">
                <ElCol :span="24">
                  <ElFormItem
                    :prop="`parameterList.${currentPluginModule}.${currentPluginHash}.${index}.name`"
                    style="margin-bottom: 8px"
                    :label="t('plugin.parameterName')"
                  >
                    <div style="display: flex; align-items: center; gap: 8px">
                      <ElInput
                        v-model="param.name"
                        :placeholder="t('plugin.parameterName')"
                        :readonly="param.type !== 'string'"
                        :disabled="param.type !== 'string'"
                        style="flex: 1"
                        @input="handleParameterListChange(currentPluginModule, currentPluginHash)"
                      />
                      <ElButton
                        :icon="Delete"
                        type="danger"
                        circle
                        size="small"
                        @click="removeParameter(currentPluginModule, currentPluginHash, index)"
                      />
                    </div>
                  </ElFormItem>
                </ElCol>

                <template v-if="param.type === 'string' || param.type === 'bool'">
                  <ElCol :span="24">
                    <ElFormItem
                      :prop="`parameterList.${currentPluginModule}.${currentPluginHash}.${index}.defaultValue`"
                      style="margin-bottom: 0"
                      :label="t('plugin.defaultValue')"
                    >
                      <ElSelect
                        v-if="param.type === 'bool'"
                        v-model="param.defaultValue"
                        :placeholder="t('plugin.defaultValue')"
                        @change="handleParameterListChange(currentPluginModule, currentPluginHash)"
                      >
                        <ElOption label="true" value="true" />
                        <ElOption label="false" value="false" />
                      </ElSelect>
                      <ElInput
                        v-else
                        v-model="param.defaultValue"
                        :placeholder="t('plugin.defaultValue')"
                        @input="handleParameterListChange(currentPluginModule, currentPluginHash)"
                      />
                    </ElFormItem>
                  </ElCol>
                </template>

                <template v-if="param.type === 'dict'">
                  <ElCol :span="24" style="margin-bottom: 8px">
                    <ElFormItem
                      :prop="`parameterList.${currentPluginModule}.${currentPluginHash}.${index}.dictCategory`"
                      style="margin-bottom: 0"
                      :label="t('plugin.category')"
                    >
                      <ElSelect
                        v-model="param.dictCategory"
                        :placeholder="t('plugin.category')"
                        filterable
                        @change="
                          handleDictCategoryChange(currentPluginModule, currentPluginHash, index)
                        "
                      >
                        <ElOption
                          v-for="dict in dictList"
                          :key="dict.id"
                          :label="dict.category || dict.name"
                          :value="dict.category || dict.name"
                        />
                      </ElSelect>
                    </ElFormItem>
                  </ElCol>
                  <ElCol :span="24">
                    <ElFormItem
                      :prop="`parameterList.${currentPluginModule}.${currentPluginHash}.${index}.dictName`"
                      style="margin-bottom: 0"
                      :label="t('plugin.dictName')"
                    >
                      <ElSelect
                        v-model="param.dictName"
                        :placeholder="t('plugin.dictName')"
                        filterable
                        :disabled="!param.dictCategory"
                        @change="handleParameterListChange(currentPluginModule, currentPluginHash)"
                      >
                        <ElOption
                          v-for="dict in getFilteredDictNames(param.dictCategory)"
                          :key="dict.id"
                          :label="dict.name"
                          :value="dict.name"
                        />
                      </ElSelect>
                    </ElFormItem>
                  </ElCol>
                </template>

                <template v-if="param.type === 'port'">
                  <ElCol :span="24">
                    <ElFormItem
                      :prop="`parameterList.${currentPluginModule}.${currentPluginHash}.${index}.portName`"
                      style="margin-bottom: 0"
                      :label="t('plugin.portName')"
                    >
                      <ElSelect
                        v-model="param.portName"
                        :placeholder="t('plugin.portName')"
                        filterable
                        @change="handleParameterListChange(currentPluginModule, currentPluginHash)"
                      >
                        <ElOption
                          v-for="port in portList"
                          :key="port.id"
                          :label="port.name"
                          :value="port.name"
                        />
                      </ElSelect>
                    </ElFormItem>
                  </ElCol>
                </template>
              </ElRow>
            </div>
          </ElCol>
        </template>
      </ElRow>
      <div v-else style="text-align: center; padding: 40px; color: #909399">
        {{ t('plugin.noParameters') }}
      </div>

      <ElRow style="margin-top: 20px">
        <ElCol :span="24">
          <ElButton
            :icon="Plus"
            style="width: 100%"
            @click="addParameter(currentPluginModule, currentPluginHash)"
          >
            {{ t('plugin.addParameter') }}
          </ElButton>
        </ElCol>
      </ElRow>

      <ElFormItem :label="t('plugin.parameter')" style="margin-top: 20px">
        <ElInput
          :model-value="parameters[currentPluginModule]?.[currentPluginHash] || ''"
          readonly
          style="background-color: #f5f5f5"
          type="textarea"
          :rows="3"
        />
        <div style="font-size: 12px; color: #909399; margin-top: 4px">
          {{ t('plugin.parameterTip') }}
        </div>
      </ElFormItem>
    </ElForm>

    <template #footer>
      <div style="text-align: right">
        <ElButton @click="closeParameterDialog">{{ t('common.cancel') }}</ElButton>
        <ElButton type="primary" @click="closeParameterDialog" style="margin-left: 10px">
          {{ t('common.confirmed') }}
        </ElButton>
      </div>
    </template>
  </ElDrawer>
</template>

<style scoped>
.ElFormItem {
  margin-bottom: 20px;
}

.module-card {
  margin-bottom: 24px;
  border-radius: 4px;
  border: 1px solid #ebeef5;
  background: #ffffff;
  box-shadow: none;
}

.module-card:hover {
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
}

.plugins-container {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.plugin-card {
  border-radius: 4px;
  border: 1px solid #e4e7ed;
  background-color: #ffffff;
  transition: all 0.2s ease-in-out;
  position: relative;
  border-left: 4px solid var(--card-accent-color, #409eff);
}

.plugin-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  border-color: #c0c4cc;
}

.plugin-card-enabled {
  background: #fcfcfc;
}

.plugin-card-header {
  padding: 12px 16px;
  border-bottom: 1px solid #f2f6fc;
  background: #fff;
  border-radius: 4px 4px 0 0;
}

.plugin-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.plugin-name-wrapper {
  flex: 1;
  overflow: hidden;
  margin-right: 12px;
}

.plugin-name {
  font-weight: 600;
  font-size: 14px;
  color: #303133;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  display: block;
}

.plugin-desc {
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
}

.text-truncate {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
}

.plugin-card-body {
  padding: 16px;
  background-color: #ffffff;
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  border-radius: 0 0 4px 4px;
}

.plugin-switch {
  --el-switch-on-color: var(--card-accent-color);
}
/* Customizing the parameter input area */
.plugin-card-body :deep(.el-textarea__inner) {
  background-color: #f8fafc !important;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-family: 'Menlo', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  color: #475569;
  box-shadow: none;
}

.plugin-card-body :deep(.el-textarea__inner:hover) {
  border-color: #cbd5e1;
}

.plugin-card-body :deep(.el-button--small) {
  padding: 8px 12px;
  font-weight: 500;
}
</style>
