import { redirect } from 'next/navigation'

type Props = { params: Promise<{ locale: string }> }

// Color and prefix catalogs live on their own sidebar sub-pages.
export default async function AdminCosmeticsPage({ params }: Props) {
  const { locale } = await params
  redirect(`/${locale}/admin/cosmetics/colors`)
}
