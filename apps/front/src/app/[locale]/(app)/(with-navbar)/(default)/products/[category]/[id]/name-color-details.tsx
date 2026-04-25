import McUsername from '@/components/ui/mc-username'
import { NameColorProductMetadata, Product } from '@/models/product'
import NameColorTryForm from './name-color-try-form'
import { cn } from '@/lib/utils'

type Props = React.ComponentProps<'div'> & {
  item: Product<NameColorProductMetadata>
}

export default function NameColorProductDetails({
  item,
  className,
  ...props
}: Props) {
  return (
    <div className={cn('flex flex-col', className)} {...props}>
      <McUsername
        username={item.name}
        colors={item.metadata.colors}
        className="scroll-m-20 text-center text-6xl font-bold"
      />
      <div className="mt-2">
        <p>
          Градиентный цвет для вашего никнейма в игре. Имя вашего персонажа с
          этим цветом будет отображаться:
        </p>
        <ul className="mt-5 ml-6 list-disc [&>li]:mt-1">
          <li>В списке игроков</li>
          <li>В чате</li>
          <li>В профиле</li>
          <li>Над головой игрока</li>
          <li>и больше...</li>
        </ul>
      </div>
      <h2 className="mt-8 scroll-m-20 text-3xl font-semibold tracking-tight first:mt-0">
        Попробуйте имя вашего персонажа!
      </h2>
      <div className="mt-6">
        <NameColorTryForm
          defaultUsername={item.name ?? 'Steve'}
          colors={item.metadata.colors}
        />
      </div>
    </div>
  )
}
