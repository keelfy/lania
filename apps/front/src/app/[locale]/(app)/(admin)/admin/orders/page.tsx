import { Badge } from '@/components/ui/badge'
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
import { getAdminOrders } from '@/lib/api-endpoints'
import { CURRENCY_SYMBOLS, DEFAULT_CURRENCY } from '@/lib/currency'
import { serverApiFetcher } from '@/lib/server'
import { cn } from '@/lib/utils'
import { OrderAmount, OrderStatus } from '@/models/order'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'
import AdminPageHeader from '../admin-page-header'
import AdminSearch from '../admin-search'
import { formatDateTime } from '../format'

// A full order ID. The API matches only the first 16 characters, as a prefix.
const MAX_SEARCH_LENGTH = 36

// In the order an order goes through them.
const STATUSES: OrderStatus[] = ['created', 'processing', 'completed', 'failed']

const STATUS_BADGE = {
  created: 'outline',
  processing: 'secondary',
  completed: 'default',
  failed: 'destructive',
} as const

type Props = {
  params: Promise<{
    locale: string
  }>
  searchParams: Promise<{
    q?: string
    page?: string
    status?: string
  }>
}

function formatAmount(amounts: OrderAmount[]) {
  const amount =
    amounts.find((amount) => amount.currency === DEFAULT_CURRENCY) ?? amounts[0]
  if (!amount) return '—'
  return `${amount.amount} ${CURRENCY_SYMBOLS[amount.currency] ?? amount.currency}`
}

export default async function AdminOrdersPage({ params, searchParams }: Props) {
  await requireAdmin()
  const { locale } = await params
  const { q, page: pageParam, status: statusParam } = await searchParams
  const t = await getTranslations({ locale, namespace: 'admin.orders' })
  const search = q?.trim() ?? ''
  const status = STATUSES.find((status) => status === statusParam)
  // Zero-based, the URL shows it starting from 1.
  const page = Math.max(0, (parseInt(pageParam ?? '') || 1) - 1)

  const orders = await getAdminOrders(
    serverApiFetcher,
    page,
    search,
    status,
  ).catch((error) => {
    console.error(error)
    return undefined
  })

  const path = `/${locale}/admin/orders`
  const hrefFor = (target: { page?: number; status?: OrderStatus | null }) => {
    const params = new URLSearchParams()
    if (search) params.set('q', search)
    const targetStatus = target.status === undefined ? status : target.status
    if (targetStatus) params.set('status', targetStatus)
    if (target.page) params.set('page', (target.page + 1).toString())
    return `${path}?${params.toString()}`
  }
  const total = orders
    ? STATUSES.reduce((sum, status) => sum + orders.statusCounts[status], 0)
    : undefined

  return (
    <div className="flex flex-col gap-6">
      <AdminPageHeader
        title={t('title')}
        description={t('description')}
        actions={
          <AdminSearch
            path={path}
            defaultValue={search}
            placeholder={t('searchPlaceholder')}
            maxLength={MAX_SEARCH_LENGTH}
            params={status ? { status } : undefined}
          />
        }
      />
      <nav className="flex flex-wrap gap-2" aria-label={t('columns.status')}>
        <Button
          asChild
          size="sm"
          variant={status ? 'ghost' : 'secondary'}
          aria-current={status ? undefined : 'page'}
        >
          <Link href={hrefFor({ status: null })}>
            {t('allStatuses')}
            {total !== undefined && (
              <span className="text-muted-foreground tabular-nums">
                {total}
              </span>
            )}
          </Link>
        </Button>
        {STATUSES.map((option) => {
          const count = orders?.statusCounts[option]
          // A failed order is paid but not granted, so it waits for an admin.
          const needsAttention = option === 'failed' && !!count
          return (
            <Button
              key={option}
              asChild
              size="sm"
              variant={status === option ? 'secondary' : 'ghost'}
              aria-current={status === option ? 'page' : undefined}
            >
              <Link href={hrefFor({ status: option })}>
                {t(`statuses.${option}`)}
                {count !== undefined && (
                  <span
                    className={cn(
                      'tabular-nums',
                      needsAttention
                        ? 'bg-destructive rounded-sm px-1.5 text-white'
                        : 'text-muted-foreground',
                    )}
                  >
                    {count}
                  </span>
                )}
              </Link>
            </Button>
          )
        })}
      </nav>
      {!orders ? (
        <p className="text-destructive py-10 text-center">{t('loadFailed')}</p>
      ) : orders.content.length === 0 ? (
        <p className="text-muted-foreground py-10 text-center">
          {t('noResults')}
        </p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('columns.order')}</TableHead>
              <TableHead>{t('columns.items')}</TableHead>
              <TableHead>{t('columns.buyer')}</TableHead>
              <TableHead className="text-right">
                {t('columns.amount')}
              </TableHead>
              <TableHead>{t('columns.status')}</TableHead>
              <TableHead>{t('columns.createdAt')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {orders.content.map((order) => (
              <TableRow key={order.id} className="align-top">
                <TableCell>
                  <div className="flex flex-col gap-1">
                    <span className="font-mono text-xs" title={order.id}>
                      {order.id.slice(0, 8)}
                    </span>
                    <span className="text-muted-foreground text-xs">
                      {order.externalId
                        ? t('payment', { id: order.externalId })
                        : t('noPayment')}
                    </span>
                  </div>
                </TableCell>
                <TableCell>
                  <ul className="flex flex-col gap-1">
                    {order.items.map((item) => (
                      <li key={item.id} className="text-sm">
                        <span className="font-medium">
                          {item.productName || t('unknownProduct')}
                        </span>{' '}
                        <Link
                          href={`/${locale}/admin/profiles/${item.profileId}`}
                          className="hover:underline"
                        >
                          {item.username || item.profileId}
                        </Link>
                        {item.seasonName && (
                          <span className="text-muted-foreground">
                            , {item.seasonName}
                          </span>
                        )}
                      </li>
                    ))}
                  </ul>
                </TableCell>
                <TableCell>
                  <Link
                    href={`/${locale}/admin/users/${order.userId}`}
                    className="text-muted-foreground font-mono text-xs hover:underline"
                    title={order.userId}
                  >
                    {order.userId.slice(0, 8)}
                  </Link>
                </TableCell>
                <TableCell className="text-right font-medium tabular-nums">
                  {formatAmount(order.amounts)}
                </TableCell>
                <TableCell>
                  <div className="flex flex-col gap-1">
                    <Badge variant={STATUS_BADGE[order.status]}>
                      {t(`statuses.${order.status}`)}
                    </Badge>
                    {order.updatedAt !== order.createdAt && (
                      <span className="text-muted-foreground text-xs">
                        {t('updatedAt', {
                          date: formatDateTime(order.updatedAt, locale),
                        })}
                      </span>
                    )}
                  </div>
                </TableCell>
                <TableCell className="whitespace-nowrap">
                  {formatDateTime(order.createdAt, locale)}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
      {orders && orders.totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          {page > 0 && (
            <Button asChild variant="outline">
              <Link href={hrefFor({ page: page - 1 })}>
                {t('previousPage')}
              </Link>
            </Button>
          )}
          <span className="text-muted-foreground text-sm">
            {t('pageOf', { page: page + 1, total: orders.totalPages })}
          </span>
          {page < orders.totalPages - 1 && (
            <Button asChild variant="outline">
              <Link href={hrefFor({ page: page + 1 })}>{t('nextPage')}</Link>
            </Button>
          )}
        </div>
      )}
    </div>
  )
}
