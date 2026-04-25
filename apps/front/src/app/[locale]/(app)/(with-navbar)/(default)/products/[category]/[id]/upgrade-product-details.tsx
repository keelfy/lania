import { cn } from '@/lib/utils'
import { Product, UpgradeProductMetadata } from '@/models/product'
import { UserCheckIcon } from 'lucide-react'
import Link from 'next/link'

type Props = React.ComponentProps<'div'> & {
  item: Product<UpgradeProductMetadata>
}

const ICONS = {
  season_access: UserCheckIcon,
}

const DESCRIPTIONS = {
  season_access: (
    <>
      <p>Доступ к серверу одному игроку до конца текущего сезона.</p>
      <p className="mt-2">
        Вы можете иметь до 2 активных проходок на сервере единовременно.
      </p>
      <p className="text-muted-foreground mt-4">
        Получая доступ к серверу, вы соглашаетесь с&nbsp;
        <Link
          href="/wiki/rules"
          className="font-medium underline underline-offset-2"
        >
          игровыми правилами
        </Link>
        ,&nbsp;
        <Link
          href="/wiki/legal/tos"
          className="font-medium underline underline-offset-2"
        >
          пользовательским соглашением
        </Link>
        ,&nbsp;
        <Link
          href="/wiki/legal/privacy"
          className="font-medium underline underline-offset-2"
        >
          политикой конфиденциальности
        </Link>
        &nbsp;и&nbsp;
        <Link
          href="/wiki/legal/refund"
          className="font-medium underline underline-offset-2"
        >
          политикой возврата денежных средств
        </Link>
        &nbsp;и&nbsp;
        <Link
          href="/wiki/legal/payments"
          className="font-medium underline underline-offset-2"
        >
          условиями оплаты
        </Link>
        &nbsp;и перед покупкой.
      </p>
    </>
  ),
}

export default function NameColorProductDetails({
  item,
  className,
  ...props
}: Props) {
  const Icon = ICONS[item.metadata.action as keyof typeof ICONS]
  return (
    <div className={cn('flex flex-col', className)} {...props}>
      <div className="flex items-center gap-2">
        {Icon && <Icon className="size-12 stroke-2" />}
        <p className="scroll-m-20 text-6xl font-extrabold">{item.name}</p>
      </div>
      <div className="mt-4">
        {DESCRIPTIONS[item.metadata.action as keyof typeof DESCRIPTIONS]}
      </div>
    </div>
  )
}
