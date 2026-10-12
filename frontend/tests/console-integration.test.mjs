import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import vm from 'node:vm'
import { createRequire } from 'node:module'
import { transformSync } from 'esbuild'

const root = path.resolve(import.meta.dirname, '..')
const require = createRequire(path.join(root, 'package.json'))
const source = fs.readFileSync(path.join(root, 'src/integration/console.ts'), 'utf8')

function loadConsole({
  pathname = '/recon/',
  search = '?embedded=1',
  cookie = '',
  layout = false,
  fetch = async () => new Response('{}')
} = {}) {
  const events = new Map()
  const classes = new Set()
  const properties = new Map()
  const messages = []
  const parent = { postMessage: (...args) => messages.push(args) }
  const window = {
    location: { pathname, search, origin: 'http://console.test' },
    parent,
    setTimeout,
    clearTimeout,
    addEventListener: (type, listener) => events.set(type, listener)
  }
  const document = {
    cookie,
    documentElement: {
      classList: {
        add: (value) => classes.add(value),
        toggle: (value, enabled) => (enabled ? classes.add(value) : classes.delete(value))
      },
      style: { setProperty: (key, value) => properties.set(key, value) }
    }
  }
  let height = 1600
  const content = { getBoundingClientRect: () => ({ height }), offsetHeight: 0, scrollHeight: 0 }
  const observers = []
  const overlays = []
  if (layout) {
    document.querySelector = (selector) => selector === '.recon-embedded-view' ? content : null
    document.querySelectorAll = (selector) => overlays.filter(element => selector.includes(element.selector))
    document.documentElement.getBoundingClientRect = () => ({ height: 0 })
    window.requestAnimationFrame = (callback) => { callback(); return 1 }
    window.cancelAnimationFrame = () => {}
  }
  const module = { exports: {} }
  vm.runInNewContext(
    transformSync(source, {
      loader: 'ts',
      format: 'cjs',
      define: { 'import.meta.env': JSON.stringify({ VITE_APP_TITLE: 'Recon', BASE_URL: '/' }) }
    }).code,
    {
      require,
      module,
      exports: module.exports,
      window,
      document,
      fetch,
      URLSearchParams,
      AbortController,
      ...(layout ? {
        ResizeObserver: class { constructor(callback) { observers.push(callback) } observe() {} disconnect() {} },
        MutationObserver: class { constructor(callback) { observers.push(callback) } observe() {} disconnect() {} },
        getComputedStyle: (element) => ({ display: 'block', visibility: 'visible', marginTop: '0px', marginBottom: '0px', ...element.style })
      } : {})
    }
  )
  return { ...module.exports, window, document, events, classes, properties, messages, observers,
    resize: (value) => { height = value; observers.forEach(callback => callback()) },
    overlay: (selector, visible) => {
      overlays.splice(0, overlays.length, { selector, style: { display: visible ? 'block' : 'none' }, getBoundingClientRect: () => ({ height: 600 }) })
      observers.forEach(callback => callback())
    }
  }
}

const adminResponse = () =>
  new Response(
    JSON.stringify({ user: { id: 'admin-fixture', email: 'admin@example.test', role: 'admin' } })
  )

test('gateway and embedded mode require the configured path and explicit flag', () => {
  assert.equal(loadConsole().isEmbedded, true)
  assert.equal(loadConsole({ search: '' }).isEmbedded, false)
  assert.equal(loadConsole({ pathname: '/recon-other/' }).isConsoleGateway, false)
  assert.equal(loadConsole({ pathname: '/' }).isConsoleGateway, false)
})

test('administrator bootstrap uses the console session and reports readiness without token state', async () => {
  const requests = []
  const app = loadConsole({
    fetch: async (...args) => {
      requests.push(args)
      return adminResponse()
    }
  })
  const user = await app.bootstrapConsoleSession()
  assert.equal(user.email, 'admin@example.test')
  assert.equal(app.consoleSession.status, 'ready')
  assert.equal(requests[0][0], '/api/auth/me')
  assert.equal(requests[0][1].credentials, 'same-origin')
  assert.equal(requests[0][1].cache, 'no-store')
  assert.deepEqual(Object.keys(user).sort(), ['email', 'id', 'role'])
  assert.equal(app.messages[0][0].type, 'recon:ready')
  assert.equal(app.messages[0][1], 'http://console.test')
  await app.bootstrapConsoleSession()
  assert.equal(requests.length, 1)
})

