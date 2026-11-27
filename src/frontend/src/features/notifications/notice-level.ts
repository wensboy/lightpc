import { CircleAlertIcon, CircleCheckIcon, CircleXIcon, InfoIcon, TriangleAlertIcon, type LucideIcon } from 'lucide-react'
import type { NoticeLevelDto } from '@/services/notice'

// The five severities share one vocabulary across the sheet, the toast and the footer. Keeping the tables in their own
// module means a level is defined once: the sheet and the toast render different shapes, but never disagree about what
// "fail" looks like.
export const noticeLevelIcons: Record<NoticeLevelDto, LucideIcon> = {
  info: InfoIcon, // 信息
  success: CircleCheckIcon, // 成功
  fail: CircleXIcon, // 失败
  error: CircleAlertIcon, // 错误
  warn: TriangleAlertIcon, // 警告
}

// Tones follow the order of escalation: blue for information, green for a success, orange for a failure that was
// handled, red for an error, yellow for a warning worth reading but not acting on.
export const noticeLevelTones: Record<NoticeLevelDto, string> = {
  info: 'text-blue-400',
  success: 'text-emerald-400',
  fail: 'text-orange-400',
  error: 'text-red-400',
  warn: 'text-amber-400',
}

// noticeLevelOrder is the order the footer counts appear in, from the mildest to the most severe.
export const noticeLevelOrder = ['fail', 'error', 'warn'] as const
