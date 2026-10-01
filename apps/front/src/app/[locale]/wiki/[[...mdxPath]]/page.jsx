import { getWikiSeasons } from '@/lib/wiki-seasons'
import { getWikiSeasonFolders } from '@/lib/wiki-page-map'
import { useMDXComponents as getMDXComponents } from '@/mdx-components'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'
import { notFound, redirect } from 'next/navigation'
import { Callout } from 'nextra/components'
import { getPageMap } from 'nextra/page-map'
import { generateStaticParamsFor, importPage } from 'nextra/pages'

export const generateStaticParams = generateStaticParamsFor('mdxPath', 'locale')

const findPage = (mdxPath, locale) =>
  importPage(mdxPath, locale).catch(() => undefined)

export async function generateMetadata(props) {
  const params = await props.params
  const page = await findPage(params.mdxPath, params.locale)
  return page?.metadata ?? {}
}

const Wrapper = getMDXComponents().wrapper

export default async function Page(props) {
  const params = await props.params
  const mdxPath = params.mdxPath ?? []
  const seasonFolders = getWikiSeasonFolders(
    await getPageMap(`/${params.locale}/wiki`),
  )
  const seasons = await getWikiSeasons(seasonFolders)
  const primary = seasons.find((season) => season.isPrimary)
  const inSeason = seasonFolders.includes(mdxPath[0])

  const page = await findPage(params.mdxPath, params.locale)

  if (!page) {
    // Links from before the wiki had seasons (/wiki/commands) lead to the current season's page.
    if (
      primary &&
      !inSeason &&
      (await findPage([primary.slug, ...mdxPath], params.locale))
    ) {
      redirect(`/${params.locale}/wiki/${[primary.slug, ...mdxPath].join('/')}`)
    }
    notFound()
  }

  const { default: MDXContent, toc, metadata, sourceCode } = page
  const season = seasons.find((season) => season.slug === mdxPath[0])
  const t = await getTranslations({ locale: params.locale, namespace: 'wiki' })

  return (
    <Wrapper toc={toc} metadata={metadata} sourceCode={sourceCode}>
      {season && primary && !season.isPrimary && (
        <Callout type="warning">
          {t.rich('season.archived', {
            name: season.name,
            current: primary.name,
            link: (chunks) => (
              <Link href={`/${params.locale}/wiki/${primary.slug}`}>
                {chunks}
              </Link>
            ),
          })}
        </Callout>
      )}
      <MDXContent {...props} params={params} />
    </Wrapper>
  )
}
