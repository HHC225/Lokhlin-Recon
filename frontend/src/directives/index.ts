import type { App } from 'vue'
import { setupPermissionDirective } from './permission/hasPermi'

export const setupPermission = (app: App<Element>) => {
  setupPermissionDirective(app)
}
