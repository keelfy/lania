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
  UsersIcon,
  MapPinIcon,
  PackageSearchIcon,
} from 'lucide-react'
import { MetaRecord } from 'nextra'

const meta: MetaRecord = {
  survival: { type: 'separator', title: 'Survival' },
  qol: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <PickaxeIcon size={14} />
        <p>Quality of life</p>
      </div>
    ),
  },
  veinminer: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <PickaxeIcon size={14} />
        <p>Vein mining</p>
      </div>
    ),
  },
  adorena: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <UsersIcon size={14} />
        <p>Character size</p>
      </div>
    ),
  },
  cauldrons: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <CookingPotIcon size={14} />
        <p>Cauldrons</p>
      </div>
    ),
  },
  mobs: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <PawPrintIcon size={14} />
        <p>Mobs</p>
      </div>
    ),
  },
  'loot-markers': {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <PackageSearchIcon size={14} />
        <p>Loot container markers</p>
      </div>
    ),
  },
  'nether-portals': {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <FlameIcon size={14} />
        <p>Nether portals</p>
      </div>
    ),
  },
  bacap: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <MilestoneIcon size={14} />
        <p>New advancements</p>
      </div>
    ),
  },
  building: { type: 'separator', title: 'Building and decoration' },
  miniblocks: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <BoxIcon size={14} />
        <p>Mini blocks</p>
      </div>
    ),
  },
  frames: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <FrameIcon size={14} />
        <p>Invisible item frames</p>
      </div>
    ),
  },
  signs: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <AppWindowMacIcon size={14} />
        <p>Signs</p>
      </div>
    ),
  },
  'debug-stick': {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <BugIcon size={14} />
        <p>Debug stick</p>
      </div>
    ),
  },
  'light-block': {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <LightbulbIcon size={14} />
        <p>Light block</p>
      </div>
    ),
  },
  'void-block': {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <BoxIcon size={14} />
        <p>Barrier</p>
      </div>
    ),
  },
  chat: { type: 'separator', title: 'Social' },
  patpat: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <PawPrintIcon size={14} />
        <p>Petting</p>
      </div>
    ),
  },
  'ping-wheel': {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <MapPinIcon size={14} />
        <p>Pings</p>
      </div>
    ),
  },
  streamotes: {
    title: (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
        <SmileIcon size={14} />
        <p>7TV emotes</p>
      </div>
    ),
  },
}

export default meta
