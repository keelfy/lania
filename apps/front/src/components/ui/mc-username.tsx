import { cn } from '@/lib/utils'

type Props = React.ComponentProps<'p'> & {
  username: string | undefined
  colors?: string[] | string
}

export default function McUsername({
  username,
  colors,
  className,
  ...props
}: Props) {
  const colorsArray = colors ? (Array.isArray(colors) ? colors : [colors]) : []
  return (
    <p
      className={cn(
        'tracking-mc w-fit translate-y-0.5 font-[Minecraft] text-transparent drop-shadow-[0_1.2px_1.2px_rgba(255,255,255,0.2)]',
        colorsArray.length > 1
          ? 'bg-gradient-to-r bg-clip-text'
          : 'text-primary',
        className,
      )}
      style={{
        backgroundImage:
          colorsArray.length > 1
            ? `linear-gradient(to right, ${colorsArray.join(', ')})`
            : undefined,
        color: colorsArray.length === 1 ? colorsArray[0] : undefined,
      }}
      {...props}
    >
      {username ?? ''}
    </p>
  )
}
