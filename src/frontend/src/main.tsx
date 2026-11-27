import React from 'react'
import ReactDOM from 'react-dom/client'
import { createHashHistory, createRouter, RouterProvider } from '@tanstack/react-router'
import { applyAllSettingEffects } from '@/features/settings/effects'
import { applySystemColorMode } from './lib/color-mode'
import { settingService } from '@/services/setting'
import { routeTree } from './routeTree.gen'
import './styles/index.css'

// The window is served by the Wails asset server, which maps a path onto a file and falls back to index.html
// when there is none. A hash history keeps every route inside that one document, so a reload never depends on
// the fallback being reached.
const router = createRouter({ routeTree, history: createHashHistory() })

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

applySystemColorMode()

// The stored mode decides the first paint: the system default covers the moment before the document
// arrives, then the effect registry applies whatever the user layer holds. A failure here only leaves the
// default theme in place, so it is allowed to stay silent.
void settingService.load()
  .then((document) => applyAllSettingEffects(document.values, document.themes))
  .catch(() => undefined)

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <RouterProvider router={router} />
  </React.StrictMode>,
)
