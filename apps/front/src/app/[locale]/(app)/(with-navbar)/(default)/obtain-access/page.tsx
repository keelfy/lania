'use client'

import { Button } from '@/components/ui/button'
import { Form } from '@/components/ui/form'
import LoadingSpinner from '@/components/ui/loading-spinner'
import RichText from '@/components/ui/rich-text'
import {
  Table,
  TableBody,
  TableCell,
  TableFooter,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import SeasonSelect from '@/components/ui/season-select'
import { getAccessMode, isFreeAccess } from '@/lib/access-mode'
import { getSelectableSeasons, pickSeason } from '@/lib/seasons'
import {
  getSeasons,
  registerForPrimarySeason,
  requestAccess,
  requestFreeAccess,
} from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { useAuthStore } from '@/providers/auth-store'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  ArrowRightIcon,
  CheckIcon,
  CalendarIcon,
  CreditCardIcon,
  Gamepad2Icon,
  UserCheckIcon,
  ZapIcon,
} from 'lucide-react'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import { parseAsString, useQueryState } from 'nuqs'
import React from 'react'
import { useForm, useWatch } from 'react-hook-form'
import { toast } from 'sonner'
import SignInButton from '../../components/sign-in-button'
import { createObtainAccessFormSchema, ObtainAccessFormValues } from './form'
import { UsernameField } from './username-field'
import { PremiumNotice } from './premium-notice'
import { useBasket } from '@/context/basket'
import { Season } from '@/models/season'

type Props = {
  params: Promise<{
    locale: string
  }>
}

const lastSteps = {
  preregistration: 'steps.step3Pre',
  free: 'steps.step3Free',
  paid: 'steps.step3',
} as const

