export type WikiLang = 'ru' | 'en'

// Item icons live in public/images/wiki/items/<minecraft_type>.png (16x16).
export const WIKI_ITEMS = {
  glowstone_dust: { ru: 'Светокаменная пыль', en: 'Glowstone Dust' },
  stick: { ru: 'Палка', en: 'Stick' },
  purple_dye: { ru: 'Фиолетовый краситель', en: 'Purple Dye' },
  brick: { ru: 'Кирпич', en: 'Brick' },
  nether_star: { ru: 'Звезда Незера', en: 'Nether Star' },
  diamond: { ru: 'Алмаз', en: 'Diamond' },
  ender_pearl: { ru: 'Жемчуг Края', en: 'Ender Pearl' },
  glowstone: { ru: 'Светокамень', en: 'Glowstone' },
  debug_stick: { ru: 'Дебаг палочка', en: 'Debug Stick' },
  light: { ru: 'Блок света', en: 'Light Block' },
  barrier: { ru: 'Барьер', en: 'Barrier' },
  end_rod: { ru: 'Стержень Энда', en: 'End Rod' },
} as const

export type WikiItem = keyof typeof WIKI_ITEMS
