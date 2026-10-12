import request from '@/axios'

export interface MapAsset {
  ip?: string
  port?: number | string
  service?: string
  url?: string
  title?: string
  status?: number
  ResponseBodyHash?: string
  [key: string]: unknown
}

export const getMapHostsApi = (rootDomain: string, pageIndex: number, pageSize = 10) =>
  request.post<{ list: string[]; total: number }>({
    url: '/api/assets/asset/hosts',
    data: { rootDomain, pageIndex, pageSize }
  })

export const getMapAssetsApi = (host: string, pageIndex = 1, pageSize = 10) =>
  request.post<{ list: MapAsset[]; total: number }>({
    url: '/api/assets/asset/by-host',
    data: { host, pageIndex, pageSize }
  })

export const getMapSensitiveApi = (url: string) =>
  request.post<{ list: unknown[] }>({
    url: '/api/assets/sensitive/names',
    data: { search: `url =${JSON.stringify(url)}`, filter: { project: [] } }
  })

export const getMapVulnerabilityApi = (url: string, pageIndex = 1) =>
  request.post<{ list: unknown[]; total: number }>({
    url: '/api/assets/vulnerability',
    data: {
      search: `url =${JSON.stringify(url)}`,
      pageIndex,
      pageSize: 20,
      filter: { project: [] }
    }
  })
