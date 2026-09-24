'use client'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { setProfileRole } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { ProfileRole } from '@/models/profile'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import React from 'react'
import { toast } from 'sonner'

const ROLES: ProfileRole[] = ['owner', 'admin', 'mod', 'player']

type Props = {
  profileId: string
  role: ProfileRole
}

export default function RoleCard({ profileId, role }: Props) {
  const t = useTranslations('admin.profiles.role')
  const router = useRouter()
  const [selected, setSelected] = React.useState<ProfileRole>(role)
  const [isPending, startTransition] = React.useTransition()

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault()
    if (isPending) return

    startTransition(async () => {
      try {
        await setProfileRole(clientApiFetcher, profileId, selected)
        toast.success(t('saved', { role: t(`names.${selected}`) }))
        router.refresh()
      } catch (error) {
        errorToast(t('saveFailed'), error)
      }
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">{t('title')}</CardTitle>
        <CardDescription>{t('description')}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="flex flex-wrap gap-2">
          <Select
            value={selected}
            onValueChange={(value) => setSelected(value as ProfileRole)}
          >
            <SelectTrigger
              aria-label={t('select')}
              className="flex-1 sm:min-w-64"
            >
              <SelectValue placeholder={t('select')} />
            </SelectTrigger>
            <SelectContent>
              {ROLES.map((item) => (
                <SelectItem key={item} value={item}>
                  {t(`names.${item}`)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button type="submit" disabled={isPending}>
            {t('save')}
          </Button>
        </form>
      </CardContent>
    </Card>
  )
}
