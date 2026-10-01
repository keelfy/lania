import { ArrowRightIcon, CommandIcon, TextIcon } from 'lucide-react'
import { MetaRecord } from 'nextra'

const meta: MetaRecord = {
  index: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <ArrowRightIcon size={14} />
        <p>Введение</p>
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
  gameplay: { display: 'children' },
  useful: { display: 'children' },
}

export default meta
