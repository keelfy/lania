import { Badge } from '@/components/ui/badge'
import { ProfileRole } from '@/models/profile'

export default function RoleBadge({ role }: { role: ProfileRole }) {
  return (
    <Badge
      variant={role === 'owner' || role === 'admin' ? 'default' : 'secondary'}
    >
      {role}
    </Badge>
  )
}