test('expired, denied and member sessions cannot become collection administrators', async () => {
  for (const status of [401, 403]) {
    const app = loadConsole({ fetch: async () => new Response('{}', { status }) })
    assert.equal(await app.bootstrapConsoleSession(), null)
    assert.equal(app.consoleSession.status, status === 401 ? 'expired' : 'denied')
    assert.equal(app.messages.length, status === 401 ? 1 : 0)
  }
  const app = loadConsole({
    fetch: async () =>
      new Response(
        JSON.stringify({
          user: { id: 'member-fixture', email: 'member@example.test', role: 'member' }
        })
      )
  })
  assert.equal(await app.bootstrapConsoleSession(), null)
  assert.equal(app.consoleSession.status, 'denied')
})

test('failed bootstrap can retry and simultaneous navigations share one check', async () => {
  let succeed = false
  let calls = 0
  const app = loadConsole({
    fetch: async () => {
      calls++
      if (!succeed) throw new Error('Offline')
      return adminResponse()
    }
  })
  await app.bootstrapConsoleSession()
  assert.equal(app.consoleSession.status, 'error')
  succeed = true
  await Promise.all([app.bootstrapConsoleSession(true), app.bootstrapConsoleSession(true)])
  assert.equal(app.consoleSession.status, 'ready')
  assert.equal(calls, 2)
})

test('mutation CSRF follows secure-cookie priority and never appears on read requests', () => {
  const app = loadConsole({ cookie: 'lokhlin_csrf=fallback; __Host-lokhlin_csrf=fixture%2Dcsrf' })
  assert.equal(app.csrfHeaders('POST')['X-CSRF-Token'], 'fixture-csrf')
  assert.equal(app.csrfHeaders('DELETE')['X-CSRF-Token'], 'fixture-csrf')
  assert.equal(Object.keys(app.csrfHeaders('GET')).length, 0)
  assert.equal(
    Object.keys(loadConsole({ cookie: 'lokhlin_csrf=%E0%A4' }).csrfHeaders('POST')).length,
    0
  )
  assert.equal(
    Object.keys(loadConsole({ pathname: '/', cookie: 'lokhlin_csrf=fixture' }).csrfHeaders('POST'))
      .length,
    0
  )
})

test('stored images and direct uploads retain the gateway prefix', () => {
  const app = loadConsole()
  assert.equal(app.resourcePath('/images/icon/hash.png'), '/recon/images/icon/hash.png')
  assert.equal(app.resourcePath('/api/poc/data/import'), '/recon/api/poc/data/import')
  app.handleConsoleUploadError({ status: 401 })
  assert.equal(app.consoleSession.status, 'expired')
  assert.equal(app.messages.at(-1)[0].type, 'recon:session-expired')
})

test('host theme messages require both the same origin and parent source', () => {
  const app = loadConsole()
  app.setupConsoleIntegration()
  const listener = app.events.get('message')
  const data = {
    type: 'recon:theme',
    theme: 'dark',
    tokens: { background: '#0b0b0b', radius: '4px', unknown: 'ignored' }
  }
  listener({ origin: 'http://other.test', source: app.window.parent, data })
  assert.equal(app.classes.has('dark'), false)
  listener({ origin: app.window.location.origin, source: {}, data })
  assert.equal(app.classes.has('dark'), false)
  listener({ origin: app.window.location.origin, source: app.window.parent, data })
  assert.equal(app.classes.has('dark'), true)
  assert.equal(app.consoleTheme.dark, true)
  assert.equal(app.properties.get('--recon-background'), '#0b0b0b')
  assert.equal(app.properties.get('--recon-radius'), '4px')
  assert.equal(app.properties.has('--recon-unknown'), false)
})

