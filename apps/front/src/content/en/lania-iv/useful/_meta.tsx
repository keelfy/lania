import { ClockIcon, PackageIcon, UsersIcon } from 'lucide-react'
import { MetaRecord } from 'nextra'

const meta: MetaRecord = {
  client: { type: 'separator', title: 'Client setup' },
  modpack: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <PackageIcon size={14} />
        <p>keelfy&apos;s modpack</p>
      </div>
    ),
  },
  server: { type: 'separator', title: 'Server information' },
  restarts: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <ClockIcon size={14} />
        <p>Server restarts</p>
      </div>
    ),
  },
  staff: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <UsersIcon size={14} />
        <p>Server staff</p>
      </div>
    ),
  },
}

export default meta
