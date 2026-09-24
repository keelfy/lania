import { redirect } from 'next/navigation'

// Claims live on the map of each world now.
export default function ClaimsPage() {
  redirect('/worlds')
}
