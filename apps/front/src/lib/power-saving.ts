/**
 * detect iOS/iPad/macOS Low Power Mode, Chromium Energy Saver Mode, and maybe future Firefox power saving mode
 * @async
 * @method detectPowerSavingMode
 * @returns {boolean|undefined} return `undefined` if not sure
 */
export function detectPowerSavingMode(): Promise<boolean | undefined> {
  // for iOS/iPadOS Safari, and maybe MacBook macOS Safari (not tested)
  if (/(iP(?:hone|ad|od)|Mac OS X)/.test(navigator.userAgent)) {
    // In Low Power Mode, cumulative delay effect happens on setInterval()
    return new Promise((resolve) => {
      const fps = 60
      const interval = 1000 / fps
      const numFrames = 30
      const startTime = performance.now()
      let i = 0
      const handle = setInterval(() => {
        if (i < numFrames) {
          i++
          return
        }
        clearInterval(handle)
        const actualInterval = (performance.now() - startTime) / numFrames
        const ratio = actualInterval / interval // 1.3x or more in Low Power Mode, 1.1x otherwise
        // alert(actualInterval+' '+interval);
        console.log(actualInterval, interval, ratio)
        resolve(ratio > 1.3)
      }, interval)
    })
  }
  // for Safari, Chromium, and maybe future Firefox
  return detectFrameRate().then((frameRate: number) => {
    // In Battery Saver Mode frameRate will be about 30fps or 20fps,
    // otherwise frameRate will be closed to monitor refresh rate (typically 60Hz)
    if (frameRate < 34) {
      return true
    }
    return undefined
  })
}

export function detectFrameRate(): Promise<number> {
  return new Promise((resolve) => {
    const numFrames = 30
    const startTime = performance.now()
    let i = 0
    const tick = () => {
      if (i < numFrames) {
        i++
        requestAnimationFrame(tick)
        return
      }
      const frameRate = numFrames / ((performance.now() - startTime) / 1000)
      resolve(frameRate)
    }
    requestAnimationFrame(() => {
      tick()
    })
  })
}
