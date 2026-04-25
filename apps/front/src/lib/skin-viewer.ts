import * as skinview3d from 'skinview3d'

function reloadSkin(
  url: string,
  isSlim: boolean,
  skinViewer: skinview3d.SkinViewer,
): void {
  if (url === '') {
    skinViewer.loadSkin(null)
  } else {
    skinViewer
      .loadSkin(url, {
        model: isSlim ? 'slim' : 'default',
      })
      .catch((e) => {
        console.error(e)
      })
  }
}

function reloadCape(url: string, skinViewer: skinview3d.SkinViewer): void {
  if (url === '') {
    skinViewer.loadCape(null)
  } else {
    skinViewer
      .loadCape(url, {
        backEquipment: 'cape',
      })
      .catch((e) => {
        console.error(e)
      })
  }
}

function updateBackground(
  skinViewer: skinview3d.SkinViewer,
  color: string,
): void {
  skinViewer.background = color
}

export function initializeViewer(
  skinUrl: string,
  capeUrl: string,
  isSlim: boolean,
  theme: 'light' | 'dark',
): void {
  const skinContainer = document.getElementById(
    'skin_container',
  ) as HTMLCanvasElement
  if (!skinContainer) {
    throw new Error('Canvas element not found')
  }

  const skinViewer = new skinview3d.SkinViewer({
    canvas: skinContainer,
  })

  skinViewer.width = 300
  skinViewer.height = 200
  skinViewer.fov = 40
  skinViewer.zoom = 0.9
  skinViewer.globalLight.intensity = 3
  skinViewer.cameraLight.intensity = 0.6
  skinViewer.autoRotate = true
  skinViewer.autoRotateSpeed = 0.5

  skinViewer.controls.enableRotate = true
  skinViewer.controls.enableZoom = false
  skinViewer.controls.enablePan = false

  // for (const part of skinParts) {
  //   for (const layer of skinLayers) {
  //     if (skinViewer.playerObject.skin) {
  //       skinViewer.playerObject.skin.head
  //       skinViewer.playerObject.skin[part][layer].visible = true
  //     }
  //   }
  // }

  reloadSkin(skinUrl, isSlim, skinViewer)
  reloadCape(capeUrl, skinViewer)
  updateBackground(skinViewer, theme === 'light' ? '#FFFFFF' : '#18181b')
}
