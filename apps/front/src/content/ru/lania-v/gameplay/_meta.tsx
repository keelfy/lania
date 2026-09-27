import {
  AppWindowMacIcon,
  BoxIcon,
  BugIcon,
  CookingPotIcon,
  FlameIcon,
  FrameIcon,
  LightbulbIcon,
  MilestoneIcon,
  PawPrintIcon,
  PickaxeIcon,
  SmileIcon,
} from 'lucide-react'
import { MetaRecord } from 'nextra'

const meta: MetaRecord = {
  qol: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <PickaxeIcon size={14} />
        <p>Удобства</p>
      </div>
    ),
  },
  cauldrons: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <CookingPotIcon size={14} />
        <p>Котлы</p>
      </div>
    ),
  },
  mobs: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <PawPrintIcon size={14} />
        <p>Мобы</p>
      </div>
    ),
  },
  'nether-portals': {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <FlameIcon size={14} />
        <p>Порталы в Нижний мир</p>
      </div>
    ),
  },
  streamotes: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <SmileIcon size={14} />
        <p>Смайлики 7TV</p>
      </div>
    ),
  },
  bacap: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <MilestoneIcon size={14} />
        <p>Новые достижения</p>
      </div>
    ),
  },
  frames: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <FrameIcon size={14} />
        <p>Невидимые рамки</p>
      </div>
    ),
  },
  signs: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <AppWindowMacIcon size={14} />
        <p>Таблички</p>
      </div>
    ),
  },
  miniblocks: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <BoxIcon size={14} />
        <p>Миниблоки</p>
      </div>
    ),
  },
  'debug-stick': {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <BugIcon size={14} />
        <p>Дебаг палочка</p>
      </div>
    ),
  },
  'light-block': {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <LightbulbIcon size={14} />
        <p>Блок света</p>
      </div>
    ),
  },
  'void-block': {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <BoxIcon size={14} />
        <p>Барьер</p>
      </div>
    ),
  },
}

export default meta
