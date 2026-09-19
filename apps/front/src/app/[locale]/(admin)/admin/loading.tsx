import { Skeleton } from '@/components/ui/skeleton'
export default function Loading() {
 return <div className="flex flex-col gap-6" role="status" aria-label="Загрузка админки"><Skeleton className="h-9 w-60"/><Skeleton className="h-10 w-full max-w-lg"/><Skeleton className="h-80 w-full"/></div>
}
