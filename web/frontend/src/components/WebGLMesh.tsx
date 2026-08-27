import { useEffect, useRef } from 'react'

// ponytail: 2d mesh not WebGL, full WebGL if jank matters — one canvas, draw once, no loop
export default function WebGLMesh() {
  const ref = useRef<HTMLCanvasElement>(null)
  useEffect(() => {
    const c = ref.current
    if (!c) return
    const ctx = c.getContext('2d')
    if (!ctx) return
    const draw = () => {
      const dpr = Math.min(window.devicePixelRatio || 1, 2)
      const w = window.innerWidth
      const h = window.innerHeight
      c.width = w * dpr
      c.height = h * dpr
      c.style.width = w + 'px'
      c.style.height = h + 'px'
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
      ctx.clearRect(0, 0, w, h)
      // radial cyan 4% top + violet 3% bottom-right, once
      const g1 = ctx.createRadialGradient(w * 0.5, 0, 0, w * 0.5, 0, w * 0.8)
      g1.addColorStop(0, 'rgba(90,211,227,0.04)')
      g1.addColorStop(1, 'transparent')
      ctx.fillStyle = g1
      ctx.fillRect(0, 0, w, h)
      const g2 = ctx.createRadialGradient(w, h, 0, w, h, w * 0.6)
      g2.addColorStop(0, 'rgba(180,138,224,0.03)')
      g2.addColorStop(1, 'transparent')
      ctx.fillStyle = g2
      ctx.fillRect(0, 0, w, h)
    }
    draw()
    const onResize = () => draw()
    window.addEventListener('resize', onResize)
    return () => { window.removeEventListener('resize', onResize); ctx.globalAlpha = 1 }
  }, [])
  return (
    <canvas
      ref={ref}
      id="mesh"
      aria-hidden="true"
      style={{ position: 'fixed', inset: '-1px', opacity: 0.015, pointerEvents: 'none', mixBlendMode: 'overlay' }}
    />
  )
}