test('embedded content reports intrinsic height and can shrink without a viewport floor', () => {
  const app = loadConsole({ layout: true })
  app.setupConsoleIntegration()
  const reports = () => app.messages.filter(([message]) => message.type === 'recon:layout')
  assert.equal(reports().at(-1)?.[0].height, 1600)
  app.resize(360)
  assert.equal(reports().at(-1)?.[0].height, 360)
  assert.equal(reports().at(-1)?.[1], app.window.location.origin)
  const previous = reports().length
  app.resize(360)
  assert.equal(reports().length, previous)
  const standalone = loadConsole({ pathname: '/', layout: true })
  standalone.setupConsoleIntegration()
  assert.equal(standalone.messages.length, 0)
})

test('parent viewport coordinates are trusted only from the same-origin host and finite bounds', () => {
  const app = loadConsole()
  app.setupConsoleIntegration()
  const receive = app.events.get('message')
  const data = { type: 'recon:viewport', top: 900, height: 600 }
  receive({ origin: 'http://other.test', source: app.window.parent, data })
  assert.equal(app.properties.has('--recon-viewport-top'), false)
  receive({ origin: app.window.location.origin, source: {}, data })
  assert.equal(app.properties.has('--recon-viewport-top'), false)
  receive({ origin: app.window.location.origin, source: app.window.parent, data })
  assert.equal(app.properties.get('--recon-viewport-top'), '900px')
  assert.equal(app.properties.get('--recon-viewport-height'), '600px')
  receive({ origin: app.window.location.origin, source: app.window.parent, data: { ...data, top: Infinity } })
  assert.equal(app.properties.get('--recon-viewport-top'), '900px')
})

test('back-forward cache restoration retains live size and overlay observation', () => {
  const app = loadConsole({ layout: true })
  app.setupConsoleIntegration()
  app.events.get('pagehide')({ persisted: true })
  app.resize(720)
  assert.equal(app.messages.filter(([message]) => message.type === 'recon:layout').at(-1)?.[0].height, 720)
  app.events.get('pagehide')({ persisted: false })
  app.resize(960)
  assert.equal(app.messages.filter(([message]) => message.type === 'recon:layout').at(-1)?.[0].height, 720)
})

test('image previews and dialogs lock the host only while visible', () => {
  const app = loadConsole({ layout: true })
  app.setupConsoleIntegration()
  const report = () => app.messages.filter(([message]) => message.type === 'recon:layout').at(-1)?.[0]
  for (const selector of ['.el-image-viewer__wrapper', '.el-overlay']) {
    app.overlay(selector, true)
    assert.equal(report().overlay, true)
    app.overlay(selector, false)
    assert.equal(report().overlay, false)
    assert.equal(report().height, 1600)
  }
})

function locale(name) {
  const module = { exports: {} }
  vm.runInNewContext(
    transformSync(fs.readFileSync(path.join(root, `src/locales/${name}.ts`), 'utf8'), {
      loader: 'ts',
      format: 'cjs'
    }).code,
    { module, exports: module.exports }
  )
  return module.exports.default
}

test('Korean defaults and English fallback cover the collection screens and session failures', () => {
  const ko = locale('ko')
  const en = locale('en')
  for (const group of [
    'console',
    'common',
    'router',
    'task',
    'asset',
    'project',
    'node',
    'plugin',
    'configuration'
  ]) {
    for (const key of Object.keys(ko[group]))
      assert.equal(typeof en[group][key], 'string', `${group}.${key} requires an English fallback`)
  }
  for (const key of ['loadingTitle', 'expiredTitle', 'deniedTitle', 'errorTitle'])
    assert.match(ko.console[key], /[가-힣]/)
  assert.match(
    fs.readFileSync(path.join(root, 'src/plugins/vueI18n/index.ts'), 'utf8'),
    /fallbackLocale: 'en'/
  )
  assert.match(
    fs.readFileSync(path.join(root, 'src/store/modules/locale.ts'), 'utf8'),
    /getStorage\('lang'\) === 'en' \? 'en' : 'ko'/
  )
})

