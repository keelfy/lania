import {
  ClockIcon,
  HistoryIcon,
  MicIcon,
  PackageIcon,
  ShirtIcon,
  UsersIcon,
} from 'lucide-react'
import { MetaRecord } from 'nextra'

const meta: MetaRecord = {
  client: { type: 'separator', title: 'Настройка клиента' },
  modpack: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <PackageIcon size={14} />
        <p>Модпак от keelfy</p>
      </div>
    ),
  },
  voice: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <MicIcon size={14} />
        <p>Голосовой чат</p>
      </div>
    ),
  },
  skins: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <ShirtIcon size={14} />
        <p>Скины</p>
      </div>
    ),
  },
  server: { type: 'separator', title: 'О сервере' },
  coreprotect: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <HistoryIcon size={14} />
        <p>История блоков</p>
      </div>
    ),
  },
  restarts: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <ClockIcon size={14} />
        <p>Перезагрузки сервера</p>
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
