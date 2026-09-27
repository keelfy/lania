import { ALargeSmallIcon, FileIcon, ShieldIcon } from 'lucide-react'
import { MetaRecord } from 'nextra'

const meta: MetaRecord = {
  terminology: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <ALargeSmallIcon size={14} />
        <p>Terminology</p>
      </div>
    ),
  },
  rules: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <ShieldIcon size={14} />
        <p>Game rules</p>
      </div>
    ),
  },
  legal: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <FileIcon size={14} />
        <p>Legal information</p>
      </div>
    ),
  },
}

export default meta
