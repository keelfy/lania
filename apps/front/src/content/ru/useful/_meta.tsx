import { ClockIcon, PackageIcon, UsersIcon } from 'lucide-react'
import { MetaRecord } from 'nextra'

const meta: MetaRecord = {
  restarts: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <ClockIcon size={14} />
        <p>Перезагрузки сервера</p>
      </div>
    ),
  },
  modpack: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <PackageIcon size={14} />
        <p>Модпак от keelfy</p>
      </div>
    ),
  },
  staff: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <UsersIcon size={14} />
        <p>Команда сервера</p>
      </div>
    ),
  },
}

export default meta
