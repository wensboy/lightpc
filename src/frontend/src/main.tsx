import React from 'react'
import ReactDOM from 'react-dom/client'
import { createRouter, RouterProvider } from '@tanstack/react-router'
import { applyAllSettingEffects } from '@/features/settings/effects'
import { applySystemColorMode } from './lib/color-mode'
import { loadSettings } from '@/mocks'
import { routeTree } from './routeTree.gen'
import './styles/index.css'

const router = createRouter({ routeTree })

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

applySystemColorMode()

// The stored mode decides the first paint: the system default covers the moment before the document
// arrives, then the effect registry applies whatever the user layer holds.
void loadSettings()
  .then((document) => applyAllSettingEffects(document.values, document.themes))
  .catch(() => undefined)

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <RouterProvider router={router} />
  </React.StrictMode>,
)
