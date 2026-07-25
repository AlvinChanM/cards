// Tiny WebAudio-based sound effects -- no audio asset files needed.
// A single shared AudioContext is created lazily on first user
// interaction (required by browser autoplay policies).

let ctx: AudioContext | null = null

function getContext(): AudioContext {
  if (!ctx) {
    ctx = new AudioContext()
  }
  if (ctx.state === 'suspended') {
    void ctx.resume()
  }
  return ctx
}

function beep(freq: number, durationMs: number, type: OscillatorType = 'sine', volume = 0.15) {
  try {
    const audioCtx = getContext()
    const osc = audioCtx.createOscillator()
    const gain = audioCtx.createGain()
    osc.type = type
    osc.frequency.value = freq
    gain.gain.value = volume
    osc.connect(gain)
    gain.connect(audioCtx.destination)
    const now = audioCtx.currentTime
    gain.gain.setValueAtTime(volume, now)
    gain.gain.exponentialRampToValueAtTime(0.001, now + durationMs / 1000)
    osc.start(now)
    osc.stop(now + durationMs / 1000)
  } catch {
    // Audio can fail silently (e.g. no user gesture yet) -- never let
    // sound effects break gameplay.
  }
}

export function playSelectSound() {
  beep(660, 60, 'triangle', 0.12)
}

export function playDeselectSound() {
  beep(330, 60, 'triangle', 0.1)
}

export function playCardSound() {
  beep(880, 120, 'sine', 0.15)
}

export function passSound() {
  beep(220, 150, 'sine', 0.12)
}

export function roundOverSound() {
  beep(523, 90, 'sine', 0.18)
  setTimeout(() => beep(659, 90, 'sine', 0.18), 90)
  setTimeout(() => beep(784, 180, 'sine', 0.18), 180)
}

export function dealSound() {
  // A quick rising flurry of clicks to suggest a card-shuffling/dealing burst.
  for (let i = 0; i < 6; i++) {
    setTimeout(() => beep(300 + i * 40, 40, 'square', 0.06), i * 55)
  }
}
