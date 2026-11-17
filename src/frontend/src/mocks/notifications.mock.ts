import type { NotificationEntry } from '@/features/notifications/store/useNotificationStore'

// notificationsMock covers the three severities plus long copy, so truncation and scrolling can be
// verified without a backend.
export const notificationsMock: NotificationEntry[] = [
  {
    id: 'task-completed',
    type: 'info',
    title: 'Task Completed: Task has been completed successfully. You can review the results in the editor.',
    description:
      'Task has been completed successfully. You can review the results in the editor, then continue with the next step from the command palette.',
    source: 'Tencent Cloud CodeBuddy',
    isExpanded: true,
  },
  {
    id: 'tool-permission',
    type: 'warning',
    title: 'Tool Permission Required: CodeBuddy needs permission to use external tools',
    description:
      'CodeBuddy needs permission to use external tools before it can continue. Grant access from the settings dialog to let the run finish automatically.',
    source: 'Tencent Cloud CodeBuddy',
    isExpanded: false,
  },
  {
    id: 'build-failed',
    type: 'error',
    title: 'Build Failed: three type errors were reported',
    description:
      'The production build failed because three files contain type errors. Open the first one to inspect the failing declaration.',
    source: 'lightpc',
    isExpanded: false,
  },
  {
    id: 'settings-updated',
    type: 'info',
    title: 'Settings Updated',
    description: 'Editor font size changed to 14 and word wrap was enabled.',
    source: 'lightpc',
    isExpanded: false,
  },
  {
    id: 'extension-installed',
    type: 'info',
    title: 'Extension Installed: markdown preview',
    description:
      'The markdown preview extension was installed and is ready to use from the command palette or the editor context menu.',
    source: 'Extension Registry',
    isExpanded: false,
  },
  {
    id: 'network-slow',
    type: 'warning',
    title: 'Network Latency Detected',
    description:
      'Requests to the package registry took longer than two seconds. Check the proxy configuration if this persists.',
    source: 'lightpc',
    isExpanded: false,
  },
  {
    id: 'do-not-disturb',
    type: 'info',
    title: 'Do Not Disturb is enabled',
    description: 'Notifications are collected silently while do not disturb is on.',
    source: 'lightpc',
    isExpanded: false,
  },
  {
    id: 'renderer-crash',
    type: 'error',
    title: 'Renderer Process Restarted',
    description:
      'The renderer process was restarted after an unexpected exit. Unsaved editor state may have been lost.',
    source: 'lightpc',
    isExpanded: false,
  },
]
