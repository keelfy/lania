import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { requireAdmin } from '@/lib/admin'
import { getAdminProfiles } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'
import AdminPageHeader from '../admin-page-header'
import AdminSearch from '../admin-search'
import { formatDate } from '../format'
import RoleBadge from '../role-badge'
import RekeyPremiumButton from './rekey-premium-button'

// The longest Minecraft username.
const MAX_SEARCH_LENGTH = 16

type Props = {
  params: Promise<{
    locale: string
  }>
  searchParams: Promise<{
    q?: string
    page?: string
  }>
}

export default async function AdminProfilesPage({
  params,
  searchParams,
}: Props) {
  await requireAdmin()
  const { locale } = await params
  const { q, page: pageParam } = await searchParams
  const t = await getTranslations({ locale, namespace: 'admin.profiles' })
  const search = q?.trim() ?? ''
  // Zero-based, the URL shows it starting from 1.
  const page = Math.max(0, (parseInt(pageParam ?? '') || 1) - 1)

  const profiles = await getAdminProfiles(serverApiFetcher, page, search).catch(
    (error) => {
      console.error(error)
      return undefined
    },
  )

  const path = `/${locale}/admin/profiles`
  const hrefFor = (target: number) => {
    const params = new URLSearchParams()
    if (search) params.set('q', search)
    params.set('page', (target + 1).toString())
    return `${path}?${params.toString()}`
  }

  return (
    <div className="flex flex-col gap-6">
      <AdminPageHeader
        title={t('title')}
        actions={
          <div className="flex flex-wrap items-center gap-2">
            <AdminSearch
              path={path}
              defaultValue={search}
              placeholder={t('searchPlaceholder')}
              maxLength={MAX_SEARCH_LENGTH}
            />
            <RekeyPremiumButton />
          </div>
        }
      />
      {!profiles ? (
        <p className="text-destructive py-10 text-center">{t('loadFailed')}</p>
      ) : profiles.content.length === 0 ? (
        <p className="text-muted-foreground py-10 text-center">
          {t('noResults')}
        </p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('columns.username')}</TableHead>
              <TableHead>{t('columns.owner')}</TableHead>
              <TableHead>{t('columns.role')}</TableHead>
              <TableHead>{t('columns.createdAt')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {profiles.content.map((profile) => (
              <TableRow key={profile.id}>
                <TableCell>
                  <Link
                    href={`/${locale}/admin/profiles/${profile.id}`}
                    className="font-medium hover:underline"
                  >
                    {profile.username}
                  </Link>
                </TableCell>
                <TableCell>
                  {profile.ownerUserId ? (
                    <Link
                      href={`/${locale}/admin/users/${profile.ownerUserId}`}
                      className="text-muted-foreground font-mono text-xs hover:underline"
                    >
                      {profile.ownerUserId}
                    </Link>
                  ) : (
                    <span className="text-muted-foreground">
                      {t('noOwner')}
                    </span>
                  )}
                </TableCell>
                <TableCell>
                  <RoleBadge role={profile.role} />
                </TableCell>
                <TableCell>{formatDate(profile.createdAt, locale)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
      {profiles && profiles.totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          {page > 0 && (
            <Button asChild variant="outline">
              <Link href={hrefFor(page - 1)}>{t('previousPage')}</Link>
            </Button>
          )}
          <span className="text-muted-foreground text-sm">
            {t('pageOf', { page: page + 1, total: profiles.totalPages })}
          </span>
          {page < profiles.totalPages - 1 && (
            <Button asChild variant="outline">
              <Link href={hrefFor(page + 1)}>{t('nextPage')}</Link>
            </Button>
          )}
        </div>
      )}
    </div>
  )
}
