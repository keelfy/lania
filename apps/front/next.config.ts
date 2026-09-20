import type { NextConfig } from 'next'
import nextra from 'nextra'
import createNextIntlPlugin from 'next-intl/plugin'

const nextConfig: NextConfig = {
  output: 'standalone',
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'crafatar-pub.neodium.fr',
      },
      {
        protocol: 'https',
        hostname: 'czx1jtlf2o.ufs.sh',
      },
      {
        protocol: 'https',
        hostname: 'images.lania.network',
      },
    ],
    loader: "custom",
    loaderFile: "./src/lib/imgproxyImageLoader.tsx",
  },
  /* config options here */
  i18n: {
    locales: ['ru', 'en'],
    defaultLocale: 'en',
  },
}

const withNextra = nextra({
  contentDirBasePath: '/wiki',
})

const withNextIntl = createNextIntlPlugin()
export default withNextIntl(withNextra(nextConfig))
