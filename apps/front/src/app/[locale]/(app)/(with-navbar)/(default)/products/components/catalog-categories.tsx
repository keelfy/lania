import type { CSSProperties } from 'react'
import { LayoutGridIcon, PaletteIcon, ShieldIcon, TagIcon } from 'lucide-react'
import { Link } from '@/i18n/navigation'
import { cn } from '@/lib/utils'
import styles from './catalog-categories.module.css'

const CATEGORIES = [
  { id: 'all', icon: LayoutGridIcon },
  { id: 'upgrade', icon: ShieldIcon },
  { id: 'name-color', icon: PaletteIcon },
  { id: 'name-prefix', icon: TagIcon },
]

export default function CatalogCategories({
  category,
  labels,
  counts,
  previewOnly,
  label,
}: {
  category: string
  labels: Record<string, string>
  counts: Record<string, number>
  previewOnly: boolean
  label: string
}) {
  const activeIndex = Math.max(
    0,
    CATEGORIES.findIndex((item) => item.id === category),
  )

  return (
    <nav
      aria-label={label}
      className="relative grid grid-cols-2 gap-2 sm:grid-cols-1"
      style={
        {
          '--selected': activeIndex,
          '--mobile-col': activeIndex % 2,
          '--mobile-row': Math.floor(activeIndex / 2),
        } as CSSProperties
      }
    >
      <span aria-hidden="true" className={styles.marker} />
      {CATEGORIES.map(({ id, icon: Icon }) => (
        <Link
          key={id}
          href={{
            pathname: '/products',
            query: {
              ...(id !== 'all' && { category: id }),
              ...(previewOnly && { mock: '1' }),
            },
          }}
          scroll={false}
          aria-current={category === id ? 'page' : undefined}
          className={cn(
            'focus-visible:ring-ring hover:bg-primary/5 relative flex h-16 items-center gap-2 rounded-md border border-transparent px-3 text-sm transition-colors outline-none focus-visible:ring-2 motion-reduce:transition-none sm:text-base',
            category === id
              ? 'text-primary-foreground'
              : 'text-muted-foreground hover:text-foreground',
          )}
        >
          <Icon className="size-4 shrink-0" />
          <span className="font-semibold">{labels[id]}</span>
          <span className="ml-auto text-xs tabular-nums opacity-70">
            {counts[id] ?? 0}
          </span>
        </Link>
      ))}
    </nav>
  )
}
