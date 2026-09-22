import LoadingSpinner from '@/components/ui/loading-spinner'

export default function Loading() {
  return (
    <div className="flex h-full w-full items-center justify-center py-10">
      <LoadingSpinner />
    </div>
  )
}
