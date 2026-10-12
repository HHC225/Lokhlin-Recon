import { reactive } from 'vue'

export const isConsoleGateway = /^\/recon(?:\/|$)/.test(window.location.pathname)
export const isEmbedded =
  isConsoleGateway && new URLSearchParams(window.location.search).get('embedded') === '1'
export const displayName = import.meta.env.VITE_APP_TITLE || 'Recon'
export const consoleTheme = reactive({ dark: false })

interface ConsoleUser {
  id: string
  email: string
  role: 'admin' | 'member'
}

export const consoleSession = reactive({
  status: 'loading' as 'loading' | 'ready' | 'expired' | 'denied' | 'error',
  user: null as ConsoleUser | null
})

const notifyHost = (type: 'recon:ready' | 'recon:session-expired') => {
  if (isEmbedded && window.parent !== window) {
    window.parent.postMessage({ type }, window.location.origin)
  }
}

export const expireConsoleSession = (status = 401) => {
  consoleSession.user = null
  consoleSession.status = status === 403 ? 'denied' : 'expired'
  if (status === 401) notifyHost('recon:session-expired')
}

export const handleConsoleUploadError = (error: unknown) => {
  const status =
    typeof error === 'object' && error !== null && 'status' in error ? Number(error.status) : 0
  if (isConsoleGateway && [401, 403].includes(status)) expireConsoleSession(status)
}

let pending: Promise<ConsoleUser | null> | null = null

export const bootstrapConsoleSession = (force = false): Promise<ConsoleUser | null> => {
  if (pending) return pending
  if (!force && consoleSession.status === 'ready') return Promise.resolve(consoleSession.user)
  consoleSession.status = 'loading'
  const controller = new AbortController()
  const timer = window.setTimeout(() => controller.abort(), 15000)
  pending = (async () => {
    try {
      const response = await fetch('/api/auth/me', {
        credentials: 'same-origin',
        cache: 'no-store',
        signal: controller.signal
      })
      if (response.status === 401 || response.status === 403) {
        expireConsoleSession(response.status)
        return null
      }
      if (!response.ok) throw new Error('Console session request failed')
      const { user } = await response.json()
      if (!user || typeof user.id !== 'string' || !user.id || typeof user.email !== 'string') {
        throw new Error('Invalid console session')
      }
      if (user.role !== 'admin') {
        expireConsoleSession(403)
        return null
      }
      consoleSession.user = { id: user.id, email: user.email, role: user.role }
      consoleSession.status = 'ready'
      notifyHost('recon:ready')
      return consoleSession.user
    } catch {
      consoleSession.user = null
      consoleSession.status = 'error'
      return null
    } finally {
      window.clearTimeout(timer)
      pending = null
    }
  })()
  return pending
}

export const csrfHeaders = (method = 'GET'): Record<string, string> => {
  if (!isConsoleGateway || ['GET', 'HEAD', 'OPTIONS'].includes(method.toUpperCase())) return {}
  for (const name of ['__Host-lokhlin_csrf', 'lokhlin_csrf']) {
    const entry = document.cookie
      .split(';')
      .map((value) => value.trim())
      .find((value) => value.startsWith(`${name}=`))
    if (entry) {
      try {
        return { 'X-CSRF-Token': decodeURIComponent(entry.slice(name.length + 1)) }
      } catch {
        return {}
      }
    }
  }
  return {}
}

export const resourcePath = (path: string) =>
  `${isConsoleGateway ? '/recon' : import.meta.env.BASE_URL.replace(/\/$/, '')}${path}`

const themeTokens = {
  background: 'background',
  foreground: 'foreground',
  card: 'card',
  primary: 'primary',
  primaryForeground: 'primary-foreground',
  muted: 'muted',
  mutedForeground: 'muted-foreground',
  border: 'border',
  input: 'input',
  ring: 'ring',
  radius: 'radius',
  fontFamily: 'font-family'
} as const

