import DeerIcon from '@/components/icons/DeerIcon'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { Button } from '@/components/ui/button'
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from '@/components/ui/hover-card'
import RichText from '@/components/ui/rich-text'
import {
  ArrowRightIcon,
  AsteriskIcon,
  BanknoteXIcon,
  CreditCardIcon,
  CuboidIcon,
  GemIcon,
  HandCoinsIcon,
  HandshakeIcon,
  MousePointerClickIcon,
  PaletteIcon,
  ScaleIcon,
  UsersIcon,
} from 'lucide-react'
import { RichTagsFunction } from 'next-intl'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'
import LandingCommunityCarousel from './components/community-carousel'
import CopyIPButton from './components/copy-ip-button'
import LandingSidebar from './components/landing-sidebar'
import LiquidGlassCard from './components/liquid-glass-card'
import SeeMoreLandingButton from './components/see-more-landing-button'
import SidebarSection from './components/sidebar-section'

const gallery = [
  {
    src: 'https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cuK6PLjxsKBdEOXRq79j3Q2ebWZmYSnoMa5iC',
    alt: 'Dragons (Season 3)',
    season: 3,
    authors: ['Asdqqq', 'Foxizans'],
  },
  {
    src: 'https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cbjQr2f9jT7i9h3X0yLRptdlZCvF5VwOSKfU8',
    alt: 'Dark Church (Season 3)',
    season: 3,
    authors: ['keelfy'],
  },
  {
    src: 'https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cbKOKTc9jT7i9h3X0yLRptdlZCvF5VwOSKfU8',
    alt: 'Flying Base (Season 3)',
    season: 3,
    authors: ['Quermercus_'],
  },
  {
    src: 'https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cPU3bWHep2GIFY0RMkWP3O9fQzVBLAKolCseX',
    alt: 'Spawn (Season 3)',
    season: 3,
    authors: ['frogot123', 'keelfy'],
  },
  {
    src: 'https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cxucvXLJjowQqVUy7ZDtGeB9iHgs1vTAchKzW',
    alt: 'Overgrown Base (Season 2)',
    season: 2,
    authors: ['keelfy', 'loonatya'],
  },
  {
    src: 'https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cd01uvQimQoVFsTb8JXRqtDwOZxyUCYIS7luN',
    alt: 'The End (Season 3)',
    season: 3,
    authors: [],
  },
  {
    src: 'https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cHC575ZA8osiSyeTWCFqrgvtlnPIx9d5GaVLX',
    alt: 'Snowmans (Season 3)',
    season: 3,
    authors: [],
  },
  {
    src: 'https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cxmeBarJjowQqVUy7ZDtGeB9iHgs1vTAchKzW',
    alt: 'Farms (Season 1)',
    season: 1,
  },
  {
    src: 'https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cJGZy0E5VXE04WIcQBZrqeYRp5GjxiFL6ty1T',
    alt: 'keelfy base (Season 1)',
    season: 1,
  },
  {
    src: 'https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cd2tyu2imQoVFsTb8JXRqtDwOZxyUCYIS7luN',
    alt: 'Boats attack (Season 3)',
    season: 3,
  },
  {
    src: 'https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cW1Jd4BC0b8RNXgu7hPxajkedqcOpDJMwf3Cv',
    alt: 'Angel (Season 1)',
    season: 1,
    authors: ['BiruSan'],
  },
  {
    src: 'https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8ckcfvAHLExFzSaI68v4fT1c3yn0Zstg9ojGXm',
    alt: 'Overgrown Base 2 (Season 2)',
    season: 2,
    authors: ['loonatya', 'keelfy'],
  },
  {
    src: 'https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cswRNljPNZdrg5V9nlMTAeyvGcPBu6KkRWODp',
    alt: 'Foxizans humor (Season 3)',
    season: 3,
    authors: ['Foxizans', 'frogot123', 'Asdqqq'],
  },
  {
    src: 'https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cu5jHnCcxsKBdEOXRq79j3Q2ebWZmYSnoMa5i',
    alt: 'Market (Season 3)',
    season: 3,
    authors: ['loonatya'],
  },
  {
    src: 'https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8crrv0At2YLQXVF6oKxqGPsUNEMD2fBz5beSgk',
    alt: 'Cherry Blossom Base (Season 1)',
    season: 1,
    authors: ['loonatya', 'Eugene'],
  },
  {
    src: 'https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cxXkfLxXJjowQqVUy7ZDtGeB9iHgs1vTAchKz',
    alt: 'Bloody Night (Season 3)',
    season: 3,
    authors: ['frogot123'],
  },
]