test('project text contains no removed language or obsolete project metadata', () => {
  function walk(directory) {
    for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
      if (['node_modules', 'dist-pro', 'dist-dev', 'dist-test', '.git'].includes(entry.name))
        continue
      const file = path.join(directory, entry.name)
      if (entry.isDirectory()) walk(file)
      else if (/\.(ts|tsx|vue|js|css|less|svg|md|html)$/.test(file))
        assert.equal(
          /\p{Script=Han}/u.test(fs.readFileSync(file, 'utf8')),
          false,
          path.relative(root, file)
        )
    }
  }
  walk(root)
  const pkg = JSON.parse(fs.readFileSync(path.join(root, 'package.json'), 'utf8'))
  for (const key of ['author', 'license', 'repository', 'homepage', 'bugs'])
    assert.equal(pkg[key], undefined)
  assert.equal(fs.existsSync(path.join(root, 'LICENSE')), false)
})

test('literal translation references resolve to meaningful Korean or English text', () => {
  const ko = locale('ko')
  const en = locale('en')
  function walk(directory) {
    for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
      const file = path.join(directory, entry.name)
      if (entry.isDirectory()) walk(file)
      else if (/\.(ts|tsx|vue)$/.test(file)) {
        for (const match of fs.readFileSync(file, 'utf8').matchAll(/\bt\(['"]([\w.]+)['"]/g)) {
          const key = match[1]
          if (!key.includes('.')) continue
          const lookup = (messages) =>
            key.split('.').reduce((value, part) => value?.[part], messages)
          assert.equal(
            typeof (lookup(ko) ?? lookup(en)),
            'string',
            `${path.relative(root, file)}: ${key}`
          )
        }
      }
    }
  }
  walk(path.join(root, 'src'))
})

test('maintained topology requests preserve the inherited pagination and query contract', async () => {
  const requests = []
  const module = { exports: {} }
  vm.runInNewContext(
    transformSync(fs.readFileSync(path.join(root, 'src/api/asset/map.ts'), 'utf8'), {
      loader: 'ts',
      format: 'cjs'
    }).code,
    {
      module,
      exports: module.exports,
      require: () => ({
        post: async (config) => {
          requests.push(config)
          return config
        }
      })
    }
  )
  await module.exports.getMapHostsApi('example.test', 2)
  await module.exports.getMapAssetsApi('app.example.test', 3)
  await module.exports.getMapSensitiveApi('https://app.example.test/')
  await module.exports.getMapVulnerabilityApi('https://app.example.test/', 4)
  assert.equal(requests[0].url, '/api/assets/asset/hosts')
  assert.equal(requests[0].data.rootDomain, 'example.test')
  assert.equal(requests[0].data.pageSize, 10)
  assert.equal(requests[1].url, '/api/assets/asset/by-host')
  assert.equal(requests[1].data.host, 'app.example.test')
  assert.equal(requests[1].data.pageIndex, 3)
  assert.equal(requests[2].data.search, 'url ="https://app.example.test/"')
  assert.equal(requests[3].data.pageIndex, 4)
  assert.equal(requests[3].data.pageSize, 20)
})

test('empty project options are arrays and populated project trees remain intact', async () => {
  for (const list of [null, undefined, [], [{ value: 'project', label: 'Fixture', children: [{ value: 'target', label: 'Target' }] }]]) {
    const requests = []
    const response = { code: 200, data: { list } }
    const module = { exports: {} }
    vm.runInNewContext(
      transformSync(fs.readFileSync(path.join(root, 'src/api/project/index.ts'), 'utf8'), { loader: 'ts', format: 'cjs' }).code,
      { module, exports: module.exports, require: () => ({ get: async (request) => { requests.push(request); return response } }) }
    )
    const result = await module.exports.getProjectAllApi()
    assert.equal(Array.isArray(result.data.list), true)
    assert.equal(result.data.list.length, list?.length ?? 0)
    if (list) assert.equal(result.data.list, list)
    assert.equal(requests[0].url, '/api/project/all')
  }
})

test('map entry preserves the embedded query and uses the authenticated Vue screen', () => {
  const html = fs.readFileSync(path.join(root, 'public/map.html'), 'utf8')
  const script = html.match(/<script>([\s\S]*?)<\/script>/)[1]
  let destination
  vm.runInNewContext(script, {
    location: {
      pathname: '/recon/map',
      search: '?embedded=1&rootDomain=example.test',
      replace: (value) => {
        destination = value
      }
    }
  })
  assert.equal(destination, '/recon/?embedded=1&rootDomain=example.test#/asset-map/index')
})
