import { footerSocialLinks } from '@/app/[locale]/(app)/components/footer'
import { Button } from '@/components/ui/button'
import RichText from '@/components/ui/rich-text'
import { getMetadataLocale } from '@/i18n/metadata-locale'
import { Link } from '@/i18n/navigation'
import { AccessMode, getAccessMode } from '@/lib/access-mode'
import { getPrimarySeason, getProducts } from '@/lib/api-endpoints'
import {
  Currency,
  CURRENCY_COOKIE,
  CURRENCY_SYMBOLS,
  DEFAULT_CURRENCY,
} from '@/lib/currency'
import { serverApiFetcher } from '@/lib/server'
import { ProductCategory, UpgradeProductMetadata } from '@/models/product'
import { ArrowRightIcon, BookOpenIcon, ShieldIcon } from 'lucide-react'
import { Metadata } from 'next'
import { getTranslations } from 'next-intl/server'
import { cookies } from 'next/headers'
import { cache } from 'react'
import { pingServer } from '../worlds/ping-server'
import ServerCard from '../worlds/server-card'

type Props = {
  params: Promise<{
    locale: string
  }>
}

// The price of the season pass in the currency, formatted; undefined when the pass is not on sale.
async function getPassPrice(locale: string, currency: Currency) {
  const upgrades = await getProducts(
    serverApiFetcher,
    ProductCategory.Upgrade,
    locale,
    currency,
  ).catch(() => [])
  const pass = upgrades.find(
    (product) =>
      (product.metadata as UpgradeProductMetadata).action === 'season_access',
  )
  return pass ? `${pass.price} ${CURRENCY_SYMBOLS[currency]}` : undefined
}

async function getCurrency() {
  return (
    ((await cookies()).get(CURRENCY_COOKIE)?.value as Currency) ??
    DEFAULT_CURRENCY
  )
}

// What a newcomer needs to know before joining: the primary season and how to get into it.
// Cached per request: the metadata and the page read the same.
const getWelcomeInfo = cache(async (locale: string) => {
  const [season, currency] = await Promise.all([
    getPrimarySeason(serverApiFetcher).catch(() => undefined),
    getCurrency(),
  ])
  const mode = getAccessMode(season)
  const [price, status] = await Promise.all([
    mode === 'paid' ? getPassPrice(locale, currency) : undefined,
    season?.publicAddress ? pingServer(season.publicAddress) : undefined,
  ])
  return { season, mode, price, status }
})

function accessKey(mode: AccessMode, price?: string) {
  if (mode === 'paid' && !price) return 'paidNoPrice'
  return mode
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const locale = await getMetadataLocale(params)
  const [t, { season, mode, price }] = await Promise.all([
    getTranslations({ locale, namespace: 'welcome.metadata' }),
    getWelcomeInfo(locale),
  ])
  // Link previews in chats show this text, so it carries the key facts.
  const description = t(`description.${accessKey(mode, price)}`, {
    version: season?.gameVersion ?? '—',
    price: price ?? '',
  })

  return {
    title: t('title'),
    description,
    openGraph: {
      type: 'website',
      url: process.env.NEXT_PUBLIC_DOMAIN + '/welcome',
      title: t('title'),
      description,
      siteName: 'Lania Network',
    },
  }
}

