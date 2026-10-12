<script setup lang="tsx">
import { ContentWrap } from '@/components/ContentWrap'
import {
  ElButton,
  ElTabPane,
  ElTabs,
  ElInput,
  ElRow,
  ElCol,
  ElSwitch,
  ElText,
  ElFormItem,
  ElForm,
  ElDropdown,
  ElDropdownItem,
  ElMessageBox
} from 'element-plus'
import ProjectList from './components/ProjectList.vue'
import AddProject from './components/AddProject.vue'
import { useI18n } from '@/hooks/web/useI18n'
import { h, reactive, ref } from 'vue'
import { Dialog } from '@/components/Dialog'
import { deleteProjectApi, getProjectDataApi } from '@/api/project'
import { useIcon } from '@/hooks/web/useIcon'
const { t } = useI18n()
let allProjectData = reactive({})
let tabNames = ref<string[]>([])
let tagNum = reactive({})
const projectListLoading = ref(false)
const getProjectTag = async (pageIndex: number, pageSize: number) => {
  if (pageIndex === 0) {
    pageIndex = currentPage.value
    pageSize = currentpageSize.value
  } else {
    currentPage.value = pageIndex
    currentpageSize.value = pageSize
  }
  try {
    const res = await getProjectDataApi(search.value, pageIndex, pageSize)

    Object.assign(allProjectData, res.data.result)
    tabNames.value = Object.keys(res.data.tag)
    Object.assign(tagNum, res.data.tag)
    const index = tabNames.value.indexOf('All')
    if (index !== -1) {
      tabNames.value.splice(index, 1)
    }
  } catch (error) {
    console.error('An error occurred:', error)
  }
}
const dialogVisible = ref(false)
const addProject = async () => {
  dialogVisible.value = true
}
const closeDialog = () => {
  dialogVisible.value = false
}
const search = ref('')
const searchicon = useIcon({ icon: 'iconoir:search' })
const currentPage = ref(1)
const currentpageSize = ref(50)

const loading = ref(false)
const handleSearch = async () => {
  loading.value = true
  projectListLoading.value = true
  await getProjectTag(currentPage.value, currentpageSize.value)
  loading.value = false
  projectListLoading.value = false
}
handleSearch()
const multipleSelection = ref(false)
const deleteicon = useIcon({ icon: 'openmoji:delete' })
const delSelect = async () => {
  const deleteAsset = ref<boolean>(false)
  ElMessageBox({
    title: t('common.delete'),
    draggable: true,
    // Should pass a function if VNode contains dynamic props
    message: () =>
      h('div', { style: { display: 'flex', alignItems: 'center' } }, [
        h('p', { style: { margin: '0 10px 0 0' } }, t('task.delAsset')),
        h(ElSwitch, {
          modelValue: deleteAsset.value,
          'onUpdate:modelValue': (val: boolean) => {
            deleteAsset.value = val
          }
        })
      ])
  }).then(async () => {
    await deleteProjectApi(selectedRowIds.value, deleteAsset.value)
    getProjectTag(currentPage.value, currentpageSize.value)
  })
}
const elDropdownicon = useIcon({ icon: 'ri:arrow-drop-down-line' })
const selectedRowIds = ref([])
</script>

<template>
  <ContentWrap>
    <ElRow class="recon-project-search" :gutter="20">
      <ElCol :span="0.5">
        <ElText class="mx-1" style="position: relative; top: 8px">{{ t('form.input') }}:</ElText>
      </ElCol>
      <ElCol :span="5">
        <ElInput v-model="search" :placeholder="t('common.inputText')" style="height: 38px" />
      </ElCol>
      <ElCol :span="5" style="position: relative; left: 16px">
        <ElButton
          :loading="loading"
          type="primary"
          :icon="searchicon"
          :aria-label="t('common.search')"
          size="large"
          style="height: 100%"
          @click="handleSearch"
        />
      </ElCol>
    </ElRow>
    <ElRow class="recon-project-actions">
      <ElCol :span="2">
        <div class="mb-10px">
          <ElButton type="primary" @click="addProject">{{ t('project.addProject') }}</ElButton>
        </div>
      </ElCol>
      <ElCol :span="2">
        <ElForm>
          <ElFormItem :label="t('common.multipleSelection')">
            <ElSwitch
              v-model="multipleSelection"
              class="mb-2"
              inline-prompt
              :active-text="t('common.on')"
              :inactive-text="t('common.off')"
            />
          </ElFormItem>
        </ElForm>
      </ElCol>
      <ElCol v-if="multipleSelection" :span="1">
        <ElDropdown trigger="click">
          <ElButton plain class="custom-button align-bottom">
            {{ t('common.operation') }}
            <ElIcon class="el-icon--right"><elDropdownicon /></ElIcon>
          </ElButton>
          <template #dropdown>
            <ElDropdownMenu>
              <ElDropdownItem :icon="deleteicon" @click="delSelect">{{
                t('common.delete')
              }}</ElDropdownItem>
            </ElDropdownMenu>
          </template>
        </ElDropdown>
      </ElCol>
    </ElRow>
    <ElTabs class="demo-tabs" v-loading="projectListLoading">
      <ElTabPane :label="`${t('common.all')} (${tagNum['All'] || 0})`"
        ><ProjectList
          :tableDataList="allProjectData['All']"
          :getProjectTag="getProjectTag"
          :total="tagNum['All']"
          :multipleSelection="multipleSelection"
          v-model:selectedRows="selectedRowIds"
      /></ElTabPane>
      <ElTabPane
        v-for="tagName in tabNames"
        :label="`${tagName} (${tagNum[tagName]})`"
        :key="tagName"
        ><ProjectList
          :tableDataList="allProjectData[tagName]"
          :getProjectTag="getProjectTag"
          :total="tagNum[tagName]"
          :multipleSelection="multipleSelection"
          v-model:selectedRows="selectedRowIds"
      /></ElTabPane>
    </ElTabs>
  </ContentWrap>
  <Dialog
    v-model="dialogVisible"
    :title="t('project.addProject')"
    center
    style="border-radius: 15px; box-shadow: 5px 5px 10px rgba(0, 0, 0, 0.3)"
  >
    <AddProject
      :closeDialog="closeDialog"
      projectid=""
      :getProjectData="getProjectTag"
      :schedule="false"
    />
  </Dialog>
</template>
<style>
.demo-tabs > .el-tabs__content {
  padding: 16px;
  color: var(--recon-foreground);
  font-size: 14px;
  font-weight: 600;
}
</style>
