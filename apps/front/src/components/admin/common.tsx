import Link from 'next/link'
import { Users, ShieldCheck } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert'
import { Empty, EmptyHeader, EmptyMedia, EmptyTitle, EmptyDescription } from '@/components/ui/empty'
import { AdminError } from '@/lib/admin/server'

export function AdminHeading({ title, description }: { title: string; description: string }) {
 return <div className="flex flex-col gap-2"><h1 className="text-3xl font-semibold tracking-tight">{title}</h1><p className="text-sm text-muted-foreground">{description}</p></div>
}
export function AdminEmpty({ title, description }: { title: string; description: string }) {
 return <Empty><EmptyHeader><EmptyMedia variant="icon"><Users/></EmptyMedia><EmptyTitle>{title}</EmptyTitle><EmptyDescription>{description}</EmptyDescription></EmptyHeader></Empty>
}
export function AdminIssue({ error, base }: { error: unknown; base: string }) {
 const status=error instanceof AdminError?error.status:503
 const message=error instanceof AdminError?error.message:'Не удалось загрузить данные. Повторите позже.'
 const login=`${process.env.NEXT_PUBLIC_ORY_SDK_URL}/self-service/login/browser?${new URLSearchParams({return_to:`${process.env.ADMIN_ORIGIN}${base}`})}`
 return <div className="mx-auto flex w-full max-w-xl flex-col gap-6 py-12"><ShieldCheck className="size-10 text-muted-foreground"/><AdminHeading title={status===401?'Войти в управление':status===403?'Доступ ограничен':'Данные недоступны'} description="Административное пространство Lania"/><Alert variant="destructive"><AlertTitle>{status===403?'Нужен аккаунт администратора':'Не удалось открыть раздел'}</AlertTitle><AlertDescription>{message}</AlertDescription></Alert><div className="flex flex-wrap gap-3">{status===401||status===403?<Button asChild><a href={login}>Войти через Lania</a></Button>:<Button asChild variant="outline"><Link href={base}>Открыть аккаунты</Link></Button>}<Button asChild variant="ghost"><Link href="/">На сайт</Link></Button></div></div>
}
