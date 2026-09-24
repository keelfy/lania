import { requireAdmin } from '@/lib/admin'
import { redirect } from 'next/navigation'

type Props = {
  params: Promise<{
    locale: string
  }>
}

export default async function AdminPage({ params }: Props) {
  await requireAdmin()
  const { locale } = await params
  redirect(`/${locale}/admin/users`)
}
