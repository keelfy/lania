import { WIKI_ITEMS, type WikiItem, type WikiLang } from '@/lib/wiki-items'
import styles from './crafting-recipe.module.css'

type Props = {
  lang: WikiLang
  // Three rows of three characters; a space is an empty slot.
  shape: [string, string, string]
  // Shape character -> item.
  ingredients: Record<string, WikiItem>
  result: WikiItem
  amount?: number
}

const TITLE = { ru: 'Верстак', en: 'Crafting' } as const

function Slot({
  item,
  lang,
  amount,
}: {
  item?: WikiItem
  lang: WikiLang
  amount?: number
}) {
  const name = item && WIKI_ITEMS[item][lang]
  return (
    <div className={styles.slot} title={name}>
      {item && (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          className={styles.icon}
          src={`/images/wiki/items/${item}.png`}
          alt={name}
          width={16}
          height={16}
        />
      )}
      {amount && amount > 1 ? (
        <span className={styles.amount}>{amount}</span>
      ) : null}
    </div>
  )
}

export function CraftingRecipe({
  lang,
  shape,
  ingredients,
  result,
  amount,
}: Props) {
  return (
    <figure className={styles.table}>
      <figcaption className={styles.title}>{TITLE[lang]}</figcaption>
      <div className={styles.body}>
        <div className={styles.grid}>
          {shape.flatMap((row, y) =>
            [0, 1, 2].map((x) => (
              <Slot
                key={`${y}${x}`}
                lang={lang}
                item={ingredients[row[x] ?? ' ']}
              />
            )),
          )}
        </div>
        <span className={styles.arrow} aria-hidden />
        <Slot lang={lang} item={result} amount={amount} />
      </div>
    </figure>
  )
}
