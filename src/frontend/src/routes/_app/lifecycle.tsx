import { createFileRoute } from '@tanstack/react-router'
import { PagePlaceholder } from '@/features/dashboard/components/PagePlaceholder'

export const Route = createFileRoute('/_app/lifecycle')({
  component: () => <PagePlaceholder title="Lifecycle" />,
})
