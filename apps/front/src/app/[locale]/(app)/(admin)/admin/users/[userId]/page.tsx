import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { requireAdmin } from '@/lib/admin'
import { getAdminUser } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'
import { notFound } from 'next/navigation'
import AdminPageHeader from '../../admin-page-header'
import { formatDate } from '../../format'
import RoleBadge from '../../role-badge'

type Props = {
  params: Promise<{
    locale: string
    userId: string
  }>
}

export default async function AdminUserPage({ params }: Props) {
  await requireAdmin()
  const { locale, userId } = await params
  const t = await getTranslations({ locale, namespace: 'admin.users' })

  const user = await getAdminUser(serverApiFetcher, userId).catch((error) => {
    console.error(error)
    return undefined
  })
  if (!user) notFound()

  return (
    <div className="flex flex-col gap-6">
      <AdminPageHeader
        backHref={`/${locale}/admin/users`}
        backLabel={t('back')}
      />
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">{user.email || user.id}</CardTitle>
          <CardDescription className="font-mono text-xs">
            {user.id}
          </CardDescription>
        </CardHeader>
        <CardContent className="grid grid-cols-2 gap-2 text-sm">
          <span className="text-muted-foreground">{t('columns.username')}</span>
          <span className="text-right">{user.username || '—'}</span>
          <span className="text-muted-foreground">{t('columns.role')}</span>
          <span className="flex justify-end">
            <RoleBadge role={user.role} />
          </span>
          <span className="text-muted-foreground">
            {t('columns.createdAt')}
          </span>
          <span className="text-right">
            {formatDate(user.createdAt, locale)}
          </span>
        </CardContent>
      </Card>
      <div className="flex flex-col gap-2">
        <h2 className="text-xl font-bold">{t('profiles')}</h2>
        {user.profiles.length === 0 ? (
          <p className="text-muted-foreground py-6 text-center">
            {t('noProfiles')}
          </p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('profileColumns.username')}</TableHead>
                <TableHead>{t('profileColumns.role')}</TableHead>
                <TableHead>{t('profileColumns.createdAt')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {user.profiles.map((profile) => (
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
                    <RoleBadge role={profile.role} />
                  </TableCell>
                  <TableCell>{formatDate(profile.createdAt, locale)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </div>
    </div>
  )
}
