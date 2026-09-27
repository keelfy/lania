import {
  ArrowRightIcon,
  CommandIcon,
  GamepadIcon,
  TextIcon,
  RocketIcon,
} from 'lucide-react'
import { MetaRecord } from 'nextra'

const meta: MetaRecord = {
  index: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <ArrowRightIcon size={14} />
        <p>Introduction</p>
      </div>
    ),
  },
  commands: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <CommandIcon size={14} />
        <p>Commands</p>
      </div>
    ),
  },
  formatting: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <TextIcon size={14} />
        <p>Formatting</p>
      </div>
    ),
  },
  gameplay: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <GamepadIcon size={14} />
        <p>Gameplay</p>
      </div>
    ),
  },
  useful: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <RocketIcon size={14} />
        <p>Useful</p>
      </div>
    ),
  },
}

export default meta
