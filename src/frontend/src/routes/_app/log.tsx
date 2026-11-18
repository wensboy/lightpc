import { createFileRoute } from '@tanstack/react-router'
import { PagePlaceholder } from '@/features/dashboard/components/PagePlaceholder'

export const Route = createFileRoute('/_app/log')({
  component: () => <PagePlaceholder title="Log" />,
})
