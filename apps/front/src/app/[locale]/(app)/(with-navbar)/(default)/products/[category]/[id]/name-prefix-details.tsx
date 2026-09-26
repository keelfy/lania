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
      {/* The icon is sized in lines so it keeps up with the title and sits on its first line. */}
      <div className="flex items-start gap-4 text-4xl leading-tight sm:text-5xl lg:text-6xl">
        <Image
          src={item.metadata.prefix}
          alt=""
          width={64}
          height={64}
          className="size-[1lh] shrink-0 p-[0.1lh] [image-rendering:pixelated]"
        />
        <McUsername
          username={item.name}
          className="scroll-m-20 font-bold text-balance"
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
