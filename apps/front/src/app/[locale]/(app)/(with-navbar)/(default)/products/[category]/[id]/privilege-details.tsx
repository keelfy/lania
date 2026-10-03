import { cn } from '@/lib/utils'
import { PrivilegeProductMetadata, Product } from '@/models/product'
import { KeyRoundIcon } from 'lucide-react'

type Props = React.ComponentProps<'div'> & {
  item: Product<PrivilegeProductMetadata>
}

export default function PrivilegeProductDetails({
  item,
  className,
  ...props
}: Props) {
  return (
    <div className={cn('flex flex-col', className)} {...props}>
      <div className="flex items-center gap-3 text-4xl leading-tight sm:text-5xl lg:text-6xl">
        <KeyRoundIcon className="size-[1em] shrink-0 stroke-2" />
        <p className="scroll-m-20 font-extrabold text-balance">{item.name}</p>
      </div>
      <p className="mt-4 max-w-prose">{item.description}</p>
    </div>
  )
}
