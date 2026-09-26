import McUsername from '@/components/ui/mc-username'
import { cn } from '@/lib/utils'
import {
  NameColorProductMetadata,
  NamePrefixProductMetadata,
  Product,
  ProductCategory,
} from '@/models/product'
import { useTranslations } from 'next-intl'
import NameTryOn from './name-try-on'

type Props = React.ComponentProps<'div'> & {
  item: Product<NameColorProductMetadata>
  // Glyphs on sale, to try the color with.
  pairings: Product<NamePrefixProductMetadata>[]
}

export default function NameColorProductDetails({
  item,
  pairings,
  className,
  ...props
}: Props) {
  const t = useTranslations('products.page')

  return (
    <div className={cn('flex flex-col gap-8', className)} {...props}>
      <div className="flex flex-col gap-3">
        <McUsername
          username={item.name}
          colors={item.metadata.colors}
          className="scroll-m-20 text-4xl leading-tight font-bold text-balance sm:text-5xl lg:text-6xl"
        />
        <p className="max-w-prose">{t('nameColor.description')}</p>
      </div>
      <NameTryOn
        category={ProductCategory.NameColor}
        item={item}
        pairings={pairings}
      />
    </div>
  )
}
