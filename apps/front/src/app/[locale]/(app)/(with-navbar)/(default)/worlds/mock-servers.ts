import type { Season, SeasonWorld } from '@/models/season'
import type { Data } from 'minecraft-pinger'

// Visual fixtures, enabled only by ?mock=1 in development.
const season: Season = {
  id: 'preview-primary',
  name: 'Lania IV',
  startDate: 1756684800000,
  publicAddress: 'play.lania.example',
  isActive: true,
  isPrimary: true,
  onlineAvailable: true,
  preregistration: false,
  freeRegistration: true,
  gameVersion: '1.21.1',
}

function world(
  id: string,
  name: string,
  image: string,
  claims = true,
): SeasonWorld {
  return {
    id,
    seasonId: season.id,
    slug: id,
    name,
    previewImage: `s3://lania-web-134312503254-eu-central-1-an/${image}`,
    mapUrl: 'https://map.lania.example',
    claimLimit: claims ? 100 : 0,
    claimDimensions: claims ? ['world'] : [],
    hiddenDimensions: [],
    claimMinPlaytimeHours: 0,
    position: 0,
  }
}

export const mockServers: {
  server: Season
  worlds: SeasonWorld[]
  status: Data | undefined
}[] = [
  {
    server: season,
    status: {
      description: {
        text: '§bLania IV §f— новый сезон\n§7Строим историю вместе.',
      },
      players: { online: 42, max: 100 },
      version: { name: '1.21.1', protocol: 767 },
      ping: 28,
    },
    worlds: [
      world('survival', 'Выживание', '2025-04-17_16.59.00_2.png'),
      world('farms', 'Фермы', '2025-08-29_21.30.10.png'),
    ],
  },
  {
    server: {
      ...season,
      id: 'preview-secondary',
      name: 'Lania III',
      publicAddress: 'classic.lania.example',
      isPrimary: false,
    },
    status: {
      description: { text: '§bLania III\n§7Мир, в который можно вернуться.' },
      players: { online: 7, max: 50 },
      version: { name: '1.21.1', protocol: 767 },
      ping: 35,
    },
    worlds: [
      {
        ...world('classic', 'Основной мир', '2025-08-19_11.07.36.png'),
        seasonId: 'preview-secondary',
      },
    ],
  },
]
