import { AxiosResponse, AxiosRequestHeaders, InternalAxiosRequestConfig } from './types'
import { ElMessage } from 'element-plus'
import qs from 'qs'
import { SUCCESS_CODE } from '@/constants'
import { useUserStoreWithOut } from '@/store/modules/user'
import { isConsoleGateway, expireConsoleSession } from '@/integration/console'

const defaultRequestInterceptors = (config: InternalAxiosRequestConfig) => {
  if (
    config.method === 'post' &&
    (config.headers as AxiosRequestHeaders)['Content-Type'] === 'application/x-www-form-urlencoded'
  ) {
    config.data = qs.stringify(config.data)
  }
  if (config.method === 'get' && config.params) {
    let url = config.url as string
    url += '?'
    const keys = Object.keys(config.params)
    for (const key of keys) {
      if (config.params[key] !== void 0 && config.params[key] !== null) {
        url += `${key}=${encodeURIComponent(config.params[key])}&`
      }
    }
    url = url.substring(0, url.length - 1)
    config.params = {}
    config.url = url
  }
  return config
}

const defaultResponseInterceptors = (response: AxiosResponse) => {
  if (response?.headers['content-type'] == 'application/octet-stream') {
    return response
  }
  if (response?.config?.responseType === 'blob') {
    return response
  } else if (response.data.code === SUCCESS_CODE) {
    if (response?.data?.message) {
      ElMessage.success(response?.data?.message)
    }
    return response.data
  } else {
    if (isConsoleGateway && [401, 403].includes(response?.data?.code)) {
      expireConsoleSession(response.data.code)
      return Promise.reject(new Error('Console session is unavailable'))
    }
    ElMessage.error(response?.data?.message)
    if (response?.data?.code === 401) {
      const userStore = useUserStoreWithOut()
      userStore.logout()
    }
    return response.data
  }
}

export { defaultResponseInterceptors, defaultRequestInterceptors }
