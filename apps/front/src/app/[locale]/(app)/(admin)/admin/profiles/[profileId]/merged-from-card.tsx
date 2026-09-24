import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { ProfileMerge } from '@/models/admin'
import { getTranslations } from 'next-intl/server'
import { formatDate } from '../../format'

type Props = {
  merges: ProfileMerge[]
  locale: string
}

// Every profile an admin merged into this one, for support history. The source profiles no longer exist.
export default async function MergedFromCard({ merges, locale }: Props) {
  const t = await getTranslations({
    locale,
    namespace: 'admin.profiles.merge',
  })

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">{t('mergedFromTitle')}</CardTitle>
        <CardDescription>{t('mergedFromDescription')}</CardDescription>
      </CardHeader>
      <CardContent>
        <ul className="flex flex-col gap-1 text-sm">
          {merges.map((merge) => (
            <li
              key={merge.id}
              className="flex items-center justify-between gap-2"
            >
              <span className="font-medium">{merge.sourceUsername}</span>
              <span className="text-muted-foreground">
                {formatDate(merge.createdAt, locale)}
              </span>
            </li>
          ))}
        </ul>
      </CardContent>
    </Card>
  )
}
