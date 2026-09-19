import Link from 'next/link'
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

export default async function ProfilesPage({params,searchParams}:{params:Promise<{locale:string}>;searchParams:Promise<Record<string,string|string[]|undefined>>}) {
 const [{locale},query]=await Promise.all([params,searchParams])
 const base=`/${locale}/admin`
 let data
 try {data=await adminData(`/profiles?${queryString(query,['q','owner','offset'])}`)} catch(error) {return <AdminIssue error={error} base={base}/>}
 const profiles=data.profiles??[]
 return <>
  <AdminHeading title={data.account?'Профили аккаунта':'Игровые профили'} description="Найдите игрока, настройте владельца и продукты."/>
  {data.account?<Alert><AlertTitle>{data.account.traits.email}</AlertTitle><AlertDescription><CopyID value={data.account.id}/><Link href={`${base}/profiles`}>Показать все профили</Link></AlertDescription></Alert>:null}
  <form action={`${base}/profiles`}><FieldGroup className="flex-row flex-wrap items-end"><input type="hidden" name="owner" value={data.owner}/><Field className="min-w-52 flex-1"><FieldLabel htmlFor="profile-search">Никнейм или UUID профиля</FieldLabel><Input id="profile-search" name="q" type="search" defaultValue={data.search} placeholder="Поиск игрока"/></Field><Button type="submit">Найти профиль</Button>{data.search?<Button asChild variant="ghost"><Link href={`${base}/profiles?owner=${data.owner}`}>Сбросить</Link></Button>:null}</FieldGroup></form>
  <Card><CardHeader><CardTitle>Список профилей</CardTitle><CardDescription>Профили с владельцем и без владельца</CardDescription></CardHeader><CardContent>{profiles.length?<Table><TableHeader><TableRow><TableHead>Игрок</TableHead><TableHead>Аккаунт владельца</TableHead><TableHead><span className="sr-only">Действия</span></TableHead></TableRow></TableHeader><TableBody>{profiles.map(profile=><TableRow key={profile.id}><TableCell><div className="flex items-center gap-3"><Avatar><AvatarFallback>{profile.username[0]?.toUpperCase()}</AvatarFallback></Avatar><div><Link href={`${base}/profiles/${profile.id}`} className="font-medium">{profile.username}</Link><CopyID value={profile.id}/></div></div></TableCell><TableCell>{profile.ownerId?<Link href={`${base}/profiles?owner=${profile.ownerId}`}><Badge variant="secondary">Привязан</Badge><code className="mt-2 block text-[11px] text-muted-foreground">{profile.ownerId}</code></Link>:<Badge variant="outline">Без владельца</Badge>}</TableCell><TableCell><Button asChild variant="ghost" size="sm"><Link href={`${base}/profiles/${profile.id}`}>Управлять</Link></Button></TableCell></TableRow>)}</TableBody></Table>:<AdminEmpty title="Профили не найдены" description="Измените поиск или откройте список всех профилей."/>}</CardContent><CardFooter className="flex flex-wrap justify-between gap-3"><span className="text-xs text-muted-foreground">На этой странице: {profiles.length}</span><div className="flex gap-2">{data.previous?<Button asChild size="sm" variant="outline"><Link href={adminURL(base,data.previous)}>Назад</Link></Button>:null}{data.next?<Button asChild size="sm" variant="outline"><Link href={adminURL(base,data.next)}>Далее</Link></Button>:null}</div></CardFooter></Card>
 </>
}
