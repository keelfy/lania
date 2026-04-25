import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  SiDiscord,
  SiDiscordHex,
  SiTelegram,
  SiTelegramHex,
} from '@icons-pack/react-simple-icons'
import {
  ArrowUpRightIcon,
  BriefcaseBusinessIcon,
  HeadsetIcon,
  MailIcon,
} from 'lucide-react'
import Link from 'next/link'

export default function ContactPage() {
  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-4xl font-extrabold tracking-tight">Контакты</h1>
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <HeadsetIcon className="size-6" />
              <p className="text-3xl font-semibold tracking-tight">Поддержка</p>
            </CardTitle>
            <CardDescription className="text-lg">
              Мы можем помочь вам с любыми вопросами, связанными с сервером,
              сайтом, оплатой и т.д.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex h-full items-end gap-2">
            <Button asChild variant="default" size="lg">
              <Link href={process.env.NEXT_PUBLIC_SUPPORT_URL ?? '#'}>
                <p>
                  Обратиться в <span className="font-bold">поддержку</span>
                </p>
                <ArrowUpRightIcon className="size-4" />
              </Link>
            </Button>
            <Button asChild variant="secondary" size="icon" className="size-10">
              <Link
                href={`mailto:${process.env.NEXT_PUBLIC_SUPPORT_EMAIL ?? ''}`}
              >
                <MailIcon className="size-4" />
              </Link>
            </Button>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <SiTelegram className="size-6" color={SiTelegramHex} />
              <p className="text-3xl font-semibold tracking-tight">Telegram</p>
            </CardTitle>
            <CardDescription className="text-lg">
              Канал с новостями и обновлениями
            </CardDescription>
          </CardHeader>
          <CardContent className="flex h-full items-end">
            <Button asChild variant="default" size="lg">
              <Link href={process.env.NEXT_PUBLIC_TELEGRAM_CHANNEL_URL ?? '#'}>
                <p>
                  Открыть&nbsp;<span className="font-bold">@laniamc</span>
                </p>
                <ArrowUpRightIcon className="size-4" />
              </Link>
            </Button>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <SiDiscord className="size-6" color={SiDiscordHex} />
              <p className="text-3xl font-semibold tracking-tight">Discord</p>
            </CardTitle>
            <CardDescription className="text-lg">
              Сервер сообщества в Discord
            </CardDescription>
          </CardHeader>
          <CardContent className="flex h-full items-end">
            <Button asChild variant="default" size="lg">
              <Link href={process.env.NEXT_PUBLIC_DISCORD_SERVER_URL ?? '#'}>
                <p>Вступить в сообщество</p>
                <ArrowUpRightIcon className="size-4" />
              </Link>
            </Button>
          </CardContent>
        </Card>
      </div>
      <Card className="mt-4">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <BriefcaseBusinessIcon className="size-6" />
            <p className="text-3xl font-semibold tracking-tight">
              Самозанятый Кузьмин Егор Олегович
            </p>
          </CardTitle>
          <CardDescription className="text-lg">
            Реквизиты самозанятого РФ
          </CardDescription>
        </CardHeader>
        <CardContent className="px-0">
          <table className="border-separate border-spacing-x-6 border-spacing-y-2">
            <thead>
              <tr>
                <th className="text-muted-foreground text-start text-sm">
                  ИНН
                </th>
                <th className="text-muted-foreground text-start text-sm">
                  Юридический адрес
                </th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>773718443054</td>
                <td>г. Москва</td>
              </tr>
            </tbody>
          </table>
        </CardContent>
      </Card>
    </div>
  )
}
