import { createFileRoute } from '@tanstack/react-router'
import { PagePlaceholder } from '@/features/dashboard/components/PagePlaceholder'

export const Route = createFileRoute('/_app/projects')({
  component: () => <PagePlaceholder title="Projects" />,
})
