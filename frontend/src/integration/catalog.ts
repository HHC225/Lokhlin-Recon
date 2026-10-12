import { useI18n } from '@/hooks/web/useI18n'

export const catalogApiBase = (import.meta.env.VITE_CATALOG_API_URL || '').replace(/\/$/, '')
export const catalogWebBase = (import.meta.env.VITE_CATALOG_WEB_URL || '').replace(/\/$/, '')
export const catalogConfigured = Boolean(catalogApiBase)

export const catalogApiPath = (path: string) => {
  if (!catalogApiBase) throw new Error(useI18n().t('console.catalogUnavailable'))
  return `${catalogApiBase}${path}`
}

export const catalogWebPath = (path: string) =>
  catalogWebBase ? `${catalogWebBase}${path}` : undefined
