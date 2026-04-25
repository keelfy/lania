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
import { requestAccess } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { useAuthStore } from '@/providers/auth-store'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  ArrowRightIcon,
  CheckIcon,
  CreditCardIcon,
  Gamepad2Icon,
  UserCheckIcon,
} from 'lucide-react'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import { parseAsString, useQueryState } from 'nuqs'
import React from 'react'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import z from 'zod'
import SignInButton from '../../components/sign-in-button'
import { obtainAccessFormSchema } from './form'
import { UsernameField } from './username-field'
import { useBasket } from '@/context/basket'

type Props = {
  params: Promise<{
    locale: string
  }>
}

export default function ObtainAccessPage({ params }: Props) {
  const { locale } = React.use(params)
  const t = useTranslations('obtainAccess')

  const [queryUsernames] = useQueryState('u', parseAsString.withDefault(''))
  const session = useAuthStore((state) => state.session)
  const router = useRouter()
  const { refresh } = useBasket()

  const form = useForm<z.infer<typeof obtainAccessFormSchema>>({
    resolver: zodResolver(obtainAccessFormSchema),
    defaultValues: {
      username: queryUsernames,
    },
  })

  const [isSubmitting, startObtainingAccess] = React.useTransition()

  const onSubmit = form.handleSubmit((data) => {
    startObtainingAccess(async () => {
      try {
        await requestAccess(clientApiFetcher, [data.username])
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
                <SignInButton />
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
            </li>
            {/** step 3: create profile, add access to the basket and redirect to the basket  */}
            <li className="ms-6 mb-10">
              <span className="bg-primary-foreground absolute -start-3 flex h-6 w-6 items-center justify-center rounded-full ring-8 ring-teal-900/30">
                <CreditCardIcon className="size-4" />
              </span>
              <h3 className="mb-1 text-lg font-semibold text-gray-900 dark:text-white">
                {t('steps.step3.title')}
              </h3>
              <p className="mb-6 text-base font-normal text-gray-500 dark:text-gray-400">
                <RichText>
                  {(tags) => t.rich('steps.step3.description', { ...tags })}
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
                  // disabled={isSubmitting || !session?.active}
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
                  <p>{t('steps.step3.action')}</p>
                </Button>
              </div>
            </li>
          </ol>
        </form>
      </Form>
    </div>
  )
}
