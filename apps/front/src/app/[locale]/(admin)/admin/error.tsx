'use client'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
export default function AdminError({reset}:{reset:()=>void}) {
 return <Alert variant="destructive"><AlertTitle>Не удалось открыть раздел</AlertTitle><AlertDescription><p>Повторите загрузку страницы.</p><Button variant="outline" onClick={reset}>Повторить</Button></AlertDescription></Alert>
}
