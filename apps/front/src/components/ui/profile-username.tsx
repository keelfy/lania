import { Profile } from '@/models/profile'
import McUsername from './mc-username'
import { Popover, PopoverContent, PopoverTrigger } from './popover'
import PlayerFace from './player-face'
import { cn } from '@/lib/utils'
import PlayerCard from './player-card'

type Props = React.ComponentProps<'p'> & {
  colors?: string[] | string
  profile?: Profile
}

export default function ProfileUsername({
  className,
  profile,
  colors,
  ...props
}: Props) {
  return (
    <Popover>
      <PopoverTrigger
        asChild
        className="group hover:bg-accent cursor-pointer rounded-xs px-1"
      >
        <span>
          <PlayerFace player={profile} className="mr-1 inline-block size-5" />
          <McUsername
            className={cn(className)}
            username={profile?.username ?? ''}
            colors={colors ?? profile?.cosmetics.name.colors.colors}
            {...props}
          />
        </span>
      </PopoverTrigger>
      <PopoverContent className="max-w-sm min-w-max p-0">
        {profile?.id && (
          <PlayerCard
            profileId={profile.id}
            nameCosmetics={profile.cosmetics.name}
            className="border-none"
          />
        )}
      </PopoverContent>
    </Popover>
  )
}
