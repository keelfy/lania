import { cn } from '@/lib/utils'
import { ReactNode } from 'react'

type Tag = 'p' | 'b' | 'i' | 'br'

type Props = {
  children(tags: Record<Tag, (chunks: ReactNode) => ReactNode>): ReactNode
  className?: string
}

export default function RichText({ children, className }: Props) {
  return (
    <span className={cn('prose', className)}>
      {children({
        p: (chunks: ReactNode) => <p>{chunks}</p>,
        b: (chunks: ReactNode) => <b className="font-semibold">{chunks}</b>,
        i: (chunks: ReactNode) => <i className="italic">{chunks}</i>,
        br: () => <br />,
      })}
    </span>
  )
}
