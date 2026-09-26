import McUsername from '@/components/ui/mc-username'
import { cn } from '@/lib/utils'
import {
  NameColorProductMetadata,
  NamePrefixProductMetadata,
  Product,
  ProductCategory,
} from '@/models/product'
import { useTranslations } from 'next-intl'
import Image from 'next/image'
import NameTryOn from './name-try-on'

type Props = React.ComponentProps<'div'> & {
  item: Product<NamePrefixProductMetadata>
  // Name colors on sale, to try the glyph with.
  pairings: Product<NameColorProductMetadata>[]
}

export default function NamePrefixProductDetails({
  item,
  pairings,
  className,
  ...props
}: Props) {
  const t = useTranslations('products.page')

  return (
    <div className={cn('flex flex-col gap-8', className)} {...props}>
      <div className="flex flex-col gap-3">
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
        <p className="max-w-prose">{t('namePrefix.description')}</p>
      </div>
      <NameTryOn
        category={ProductCategory.NamePrefix}
        item={item}
        pairings={pairings}
      />
    </div>
  )
}
