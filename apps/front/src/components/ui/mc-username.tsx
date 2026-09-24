import { cn } from '@/lib/utils'

type Props = React.ComponentProps<'p'> & {
  username: string | undefined
  colors?: string[] | string
}

// How a name wears its colours: one colour as is, several as a left-to-right gradient.
// The gradient only paints through transparent text clipped to the background.
export function usernameColorStyle(colors?: string[] | string) {
  const colorsArray = colors ? (Array.isArray(colors) ? colors : [colors]) : []
  return {
    isGradient: colorsArray.length > 1,
    style: {
      backgroundImage:
        colorsArray.length > 1
          ? `linear-gradient(to right, ${colorsArray.join(', ')})`
          : undefined,
      color: colorsArray.length === 1 ? colorsArray[0] : undefined,
    },
  }
}

export default function McUsername({
  username,
  colors,
  className,
  ...props
}: Props) {
  const { isGradient, style } = usernameColorStyle(colors)
  return (
    <p
      className={cn(
        'tracking-mc font-minecraft w-fit translate-y-0.5 text-transparent drop-shadow-[0_1.2px_1.2px_rgba(255,255,255,0.2)]',
        isGradient ? 'bg-gradient-to-r bg-clip-text' : 'text-primary',
        className,
      )}
      style={style}
      {...props}
    >
      {username ?? ''}
    </p>
  )
}
