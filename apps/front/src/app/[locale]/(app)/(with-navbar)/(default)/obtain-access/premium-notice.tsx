import { Checkbox } from '@/components/ui/checkbox'
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormTranslatedMessage,
} from '@/components/ui/form'
import { KeyRoundIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { UseFormReturn } from 'react-hook-form'
import { ObtainAccessFormValues } from './form'

type Props = {
  form: UseFormReturn<ObtainAccessFormValues>
  username: string
}

// A premium nickname only lets its Mojang account in: the player confirms they
// own it before getting access.
export function PremiumNotice({ form, username }: Props) {
  const t = useTranslations('obtainAccess.steps.step2.premium')
  const tMessages = useTranslations('obtainAccess')
  const bold = (chunks: React.ReactNode) => (
    <b className="text-foreground font-semibold">{chunks}</b>
  )

  return (
    <FormField
      control={form.control}
      name="premiumAcknowledged"
      render={({ field }) => (
        <FormItem className="gap-3 rounded-md border border-yellow-500/40 bg-yellow-500/5 p-4">
          <div className="flex gap-3">
            <KeyRoundIcon className="mt-0.5 size-4 shrink-0 text-yellow-500" />
            <div className="flex flex-col gap-1">
              <p className="text-sm font-semibold">{t('title')}</p>
              <p className="text-muted-foreground text-sm">
                {t.rich('description', { username, b: bold })}
              </p>
            </div>
          </div>
          <div className="flex items-start gap-3 ps-7">
            <FormControl>
              <Checkbox
                className="mt-0.5"
                checked={field.value}
                onCheckedChange={(checked) => field.onChange(checked === true)}
              />
            </FormControl>
            <FormLabel className="text-sm leading-snug font-normal">
              {t('acknowledge', { username })}
            </FormLabel>
          </div>
          <FormTranslatedMessage className="ps-7" getMessage={tMessages} />
        </FormItem>
      )}
    />
  )
}
