import DeerIcon from '@/components/icons/DeerIcon'
import { LOCALE_NAMES, LOCALES } from '@/lib/locale'
import { HeartIcon } from 'lucide-react'
import { Metadata } from 'next'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'
import { Footer, Layout, Navbar } from 'nextra-theme-docs'
import 'nextra-theme-docs/style.css'
import { Button, Head } from 'nextra/components'
import { getPageMap } from 'nextra/page-map'
import { footerLegalLinkd, footerSocialLinks } from '../(app)/components/footer'

type Props = {
  params: Promise<{
    locale: string
  }>
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'wiki' })

  return {
    title: t('metadata.title'),
    description: t('metadata.description'),
    openGraph: {
      type: 'website',
      url: 'https://lania.gg/wiki',
      title: t('metadata.title'),
      description: t('metadata.description'),
      siteName: 'Lania.GG',
    },
  }
}

// const banner = <Banner storageKey="some-key">Nextra 4.0 is released 🎉</Banner>
const navbar = (
  <Navbar
    logo={
      <>
        <DeerIcon style={{ width: '2rem', height: '2rem' }} />
        <p
          style={{
            display: 'flex',
            alignItems: 'center',
            textTransform: 'uppercase',
            letterSpacing: '0.1rem',
            marginLeft: '.4em',
            fontWeight: 800,
            fontSize: '1rem',
          }}
        >
          Lania
        </p>
      </>
    }
  />
)

const footer = (
  <Footer>
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'start',
        gap: '1rem',
        width: '100%',
      }}
    >
      <div
        style={{
          display: 'flex',
          width: '100%',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: '0.5rem',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
            paddingLeft: '0.25rem',
          }}
        >
          <DeerIcon style={{ width: '2rem', height: '2rem' }} />
          <p
            style={{
              fontSize: '1rem',
              fontWeight: 600,
              fontFamily: 'monospace',
            }}
          >
            LANIA.GG © {new Date().getFullYear()}
          </p>
        </div>
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
          }}
        >
          {footerSocialLinks.map((link) => (
            <Button
              key={link.href}
              style={{
                padding: '0.25rem',
              }}
            >
              <Link href={link.href}>
                <link.icon
                  style={{ width: '1.25rem', height: '1.25rem' }}
                  color={link.color}
                />
              </Link>
            </Button>
          ))}
        </div>
      </div>
      <div
        style={{
          display: 'flex',
          flexWrap: 'wrap',
          alignItems: 'center',
          gap: '0.5rem',
        }}
      >
        {footerLegalLinkd.map((link) => (
          <Button
            key={link.href}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              padding: '0.25rem',
            }}
          >
            <Link
              href={link.href}
              style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}
            >
              <link.icon style={{ width: '1rem', height: '1rem' }} />
              <p>{link.label}</p>
            </Link>
          </Button>
        ))}
      </div>
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          width: '100%',
          gap: '0.5rem',
        }}
      >
        <p
          style={{
            color: 'var(--muted-foreground)',
            paddingLeft: '0.25rem',
            fontSize: '0.875rem',
          }}
        >
          Not an official Minecraft product. We are in no way affiliated with or
          endorsed by Mojang Synergies AB, Microsoft Corporation or other
          rightsholders.
        </p>
        <p
          style={{
            color: 'var(--muted-foreground)',
            fontSize: '0.875rem',
          }}
        >
          made by&nbsp;
          <a
            href="https://twitch.tv/keelfy"
            target="_blank"
            rel="noopener noreferrer"
            style={{
              textDecoration: 'underline',
            }}
          >
            кифли
          </a>
          &nbsp;
          <HeartIcon
            style={{ width: '1rem', height: '1rem', display: 'inline-block' }}
          />
        </p>
      </div>
    </div>
  </Footer>
)

export default async function WikiLayout({
  params,
  children,
}: React.PropsWithChildren<Props>) {
  const { locale } = await params
  const t = await getTranslations({ locale })

  const pageMap = (await getPageMap(`/${locale}/wiki`)).filter(
    (ele) => !('name' in ele && ele.name === '[locale]'),
  )

  return (
    <html
      // Not required, but good for SEO
      lang={locale}
      // Required to be set
      dir="ltr"
      // Suggested by `next-themes` package https://github.com/pacocoursey/next-themes#with-app
      suppressHydrationWarning
    >
      <Head>
        <title>{t('wiki.metadata.title')}</title>
        <meta name="description" content={t('wiki.metadata.description')} />
      </Head>
      <body>
        <Layout
          darkMode={false}
          i18n={LOCALES.map((l) => ({
            locale: l.toString(),
            name: LOCALE_NAMES[l],
          }))}
          toc={{
            backToTop: t('wiki.backToTop'),
          }}
          themeSwitch={{
            dark: t('common.theme.dark'),
            light: t('common.theme.light'),
            system: t('common.theme.system'),
          }}
          // banner={banner}
          navbar={navbar}
          pageMap={pageMap}
          docsRepositoryBase="https://github.com/shuding/nextra/tree/main/docs"
          footer={footer}
          // ... Your additional layout options
        >
          {children}
        </Layout>
      </body>
    </html>
  )
}
