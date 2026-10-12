import '@/plugins/unocss'

import '@/plugins/svgIcon'

import { setupI18n } from '@/plugins/vueI18n'

import { setupStore } from '@/store'

import { setupGlobCom } from '@/components'

import { setupElementPlus } from '@/plugins/elementPlus'

import '@/styles/index.less'

import '@/plugins/animate.css'

import { setupRouter } from './router'

import { setupPermission } from './directives'

import { createApp } from 'vue'

import App from './App.vue'

import './permission'
import { setupConsoleIntegration } from '@/integration/console'

const setupAll = async () => {
  setupConsoleIntegration()
  const app = createApp(App)

  await setupI18n(app)

  setupStore(app)

  setupGlobCom(app)

  setupElementPlus(app)

  setupRouter(app)

  setupPermission(app)

  app.mount('#app')
}

setupAll()
