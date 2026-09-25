import { cn } from '@/lib/utils'

// Marks a control that leads to something the user is expected to do. The parent must be positioned.
export default function AttentionDot({ className }: { className?: string }) {
  return (
    <span
      aria-hidden
      className={cn(
        'ring-background pointer-events-none absolute -top-0.5 -right-0.5 size-2.5 rounded-full bg-red-500 ring-2',
        className,
      )}
    />
  )
}