const features = [
  {
    title: 'vanilla',
    icon: CuboidIcon,
    getDescription: (
      rich: (
        key: string,
        args?: Record<string, string | number | RichTagsFunction | Date>,
      ) => React.ReactNode,
    ) => (
      <p>
        <RichText>
          {(tags) =>
            rich('section3.features.vanilla.description', {
              ...tags,
              features: (chunks: React.ReactNode) => (
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
        <br />
        <br />
        <span className="text-muted-foreground text-xs">
          <RichText>
            {(tags) =>
              rich('section3.features.vanilla.description2', {
                ...tags,
                requirements: (chunks: React.ReactNode) => (
                  <Link
                    href="/wiki/requirements"
                    className="underline underline-offset-2"
                  >
                    {chunks}
                  </Link>
                ),
              })
            }
          </RichText>
        </span>
      </p>
    ),
  },
  {
    title: 'noLicense',
    icon: CreditCardIcon,
    getDescription: (
      rich: (
        key: string,
        args?: Record<string, string | number | RichTagsFunction | Date>,
      ) => React.ReactNode,
    ) => (
      <p>
        {rich('section3.features.noLicense.description')}
        <br />
        <br />
        <span className="text-muted-foreground text-xs">
          {rich('section3.features.noLicense.description2')}
        </span>
      </p>
    ),
  },
  {
    title: 'staff',
    icon: ScaleIcon,
    getDescription: (
      rich: (
        key: string,
        args?: Record<string, string | number | RichTagsFunction | Date>,
      ) => React.ReactNode,
    ) => (
      <p>
        {rich('section3.features.staff.description')}
        <br />
        <br />
        <RichText className="text-muted-foreground text-xs">
          {(tags) =>
            rich('section3.features.staff.description2', {
              ...tags,
              staff: (chunks: React.ReactNode) => (
                <Link
                  href="/wiki/useful/staff"
                  className="underline underline-offset-2"
                >
                  {chunks}
                </Link>
              ),
            })
          }
        </RichText>
      </p>
    ),
  },
  {
    title: 'optimization',
    icon: GemIcon,
    getDescription: (
      rich: (
        key: string,
        args?: Record<string, string | number | RichTagsFunction | Date>,
      ) => React.ReactNode,
    ) => (
      <p>
        {rich('section3.features.optimization.description')}
        <br />
        <br />
        {rich('section3.features.optimization.description2')}
      </p>
    ),
  },
  {
    title: 'noPayToWin',
    icon: BanknoteXIcon,
    getDescription: (
      rich: (
        key: string,
        args?: Record<string, string | number | RichTagsFunction | Date>,
      ) => React.ReactNode,
    ) => (
      <HoverCard openDelay={0}>
        <RichText>
          {(tags) =>
            rich('section3.features.noPayToWin.description', {
              ...tags,
              hover: (chunks: React.ReactNode) => (
                <HoverCardTrigger>
                  <span className="cursor-default italic underline underline-offset-2">
                    {chunks}
                  </span>
                </HoverCardTrigger>
              ),
            })
          }
        </RichText>
        <HoverCardContent className="max-w-lg text-sm">
          <RichText>
            {(tags) =>
              rich('section3.features.noPayToWin.balanceHover', {
                ...tags,
              })
            }
          </RichText>
        </HoverCardContent>
      </HoverCard>
    ),
  },
  {
    title: 'promotionOfCreativity',
    icon: PaletteIcon,
    getDescription: (
      rich: (
        key: string,
        args?: Record<string, string | number | RichTagsFunction | Date>,
      ) => React.ReactNode,
    ) => <p>{rich('section3.features.promotionOfCreativity.description')}</p>,
  },
  {
    title: 'partnershipProgram',
    icon: HandCoinsIcon,
    disabled: true,
    getDescription: (
      rich: (
        key: string,
        args?: Record<string, string | number | RichTagsFunction | Date>,
      ) => React.ReactNode,
    ) => (
      <div className="grid gap-2">
        <ol className="list-inside list-disc space-y-2">
          <li>{rich('section3.features.partnershipProgram.element1')}</li>
          <li>{rich('section3.features.partnershipProgram.element2')}</li>
        </ol>
        <Link
          href="/partnership"
          className="font-semibold underline underline-offset-2"
        >
          {rich('section3.features.partnershipProgram.seeMore')}
        </Link>
      </div>
    ),
  },
]

type Props = {
  params: Promise<{
    locale: string
  }>
}

export default async function LandingPage({ params }: Props) {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'landing' })
  return (
    <div className="w-full">
      <LandingSidebar />
      <SidebarSection
        id="section-1"
        className="flex h-svh w-full snap-start flex-col items-center justify-center"
      >
        <div className="relative">
          <LiquidGlassCard>
            <div className="flex flex-col items-start gap-6">
              <div className="flex items-center gap-4">
                <DeerIcon className="size-16 sm:size-20" />
                <h1 className="cursor-default bg-gradient-to-r from-white to-teal-400/90 bg-clip-text text-5xl font-bold tracking-widest text-transparent uppercase sm:text-6xl">
                  lania
                </h1>
              </div>

              <div className="flex flex-col gap-2">
                <p className="text-lg">
                  <RichText>
                    {(tags) => t.rich('description', { ...tags })}
                  </RichText>
                </p>

                <p className="text-md">
                  <AsteriskIcon className="inline-block size-4 align-text-top text-teal-400/90" />
                  <RichText>
                    {(tags) => t.rich('version', { ...tags })}
                  </RichText>
                </p>
              </div>

              <div className="flex w-full items-center gap-4">
                <Button variant="default" size="lg" asChild className="group">
                  <Link
                    href={`/${locale}/obtain-access`}
                    className="flex w-full items-center gap-2 sm:w-auto"
                  >
                    <ArrowRightIcon className="size-4 transition-transform duration-300 group-hover:translate-x-1" />
                    {t('getAccess')}
                  </Link>
                </Button>
                <CopyIPButton
                  className="hidden sm:inline-flex"
                  locale={locale}
                />
              </div>
            </div>
          </LiquidGlassCard>
          <SeeMoreLandingButton
            sectionId="section-2"
            className="top-[calc(50%+8rem)]"
          >
            {t('section1.seeMore')}
          </SeeMoreLandingButton>
        </div>
      </SidebarSection>

      <SidebarSection
        id="section-2"
        className="relative flex h-svh w-full snap-start items-start justify-center pt-12 lg:items-center lg:pt-0"
      >
        <div className="container mx-auto flex flex-col gap-4 px-6 py-4 backdrop-blur-xs sm:rounded-2xl md:max-w-5xl">
          <div className="flex items-center gap-2">
            <UsersIcon className="size-8" />
            <h2 className="bg-gradient-to-r from-white to-teal-300 bg-clip-text text-3xl font-bold text-transparent">
              {t('section2.title')}
            </h2>
          </div>
          <p className="text-lg">{t('section2.description1')}</p>
          <div>
            <LandingCommunityCarousel gallery={gallery} />
          </div>
          <p className="text-lg">
            <RichText>
              {(tags) => t.rich('section2.description2', { ...tags })}
            </RichText>
          </p>
        </div>
        <SeeMoreLandingButton
          sectionId="section-3"
          className="bottom-10 hidden sm:flex"
        >
          {t('section2.seeMore')}
        </SeeMoreLandingButton>
      </SidebarSection>

      <SidebarSection
        id="section-3"
        className="relative flex h-[calc(max(100svh,800px))] w-full snap-start items-start justify-center pt-12 lg:items-center lg:pt-0"
      >
        <div className="container mx-auto flex max-w-5xl flex-col gap-6 px-6 py-4 backdrop-blur-xs sm:rounded-2xl">
          <div className="flex flex-col gap-2">
            <div className="flex items-center gap-2">
              <MousePointerClickIcon className="size-8" />
              <h2 className="bg-gradient-to-r from-white to-teal-300 bg-clip-text text-3xl font-bold text-transparent">
                <RichText>
                  {(tags) => t.rich('section3.title', { ...tags })}
                </RichText>
              </h2>
            </div>
            <p className="text-md text-muted-foreground">
              {t('section3.description')}
            </p>
          </div>
          <Accordion
            type="single"
            collapsible
            className="w-full"
            defaultValue="item-1"
          >
            {features
              .filter((feature) => !feature.disabled)
              .map((feature) => (
                <AccordionItem key={feature.title} value={feature.title}>
                  <AccordionTrigger className="text-md">
                    <div className="flex items-center gap-2">
                      <feature.icon className="size-5 text-teal-100" />
                      {t(`section3.features.${feature.title}.title`)}
                    </div>
                  </AccordionTrigger>
                  <AccordionContent>
                    {feature.getDescription(t.rich)}
                  </AccordionContent>
                </AccordionItem>
              ))}
          </Accordion>
        </div>
        <SeeMoreLandingButton
          sectionId="section-4"
          className="bottom-10 hidden sm:flex"
        >
          {t('section3.seeMore')}
        </SeeMoreLandingButton>
      </SidebarSection>

      <SidebarSection
        id="section-4"
        className="flex h-svh w-full snap-start items-start justify-center pt-12 lg:items-center lg:pt-0"
      >
        <div className="container mx-auto flex max-w-5xl flex-col gap-6 px-6 py-4 backdrop-blur-xs sm:rounded-2xl">
          <div className="flex flex-col gap-2">
            <div className="flex items-center gap-2">
              <MousePointerClickIcon className="size-8" />
              <h2 className="bg-gradient-to-r from-white to-teal-300 bg-clip-text text-3xl font-bold text-transparent">
                {t('section4.title')}
              </h2>
            </div>
            <p className="text-md text-muted-foreground">
              {t('section4.description')}
            </p>
          </div>
          <div className="flex flex-col items-center gap-4 md:flex-row">
            <Button
              variant="default"
              size="lg"
              asChild
              className="group w-full lg:w-auto"
            >
              <Link href={`/${locale}/obtain-access`}>
                <ArrowRightIcon className="size-4 transition-transform duration-300 group-hover:translate-x-1" />
                {t('getAccess')}
              </Link>
            </Button>
            <Button
              variant="secondary"
              size="lg"
              asChild
              className="group w-full lg:w-auto"
              disabled
            >
              <Link
                href={`/${locale}/partnership`}
                className="pointer-events-none opacity-70"
              >
                <HandshakeIcon className="size-4 transition-colors duration-300 group-hover:text-teal-100" />
                {t('becomePartner')}
              </Link>
            </Button>
            <CopyIPButton className="w-full lg:w-auto" locale={locale} />
          </div>
        </div>
      </SidebarSection>
    </div>
  )
}
