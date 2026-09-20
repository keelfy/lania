'use client'

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
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
import { revokeGrant } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { AdminCosmeticsCatalog, AdminGrant, AdminSeason } from '@/models/admin'
import { Product, ProductMetadata } from '@/models/product'
import { useLocale, useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import React from 'react'
import { toast } from 'sonner'
import { formatDateTime } from '../../format'
import GrantCosmeticDialog from './grant-cosmetic-dialog'
import GrantProductDialog from './grant-product-dialog'

type Props = {
  profileId: string
  // Missing when the grants could not be loaded.
  grants: AdminGrant[] | undefined
  seasons: AdminSeason[]
  products: Product<ProductMetadata>[]
  // Missing when the catalog could not be loaded.
  catalog: AdminCosmeticsCatalog | undefined
}

export default function GrantsCard({
  profileId,
  grants,
  seasons,
  products,
  catalog,
}: Props) {
  const t = useTranslations('admin.profiles.grants')
  const locale = useLocale()
  const router = useRouter()
  const [isPending, startTransition] = React.useTransition()

  const seasonNumber = (seasonId: string | undefined) =>
    seasons.find((season) => season.id === seasonId)?.seasonNumber

  const handleRevoke = (grant: AdminGrant) => {
    startTransition(async () => {
      try {
        await revokeGrant(clientApiFetcher, profileId, grant.type, grant.id)
        toast.success(grant.revokedAt ? t('synced') : t('revoked'))
        router.refresh()
      } catch (error) {
        errorToast(t('revokeFailed'), error)
        // The grant may be revoked even though the game server did not answer.
        router.refresh()
      }
    })
  }

  const sourceLabel = (grant: AdminGrant) => {
    if (grant.source)
      return t.has(`sources.${grant.source}`)
        ? t(`sources.${grant.source}`)
        : grant.source
    return grant.orderItemId ? t('sources.order') : t('sources.manual')
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-2">
        <div className="flex flex-col gap-1.5">
          <CardTitle className="text-lg">{t('title')}</CardTitle>
          <CardDescription>{t('description')}</CardDescription>
        </div>
        <div className="flex gap-2">
          <GrantCosmeticDialog
            profileId={profileId}
            seasons={seasons}
            catalog={catalog}
          />
          <GrantProductDialog
            profileId={profileId}
            seasons={seasons}
            products={products}
          />
        </div>
      </CardHeader>
      <CardContent>
        {!grants ? (
          <p className="text-destructive py-6 text-center">{t('loadFailed')}</p>
        ) : grants.length === 0 ? (
          <p className="text-muted-foreground py-6 text-center">{t('empty')}</p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('columns.grant')}</TableHead>
                <TableHead>{t('columns.season')}</TableHead>
                <TableHead>{t('columns.source')}</TableHead>
                <TableHead>{t('columns.status')}</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {grants.map((grant) => (
                <TableRow key={`${grant.type}:${grant.id}`}>
                  <TableCell>
                    <div className="flex flex-col gap-1">
                      <span className="font-medium">
                        {grant.name ?? t('seasonAccess')}
                      </span>
                      <Badge variant="outline">
                        {t(`types.${grant.type}`)}
                        {grant.prefixType &&
                          ` · ${t(`prefixTypes.${grant.prefixType}`)}`}
                      </Badge>
                    </div>
                  </TableCell>
                  <TableCell>
                    {seasonNumber(grant.seasonId) ?? t('anySeason')}
                  </TableCell>
                  <TableCell>{sourceLabel(grant)}</TableCell>
                  <TableCell>
                    <div className="flex flex-col gap-1 text-xs">
                      <span>{formatDateTime(grant.createdAt, locale)}</span>
                      {grant.revokedAt ? (
                        <Badge variant="destructive">
                          {t('revokedAt', {
                            date: formatDateTime(grant.revokedAt, locale),
                          })}
                        </Badge>
                      ) : (
                        <Badge>{t('active')}</Badge>
                      )}
                    </div>
                  </TableCell>
                  <TableCell className="text-right">
                    <AlertDialog>
                      <AlertDialogTrigger asChild>
                        <Button
                          variant={grant.revokedAt ? 'ghost' : 'outline'}
                          size="sm"
                          disabled={isPending}
                        >
                          {grant.revokedAt ? t('sync') : t('revoke')}
                        </Button>
                      </AlertDialogTrigger>
                      <AlertDialogContent>
                        <AlertDialogHeader>
                          <AlertDialogTitle>
                            {grant.revokedAt
                              ? t('syncConfirmTitle')
                              : t('revokeConfirmTitle')}
                          </AlertDialogTitle>
                          <AlertDialogDescription>
                            {grant.revokedAt
                              ? t('syncConfirmDescription')
                              : t('revokeConfirmDescription')}
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                          <AlertDialogCancel>{t('cancel')}</AlertDialogCancel>
                          <AlertDialogAction
                            onClick={() => handleRevoke(grant)}
                          >
                            {grant.revokedAt ? t('sync') : t('revoke')}
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  )
}
