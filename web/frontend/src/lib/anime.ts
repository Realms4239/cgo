import { animate, createTimeline, stagger, svg, text, createDrawable, splitText, cubicBezier } from 'animejs'

// anime v4 a retiré la forme CHAÎNE 'cubicBezier(...)' (warning console,
// easing silencieusement ignoré) — la fonction importée, une seule fois.
const EASE = cubicBezier(0.16, 1, 0.3, 1)
const EASE_OUT = cubicBezier(0.4, 0, 0.2, 1)
export { EASE }

export const prefersReducedMotion = () =>
  typeof window !== 'undefined' &&
  typeof window.matchMedia === 'function' &&
  window.matchMedia('(prefers-reduced-motion: reduce)').matches

export function animateViewEnter() {
  if (prefersReducedMotion()) return
  const tl = createTimeline()
  // fondu simple — le blur animé sur .view entier coûtait trop au switch
  tl.add('.view:not([hidden])', { translateY: [8, 0], opacity: [0, 1], duration: 300, ease: EASE } as any, 0)
}

export function animateArmButton(el: Element) {
  if (prefersReducedMotion()) return
  const tl = createTimeline()
  tl.add(el, { scale: [0.96, 1], duration: 400, ease: 'outElastic(1, .6)' } as any, 0)
  tl.add(el, { rotate: [0.5, -0.5, 0], duration: 400, ease: 'outElastic(1, .6)' } as any, 0)
  tl.add(el, { boxShadow: ['0 0 0 rgba(213,22,0,0)', '0 0 16px rgba(213,22,0,0.35)', '0 0 0 rgba(213,22,0,0)'], duration: 800, ease: 'linear' } as any, 0)
}

export function animatePromptEnter(el: Element) {
  if (prefersReducedMotion()) return
  const tl = createTimeline({ defaults: { duration: 500, ease: EASE } } as any)
  tl.add(el, { translateY: [12, 0], opacity: [0, 1], filter: ['blur(4px)', 'blur(0)'], duration: 500, ease: EASE } as any, 0)
  const btns = (el as HTMLElement).querySelectorAll('button')
  if (btns.length) tl.add(btns, { translateY: [8, 0], opacity: [0, 1], delay: stagger(40, { start: 60 }), duration: 400, ease: EASE } as any, 0)
}

export function animatePromptExit(el: Element): Promise<void> {
  if (prefersReducedMotion()) return Promise.resolve()
  return new Promise((resolve) => {
    const tl = createTimeline() as any
    tl.add(el, { translateY: [0, 12], opacity: [1, 0], duration: 300, ease: EASE_OUT } as any, 0)
    setTimeout(resolve, 300)
  })
}

export function animateShake(el: Element) {
  if (prefersReducedMotion()) return
  animate(el as HTMLElement, { translateX: [-4, 4, 0], duration: 400, ease: 'outElastic(1, .6)' } as any)
}

export function animateFlash(el: Element) {
  if (prefersReducedMotion()) return
  animate(el as HTMLElement, { translateY: [-16, 0], opacity: [0, 1], duration: 300, ease: EASE } as any)
}

export function animateToasts(els: Element[]) {
  if (prefersReducedMotion() || !els.length) return
  const tl = createTimeline() as any
  tl.add(els, { translateY: [16, 0], opacity: [0, 1], delay: stagger(20), duration: 300, ease: EASE } as any, 0)
}

// --- dense: svg/text — animejs@4.5.0, garde de source animeSvg.test.ts ---

export function animateMeteolinkShimmer(el: Element) {
  if (prefersReducedMotion()) return
  const path = el.querySelector('.ml-icon path') as SVGGeometryElement | null
  if (path) {
    const drawable = svg.createDrawable(path)
    animate(drawable as any, { draw: ['0 0', '0 1'], duration: 800, ease: 'linear' } as any)
  }
  const textEl = el.querySelector('.wordmark-text') as HTMLElement | null
  const target = (textEl ?? el) as HTMLElement
  const splitter = text.splitText(target as any, { chars: true } as any) as any
  void splitText; void createDrawable
  const chars: Element[] = splitter?.chars ?? []
  if (chars.length) {
    animate(chars as any, { translateY: [8, 0], opacity: [0, 1], duration: 600, ease: EASE, delay: stagger(30, { grid: [4, 2], from: 'center' } as any) } as any)
  }
}
