export default function DefaultLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="container mx-auto max-w-5xl flex-1 px-4 sm:px-0">
      {children}
    </div>
  )
}
