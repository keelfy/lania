import {
  ALargeSmallIcon,
  ArrowRightIcon,
  CalendarIcon,
  CommandIcon,
  FileIcon,
  GamepadIcon,
  MapIcon,
  ShoppingBagIcon,
  ShieldIcon,
  TextIcon,
  RocketIcon,
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
  index: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <ArrowRightIcon size={14} />
        <p>Введение</p>
      </div>
    ),
  },
  terminology: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <ALargeSmallIcon size={14} />
        <p>Терминология</p>
      </div>
    ),
  },
  commands: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <CommandIcon size={14} />
        <p>Команды</p>
      </div>
    ),
  },
  formatting: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <TextIcon size={14} />
        <p>Форматирование</p>
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
  gameplay: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <GamepadIcon size={14} />
        <p>Игровые механики</p>
      </div>
    ),
  },
  useful: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <RocketIcon size={14} />
        <p>Полезное</p>
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
