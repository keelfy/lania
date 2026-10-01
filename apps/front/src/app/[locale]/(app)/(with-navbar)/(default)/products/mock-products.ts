import { Currency } from '@/lib/currency'
import {
  Product,
  ProductCategory,
  NameColorProductMetadata,
  NamePrefixProductMetadata,
  UpgradeProductMetadata,
} from '@/models/product'

type MockProduct = Product<
  NameColorProductMetadata | NamePrefixProductMetadata | UpgradeProductMetadata
>

// Demonstration prices only; these are not live exchange rates.
const PRICE_FACTOR: Record<Currency, number> = {
  RUB: 1,
  USD: 0.012,
  EUR: 0.011,
  BRL: 0.06,
  TRY: 0.45,
  PLN: 0.045,
}

export function getMockProducts(
  category: string | undefined,
  locale: string,
  currency: Currency,
): MockProduct[] {
  const ru = locale === 'ru'
  const old = new Date('2025-01-01T00:00:00Z')
  const recent = new Date()
  const colorDescription = ru
    ? 'Градиентный цвет вашего ника в чате и списке игроков.'
    : 'A gradient for your name in chat and the player list.'
  const prefixDescription = ru
    ? 'Иконка перед вашим ником. Можно сочетать с цветом имени.'
    : 'An icon before your name. Combine it with a name color.'
  const products: MockProduct[] = [
    {
      id: 'preview-season-access',
      name: ru ? 'Проходка' : 'Season access',
      description: ru
        ? 'Доступ к серверу на текущий сезон. Исследуйте мир и стройте вместе с сообществом.'
        : 'Join the current season. Explore the world and build with the community.',
      price: 49,
      category: ProductCategory.Upgrade,
      metadata: { action: 'season_access' },
      soldCount: 128,
      createdAt: old,
    },
    ...[
      {
        name: ru ? 'Сияние далёких галактик' : 'Glow of distant galaxies',
        colors: ['#4343a8', '#b35cff'],
      },
      { name: 'Frozen', colors: ['#57bbff', '#d2fff9'] },
      { name: 'Kyoto', colors: ['#ff776e', '#ffcf87'] },
      { name: 'Magic', colors: ['#ef78ff', '#877dff'] },
      { name: 'Northern Lights', colors: ['#54e5ac', '#67bdff', '#bc7dff'] },
      { name: 'CocoaaIce', colors: ['#be8b70', '#f1e4d0'] },
    ].map(
      ({ name, colors }, index): MockProduct => ({
        id: `preview-color-${index}`,
        name,
        description: colorDescription,
        price: 199,
        category: ProductCategory.NameColor,
        metadata: { colors, nameColorId: `preview-name-color-${index}` },
        soldCount: 32 + index * 7,
        createdAt: index === 3 ? recent : old,
      }),
    ),
    ...[
      {
        name: ru ? 'Рюкзак путешественника' : 'Traveler’s backpack',
        image: 'crystal',
      },
      {
        name: 'FoxFace',
        image: 'fox',
      },
      {
        name: 'Popcat',
        image: 'cat',
      },
    ].map(
      ({ name, image }, index): MockProduct => ({
        id: `preview-prefix-${index}`,
        name,
        description: prefixDescription,
        price: 99,
        category: ProductCategory.NamePrefix,
        metadata: {
          prefix: `/images/mock-products/${image}.svg`,
          noSpace: index === 2,
          namePrefixId: `preview-name-prefix-${index}`,
        },
        soldCount: 18 + index * 5,
        createdAt: index === 2 ? recent : old,
      }),
    ),
  ]

  return products
    .filter(
      (product) =>
        !category || category === 'all' || product.category === category,
    )
    .map((product) => ({
      ...product,
      price:
        Math.round(product.price * (PRICE_FACTOR[currency] ?? 1) * 100) / 100,
    }))
}
