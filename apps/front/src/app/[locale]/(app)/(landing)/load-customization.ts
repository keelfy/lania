import config from '../../../../../public/landing-customization.json'
import { cacheLife } from 'next/cache'
import { getProducts } from '@/lib/api-endpoints'
import { publicApiFetcher } from '@/lib/public-api'
import {
  ProductCategory,
  type NameColorProductMetadata,
  type NamePrefixProductMetadata,
} from '@/models/product'

// The products are public and the landing page shows no prices, so every visitor shares one cached result.
// A failed read throws and is not cached, so the caller falls back.
export async function loadCustomization(locale: string) {
  'use cache'
  cacheLife('minutes')
  const [colorProducts, glyphProducts] = await Promise.all([
    getProducts(publicApiFetcher, ProductCategory.NameColor, locale),
    getProducts(publicApiFetcher, ProductCategory.NamePrefix, locale),
  ])
  const colors = colorProducts.map((product) => {
    const metadata = product.metadata as NameColorProductMetadata
    return {
      id: metadata.nameColorId,
      name: product.name,
      colors: metadata.colors,
    }
  })
  const glyphs = glyphProducts.map((product) => {
    const metadata = product.metadata as NamePrefixProductMetadata
    return {
      id: metadata.namePrefixId,
      name: product.name,
      image: metadata.prefix,
      noSpace: metadata.noSpace,
    }
  })

  return {
    colors: config.nameColorIds.flatMap((id) => {
      const color = colors.find((color) => color.id === id)
      return color && color.colors.length > 0 ? [{ ...color, id }] : []
    }),
    glyphs: config.glyphIds.flatMap((id) => {
      const glyph = glyphs.find((glyph) => glyph.id === id)
      return glyph?.image ? [glyph] : []
    }),
  }
}

export const noCustomization = { colors: [], glyphs: [] }
