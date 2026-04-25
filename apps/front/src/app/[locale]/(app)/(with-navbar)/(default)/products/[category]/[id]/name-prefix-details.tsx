import McUsername from '@/components/ui/mc-username'
import { cn } from '@/lib/utils'
import { NamePrefixProductMetadata, Product } from '@/models/product'
import Image from 'next/image'

type Props = React.ComponentProps<'div'> & {
  item: Product<NamePrefixProductMetadata>
}

export default function NamePrefixProductDetails({
  item,
  className,
  ...props
}: Props) {
  return (
    <div className={cn('flex flex-col', className)} {...props}>
      <div className="flex items-center gap-4">
        <Image
          src={item.metadata.prefix}
          alt={item.metadata.prefix}
          width={40}
          height={40}
          unoptimized
          className="-translate-y-1"
        />
        <McUsername
          username={item.name}
          className="scroll-m-20 text-center text-6xl font-bold"
        />
      </div>
      <div className="mt-2">
        <p>
          Префикс для вашего никнейма в игре. Имя вашего персонажа с этим
          префиксом будет отображаться:
        </p>
        <ul className="mt-5 ml-6 list-disc [&>li]:mt-1">
          <li>В списке игроков</li>
          <li>В чате</li>
          <li>В профиле</li>
          <li>Над головой игрока</li>
          <li>и больше...</li>
        </ul>
      </div>
    </div>
  )
}
