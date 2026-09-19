import Link from 'next/link'
import { adminData, queryString, AdminError } from '@/lib/admin/server'
import { categoryLabels } from '@/lib/admin/types'
import { AdminHeading, AdminEmpty, AdminIssue } from '@/components/admin/common'
import { CopyID, ConfirmMutation, AttachProfile, GrantProduct, SyncProfile } from '@/components/admin/actions'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card'
import { Table, TableHeader, TableHead, TableBody, TableRow, TableCell } from '@/components/ui/table'

export default async function ProfilePage({params,searchParams}:{params:Promise<{locale:string;id:string}>;searchParams:Promise<Record<string,string|string[]|undefined>>}) {
 const [{locale,id},query]=await Promise.all([params,searchParams])
 const base=`/${locale}/admin`
 let data
 try {data=await adminData(`/profiles/${encodeURIComponent(id)}?${queryString(query,['candidate'])}`)} catch(error) {return <AdminIssue error={error} base={base}/>}
 const profile=data.profile
 if(!profile) return <AdminIssue error={new AdminError(404,'Профиль не найден.')} base={base}/>
 const grants=data.grants??[]
 return <>
  <Link className="text-sm text-muted-foreground" href={`${base}/profiles`}>Все профили</Link>
  <div className="flex flex-wrap items-center justify-between gap-4"><AdminHeading title={profile.username} description="Игровой профиль Minecraft"/><Badge variant={profile.ownerId?'secondary':'outline'}>{profile.ownerId?'Профиль привязан':'Без владельца'}</Badge></div>
  <dl className="flex flex-wrap gap-5 rounded-lg bg-muted/50 p-4"><div><dt className="text-xs text-muted-foreground">UUID профиля</dt><dd><CopyID value={profile.id} label="Скопировать UUID профиля"/></dd></div><div><dt className="text-xs text-muted-foreground">Minecraft UUID</dt><dd><CopyID value={profile.minecraftUUID} label="Скопировать Minecraft UUID"/></dd></div></dl>
  <div className="grid min-w-0 items-start gap-6 xl:grid-cols-[minmax(0,1fr)_320px]">
   <div className="flex min-w-0 flex-col gap-6"><Card><CardHeader><CardTitle>Продукты профиля <Badge variant="secondary">{grants.length}</Badge></CardTitle><CardDescription>Выданные права и срок их действия</CardDescription></CardHeader><CardContent>{grants.length?<Table><TableHeader><TableRow><TableHead>Продукт</TableHead><TableHead>Срок</TableHead><TableHead><span className="sr-only">Действия</span></TableHead></TableRow></TableHeader><TableBody>{grants.map(grant=><TableRow key={grant.id}><TableCell><strong className="font-medium">{grant.name}</strong><p className="text-xs text-muted-foreground">{categoryLabels[grant.kind]}</p></TableCell><TableCell><Badge variant="outline">{grant.season?`Сезон ${grant.season}`:'Постоянно'}</Badge></TableCell><TableCell>{grant.protected?<span className="text-xs text-muted-foreground">Базовый цвет</span>:<ConfirmMutation profileId={id} action="revoke" fields={{grant:grant.id,kind:grant.kind,confirm:'yes'}} label="Отозвать" title={`Отозвать «${grant.name}»?`} description={`${profile.username} потеряет ${categoryLabels[grant.kind].toLowerCase()} — ${grant.season?`сезон ${grant.season}`:'постоянное право'}. Оплаченный заказ сохранится. Возврат денег не выполняется.`}/>}</TableCell></TableRow>)}</TableBody></Table>:<AdminEmpty title="Продукты ещё не выданы" description="Выберите продукт и сезон в форме выдачи."/>}</CardContent><CardFooter><p className="text-xs text-muted-foreground">Отзыв убирает право на продукт. История заказов сохраняется.</p></CardFooter></Card>
   <Card><CardHeader><CardTitle>Обновление Minecraft</CardTitle><CardDescription>Если сервер был недоступен, повторите синхронизацию после восстановления связи.</CardDescription></CardHeader><CardContent><SyncProfile profileId={id}/></CardContent></Card></div>
   <aside className="grid min-w-0 gap-6 md:grid-cols-2 xl:grid-cols-1" aria-label="Действия с профилем"><Card><CardHeader><CardTitle>Владелец профиля</CardTitle><CardDescription>{profile.ownerId?'Аккаунт, управляющий профилем':data.candidate?'Проверьте выбранный аккаунт':'Выберите аккаунт владельца'}</CardDescription></CardHeader><CardContent className="flex flex-col gap-4">{profile.ownerId?<><p className="break-all font-medium">{data.account?.traits.email??'Аккаунт удалён'}</p><CopyID value={profile.ownerId}/>{data.account?<Link className="text-sm text-primary underline" href={`${base}/profiles?owner=${profile.ownerId}`}>Профили этого аккаунта</Link>:null}<ConfirmMutation profileId={id} action="owner" fields={{owner:'',expected_owner:profile.ownerId}} label="Отвязать профиль" title={`Отвязать ${profile.username}?`} description="Текущий аккаунт потеряет управление профилем. Игровой профиль и продукты сохранятся."/></>:data.candidate?<><p className="break-all font-medium">{data.candidate.traits.email}</p><CopyID value={data.candidate.id}/><AttachProfile profileId={id} ownerId={data.candidate.id}/><Link className="text-sm text-primary underline" href={`${base}?attach_to=${id}`}>Выбрать другой аккаунт</Link></>:<><p className="text-sm text-muted-foreground">Профиль ещё не привязан. Найдите владельца по email или UUID.</p><Button asChild variant="outline"><Link href={`${base}?attach_to=${id}`}>Выбрать аккаунт</Link></Button></>}</CardContent></Card>
   <Card><CardHeader><CardTitle>Выдать продукт</CardTitle><CardDescription>Выдача для выбранного сезона</CardDescription></CardHeader><CardContent><GrantProduct profileId={id} products={data.products??[]} seasons={data.seasons??[]} activeSeason={data.activeSeason}/></CardContent></Card></aside>
  </div>
 </>
}
