# Recon frontend

Vue 3, Pinia, Vue Router hash routes, and Element Plus collection screens.

```sh
npm ci --ignore-scripts
npm run build:pro
npm run test:integration
```

Production output is `dist-pro/`, served by the console gateway at `/recon/`. Open `/recon/?embedded=1#/project-management/index` inside the console to hide the collection header, sidebar, settings, and account tools. Full collection navigation is available at `/recon/#/project-management/index`.

The gateway uses the existing administrator console session. Bootstrap checks `/api/auth/me`; mutations send the `__Host-dast_csrf` or `dast_csrf` cookie as `X-CSRF-Token`. Gateway mode never persists collection login tokens. Private MCP and account administration stay outside the embedded workspace.

Korean is the default language and English is the fallback. A previous Chinese locale preference resets to Korean. `VITE_APP_TITLE` sets the display name, with `Recon` as its default.

The host sends same-origin messages from the parent frame: `{type: 'recon:theme', theme: 'light' | 'dark', tokens: {...}}`. Supported token keys are `background`, `foreground`, `card`, `primary`, `primaryForeground`, `muted`, `mutedForeground`, `border`, `input`, `ring`, `radius`, and `fontFamily`. The frame reports `recon:ready` after session bootstrap and `recon:session-expired` for an expired session.

Optional external plugin and fingerprint catalog support uses `VITE_CATALOG_API_URL` and `VITE_CATALOG_WEB_URL`. Without these settings the UI explains that the catalog is unavailable; local plugin import remains usable.

For standalone frontend development, `npm run dev` serves the existing standalone login flow at `/`; Vite proxies collection APIs to `127.0.0.1:8082`.
