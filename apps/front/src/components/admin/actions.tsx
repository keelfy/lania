'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { Copy, LoaderCircle, Plus, RefreshCw } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert'
import { AlertDialog, AlertDialogTrigger, AlertDialogContent, AlertDialogHeader, AlertDialogTitle, AlertDialogDescription, AlertDialogFooter, AlertDialogCancel, AlertDialogAction } from '@/components/ui/alert-dialog'
import { FieldGroup, Field, FieldLabel, FieldDescription } from '@/components/ui/field'
import { Select, SelectTrigger, SelectValue, SelectContent, SelectGroup, SelectItem } from '@/components/ui/select'
import { categoryLabels, type Product } from '@/lib/admin/types'

function useMutation(profileId: string, action: string) {
 const router=useRouter()
 const [pending,startTransition]=useTransition()
 const [error,setError]=useState('')
 function execute(fields:Record<string,string>) {
  setError('')
  startTransition(async()=>{
   try {
    const response=await fetch(`/api/admin/profiles/${profileId}/${action}`,{method:'POST',body:new URLSearchParams(fields)})
    const data=await response.json()
    if(!response.ok) {setError(data.error??'Не удалось сохранить изменения.');return}
    if(data.result==='sync_failed') {setError('Права сохранены, но Minecraft пока не обновлён. Повторите синхронизацию.');toast.warning('Minecraft не обновлён')}
    else {toast.success('Изменения сохранены')}
    router.refresh()
   } catch {setError('Нет связи с сервисом. Проверьте состояние профиля перед повтором.')}
  })
 }
 return {pending,error,execute}
}
function MutationError({message}:{message:string}) {
 return message?<Alert variant="destructive"><AlertTitle>Проверьте результат</AlertTitle><AlertDescription>{message}</AlertDescription></Alert>:null
}
export function CopyID({value,label='Скопировать UUID'}:{value:string;label?:string}) {
 return <div className="flex min-w-0 items-center gap-1"><code className="break-all text-[11px] text-muted-foreground">{value}</code><Button variant="ghost" size="icon" type="button" aria-label={label} onClick={async()=>{try{await navigator.clipboard.writeText(value);toast.success('UUID скопирован')}catch{toast.error('Скопируйте UUID вручную')}}}><Copy/></Button></div>
}
export function ConfirmMutation({profileId,action,fields,label,title,description}:{profileId:string;action:'owner'|'revoke';fields:Record<string,string>;label:string;title:string;description:string}) {
 const {pending,error,execute}=useMutation(profileId,action)
 return <div className="flex flex-col gap-3"><AlertDialog><AlertDialogTrigger asChild><Button variant="outline" size="sm" disabled={pending}>{pending?<LoaderCircle data-icon="inline-start" className="animate-spin"/>:null}{label}</Button></AlertDialogTrigger><AlertDialogContent><AlertDialogHeader><AlertDialogTitle>{title}</AlertDialogTitle><AlertDialogDescription>{description}</AlertDialogDescription></AlertDialogHeader><AlertDialogFooter><AlertDialogCancel>Отмена</AlertDialogCancel><AlertDialogAction asChild><Button variant="destructive" onClick={()=>execute(fields)}>{label}</Button></AlertDialogAction></AlertDialogFooter></AlertDialogContent></AlertDialog><MutationError message={error}/></div>
}
export function AttachProfile({profileId,ownerId}:{profileId:string;ownerId:string}) {
 const {pending,error,execute}=useMutation(profileId,'owner')
 return <div className="flex flex-col gap-3"><Button disabled={pending} onClick={()=>execute({owner:ownerId,expected_owner:''})}>{pending?<LoaderCircle data-icon="inline-start" className="animate-spin"/>:null}Привязать профиль</Button><MutationError message={error}/></div>
}
export function GrantProduct({profileId,products,seasons,activeSeason}:{profileId:string;products:Product[];seasons:{id:string;number:number}[];activeSeason:string}) {
 const [search,setSearch]=useState('')
 const [product,setProduct]=useState('')
 const [season,setSeason]=useState(seasons.some(s=>s.id===activeSeason)?activeSeason:'')
 const {pending,error,execute}=useMutation(profileId,'give')
 const visible=products.filter(p=>`${p.name} ${categoryLabels[p.category]}`.toLocaleLowerCase('ru').includes(search.toLocaleLowerCase('ru')))
 return <form onSubmit={event=>{event.preventDefault();if(product&&season)execute({product,season})}} aria-busy={pending}><FieldGroup>
  <Field><FieldLabel htmlFor="product-search">Поиск продукта</FieldLabel><Input id="product-search" type="search" placeholder="Название продукта" value={search} onChange={event=>{setSearch(event.target.value);setProduct('')}}/></Field>
  <Field><FieldLabel htmlFor="product">Продукт</FieldLabel><Select value={product} onValueChange={setProduct} disabled={pending||visible.length===0}><SelectTrigger id="product" className="w-full"><SelectValue placeholder="Выберите продукт"/></SelectTrigger><SelectContent><SelectGroup>{visible.map(p=><SelectItem key={p.id} value={p.id}>{p.name} · {categoryLabels[p.category]}</SelectItem>)}</SelectGroup></SelectContent></Select>{visible.length===0?<FieldDescription>Продукты не найдены. Измените поиск.</FieldDescription>:null}</Field>
  <Field><FieldLabel htmlFor="season">Сезон</FieldLabel><Select value={season} onValueChange={setSeason} disabled={pending}><SelectTrigger id="season" className="w-full"><SelectValue placeholder="Выберите сезон"/></SelectTrigger><SelectContent><SelectGroup>{seasons.map(s=><SelectItem key={s.id} value={s.id}>Сезон {s.number}{s.id===activeSeason?' · текущий':''}</SelectItem>)}</SelectGroup></SelectContent></Select><FieldDescription>Продукт действует в выбранном сезоне. Цвет и префикс игрок выбирает самостоятельно.</FieldDescription></Field>
  <MutationError message={error}/><Button type="submit" disabled={pending||!product||!season}>{pending?<LoaderCircle data-icon="inline-start" className="animate-spin"/>:<Plus data-icon="inline-start"/>}{pending?'Сохраняем…':'Выдать продукт'}</Button>
 </FieldGroup></form>
}
export function SyncProfile({profileId}:{profileId:string}) {
 const {pending,error,execute}=useMutation(profileId,'sync')
 return <div className="flex flex-col items-start gap-3"><Button variant="outline" disabled={pending} onClick={()=>execute({})}>{pending?<LoaderCircle data-icon="inline-start" className="animate-spin"/>:<RefreshCw data-icon="inline-start"/>}Повторить синхронизацию</Button><MutationError message={error}/></div>
}