export const applyConsoleTheme = (
  theme: 'light' | 'dark',
  tokens: Record<string, unknown> = {}
) => {
  const root = document.documentElement
  consoleTheme.dark = theme === 'dark'
  root.classList.toggle('dark', theme === 'dark')
  root.classList.toggle('light', theme !== 'dark')
  root.style.colorScheme = theme
  for (const [key, variable] of Object.entries(themeTokens)) {
    const value = tokens[key]
    if (typeof value === 'string' && value.length < 512 && !/[;{}]/.test(value)) {
      root.style.setProperty(`--recon-${variable}`, value)
    }
  }
}

const setupEmbeddedLayout = () => {
  if (!isEmbedded || window.parent === window || typeof ResizeObserver === 'undefined') return
  let previousHeight = 0
  let previousOverlay = false
  let scheduled = false
  let active = true
  const observed = new Set<Element>()
  const report = () => {
    const content = document.querySelector('.recon-embedded-view') || document.querySelector('.recon-session') || document.querySelector('#app')
    if (!content || !active) return
    const box = content.getBoundingClientRect()
    const style = getComputedStyle(content)
    const height = Math.ceil(box.height + (parseFloat(style.marginTop) || 0) + (parseFloat(style.marginBottom) || 0))
    const overlay = Array.from(document.querySelectorAll('.el-overlay, .el-image-viewer__wrapper')).some((element) => {
      const style = getComputedStyle(element)
      return style.display !== 'none' && style.visibility !== 'hidden' && element.getBoundingClientRect().height > 0
    })
    if (height > 0 && (height !== previousHeight || overlay !== previousOverlay)) {
      previousHeight = height
      previousOverlay = overlay
      window.parent.postMessage({ type: 'recon:layout', height, overlay }, window.location.origin)
    }
  }
  const schedule = () => {
    if (scheduled || !active) return
    scheduled = true
    window.requestAnimationFrame(() => { scheduled = false; report() })
  }
  const resize = new ResizeObserver(schedule)
  const observe = () => {
    for (const selector of ['.recon-embedded-view', '.recon-session', '#app']) {
      const element = document.querySelector(selector)
      if (element && !observed.has(element)) { resize.observe(element); observed.add(element) }
    }
    schedule()
  }
  const mutations = new MutationObserver(observe)
  mutations.observe(document.documentElement, { childList: true, subtree: true, attributes: true, characterData: true })
  window.addEventListener('resize', schedule)
  window.addEventListener('pagehide', (event) => {
    if (event.persisted) return
    active = false
    resize.disconnect()
    mutations.disconnect()
    observed.clear()
  })
  observe()
}

export const setupConsoleIntegration = () => {
  if (!isConsoleGateway) return
  document.documentElement.classList.add('recon-console')
  if (isEmbedded) document.documentElement.classList.add('recon-embedded')
  const theme = new URLSearchParams(window.location.search).get('theme')
  applyConsoleTheme(theme === 'dark' ? 'dark' : 'light')
  window.addEventListener('message', (event) => {
    if (event.origin !== window.location.origin || event.source !== window.parent) return
    if (event.data?.type === 'recon:theme' && ['light', 'dark'].includes(event.data.theme)) {
      applyConsoleTheme(event.data.theme, event.data.tokens || {})
    }
    if (event.data?.type === 'recon:viewport') {
      const { top, height } = event.data
      if (typeof top === 'number' && typeof height === 'number' && Number.isFinite(top) && Number.isFinite(height) && top >= 0 && top <= 100000 && height >= 0 && height <= 20000) {
        document.documentElement.style.setProperty('--recon-viewport-top', `${top}px`)
        document.documentElement.style.setProperty('--recon-viewport-height', `${height}px`)
      }
    }
  })
  setupEmbeddedLayout()
  if (typeof BroadcastChannel !== 'undefined') {
    const channel = new BroadcastChannel('lokhlin-auth')
    channel.onmessage = () => void bootstrapConsoleSession(true)
  }
  window.addEventListener('pageshow', (event) => {
    if (event.persisted) void bootstrapConsoleSession(true)
  })
}
