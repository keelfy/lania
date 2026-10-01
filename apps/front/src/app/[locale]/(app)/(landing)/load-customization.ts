import config from '../../../../../public/landing-customization.json'
import { getProducts } from '@/lib/api-endpoints'
import type { ApiFetcher } from '@/lib/fetcher'
import {
  ProductCategory,
  type NameColorProductMetadata,
  type NamePrefixProductMetadata,
} from '@/models/product'

export async function loadCustomization(fetcher: ApiFetcher, locale: string) {
  const [colorProducts, glyphProducts] = await Promise.all([
    getProducts(fetcher, ProductCategory.NameColor, locale).catch(() => []),
    getProducts(fetcher, ProductCategory.NamePrefix, locale).catch(() => []),
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
