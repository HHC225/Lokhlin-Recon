import request from '@/axios'

export const getDictApi = () => {
  return request.get({ url: '/mock/dict/list' })
}

export const getDictOneApi = async () => {
  return request.get({ url: '/mock/dict/one' })
}
