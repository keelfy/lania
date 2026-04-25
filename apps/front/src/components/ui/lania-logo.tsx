import { cn } from '@/lib/utils'
import DeerIcon from '../icons/DeerIcon'

type LaniaLogoProps = {
  className?: string
}

export default function LaniaLogo({ className }: LaniaLogoProps) {
  return (
    <div className={cn('flex items-center gap-1', className)}>
      <DeerIcon className="size-5" />
      <h1 className="text-md bg-gradient-to-r from-white to-teal-400/90 bg-clip-text font-bold tracking-widest text-transparent uppercase">
        lania
      </h1>
    </div>
  )
}
