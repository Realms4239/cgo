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
  const tl = createTimeline()
  tl.add(el, { scale: [0.96, 1], duration: 400, ease: 'outElastic(1, .6)' } as any, 0)
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