export default function ObtainAccessPage({ params }: Props) {
  const { locale } = React.use(params)
  const t = useTranslations('obtainAccess')
  const [seasons, setSeasons] = React.useState<Season[]>([])
  React.useEffect(() => {
    let cancelled = false
    getSeasons(clientApiFetcher)
      .then((seasons) => {
        if (!cancelled) setSeasons(getSelectableSeasons(seasons))
      })
      .catch(console.error)
    return () => {
      cancelled = true
    }
  }, [])
  // The season comes from the `s` query param and falls back to the primary one.
  const [querySeasonId, setQuerySeasonId] = useQueryState('s', parseAsString)
  const season = pickSeason(seasons, querySeasonId)
  const accessMode = getAccessMode(season)
  const freeAccess = isFreeAccess(accessMode)
  const lastStep = lastSteps[accessMode]

  const [queryUsernames] = useQueryState('u', parseAsString.withDefault(''))
  const session = useAuthStore((state) => state.session)
  const router = useRouter()
  const { refresh } = useBasket()

  const seasonId = season?.id
  // The last checked nickname that belongs to a Mojang account.
  const [premiumUsername, setPremiumUsername] = React.useState<string | null>(
    null,
  )
  const form = useForm<ObtainAccessFormValues>({
    // Validate while typing, so a premium nickname is flagged before submit.
    mode: 'onChange',
    // react-hook-form reads the latest resolver on every validation.
    resolver: zodResolver(
      createObtainAccessFormSchema(seasonId, (checked, premium) =>
        setPremiumUsername(premium ? checked : null),
      ),
    ),
    defaultValues: {
      username: queryUsernames,
      premiumAcknowledged: false,
    },
  })

  const username = useWatch({ control: form.control, name: 'username' })
  const premium =
    !!premiumUsername &&
    premiumUsername.toLowerCase() === username?.toLowerCase()

  // The acknowledgement is about one nickname: another one asks again.
  React.useEffect(() => {
    form.setValue('premiumAcknowledged', false)
  }, [premiumUsername, form])

  // The same username can have access to one season and not to another.
  React.useEffect(() => {
    if (seasonId && form.getValues('username')) {
      void form.trigger('username')
    }
  }, [seasonId, form])

  const [isSubmitting, startObtainingAccess] = React.useTransition()

  const onSubmit = form.handleSubmit((data) => {
    if (!session?.active) {
      // Send the guest to sign in and bring them back with the username
      // prefilled through the `u` query param.
      const returnParams = new URLSearchParams({ u: data.username })
      if (season) returnParams.set('s', season.id)
      const returnTo = `${process.env.NEXT_PUBLIC_DOMAIN}/${locale}/obtain-access?${returnParams}`
      const signInParams = new URLSearchParams({ return_to: returnTo })
      // External Ory URL, not an internal Next.js page.
      // eslint-disable-next-line @next/next/no-location-assign-relative-destination
      window.location.assign(
        `${process.env.NEXT_PUBLIC_ORY_SDK_URL}/self-service/login/browser?${signInParams}`,
      )
      return
    }

    startObtainingAccess(async () => {
      try {
        if (!season) return
        if (freeAccess) {
          if (accessMode === 'preregistration') {
            await requestFreeAccess(clientApiFetcher, season.id, [
              data.username,
            ])
            toast.success(t('successPre'))
          } else {
            await registerForPrimarySeason(clientApiFetcher, season.id, [
              data.username,
            ])
            toast.success(t('successFree'))
          }
          router.push(`/${locale}/profiles`)
          return
        }

        await requestAccess(clientApiFetcher, season.id, [data.username])
        toast.success(t('success'))
        await refresh()
        router.push(`/${locale}/basket`)
      } catch (error) {
        console.error(error)
        errorToast(t('error'), error)
      }
    })
  })

  return (
    <div className="flex flex-col items-center justify-center gap-6">
      <h1 className="flex items-center gap-2 text-4xl font-extrabold tracking-tight">
        <UserCheckIcon className="size-8" />
        {t('title')}
      </h1>

      <Form {...form}>
        <form onSubmit={onSubmit} className="w-full px-6">
          <ol className="border-primary-foreground relative border-s">
            {!session?.active && (
              <li className="ms-6 mb-10">
                <span className="bg-primary-foreground absolute -start-3 flex h-6 w-6 items-center justify-center rounded-full ring-8 ring-teal-900/30">
                  <CheckIcon className="size-4 text-green-500" />
                </span>
                <h3 className="mb-1 flex items-center text-lg font-semibold text-gray-900 dark:text-white">
                  {t('steps.step1.title')}
                </h3>
                <p className="mb-4 text-base font-normal text-gray-500 dark:text-gray-400">
                  {t('steps.step1.description')}
                </p>
                <SignInButton username={username} />
              </li>
            )}
            {/** season: only asked when more than one season is running */}
            {seasons.length > 1 && (
              <li className="ms-6 mb-10">
                <span className="bg-primary-foreground absolute -start-3 flex h-6 w-6 items-center justify-center rounded-full ring-8 ring-teal-900/30">
                  <CalendarIcon className="size-4" />
                </span>
                <h3 className="mb-1 text-lg font-semibold text-gray-900 dark:text-white">
                  {t('steps.stepSeason.title')}
                </h3>
                <p className="mb-4 text-base font-normal text-gray-500 dark:text-gray-400">
                  {t('steps.stepSeason.description')}
                </p>
                <SeasonSelect
                  seasons={seasons}
                  selectedSeasonId={season?.id}
                  onSelectSeasonId={(id) => void setQuerySeasonId(id)}
                  className="w-full"
                />
              </li>
            )}
            {/** step 2: specify username */}
            <li className="ms-6 mb-10">
              <span className="bg-primary-foreground absolute -start-3 flex h-6 w-6 items-center justify-center rounded-full ring-8 ring-teal-900/30">
                <Gamepad2Icon className="size-4" />
              </span>
              <h3 className="mb-1 text-lg font-semibold text-gray-900 dark:text-white">
                {t('steps.step2.title')}
              </h3>
              <p className="mb-4 text-base font-normal text-gray-500 dark:text-gray-400">
                {t('steps.step2.description')}
              </p>
              <UsernameField form={form} />
              {premium && <PremiumNotice form={form} username={username} />}
            </li>
            {/** step 3: create the profile and either grant access right away
             * (free registration) or add the season pass to the basket */}
            <li className="ms-6 mb-10">
              <span className="bg-primary-foreground absolute -start-3 flex h-6 w-6 items-center justify-center rounded-full ring-8 ring-teal-900/30">
                {freeAccess ? (
                  <ZapIcon className="size-4" />
                ) : (
                  <CreditCardIcon className="size-4" />
                )}
              </span>
              <h3 className="mb-1 text-lg font-semibold text-gray-900 dark:text-white">
                {t(`${lastStep}.title`)}
              </h3>
              <p className="mb-6 text-base font-normal text-gray-500 dark:text-gray-400">
                <RichText>
                  {(tags) => t.rich(`${lastStep}.description`, { ...tags })}
                </RichText>
              </p>
              <div className="flex flex-col gap-4">
                <Table className="hidden">
                  <TableHeader>
                    <TableRow>
                      <TableHead>Товар</TableHead>
                      <TableHead>Кол-во</TableHead>
                      <TableHead>Сумма</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {/* {form.watch('username')?.map((name, idx) => {
                      const { error } = form.getFieldState(
                        `username` as const,
                        form.formState,
                      )
                      if (!name || error) return null
                      return (
                        <TableRow key={form.watch('username')}>
                          <TableCell>
                            <p className="font-medium">
                              Доступ на
                              <span className="font-semibold">
                                &nbsp;4 сезон&nbsp;
                              </span>
                              сервера
                              <span className="font-semibold">
                                &nbsp;Lania&nbsp;
                              </span>
                              для игрока
                              <span className="font-bold">
                                &nbsp;&quot;{name}&quot;
                              </span>
                            </p>
                          </TableCell>
                          <TableCell>1</TableCell>
                          <TableCell>{PRICE} ₽</TableCell>
                        </TableRow>
                      )
                    })} */}
                  </TableBody>
                  <TableFooter>
                    <TableRow>
                      <TableCell colSpan={3} className="text-left">
                        Итого:
                        <span className="font-bold">&nbsp;100 ₽</span>
                      </TableCell>
                    </TableRow>
                  </TableFooter>
                </Table>
                <Button
                  className="w-full"
                  type="submit"
                  disabled={isSubmitting || !season}
                >
                  {isSubmitting ? (
                    <LoadingSpinner className="size-4" />
                  ) : (
                    <ArrowRightIcon className="size-4" />
                  )}
                  {/* <p>
                    Оплатить&nbsp;
                    <span className="font-bold">{totalPrice} ₽</span>
                  </p> */}
                  <p>{t(`${lastStep}.action`)}</p>
                </Button>
              </div>
            </li>
          </ol>
        </form>
      </Form>
    </div>
  )
}
