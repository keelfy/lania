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
import { getAdminUsers } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'
import AdminPageHeader from '../admin-page-header'
import AdminSearch from '../admin-search'
import { formatDate } from '../format'
import RoleBadge from '../role-badge'

// The longest valid email address.
const MAX_SEARCH_LENGTH = 254

type Props = {
  params: Promise<{
    locale: string
  }>
  searchParams: Promise<{
    q?: string
    token?: string
  }>
}

export default async function AdminUsersPage({ params, searchParams }: Props) {
  await requireAdmin()
  const { locale } = await params
  const { q, token } = await searchParams
  const t = await getTranslations({ locale, namespace: 'admin.users' })
  const search = q?.trim() ?? ''

  const users = await getAdminUsers(serverApiFetcher, search, token).catch(
    (error) => {
      console.error(error)
      return undefined
    },
  )

  const path = `/${locale}/admin/users`
  const nextHref = users?.nextPageToken
    ? `${path}?${new URLSearchParams({
        ...(search ? { q: search } : {}),
        token: users.nextPageToken,
      })}`
    : undefined

  return (
    <div className="flex flex-col gap-6">
      <AdminPageHeader
        title={t('title')}
        actions={
          <AdminSearch
            path={path}
            defaultValue={search}
            placeholder={t('searchPlaceholder')}
            maxLength={MAX_SEARCH_LENGTH}
          />
        }
      />
      {!users ? (
        <p className="text-destructive py-10 text-center">{t('loadFailed')}</p>
      ) : users.content.length === 0 ? (
        <p className="text-muted-foreground py-10 text-center">
          {t('noResults')}
        </p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('columns.email')}</TableHead>
              <TableHead>{t('columns.username')}</TableHead>
              <TableHead>{t('columns.role')}</TableHead>
              <TableHead>{t('columns.createdAt')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {users.content.map((user) => (
              <TableRow key={user.id}>
                <TableCell>
                  <Link
                    href={`/${locale}/admin/users/${user.id}`}
                    className="font-medium hover:underline"
                  >
                    {user.email || user.id}
                  </Link>
                </TableCell>
                <TableCell>{user.username || '—'}</TableCell>
                <TableCell>
                  <RoleBadge role={user.role} />
                </TableCell>
                <TableCell>{formatDate(user.createdAt, locale)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
      {search && users && (
        <p className="text-muted-foreground text-center text-sm">
          {t('searchLimit', { count: users.content.length })}
        </p>
      )}
      {(token || nextHref) && (
        <div className="flex justify-center gap-2">
          {token && (
            <Button asChild variant="outline">
              <Link
                href={search ? `${path}?q=${encodeURIComponent(search)}` : path}
              >
                {t('firstPage')}
              </Link>
            </Button>
          )}
          {nextHref && (
            <Button asChild variant="outline">
              <Link href={nextHref}>{t('nextPage')}</Link>
            </Button>
          )}
        </div>
      )}
    </div>
  )
}
