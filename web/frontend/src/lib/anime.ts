import { animate, createTimeline, stagger, utils } from 'animejs'

export const prefersReducedMotion = () =>
  typeof window !== 'undefined' &&
  typeof window.matchMedia === 'function' &&
  window.matchMedia('(prefers-reduced-motion: reduce)').matches

export function animateViewEnter() {
  if (prefersReducedMotion()) return
  const tl = createTimeline()
  tl.add('.view', { translateY: [8, 0], opacity: [0, 1], filter: ['blur(4px)', 'blur(0)'], duration: 500, ease: 'cubicBezier(0.16,1,0.3,1)' } as any, 0)
  tl.add('.card', { translateY: [12, 0], opacity: [0, 1], delay: stagger(40, { start: 100 }), duration: 600, ease: 'cubicBezier(0.16,1,0.3,1)' } as any, 0)
}

export function animateCardStagger() {
  // ponytail: unused — animateViewEnter covers card stagger; keep for isolated card grids if needed
  if (prefersReducedMotion()) return
  const tl = createTimeline()
  tl.add('.card', { translateY: [12, 0], opacity: [0, 1], delay: stagger(40, { start: 100 }), duration: 600, ease: 'cubicBezier(0.16,1,0.3,1)' } as any, 0)
}

export function animateBar(el: Element) {
  if (prefersReducedMotion()) return
  // ponytail: simple scaleX bars, WAAPI spring if jank matters
  const h = el as HTMLElement
  h.style.transformOrigin = 'left center'
  utils.set(h, { scaleX: 0 } as any)
  animate(h, { scaleX: [0, 1], duration: 600, ease: 'cubicBezier(0.16,1,0.3,1)' } as any)
}

export function animateArmButton(el: Element) {
  if (prefersReducedMotion()) return
  // primary white → danger red spring scale 0.96→1 rotate 0.5→-0.5 + shadow pulse
  const tl = createTimeline()
  tl.add(el, { scale: [0.96, 1], duration: 400, ease: 'outElastic(1, .6)' } as any, 0)
  tl.add(el, { rotate: [0.5, -0.5, 0], duration: 400, ease: 'outElastic(1, .6)' } as any, 0)
  tl.add(el, { boxShadow: ['0 0 0 rgba(226,39,24,0)', '0 0 16px rgba(226,39,24,0.35)', '0 0 0 rgba(226,39,24,0)'], duration: 800, ease: 'linear' } as any, 0)
}

export function animateBannerPulse(el: Element) {
  if (prefersReducedMotion()) return
  // ponytail: opacity pulse only, not loud bounce
  const tl = createTimeline()
  tl.add(el, { opacity: [0.85, 1], duration: 700, ease: 'inOut(3)' } as any, 0)
}

export function animateLiveEnter() {
  if (prefersReducedMotion()) return
  // ponytail: clipPath reveal simple, WAAPI if jank matters
  const tl = createTimeline()
  tl.add('.card', { clipPath: ['inset(0 100% 0 0)', 'inset(0 0% 0 0)'], duration: 700, delay: stagger(50, { start: 80 }), ease: 'cubicBezier(0.16,1,0.3,1)' } as any, 0)
}

export function animateRail(el: Element, pinned: boolean) {
  if (prefersReducedMotion()) return
  const tl = createTimeline({ defaults: { duration: 400, ease: 'cubicBezier(0.4,0,0.2,1)' } } as any)
  tl.add(el, { width: [pinned ? 56 : 232, pinned ? 232 : 56], duration: 400, ease: 'cubicBezier(0.4,0,0.2,1)' } as any, 0)
}

export function animatePromptEnter(el: Element) {
  if (prefersReducedMotion()) return
  const tl = createTimeline({ defaults: { duration: 500, ease: 'cubicBezier(0.16,1,0.3,1)' } } as any)
  tl.add(el, { translateY: [12, 0], opacity: [0, 1], filter: ['blur(4px)', 'blur(0)'], duration: 500, ease: 'cubicBezier(0.16,1,0.3,1)' } as any, 0)
  const btns = (el as HTMLElement).querySelectorAll('button')
  if (btns.length) tl.add(btns, { translateY: [8, 0], opacity: [0, 1], delay: stagger(40, { start: 60 }), duration: 400, ease: 'cubicBezier(0.16,1,0.3,1)' } as any, 0)
}

export function animatePromptExit(el: Element): Promise<void> {
  if (prefersReducedMotion()) return Promise.resolve()
  return new Promise((resolve) => {
    const tl = createTimeline() as any
    tl.add(el, { translateY: [0, 12], opacity: [1, 0], duration: 300, ease: 'cubicBezier(0.4,0,0.2,1)' } as any, 0)
    setTimeout(resolve, 300)
  })
}

export function animateShake(el: Element) {
  if (prefersReducedMotion()) return
  animate(el as HTMLElement, { translateX: [-4, 4, 0], duration: 400, ease: 'outElastic(1, .6)' } as any)
}

export function animateFlash(el: Element) {
  if (prefersReducedMotion()) return
  animate(el as HTMLElement, { translateY: [-16, 0], opacity: [0, 1], duration: 300, ease: 'cubicBezier(0.16,1,0.3,1)' } as any)
}

export function animateToasts(els: Element[]) {
  if (prefersReducedMotion() || !els.length) return
  const tl = createTimeline() as any
  tl.add(els, { translateY: [16, 0], opacity: [0, 1], delay: stagger(20), duration: 300, ease: 'cubicBezier(0.16,1,0.3,1)' } as any, 0)
}
