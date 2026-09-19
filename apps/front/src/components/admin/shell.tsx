'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { Users, SquareUserRound, ShieldCheck } from 'lucide-react'
import DeerIcon from '@/components/icons/DeerIcon'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'

export function AdminShell({ children, base }: { children: React.ReactNode; base: string }) {
  const pathname = usePathname()
  const profiles = pathname.includes('/admin/profiles')
  return <div className="flex min-h-dvh flex-col bg-muted/30 lg:flex-row">
    <a href="#admin-content" className="sr-only focus:not-sr-only focus:absolute focus:p-4">Перейти к содержимому</a>
    <aside className="flex shrink-0 flex-col gap-8 border-b bg-sidebar p-5 lg:sticky lg:top-0 lg:h-dvh lg:w-56 lg:border-r lg:border-b-0 lg:p-6">
      <Link href={base} className="flex items-center gap-3"><DeerIcon className="size-8 text-primary"/><div><span className="text-xl font-semibold">Lania</span><p className="text-xs text-muted-foreground">Управление проектом</p></div></Link>
      <nav aria-label="Администрирование" className="flex gap-2 lg:flex-col">
        {[{href:base,label:'Аккаунты',icon:Users,active:!profiles},{href:`${base}/profiles`,label:'Профили',icon:SquareUserRound,active:profiles}].map(({href,label,icon:Icon,active})=><Button key={href} asChild variant={active?'secondary':'ghost'} className={cn('justify-start', active && 'font-semibold')}><Link href={href} aria-current={active?'page':undefined}><Icon data-icon="inline-start"/>{label}</Link></Button>)}
      </nav>
      <div className="mt-auto hidden flex-col gap-4 lg:flex"><Separator/><div className="flex items-center gap-2 text-xs text-muted-foreground"><ShieldCheck className="size-4"/>Административный доступ</div></div>
    </aside>
    <div className="flex min-w-0 flex-1 flex-col"><header className="flex h-16 items-center justify-between border-b bg-background px-6 text-xs text-muted-foreground"><span>Управление / {profiles?'Профили':'Аккаунты'}</span><Link href="/">Открыть сайт</Link></header><main id="admin-content" className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-7 p-4 md:p-8">{children}</main><footer className="px-8 py-5 text-xs text-muted-foreground">Lania · Аккаунты, игровые профили и продукты</footer></div>
  </div>
}
