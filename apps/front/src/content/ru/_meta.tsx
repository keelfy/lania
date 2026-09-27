import {
  ALargeSmallIcon,
  CalendarIcon,
  FileIcon,
  MapIcon,
  ShoppingBagIcon,
  ShieldIcon,
} from 'lucide-react'
import { MetaRecord } from 'nextra'

const meta: MetaRecord = {
  worlds: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <MapIcon size={14} />
        <p>Миры</p>
      </div>
    ),
    type: 'page',
    href: '/worlds',
  },
  shop: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <ShoppingBagIcon size={14} />
        <p>Магазин</p>
      </div>
    ),
    type: 'page',
    href: '/products',
  },
  seasons: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <CalendarIcon size={14} />
        <p>Сезоны</p>
      </div>
    ),
    type: 'page',
    href: '/seasons',
  },
  terminology: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <ALargeSmallIcon size={14} />
        <p>Терминология</p>
      </div>
    ),
  },
  rules: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <ShieldIcon size={14} />
        <p>Игровые правила</p>
      </div>
    ),
  },
  legal: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <FileIcon size={14} />
        <p>Правовая информация</p>
      </div>
    ),
  },
}

export default meta
