type LiquidGlassCardProps = React.PropsWithChildren<React.ComponentProps<'div'>>

export default function LiquidGlassCard({
  children,
  ...props
}: LiquidGlassCardProps) {
  return (
    // static glass (more performant): bg-white/5 ring-1 ring-white/10
    <div
      className="px-6 py-4 backdrop-blur-xs sm:max-w-xl sm:items-start sm:rounded-2xl"
      {...props}
    >
      {children}
    </div>
  )
}
