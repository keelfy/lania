import McUsername from '@/components/ui/mc-username'
import PlayerFace from '@/components/ui/player-face'
import { cn } from '@/lib/utils'
import { Profile } from '@/models/profile'
import { useTranslations } from 'next-intl'
import Image from 'next/image'

export type NameLook = {
  username: string
  // Missing for the plain white name the game shows by default.
  colors?: string[]
  // Image of the glyph worn before the name.
  prefixImage?: string
}

// Names the game shows without any cosmetics.
const WHITE = ['#FFFFFF']

type NameTagProps = NameLook & {
  className?: string
}

// A name the way the game draws it: the glyph, then the name, both one line tall.
function NameTag({ username, colors, prefixImage, className }: NameTagProps) {
  return (
    <span className={cn('inline-flex items-center gap-[0.3em]', className)}>
      {prefixImage && (
        <Image
          src={prefixImage}
          alt=""
          width={32}
          height={32}
          className="size-[1.1em] shrink-0 [image-rendering:pixelated]"
        />
      )}
      <McUsername
        username={username}
        colors={colors ?? WHITE}
        className="translate-y-0 truncate drop-shadow-[0.12em_0.12em_0_rgba(0,0,0,0.6)]"
      />
    </span>
  )
}

type Props = React.ComponentProps<'figure'> & {
  look: NameLook
  profile?: Profile
}

// A still of the game screen with the name in the three places players see it:
// the player list, the tag above the head and the chat.
export default function InGamePreview({
  look,
  profile,
  className,
  ...props
}: Props) {
  const t = useTranslations('products.page.preview')

  return (
    <figure
      aria-label={t('label')}
      className={cn(
        'font-minecraft relative flex min-h-96 flex-col justify-between gap-6 overflow-hidden rounded-lg p-3 text-white select-none sm:p-4',
        // Daylight sky over a strip of grass and dirt.
        'bg-[linear-gradient(to_bottom,#78a7ff_0%,#b5cfff_78%,#5d9b3a_78%,#5d9b3a_81%,#7a5535_81%)]',
        className,
      )}
      {...props}
    >
      <div className="mx-auto w-full max-w-80 bg-black/45 p-0.5 text-base sm:text-lg">
        {['Alex', look.username, 'Steve'].map((username, index) => (
          <div
            key={index}
            className="mt-px flex items-center gap-1 bg-white/15 px-0.5 first:mt-0"
          >
            <PlayerFace
              player={index === 1 ? profile : undefined}
              className="size-[1em] shrink-0 [image-rendering:pixelated]"
            />
            {index === 1 ? (
              <NameTag {...look} className="min-w-0" />
            ) : (
              <NameTag username={username} />
            )}
          </div>
        ))}
      </div>

      <div className="flex flex-col items-center gap-2">
        <div className="max-w-full bg-black/30 px-1.5 text-xl sm:text-2xl">
          <NameTag {...look} className="max-w-full" />
        </div>
        <PlayerFace
          player={profile}
          className="size-20 shadow-[0_10px_0_-4px_rgba(0,0,0,0.25)] [image-rendering:pixelated]"
        />
      </div>

      <div className="w-fit max-w-full bg-black/45 px-1.5 py-0.5 text-base leading-relaxed sm:text-xl">
        <div className="flex min-w-0 items-center">
          <NameTag {...look} className="min-w-0" />
          &#58;&nbsp;
          <span className="truncate">{t('message')}</span>
        </div>
        <div className="flex min-w-0 items-center">
          <NameTag username="Alex" className="text-[#AAAAAA]" />
          &#58;&nbsp;
          <span className="truncate">{t('reply')}</span>
        </div>
      </div>
    </figure>
  )
}
