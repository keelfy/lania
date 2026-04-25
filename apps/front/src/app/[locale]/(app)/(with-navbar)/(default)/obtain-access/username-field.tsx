import {
  FormControl,
  FormField,
  FormItem,
  FormTranslatedMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import LoadingSpinner from '@/components/ui/loading-spinner'
import { CheckIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { UseFormReturn, useFormState } from 'react-hook-form'
import z from 'zod'
import { obtainAccessFormSchema } from './form'

type Props = {
  form: UseFormReturn<z.infer<typeof obtainAccessFormSchema>>
}

export function UsernameField({ form }: Props) {
  const { isValidating } = useFormState({
    control: form.control,
    name: `username`,
  })
  const fieldState = form.getFieldState(`username`, form.formState)
  const hasError = !!fieldState.error

  const t = useTranslations('obtainAccess')

  return (
    <FormField
      control={form.control}
      name={`username`}
      render={({ field }) => (
        <FormItem>
          <FormControl>
            <div className="mb-2 flex flex-col gap-2">
              <div className="flex h-9 w-full items-center gap-2">
                <div className="flex flex-1 items-center">
                  <div className="relative w-full flex-1">
                    <Input
                      placeholder="Steve"
                      className="h-9 w-full"
                      maxLength={16}
                      {...field}
                    />
                    {isValidating ? (
                      <div className="absolute top-1/2 right-4 flex -translate-y-1/2 items-center gap-2">
                        <LoadingSpinner className="size-4" />
                        <p className="text-muted-foreground text-xs">
                          {t('steps.step2.checking')}
                        </p>
                      </div>
                    ) : !hasError && field.value ? (
                      <div className="absolute top-1/2 right-4 flex -translate-y-1/2 items-center gap-2">
                        <CheckIcon className="size-4 text-green-500" />
                        <p className="text-xs">{t('steps.step2.available')}</p>
                      </div>
                    ) : null}
                  </div>
                </div>
              </div>
              <FormTranslatedMessage getMessage={t} />
            </div>
          </FormControl>
        </FormItem>
      )}
    />
  )
}