export default async function WelcomePage({ params }: Props) {
  const { locale } = await params
  const [t, { season, mode, price, status }] = await Promise.all([
    getTranslations({ locale, namespace: 'welcome' }),
    getWelcomeInfo(locale),
  ])
  const version = season?.gameVersion ?? '—'
  const joinable = !!season?.publicAddress && mode !== 'preregistration'

  const facts = [
    { key: 'version', value: version },
    { key: 'vanilla' },
    { key: 'license' },
    { key: `access.${accessKey(mode, price)}`, value: price },
    { key: 'shop' },
  ]

  return (
    <div className="mx-auto flex max-w-3xl flex-col gap-12 pb-8">
      <header className="flex flex-col gap-6">
        <div className="flex flex-col gap-3">
          <h1 className="text-4xl font-extrabold tracking-tight text-balance sm:text-5xl">
            {t('title')}
          </h1>
          <p className="text-muted-foreground text-lg text-pretty">
            {t('lead', { season: season?.name ?? 'Lania' })}
          </p>
        </div>
        {season && joinable && (
          <ServerCard
            locale={locale}
            server={season}
            worlds={[]}
            status={status}
          />
        )}
      </header>

      <section className="flex flex-col gap-4">
        <h2 className="text-2xl font-bold tracking-tight">
          {t('facts.title')}
        </h2>
        <dl className="divide-border divide-y border-y">
          {facts.map(({ key, value }) => (
            <div
              key={key}
              className="grid gap-1 py-4 sm:grid-cols-[12rem_1fr] sm:gap-6"
            >
              <dt className="text-muted-foreground text-sm sm:pt-0.5">
                {t(`facts.${key}.label`)}
              </dt>
              <dd className="flex flex-col gap-1">
                <span className="text-lg font-semibold">
                  {t(`facts.${key}.value`, { value: value ?? '' })}
                </span>
                <span className="text-muted-foreground text-sm">
                  <RichText>
                    {(tags) =>
                      t.rich(`facts.${key}.hint`, {
                        ...tags,
                        qol: (chunks) => (
                          <Link
                            href="/wiki/gameplay/qol"
                            className="underline underline-offset-2"
                          >
                            {chunks}
                          </Link>
                        ),
                      })
                    }
                  </RichText>
                </span>
              </dd>
            </div>
          ))}
        </dl>
      </section>

      <section className="flex flex-col gap-4">
        <h2 className="text-2xl font-bold tracking-tight">
          {t('steps.title')}
        </h2>
        <ol className="flex flex-col gap-6">
          <Step number={1} title={t(`steps.access.${mode}.title`)}>
            <p>{t(`steps.access.${mode}.description`)}</p>
            <Button asChild className="self-start">
              <Link href="/obtain-access">
                <ArrowRightIcon className="size-4" />
                {t(`steps.access.${mode}.action`)}
              </Link>
            </Button>
          </Step>
          <Step number={2} title={t('steps.launch.title', { version })}>
            <p>{t('steps.launch.description')}</p>
          </Step>
          <Step
            number={3}
            title={
              joinable
                ? t('steps.join.title', { address: season.publicAddress! })
                : t('steps.join.titleSoon')
            }
          >
            <p>
              {joinable
                ? t('steps.join.description')
                : t('steps.join.descriptionSoon')}
            </p>
          </Step>
        </ol>
      </section>

      <section className="flex flex-col gap-4">
        <h2 className="text-2xl font-bold tracking-tight">{t('more.title')}</h2>
        <div className="flex flex-wrap gap-3">
          <Button variant="secondary" asChild>
            <Link href="/wiki">
              <BookOpenIcon className="size-4" />
              {t('more.wiki')}
            </Link>
          </Button>
          <Button variant="secondary" asChild>
            <Link href="/wiki/rules">
              <ShieldIcon className="size-4" />
              {t('more.rules')}
            </Link>
          </Button>
          {footerSocialLinks
            .filter(({ label }) => label !== 'support')
            .map(({ label, icon: Icon, href }) => (
              <Button key={label} variant="secondary" asChild>
                <a href={href} target="_blank" rel="noreferrer">
                  <Icon className="size-4" />
                  {t(`more.${label}`)}
                </a>
              </Button>
            ))}
        </div>
      </section>
    </div>
  )
}

function Step({
  number,
  title,
  children,
}: React.PropsWithChildren<{ number: number; title: string }>) {
  return (
    <li className="grid grid-cols-[2.5rem_1fr] gap-4">
      <span
        aria-hidden
        className="font-minecraft flex size-10 items-center justify-center rounded-sm bg-teal-500/15 text-xl text-teal-300"
      >
        {number}
      </span>
      <div className="flex flex-col gap-2 pt-1.5">
        <h3 className="text-lg font-semibold break-words">{title}</h3>
        <div className="text-muted-foreground flex flex-col gap-3">
          {children}
        </div>
      </div>
    </li>
  )
}
