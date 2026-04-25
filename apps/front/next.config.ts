import type { NextConfig } from 'next'
import nextra from 'nextra'
import createNextIntlPlugin from 'next-intl/plugin'

const nextConfig: NextConfig = {
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'crafatar.com',
      },
      {
        protocol: 'https',
        hostname: 'czx1jtlf2o.ufs.sh',
      },
    ],
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
