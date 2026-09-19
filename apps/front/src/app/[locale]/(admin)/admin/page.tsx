import Link from 'next/link'
import { ChevronRight } from 'lucide-react'
import { adminData, adminURL, queryString } from '@/lib/admin/server'
import { AdminHeading, AdminEmpty, AdminIssue } from '@/components/admin/common'
import { CopyID } from '@/components/admin/actions'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { FieldGroup, Field, FieldLabel } from '@/components/ui/field'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card'
import { Table, TableHeader, TableHead, TableBody, TableRow, TableCell } from '@/components/ui/table'

export default async function AccountsPage({params,searchParams}:{params:Promise<{locale:string}>;searchParams:Promise<Record<string,string|string[]|undefined>>}) {
 const [{locale},query]=await Promise.all([params,searchParams])
 const base=`/${locale}/admin`
 let data
 try {data=await adminData(`/?${queryString(query,['q','page_token','attach_to'])}`)} catch(error) {return <AdminIssue error={error} base={base}/>}
 const accounts=data.accounts??[]
 return <>
  <AdminHeading title={data.attachProfile?'Выбор аккаунта':'Аккаунты'} description="Участники проекта и их игровые профили."/>
  {data.attachProfile?<Alert><AlertTitle>Выберите владельца для {data.attachProfile.username}</AlertTitle><AlertDescription><p>После выбора проверьте аккаунт и подтвердите привязку.</p><Link href={`${base}/profiles/${data.attachProfile.id}`}>Отмена</Link></AlertDescription></Alert>:null}
  <form action={base}><FieldGroup className="flex-row flex-wrap items-end"><Field className="min-w-52 flex-1"><FieldLabel htmlFor="account-search">Email или UUID аккаунта</FieldLabel><Input id="account-search" name="q" type="search" defaultValue={data.search} placeholder="Точный email или UUID аккаунта"/></Field>{data.attachProfile?<input type="hidden" name="attach_to" value={data.attachProfile.id}/>:null}<Button type="submit">Найти аккаунт</Button>{data.search?<Button asChild variant="ghost"><Link href={`${base}${data.attachProfile?`?attach_to=${data.attachProfile.id}`:''}`}>Сбросить</Link></Button>:null}</FieldGroup></form>
  <Card><CardHeader><CardTitle>{data.search?'Результаты поиска':'Все аккаунты'}</CardTitle><CardDescription>Включая аккаунты без игрового профиля</CardDescription></CardHeader><CardContent>
   {accounts.length?<Table><TableHeader><TableRow><TableHead>Аккаунт</TableHead><TableHead>Статус</TableHead><TableHead>Регистрация, UTC</TableHead><TableHead><span className="sr-only">Действия</span></TableHead></TableRow></TableHeader><TableBody>{accounts.map(account=><TableRow key={account.id}><TableCell><div className="flex items-center gap-3"><Avatar><AvatarFallback>{account.traits.email?.[0]?.toUpperCase()??'?'}</AvatarFallback></Avatar><div><span className="font-medium">{account.traits.email||'Email не указан'}</span><CopyID value={account.id}/></div></div></TableCell><TableCell><Badge variant={account.state==='active'?'secondary':'outline'}>{account.state==='active'?'Активен':'Неактивен'}</Badge></TableCell><TableCell><time dateTime={account.created_at}>{account.created_at?new Date(account.created_at).toLocaleDateString('ru-RU',{timeZone:'UTC'}):'—'}</time></TableCell><TableCell><Button asChild variant="ghost" size="sm"><Link href={data.attachProfile?`${base}/profiles/${data.attachProfile.id}?candidate=${account.id}`:`${base}/profiles?owner=${account.id}`}>{data.attachProfile?'Выбрать':'Профили'}<ChevronRight data-icon="inline-end"/></Link></Button></TableCell></TableRow>)}</TableBody></Table>:<AdminEmpty title="Аккаунты не найдены" description="Проверьте полный email или UUID аккаунта либо сбросьте поиск."/>}
  </CardContent><CardFooter className="flex flex-wrap justify-between gap-3"><span className="text-xs text-muted-foreground">На этой странице: {accounts.length}</span><div className="flex gap-2">{data.previous?<Button asChild variant="outline" size="sm"><Link href={adminURL(base,data.previous)}>К началу</Link></Button>:null}{data.next?<Button asChild variant="outline" size="sm"><Link href={adminURL(base,data.next)}>Далее</Link></Button>:null}</div></CardFooter></Card>
 </>
}
