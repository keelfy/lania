import { Badge } from '@/components/ui/badge'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { ProfileUsernameChange } from '@/models/admin'
import { ArrowRightIcon } from 'lucide-react'
import { getTranslations } from 'next-intl/server'
import { formatDate } from '../../format'

type Props = {
  changes: ProfileUsernameChange[]
  locale: string
}

// Every nickname the profile had, for support: an unlicensed profile left its old UUID with each change.
export default async function UsernameChangesCard({ changes, locale }: Props) {
  const t = await getTranslations({
    locale,
    namespace: 'admin.profiles.usernameChanges',
  })

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">{t('title')}</CardTitle>
        <CardDescription>{t('description')}</CardDescription>
      </CardHeader>
      <CardContent>
        <ul className="flex flex-col gap-2 text-sm">
          {changes.map((change) => (
            <li
              key={change.id}
              className="flex flex-wrap items-center justify-between gap-2"
            >
              <span
                className="flex items-center gap-2"
                title={
                  change.oldMcUuid === change.newMcUuid
                    ? change.newMcUuid
                    : `${change.oldMcUuid} → ${change.newMcUuid}`
                }
              >
                <span className="text-muted-foreground">
                  {change.oldUsername}
                </span>
                <ArrowRightIcon className="text-muted-foreground size-3.5" />
                <span className="font-medium">{change.newUsername}</span>
                <Badge variant="outline">{t(`source.${change.source}`)}</Badge>
              </span>
              <span className="text-muted-foreground">
                {formatDate(change.createdAt, locale)}
              </span>
            </li>
          ))}
        </ul>
      </CardContent>
    </Card>
  )
}
