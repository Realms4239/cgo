import { describe, expect, it } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import LiveDeadline from './LiveDeadline'
import type { LiveFrame } from '../lib/types'

type Frame = Pick<LiveFrame, 'deadline_ok_pct' | 'deadline_measured'>
const render = (frame: Frame) => renderToStaticMarkup(<LiveDeadline frame={frame} />)

describe('LiveDeadline — absence du pourcentage d’échéances', () => {
  it('dit « — absent » quand la deadline n’est pas mesurée', () => {
    const html = render({ deadline_measured: false })
    expect(html).toContain('—')
    expect(html).toContain('absent')
    expect(html).not.toContain('%')
    expect(html).not.toContain('0,0')
  })

  it('ne montre jamais un 0 fabriqué même s’il accompagne le drapeau absent', () => {
    const html = render({ deadline_ok_pct: 0, deadline_measured: false })
    expect(html).toContain('absent')
    expect(html).not.toContain('0,0')
  })

  it('dit aussi absent quand le drapeau de mesure manque', () => {
    const html = render({ deadline_ok_pct: 96.7 })
    expect(html).toContain('absent')
    expect(html).not.toContain('96,7')
  })

  it('garde « 0,0 % » pour un zéro réellement mesuré', () => {
    const html = render({ deadline_ok_pct: 0, deadline_measured: true })
    expect(html).toContain('0,0')
    expect(html).toContain('%')
    expect(html).not.toContain('absent')
  })

  it('ne change pas le format d’un pourcentage mesuré', () => {
    expect(render({ deadline_ok_pct: 96.7, deadline_measured: true })).toContain('96,7 %')
  })
})
